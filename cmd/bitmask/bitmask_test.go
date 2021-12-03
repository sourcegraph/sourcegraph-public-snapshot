package main

import (
	"os"
	"testing"
)

var result = 0

func BenchmarkTestWriteCache(b *testing.B) {
	for j := 0; j < b.N; j++ {
		dir := os.Getenv("HOME") + "/dev/sourcegraph/sourcegraph"
		WriteCache(dir, 1_000)
	}
}

func BenchmarkTestBitmask(b *testing.B) {
	repo := ReadCache()
	b.ResetTimer()
	i := 0
	query := []byte("Pag")
	for j := 0; j < b.N; j++ {
		paths := repo.PathsMatchingQuery(query)
		for path := range paths {
			//fmt.Println(path)
			i += len(path)
		}
	}
	result = i
}
