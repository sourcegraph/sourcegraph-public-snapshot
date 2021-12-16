package main

import (
	"fmt"
	"github.com/sourcegraph/sourcegraph/cmd/fileskip"
	"math"
	"os"
	"strings"
	"sync"
	"testing"
)

const maxResults = 20

var isQueryBaseline = "true" == os.Getenv("FILESKIP_BASELINE")

func benchmarkQuery(b *testing.B, c Corpus) {
	fileskip.IsProgressBarEnabled = false
	var index *fileskip.RepoIndex
	var err error
	b.ResetTimer()
	for _, query := range c.Queries {
		b.Run(fmt.Sprintf("%v-%v", c.Name, query), func(b *testing.B) {
			if index == nil {
				index, err = c.LoadRepoIndex()
				if err != nil {
					panic(err)
				}
				b.ResetTimer()
			}
			matchingResults := map[string]struct{}{}
			falsePositives := 0
			for i := 0; i < b.N; i++ {
				if isQueryBaseline {
					benchmarkBaselineQuery(index, matchingResults, query)
				} else {
					falsePositives = benchmarkFileskipQuery(index, matchingResults, query)
				}
			}
			b.StopTimer()
			b.ReportMetric(float64(len(index.Blobs)), "index-size")
			b.ReportMetric(float64(len(matchingResults)), "result-count")
			b.ReportMetric(float64(falsePositives), "false-positives")
			b.ReportMetric(float64(falsePositives)/math.Max(1, float64(len(matchingResults))), "false-positive/true-positive")
		})
	}
}

func benchmarkFileskipQuery(index *fileskip.RepoIndex, matchingResults map[string]struct{}, query string) int {
	falsePositives := 0
	for filename := range index.FilenamesMatchingQuery(query) {
		if expensiveHasMatch(index.FS, filename, query) {
			matchingResults[filename] = struct{}{}
			if len(matchingResults) > maxResults {
				break
			}
		} else {
			falsePositives++
		}
	}
	return falsePositives
}

func benchmarkBaselineQuery(index *fileskip.RepoIndex, matchingResults map[string]struct{}, query string) {
	batchSize := 100
	matches := make(chan string, len(index.Blobs))
	var wg sync.WaitGroup
	for j := 0; j < len(index.Blobs); j += batchSize {
		k := j + batchSize
		if k > len(index.Blobs) {
			k = len(index.Blobs)
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for _, b := range index.Blobs[start:end] {
				if expensiveHasMatch(index.FS, b.Path, query) {
					matches <- b.Path
				}
			}
		}(j, k)
	}
	wg.Wait()
	close(matches)
	for path := range matches {
		matchingResults[path] = struct{}{}
		if len(matchingResults) > maxResults {
			break
		}
	}
}

func expensiveHasMatch(fs fileskip.FileSystem, filename, query string) bool {
	textBytes, err := fs.ReadRelativeFilename(filename)
	if err != nil {
		panic(err)
	}
	text := strings.ToUpper(string(textBytes))
	return strings.Index(text, query) >= 0
}

func BenchmarkQuery(b *testing.B) {
	for _, corpus := range all {
		if corpus.Name != "sourcegraph" {
			//continue
		}
		benchmarkQuery(b, corpus)
	}
}

////func BenchmarkQueryFlaskShort(b *testing.B)  { benchmarkShortQuery(b, flask) }
////func BenchmarkQueryFlaskMedium(b *testing.B) { benchmarkMediumQuery(b, flask) }
////func BenchmarkQueryFlaskLong(b *testing.B)   { benchmarkLongQuery(b, flask) }
//
//func BenchmarkQuerySourcegraphShort(b *testing.B)  { benchmarkShortQuery(b, sourcegraph) }
//func BenchmarkQuerySourcegraphMedium(b *testing.B) { benchmarkMediumQuery(b, sourcegraph) }
//func BenchmarkQuerySourcegraphLong(b *testing.B)   { benchmarkLongQuery(b, sourcegraph) }
//
//func BenchmarkQueryKubernetesShort(b *testing.B)  { benchmarkShortQuery(b, kubernetes) }
//func BenchmarkQueryKubernetesMedium(b *testing.B) { benchmarkMediumQuery(b, kubernetes) }
//func BenchmarkQueryKubernetesLong(b *testing.B)   { benchmarkLongQuery(b, kubernetes) }
//
//func BenchmarkQueryLinuxShort(b *testing.B)  { benchmarkShortQuery(b, linux) }
//func BenchmarkQueryLinuxMedium(b *testing.B) { benchmarkMediumQuery(b, linux) }
//func BenchmarkQueryLinuxLong(b *testing.B)   { benchmarkLongQuery(b, linux) }
//
//func BenchmarkQueryChromiumShort(b *testing.B)  { benchmarkShortQuery(b, chromium) }
//func BenchmarkQueryChromiumMedium(b *testing.B) { benchmarkMediumQuery(b, chromium) }
//func BenchmarkQueryChromiumLong(b *testing.B)   { benchmarkLongQuery(b, chromium) }
//
//func BenchmarkQueryMegarepoShort(b *testing.B)  { benchmarkShortQuery(b, megarepo) }
//func BenchmarkQueryMegarepoMedium(b *testing.B) { benchmarkMediumQuery(b, megarepo) }
//func BenchmarkQueryMegarepoLong(b *testing.B)   { benchmarkLongQuery(b, megarepo) }

func loadCorpus(b *testing.B, corpus Corpus) {
	if isQueryBaseline {
		return
	}
	fileskip.IsProgressBarEnabled = false
	var index *fileskip.RepoIndex
	var err error
	for i := 0; i < b.N; i++ {
		index, err = corpus.LoadRepoIndex()
		if err != nil {
			panic(err)
		}
	}
	b.StopTimer()
	//stat, err := os.Stat(corpus.indexCachePath())
	//if err != nil {
	//	panic(err)
	//}
	//b.ReportMetric(float64(stat.Size()), "index-disk-size")
	//stat, err = os.Stat(corpus.zipCachePath())
	//if err != nil {
	//	panic(err)
	//}
	//b.ReportMetric(float64(stat.Size()), "archive-disk-size")
	indexedBlobsSize := int64(0)
	bloomFilterBinaryStorageSize := 0
	for _, blob := range index.Blobs {
		statSize, err := index.FS.StatSize(blob.Path)
		if err != nil {
			panic(err)
		}
		indexedBlobsSize = indexedBlobsSize + statSize
		serializedBitmap, err := blob.Filter.MarshalBinary()
		if err != nil {
			panic(err)
		}
		bloomFilterBinaryStorageSize += len(serializedBitmap)
	}
	b.ReportMetric(float64(len(index.Blobs)), "indexed-blob-count")
	b.ReportMetric(float64(indexedBlobsSize), "indexed-blobs-size")
	b.ReportMetric(float64(bloomFilterBinaryStorageSize), "bloom-memory-size")
	b.ReportMetric(float64(bloomFilterBinaryStorageSize)/float64(indexedBlobsSize), "compression-ratio")
}

func BenchmarkLoadFlask(b *testing.B)       { loadCorpus(b, flask) }
func BenchmarkLoadSourcegraph(b *testing.B) { loadCorpus(b, sourcegraph) }
func BenchmarkLoadKubernetes(b *testing.B)  { loadCorpus(b, kubernetes) }
func BenchmarkLoadLinux(b *testing.B)       { loadCorpus(b, linux) }
func BenchmarkLoadChromium(b *testing.B)    { loadCorpus(b, chromium) }
func BenchmarkLoadMegarepo(b *testing.B)    { loadCorpus(b, megarepo) }
