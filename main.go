package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/julienschmidt/httprouter"
)

// RedisObject entrypoint for all our redis need
type RedisObject struct {
	pool *redis.Pool
}

var router *httprouter.Router
var robj *RedisObject

var formatKey = "key_%d"
var formatHash = "hash_%d"
var formatHashField = "field_%d"

var totalKeys = 1000000

// var ErrNil = errors.New("redigo: nil returned")
// var ErrPoolExhausted = errors.New("redigo: connection pool exhausted")

type rqstPayloadPopulateKey struct {
	TotalKeys int `json:"totalKeys"`
}

func main() {
	robj = NewRedisObject(":6379", 500)
	resp, err := robj.testConnection()
	log.Printf("response ping redis resp:%s err:%+v\n", resp, err)

	router = httprouter.New()
	registerRoutes(router)

	log.Printf("POC Redis Key - Fired Up and Ready to Go!!")
	log.Fatal(http.ListenAndServe(":3103", router))
}

func registerRoutes(router *httprouter.Router) {
	router.GET("/api/v1/ping", handlerPing)
	router.POST("/api/v1/redis/prune", handlerRedisPrune)

	router.GET("/api/v1/data/keys/:key_id", handlerGetKey)
	router.GET("/api/v1/data/hash/:key_id", handlerGetHash)

	router.POST("/api/v1/redis/populate/keys", handlerPopulateKeys)
	router.POST("/api/v1/redis/populate/hash", handlerPopulateHash)
}

// NewRedisObject initalizing pools
func NewRedisObject(host string, poolSize int) *RedisObject {
	pool := &redis.Pool{
		MaxIdle:         30,
		MaxActive:       poolSize,
		IdleTimeout:     1 * time.Second,
		MaxConnLifetime: 3 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", host, redis.DialConnectTimeout(time.Second))
			if err != nil {
				log.Println("error getting connection:", err)
				return nil, err
			}
			return conn, err
		},
	}

	RedisObject := RedisObject{
		pool: pool,
	}

	return &RedisObject
}

func (r *RedisObject) testConnection() (string, error) {
	conn := r.pool.Get()
	defer conn.Close()

	response, err := redis.String(conn.Do("PING"))

	return response, err
}

// Set set single value
func (r *RedisObject) Set(key, value string) error {
	conn := r.pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value)

	return err
}

// Get get single value
func (r *RedisObject) Get(key string) (string, error) {
	conn := r.pool.Get()
	defer conn.Close()

	res, err := redis.String(conn.Do("GET", key))

	return res, err
}

// HSet set a field in hash object
func (r *RedisObject) HSet(key, field, value string) error {
	conn := r.pool.Get()
	defer conn.Close()

	_, err := conn.Do("HSET", key, field, value)

	return err
}

// HGet get a field in hash object
func (r *RedisObject) HGet(hashname, field string) (string, error) {
	conn := r.pool.Get()
	defer conn.Close()

	res, err := redis.String(conn.Do("HGET", hashname, field))

	return res, err
}

// Prune delete all data in redis
func (r *RedisObject) Prune() error {
	conn := r.pool.Get()
	defer conn.Close()

	_, err := conn.Do("FLUSHALL")

	return err
}

// PipeSetKeys set key in bulk via Pipeline method
func (r *RedisObject) PipeSetKeys(totalkeys int) {
	conn := r.pool.Get()
	defer conn.Close()

	var err error
	for i := 0; i < totalkeys; i++ {
		keyID := fmt.Sprintf(formatKey, i)

		err = conn.Send("SET", keyID, i)

		if err != nil {
			log.Printf("[PipeSetKeys] error send key key_id:%s :%+v\n", keyID, err)
		}
	}

	err = conn.Flush()
	if err != nil {
		log.Printf("[PipeSetKeys] error flush :%+v\n", err)
	}

	for i := 0; i < totalkeys; i++ {
		_, err := conn.Receive()
		if err != nil {
			log.Printf("[PipeSetKeysReceive] receive :%+v\n", err)
		}
	}

	log.Printf("PipeSetKeys done")
	return
}

