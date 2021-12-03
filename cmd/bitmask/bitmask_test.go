package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var result = 0
var dir = os.Getenv("HOME") + "/dev/sourcegraph/sourcegraph"
var benchmarkCacheDir = os.Getenv("HOME") + "/dev/sourcegraph/benchmark-cache"

func BenchmarkIndex(b *testing.B) {
	for j := 0; j < b.N; j++ {
		err := WriteCache(dir, benchmarkCacheDir)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkQuery(b *testing.B) {
	query := "drivers/gpu/drm/i915/i915_perf.c"
	repo, err := ReadCache(cacheFile)
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
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
			if falsePositives == 1 {
				fmt.Println(abspath)
			}
		}
	}
	if len(matchingPaths) > len(matchingPaths) {
		println("NON DISTINCT!")

	}
	ratio := float64(falsePositives) / math.Max(1.0, float64(len(matchingPaths)))
	fmt.Printf("fp %v len %v\n", ratio, len(matchingPaths))
}

// TODO CPU: serialize, deserialize
// TODO Size: serialized file, in-memory index
