package embeddings

import (
	"container/heap"
	"fmt"
	"math"
	"sort"

	"github.com/sourcegraph/conc"
)

type nearestNeighbor struct {
	index int
	score int32
	debug string
}

type nearestNeighborsHeap struct {
	neighbors []nearestNeighbor
}

func (nn *nearestNeighborsHeap) Len() int { return len(nn.neighbors) }

func (nn *nearestNeighborsHeap) Less(i, j int) bool {
	return nn.neighbors[i].score < nn.neighbors[j].score
}

func (nn *nearestNeighborsHeap) Swap(i, j int) {
	nn.neighbors[i], nn.neighbors[j] = nn.neighbors[j], nn.neighbors[i]
}

func (nn *nearestNeighborsHeap) Push(x any) {
	nn.neighbors = append(nn.neighbors, x.(nearestNeighbor))
}

func (nn *nearestNeighborsHeap) Pop() any {
	old := nn.neighbors
	n := len(old)
	x := old[n-1]
	nn.neighbors = old[0 : n-1]
	return x
}

func (nn *nearestNeighborsHeap) Peek() nearestNeighbor {
	return nn.neighbors[0]
}

func newNearestNeighborsHeap() *nearestNeighborsHeap {
	nn := &nearestNeighborsHeap{neighbors: make([]nearestNeighbor, 0)}
	heap.Init(nn)
	return nn
}

type partialRows struct {
	start int
	end   int
}

// splitRows splits nRows into nWorkers equal (or nearly equal) sized chunks.
func splitRows(numRows int, numWorkers int, minRowsToSplit int) []partialRows {
	if numWorkers == 1 || numRows <= numWorkers || numRows <= minRowsToSplit {
		return []partialRows{{0, numRows}}
	}
	nRowsPerWorker := int(math.Ceil(float64(numRows) / float64(numWorkers)))

	rowsPerWorker := make([]partialRows, numWorkers)
	for i := 0; i < numWorkers; i++ {
		rowsPerWorker[i] = partialRows{
			start: min(i*nRowsPerWorker, numRows),
			end:   min((i+1)*nRowsPerWorker, numRows),
		}
	}
	return rowsPerWorker
}

type WorkerOptions struct {
	NumWorkers int
	// MinRowsToSplit indicates the minimum number of rows that should be split
	// among the workers. If numRows <= MinRowsToSplit, then we use a single worker
	// to process the index, regardless of the NumWorkers option.
	MinRowsToSplit int
}

// SimilaritySearch finds the `nResults` most similar rows to a query vector. It uses the cosine similarity metric.
// IMPORTANT: The vectors in the embedding index have to be normalized for similarity search to work correctly.
func (index *EmbeddingIndex) SimilaritySearch(query []int8, numResults int, workerOptions WorkerOptions, opts SearchOptions) []EmbeddingSearchResult {
	if numResults == 0 {
		return []EmbeddingSearchResult{}
	}

	numRows := len(index.RowMetadata)
	// Cannot request more results than there are rows.
	numResults = min(numRows, numResults)
	// We need at least 1 worker.
	numWorkers := max(1, workerOptions.NumWorkers)

	// Split index rows among the workers. Each worker will run a partial similarity search on the assigned rows.
	rowsPerWorker := splitRows(numRows, numWorkers, workerOptions.MinRowsToSplit)
	heaps := make([]*nearestNeighborsHeap, len(rowsPerWorker))

	if len(rowsPerWorker) > 1 {
		var wg conc.WaitGroup
		for workerIdx := 0; workerIdx < len(rowsPerWorker); workerIdx++ {
			// Capture the loop variable value so we can use it in the closure below.
			workerIdx := workerIdx
			wg.Go(func() {
				heaps[workerIdx] = index.partialSimilaritySearch(query, numResults, rowsPerWorker[workerIdx], opts)
			})
		}
		wg.Wait()
	} else {
		// Run the similarity search directly when we have a single worker to eliminate the concurrency overhead.
		heaps[0] = index.partialSimilaritySearch(query, numResults, rowsPerWorker[0], opts)
	}

	// Collect all heap neighbors from workers into a single array.
	neighbors := make([]nearestNeighbor, 0, len(rowsPerWorker)*numResults)
	for _, heap := range heaps {
		if heap != nil {
			neighbors = append(neighbors, heap.neighbors...)
		}
	}
	// And re-sort it according to the score (descending).
	sort.Slice(neighbors, func(i, j int) bool { return neighbors[i].score > neighbors[j].score })

	// Take top neighbors and return them as results.
	results := make([]EmbeddingSearchResult, numResults)

	for idx := 0; idx < min(numResults, len(neighbors)); idx++ {
		results[idx] = EmbeddingSearchResult{
			RepoEmbeddingRowMetadata: index.RowMetadata[neighbors[idx].index],
			Debug:                    neighbors[idx].debug,
		}
	}

	return results
}

