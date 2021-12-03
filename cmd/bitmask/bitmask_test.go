package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var result = 0
var dir = os.Getenv("HOME") + "/dev/sourcegraph/sourcegraph"
var cacheDir = os.Getenv("HOME") + "/dev/sourcegraph/bitmask-cache"

func BenchmarkIndex(b *testing.B) {
	for j := 0; j < b.N; j++ {
		err := WriteCache(dir, cacheDir)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkQuery(b *testing.B) {
	repo := ReadCache(cacheDir)
	b.ResetTimer()
	query := "Visitor"
	matchingPaths := make(map[string]struct{})
	for j := 0; j < b.N; j++ {
		paths := repo.PathsMatchingQuery(query)
		for path := range paths {
			matchingPaths[path] = struct{}{}
		}
	}
	b.StopTimer()
	falsePositives := 0
	for relpath := range matchingPaths {
		abspath := filepath.Join(dir, relpath)
		b, err := os.ReadFile(abspath)
		if err != nil {
			panic(err)
		}
		t := string(b)
		if strings.Index(t, query) < 0 {
			falsePositives++
		} else {
			//fmt.Println(abspath)
		}
	}
	if len(matchingPaths) > len(matchingPaths) {
		println("NON DISTINCT!")

	}
	fmt.Printf("fp %v len %v\n", float64(falsePositives)/float64(len(matchingPaths)), len(matchingPaths))
}

// TODO: serialize
// TODO: deserialize
