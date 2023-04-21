package main

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
)

func readSourcegraphEmbeddings() *embeddings.RepoEmbeddingIndex {
	return readEmbeddings("/Users/camdencheek/.sourcegraph-dev/data/blobstore-go/buckets/embeddings/github_com_sourcegraph_sourcegraph_cf360e12ff91b2fc199e75aef4ff6744.embeddingindex")
}

func readEmbeddings(path string) *embeddings.RepoEmbeddingIndex {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	rei, err := embeddings.DecodeRepoEmbeddingIndex(gob.NewDecoder(f))
	if err != nil {
		panic(err)
	}
	return rei
}

func main() {
	rei := readSourcegraphEmbeddings()

	f2, err := os.Open("/tmp/ranks_float32.gob")
	if err != nil {
		panic(err)
	}
	defer f2.Close()

	var r1 map[int][]int
	err = gob.NewDecoder(f2).Decode(&r1)
	if err != nil {
		panic(err)
	}

	r2 := map[int][]int{}
	for _, row := range []int{25, 558, 1261, 2222, 3885} {
		r2[row] = getRanks(rei.CodeIndex.SimilaritySearch(rei.CodeIndex.Row(row), 10000000, embeddings.WorkerOptions{}, embeddings.SearchOptions{}))
	}

	merged := compareResult{}

	for row, rank1 := range r1 {
		rank2, ok := r2[row]
		if !ok {
			panic(fmt.Sprintf("found unknown row %d", row))
		}
		compared := compareRanks(rank1, rank2)
		merged.count += compared.count
		merged.totalMigration += compared.totalMigration
		merged.top1same += compared.top1same
		merged.top10same += compared.top10same
		merged.top100same += compared.top100same
		merged.top1000same += compared.top1000same
		fmt.Println(compared.String())
	}

	fmt.Printf("Average: %0.2f\n", float32(merged.totalMigration)/float32(merged.count))
	fmt.Printf("Top10: %0.2f\n", float32(merged.top10same)/float32(len(r1)))
	fmt.Printf("Top100: %0.2f\n", float32(merged.top100same)/float32(len(r1)))
	fmt.Printf("Top1000: %0.2f\n", float32(merged.top1000same)/float32(len(r1)))
}

func getRanks(results []embeddings.EmbeddingSearchResult) []int {
	res := make([]int, len(results))
	for i, r := range results {
		res[i] = r.RowNum
	}
	return res
}

func compareRanks(a, b []int) compareResult {
	if len(a) != len(b) {
		panic("cannot compare slices of different lengths")
	}

	aIndexes := make(map[int]int, len(a))
	for i, val := range a {
		aIndexes[val] = i
	}

	var res compareResult
	for i, val := range b {
		prev, ok := aIndexes[val]
		if !ok {
			panic(fmt.Sprintf("found a rank that does not exist in both lists: %d", val))
		}

		diff := prev - i
		if diff < 0 {
			diff = -diff
		}
		res.count += 1
		res.totalMigration += diff

		if i == 0 && prev == 0 {
			res.top1same += 1
		}

		if i < 10 && prev < 10 {
			res.top10same += 1
		}

		if i < 100 && prev < 100 {
			res.top100same += 1
		}

		if i < 1000 && prev < 1000 {
			res.top1000same += 1
		}
	}

	return res
}

type compareResult struct {
	count          int
	totalMigration int
	top1same       int
	top10same      int
	top100same     int
	top1000same    int
}

func (cr compareResult) String() string {
	return fmt.Sprintf("Count: %d\nTotal: %d\nAverage: %0.2f\nPercent: %0.2f\nTop1: %d\nTop10: %d\nTop100: %d\nTop1000: %d\n",
		cr.count, cr.totalMigration,
		float32(cr.totalMigration)/float32(cr.count),
		float32(cr.totalMigration)/float32(cr.count)/float32(cr.count)*100,
		cr.top1same, cr.top10same, cr.top100same, cr.top1000same)
}