func (index *EmbeddingIndex) partialSimilaritySearch(query []int8, numResults int, partialRows partialRows, opts SearchOptions) *nearestNeighborsHeap {
	nRows := partialRows.end - partialRows.start
	if nRows <= 0 {
		return nil
	}
	numResults = min(nRows, numResults)

	nnHeap := newNearestNeighborsHeap()
	for i := partialRows.start; i < partialRows.start+numResults; i++ {
		score, debugInfo := index.score(query, i, opts)
		heap.Push(nnHeap, nearestNeighbor{index: i, score: score, debug: debugInfo})
	}

	for i := partialRows.start + numResults; i < partialRows.end; i++ {
		score, debugInfo := index.score(query, i, opts)
		// Add row if it has greater similarity than the smallest similarity in the heap.
		// This way we ensure keep a set of the highest similarities in the heap.
		if score > nnHeap.Peek().score {
			heap.Pop(nnHeap)
			heap.Push(nnHeap, nearestNeighbor{index: i, score: score, debug: debugInfo})
		}
	}

	return nnHeap
}

const (
	scoreFileRankWeight   int32 = 1
	scoreSimilarityWeight int32 = 2
)

func (index *EmbeddingIndex) score(query []int8, i int, opts SearchOptions) (score int32, debugInfo string) {
	addScore := func(what string, s int32) {
		score += s
		if opts.Debug {
			debugInfo += fmt.Sprintf("%s:%d, ", what, s)
		}
	}

	similarity := CosineSimilarity(index.row(i), query)

	addScore("similarity", scoreSimilarityWeight*similarity)

	// handle missing ranks
	if opts.UseDocumentRanks && len(index.Ranks) > i {
		// The file rank represents a log (base 2) count. The log ranks should be
		// bounded at 32, but we cap it just in case to ensure it falls in the range [0,
		// 1]. I am not using math.Min here to avoid the back and forth conversion
		// between float64 and float32.
		normalizedRank := index.Ranks[i] / 32.0
		if normalizedRank > 1.0 {
			normalizedRank = 1.0
		}
		addScore("rank", int32(float32(scoreFileRankWeight)*normalizedRank))
	}

	if opts.Debug {
		debugInfo = fmt.Sprintf("score: %d, %s", score, debugInfo)
	}

	return score, debugInfo
}

func CosineSimilarity(row []int8, query []int8) int32 {
	const unrollWidth = 4
	similarity := int32(0.0)
	unrolledIters := len(row) / unrollWidth
	for i := 0; i < unrolledIters; i += unrollWidth {
		// The loop here is unrolled so that the multiplications
		// can be pipelined. Got best benchmark results with unrolling
		// 4 at a time. Results may vary on other arches.
		m0 := int32(row[i]) * int32(query[i])
		m1 := int32(row[i+1]) * int32(query[i+1])
		m2 := int32(row[i+2]) * int32(query[i+2])
		m3 := int32(row[i+3]) * int32(query[i+3])
		similarity += m0 + m1 + m2 + m3
	}
	// Do any remainder element-wise. This ensures that we
	// can still handle dimensions that aren't multiples of 4.
	for i := unrolledIters * unrollWidth; i < len(row); i++ {
		similarity += int32(row[i]) * int32(query[i])
	}
	return similarity
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type SearchOptions struct {
	Debug            bool
	UseDocumentRanks bool
}