// PipeSetHashes set key in bulk via Pipeline method
func (r *RedisObject) PipeSetHashes(totalkeys int) {
	conn := r.pool.Get()
	defer conn.Close()

	var err error
	for i := 0; i < totalkeys; i++ {
		keyID := getHashShard(i)
		fieldID := fmt.Sprintf(formatHashField, i)

		err = conn.Send("HSET", keyID, fieldID, i)

		if err != nil {
			log.Printf("[PipeSetHashes] error send key key_id:%s val:%d :%+v\n", keyID, i, err)
		}
	}

	err = conn.Flush()
	if err != nil {
		log.Printf("[PipeSetHashes] error flush :%+v\n", err)
	}

	for i := 0; i < totalkeys; i++ {
		_, err := conn.Receive()
		if err != nil {
			log.Printf("[PipeSetHashes] receive :%+v\n", err)
		}
	}

	log.Printf("PipeSetHash done")
	return
}

func writeRespone(w http.ResponseWriter, httpStatusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")

	w.WriteHeader(httpStatusCode)

	json.NewEncoder(w).Encode(data)
	return
}

func handlerPing(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	writeRespone(w, http.StatusOK, "Pong")
	return
}

func handlerRedisPrune(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	err := robj.Prune()
	if err != nil {
		log.Printf("[handlerRedisPrune] error prune redis :%+v\n", err)
		writeRespone(w, http.StatusBadRequest, "Error pruning redis")
		return
	}

	writeRespone(w, http.StatusOK, "Redis pruned")
	return
}

func handlerGetKey(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	keyID, err := strconv.ParseInt(params.ByName("key_id"), 10, 64)
	if err != nil {
		log.Printf("[handlerGetKey] error parse keyid:%s :%+v\n", params.ByName("key_id"), err)
		writeRespone(w, http.StatusBadRequest, "Bad parameter passed")
		return
	}

	val, err := robj.Get(fmt.Sprintf(formatKey, keyID))
	if err != nil {
		log.Printf("[handlerGetKey] error read redis keyid:%d :%+v\n", keyID, err)
		writeRespone(w, http.StatusBadRequest, "Error load data")
		return
	}

	writeRespone(w, http.StatusOK, fmt.Sprintf("Value for key_id:%d is:%s\n", keyID, val))
	return
}

func handlerGetHash(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	fieldID, err := strconv.ParseInt(params.ByName("key_id"), 10, 64)
	if err != nil {
		log.Printf("[handlerGetHash] error parse hashID:%s :%+v\n", params.ByName("key_id"), err)
		writeRespone(w, http.StatusBadRequest, "Bad parameter passed")
		return
	}

	hashname := getHashShard(int(fieldID))

	val, err := robj.HGet(hashname, fmt.Sprintf(formatHashField, fieldID))
	if err != nil {
		log.Printf("[handlerGetHash] error read redis hashname:%s fieldID:%d :%+v\n", hashname, fieldID, err)
		writeRespone(w, http.StatusBadRequest, "Error load data")
		return
	}

	writeRespone(w, http.StatusOK, fmt.Sprintf("Value for hash_id:%d is:%s\n", fieldID, val))
	return
}

// handlerPopulateKeys populate redis with sample key data
// request payload: { totalKeys: int }
func handlerPopulateKeys(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var payload rqstPayloadPopulateKey
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("[handlerPopulateKeys] error decode payload :%+v\n", err)
		writeRespone(w, http.StatusBadRequest, "Bad payload passed")
		return
	}

	totalKeys = payload.TotalKeys

	robj.PipeSetKeys(payload.TotalKeys)

	writeRespone(w, http.StatusOK, fmt.Sprintf("Redis has been populated with %d data", payload.TotalKeys))
	return
}

// handlerPopulateHash populate redis with sample hash data
// request payload: { totalKeys: int }
func handlerPopulateHash(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	var payload rqstPayloadPopulateKey
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("[handlerPopulateHash] error decode payload :%+v\n", err)
		writeRespone(w, http.StatusBadRequest, "Bad payload passed")
		return
	}

	totalKeys = payload.TotalKeys

	robj.PipeSetHashes(payload.TotalKeys)

	writeRespone(w, http.StatusOK, fmt.Sprintf("Redis has been populated with %d data", payload.TotalKeys))
	return
}

func getHashShard(fieldID int) string {
	maxFieldPerHash := 1000

	shard := fieldID / maxFieldPerHash
	return fmt.Sprintf(formatHash, int(shard))
}
