package main

import (
	"sync"
	"testing"
)

var result = 0

func BenchmarkTestBitmask(b *testing.B) {
	indexes := ReadCache()
	b.ResetTimer()
	i := 0
	query := []byte("Pag")
	for j := 0; j < b.N; j++ {
		paths := matchingPaths(indexes, query)
		for path := range paths {
			//fmt.Println(path)
			i += len(path)
		}
	}
	result = i
}

func matchingPaths(indexes []TrigramIndex, query []byte) chan string {
	res := make(chan string, len(indexes))
	batchSize := 5_000
	var wg sync.WaitGroup
	for i := 0; i < len(indexes); i += batchSize {
		j := i + batchSize
		if j > len(indexes) {
			j = len(indexes)
		}
		batch := indexes[i:j]
		wg.Add(1)
		go func() {
			defer wg.Done()
			queryBatch(res, batch, query)
		}()
	}
	wg.Wait()
	close(res)
	return res
}

func queryBatch(results chan string, indexes []TrigramIndex, query []byte) {
	for _, index := range indexes {
		if index.Filter != nil && index.Filter.Test(query) {
			results <- index.Path
		}
	}
}
