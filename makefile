PROJECTNAME=$(shell basename "$(PWD)")

# Go related variables.
GOBASE=$(shell pwd)
GOPATH=$(GOBASE)/vendor:$(GOBASE):/home/azer/code/golang # You can remove or change the path after last colon.
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

build-linux:
	GOOS=linux GOARCH=amd64 go build -a -o poc-redis-keys-count main.go

build:
	@CGO_ENABLED=0 go build -a -o poc-redis-keys-count main.go

dockerup:
	docker-compose up --build -d

dockerdown:
	docker-compose down

docker-redis:
	docker container run --rm -p 6379:6379 -v ~/go/src/github.com/nahwinrajan/poc-redis-keys-countconfig/:/usr/local/etc/redis/redis.conf --name redis-poc redis

ltest-keys:
	vegeta attack -format=json -duration=10s -rate=100/1s -workers=130 -timeout=1s -targets=./vegeta-format-keys.json | tee vegeta-result-keys.bin | vegeta report

ltest-hash:
	vegeta attack -format=json -duration=10s -rate=100/1s -workers=130 -timeout=1s -targets=./vegeta-format-hash.json | tee vegeta-result-hash.bin | vegeta report