# POC Redis Keys Count

This project serve as proof of concept that number of keys matter for performance (time taken to find the key and response request).

## Hipothesis
Number of keys in redis do matter in performance of the redis service. If you have a lot of keys it will have longer latency than compared with same number of keys yet sharded
into hash.

A lot: we are talking about 100k keys and above,

So this Project will help you to proof that it is better to shard lot of small keys into shard of hash.


## TODO:
- [x] Build sample golang project with connection to redis
- [] Fix error whilst building golang image
- [x] Create report / article based with the number
- [] Dynamic vegeta hit list synced with total keys 


#### Docker golang image Issue, debug so far:
1. Dep ensure worked, and downloded into intermediary container whilst building
2. `Workdir` do create the new folder and `cd` to that path
3. Everything did copied to `Workdir` path


## Highlights

Something is funny albeit with the configuration difference between `SET` and `HSET` keys memory usage is bismall,
it is not anything near to the memory optimization article nor example from Instagram usage of Redis article.
Maybe the redis image is not reading the Redis Configuration file specified.