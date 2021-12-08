package main

import (
	"fmt"
	"github.com/sourcegraph/sourcegraph/cmd/fileskip"
	"math"
	"strings"
	"testing"
)

func benchmarkQuery(b *testing.B, c Corpus, query string) {
	fileskip.IsProgressBarEnabled = false
	index, err := c.LoadRepoIndex()
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	matchingResults := map[string]struct{}{}
	for i := 0; i < b.N; i++ {
		for m := range index.PathsMatchingQuery(query) {
			matchingResults[m] = struct{}{}
		}
	}
	b.StopTimer()
	falsePositives := 0
	for m := range matchingResults {
		data, err := index.FS.ReadRelativeFilename(m)
		if err != nil {
			panic(err)
		}
		text := string(data)
		isFalsePositive := strings.Index(text, query) < 0
		if isFalsePositive {
			falsePositives++
		}
	}
	b.ReportMetric(float64(len(index.Blobs)), "index-size")
	b.ReportMetric(float64(len(matchingResults)), "true-positives")
	b.ReportMetric(float64(falsePositives), "false-positives")
	b.ReportMetric(float64(falsePositives)/math.Max(1, float64(len(matchingResults))), "false-positive/true-positive")
}

func benchmarkShortQuery(b *testing.B, c Corpus)  { benchmarkQuery(b, c, c.Queries[0]) }
func benchmarkMediumQuery(b *testing.B, c Corpus) { benchmarkQuery(b, c, c.Queries[len(c.Queries)/2]) }
func benchmarkLongQuery(b *testing.B, c Corpus)   { benchmarkQuery(b, c, c.Queries[len(c.Queries)-1]) }

func BenchmarkQueryFlaskShort(b *testing.B)  { benchmarkShortQuery(b, flask) }
func BenchmarkQueryFlaskMedium(b *testing.B) { benchmarkMediumQuery(b, flask) }
func BenchmarkQueryFlaskLong(b *testing.B)   { benchmarkLongQuery(b, flask) }

func BenchmarkQuerySourcegraphShort(b *testing.B)  { benchmarkShortQuery(b, sourcegraph) }
func BenchmarkQuerySourcegraphMedium(b *testing.B) { benchmarkMediumQuery(b, sourcegraph) }
func BenchmarkQuerySourcegraphLong(b *testing.B)   { benchmarkLongQuery(b, sourcegraph) }

func BenchmarkQueryKubernetesShort(b *testing.B)  { benchmarkShortQuery(b, kubernetes) }
func BenchmarkQueryKubernetesMedium(b *testing.B) { benchmarkMediumQuery(b, kubernetes) }
func BenchmarkQueryKubernetesLong(b *testing.B)   { benchmarkLongQuery(b, kubernetes) }

func BenchmarkQueryLinuxShort(b *testing.B)  { benchmarkShortQuery(b, linux) }
func BenchmarkQueryLinuxMedium(b *testing.B) { benchmarkMediumQuery(b, linux) }
func BenchmarkQueryLinuxLong(b *testing.B)   { benchmarkLongQuery(b, linux) }

func BenchmarkQueryChromiumShort(b *testing.B)  { benchmarkShortQuery(b, chromium) }
func BenchmarkQueryChromiumMedium(b *testing.B) { benchmarkMediumQuery(b, chromium) }
func BenchmarkQueryChromiumLong(b *testing.B)   { benchmarkLongQuery(b, chromium) }

func BenchmarkQueryMegarepoShort(b *testing.B)  { benchmarkShortQuery(b, megarepo) }
func BenchmarkQueryMegarepoMedium(b *testing.B) { benchmarkMediumQuery(b, megarepo) }
func BenchmarkQueryMegarepoLong(b *testing.B)   { benchmarkLongQuery(b, megarepo) }

func loadCorpus(b *testing.B, corpus Corpus) {
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
	for _, b := range index.Blobs {
		if b.Filter.K() > 7 {
			fmt.Println(b.Filter.K())
		}
	}
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
		bloomFilterBinaryStorageSize = bloomFilterBinaryStorageSize + blob.Filter.BitSet().BinaryStorageSize()
	}
	b.ReportMetric(float64(len(index.Blobs)), "indexed-blob-count")
	b.ReportMetric(float64(indexedBlobsSize), "indexed-blobs-size")
	b.ReportMetric(float64(bloomFilterBinaryStorageSize), "bloom-storage-size")
	b.ReportMetric(float64(bloomFilterBinaryStorageSize)/float64(indexedBlobsSize), "compression-ratio")
}

func BenchmarkLoadFlask(b *testing.B)       { loadCorpus(b, flask) }
func BenchmarkLoadSourcegraph(b *testing.B) { loadCorpus(b, sourcegraph) }
func BenchmarkLoadKubernetes(b *testing.B)  { loadCorpus(b, kubernetes) }
func BenchmarkLoadLinux(b *testing.B)       { loadCorpus(b, linux) }
func BenchmarkLoadChromium(b *testing.B)    { loadCorpus(b, chromium) }
func BenchmarkLoadMegarepo(b *testing.B)    { loadCorpus(b, megarepo) }
