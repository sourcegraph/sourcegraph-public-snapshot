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

func BenchmarkIndex(b *testing.B) {
	for j := 0; j < b.N; j++ {
		err := WriteCache(dir, cacheFile)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkQuery(b *testing.B) {
	repo, err := ReadCache(cacheFile)
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	query := "Case"
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
		abspath := filepath.Join(repo.Dir, relpath)
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

// TODO CPU: serialize, deserialize
// TODO Size: serialized file, in-memory index
