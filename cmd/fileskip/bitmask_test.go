package fileskip

import (
	"fmt"
	"github.com/loov/hrtime/hrtesting"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	exampleText = `Hello world,
this is the world,
this it the time!
We all do our Best!
The World is the best.
`
)

var dir = os.Getenv("HOME") + "/dev/sourcegraph/sourcegraph"
var cacheFile = os.Getenv("HOME") + "/dev/sourcegraph/fileskip-cache"
var benchmarkCacheDir = os.Getenv("HOME") + "/dev/sourcegraph/benchmark-cache"

func BenchmarkIndex(b *testing.B) {
	bench := hrtesting.NewBenchmark(b)
	defer bench.Report()
	for bench.Next() {
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
		paths := repo.FilenamesMatchingQuery(query)
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
		panic("NON DISTINCT!")
	}
	ratio := float64(falsePositives) / math.Max(1.0, float64(len(matchingPaths)))
	fmt.Printf("fp %v len %v\n", ratio, len(matchingPaths))
}

func newRepoIndex(t *testing.T) *RepoIndex {
	fs := InMemoryFileSystem{map[string]string{"readme.md": exampleText}}
	r, err := NewInMemoryRepoIndex(&fs)
	if err != nil {
		t.Fatalf("failed to create repo index %v", err)
	}
	return r
}

//func TestFalseResults(t *testing.T) {
//	r := newRepoIndex(t)
//	for i := range exampleText {
//		for j := i + 1; j < len(exampleText); j++ {
//			query := exampleText[i:j]
//			truePositiveCount := len(r.PathsMatchingQuerySync(query))
//			if truePositiveCount == 0 {
//				t.Fatalf("query '%v' triggered a false negative", query)
//			}
//
//			falseQueries := []string{
//				exampleText[i:j] + "1",
//				strings.ToUpper(exampleText[i:j]) + strings.ToLower(exampleText[i:j]),
//			}
//			for _, falseQuery := range falseQueries {
//				falsePositiveCount := len(r.PathsMatchingQuerySync(falseQuery))
//				if falsePositiveCount > 0 {
//					t.Fatalf("query '%v' triggered a false positive", query)
//				}
//			}
//		}
//	}
//}

// TODO CPU: serialize, deserialize
// TODO Size: serialized file, in-memory index
func FalsePositive(t *testing.T) {
	files := map[string]string{
		"monitoring/definitions/git_server.go": "Repository",
		//"client/web/src/nav/UserNavItem.story.tsx": "JVM",
	}
	for file, query := range files {
		abspath := filepath.Join("/Users/olafurpg/dev/sourcegraph/sourcegraph", file)
		bytes, err := os.ReadFile(abspath)
		if err != nil {
			panic(err)
		}
		fs := InMemoryFileSystem{
			map[string]string{
				file: string(bytes),
			},
		}
		r, err := NewInMemoryRepoIndex(&fs)
		if err != nil {
			panic(err)
		}
		paths := r.PathsMatchingQuerySync(query)
		if len(paths) > 0 && strings.Index(string(bytes), query) < 0 {
			t.Fatalf("query '%v' triggered a false positive in path '%v'", query, abspath)
		}
	}
}

func BenchmarkIteration(b *testing.B) {
	fileCount := 700_000
	mask := 1231254214
	sum := 0
	for i := 0; i < b.N; i++ {
		for j := 0; j < fileCount; j++ {
			mask ^= j & i
			if mask > 0 {
				sum++
			}
		}
	}
}
