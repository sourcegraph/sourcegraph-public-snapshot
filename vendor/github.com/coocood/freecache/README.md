# FreeCache - A cache library for Go with zero GC overhead and high concurrent performance.

Long lived objects in memory introduce expensive GC overhead, With FreeCache, you can cache unlimited number of objects in memory 
without increased latency and degraded throughput. 

[![Build Status](https://travis-ci.org/coocood/freecache.png?branch=master)](https://travis-ci.org/coocood/freecache)
[![GoCover](http://gocover.io/_badge/github.com/coocood/freecache)](http://gocover.io/github.com/coocood/freecache)
[![GoDoc](https://godoc.org/github.com/coocood/freecache?status.svg)](https://godoc.org/github.com/coocood/freecache)

## Features

* Store hundreds of millions of entries
* Zero GC overhead
* High concurrent thread-safe access
* Pure Go implementation
* Expiration support
* Nearly LRU algorithm
* Strictly limited memory usage
* Come with a toy server that supports a few basic Redis commands with pipeline
* Iterator support

## Performance

Here is the benchmark result compares to built-in map, `Set` performance is about 2x faster than built-in map, `Get` performance is about 1/2x slower than built-in map. Since it is single threaded benchmark, in multi-threaded environment, 
FreeCache should be many times faster than single lock protected built-in map.

    BenchmarkCacheSet        3000000               446 ns/op
    BenchmarkMapSet          2000000               861 ns/op
    BenchmarkCacheGet        3000000               517 ns/op
    BenchmarkMapGet         10000000               212 ns/op

## Example Usage

```go
cacheSize := 100 * 1024 * 1024
cache := freecache.NewCache(cacheSize)
debug.SetGCPercent(20)
key := []byte("abc")
val := []byte("def")
expire := 60 // expire in 60 seconds
cache.Set(key, val, expire)
got, err := cache.Get(key)
if err != nil {
    fmt.Println(err)
} else {
    fmt.Println(string(got))
}
affected := cache.Del(key)
fmt.Println("deleted key ", affected)
fmt.Println("entry count ", cache.EntryCount())
```

## Notice

* Memory is preallocated.
* If you allocate large amount of memory, you may need to set `debug.SetGCPercent()`
to a much lower percentage to get a normal GC frequency.

## How it is done

FreeCache avoids GC overhead by reducing the number of pointers.
No matter how many entries stored in it, there are only 512 pointers.
The data set is sharded into 256 segments by the hash value of the key.
Each segment has only two pointers, one is the ring buffer that stores keys and values, 
the other one is the index slice which used to lookup for an entry.
Each segment has its own lock, so it supports high concurrent access.

## TODO

* Support dump to file and load from file.
* Support resize cache size at runtime.

## License

The MIT License
