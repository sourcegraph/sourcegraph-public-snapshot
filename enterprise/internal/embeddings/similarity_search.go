package embeddings

import (
	"container/heap"
	"math"
	"sort"

	"github.com/sourcegraph/conc"
)

type nearestNeighbor struct {
	index      int
	similarity float32
}

type nearestNeighborsHeap struct {
	neighbors []nearestNeighbor
}

func (nn *nearestNeighborsHeap) Len() int { return len(nn.neighbors) }

func (nn *nearestNeighborsHeap) Less(i, j int) bool {
	return nn.neighbors[i].similarity < nn.neighbors[j].similarity
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
func (index *EmbeddingIndex[T]) SimilaritySearch(query []float32, numResults int, workerOptions WorkerOptions) []*T {
	if numResults == 0 {
		return []*T{}
	}

	numRows := len(index.RowMetadata)
	// Cannot request more results then there are rows.
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
			wg.Go(func() { heaps[workerIdx] = index.partialSimilaritySearch(query, numResults, rowsPerWorker[workerIdx]) })
		}
		wg.Wait()
	} else {
		// Run the similarity search directly when we have a single worker to eliminate the concurrency overhead.
		heaps[0] = index.partialSimilaritySearch(query, numResults, rowsPerWorker[0])
	}

	// Collect all heap neighbors from workers into a single array.
	neighbors := make([]nearestNeighbor, 0, len(rowsPerWorker)*numResults)
	for _, heap := range heaps {
		if heap != nil {
			neighbors = append(neighbors, heap.neighbors...)
		}
	}
	// And re-sort it according to the similarity (descending).
	sort.Slice(neighbors, func(i, j int) bool { return neighbors[i].similarity > neighbors[j].similarity })

	// Take top neighbors and return them as results.
	results := make([]*T, numResults)
	for idx := 0; idx < min(numResults, len(neighbors)); idx++ {
		results[idx] = &index.RowMetadata[neighbors[idx].index]
	}
	return results
}

func (index *EmbeddingIndex[T]) partialSimilaritySearch(query []float32, numResults int, partialRows partialRows) *nearestNeighborsHeap {
	nRows := partialRows.end - partialRows.start
	if nRows <= 0 {
		return nil
	}
	numResults = min(nRows, numResults)

	nnHeap := newNearestNeighborsHeap()
	for i := partialRows.start; i < partialRows.start+numResults; i++ {
		similarity := CosineSimilarity(
			index.Embeddings[i*index.ColumnDimension:(i+1)*index.ColumnDimension],
			query,
		)
		heap.Push(nnHeap, nearestNeighbor{i, similarity})
	}

	for i := partialRows.start + numResults; i < partialRows.end; i++ {
		similarity := CosineSimilarity(
			index.Embeddings[i*index.ColumnDimension:(i+1)*index.ColumnDimension],
			query,
		)
		// Add row if it has greater similarity than the smallest similarity in the heap.
		// This way we ensure keep a set of highest similarities in the heap.
		if similarity > nnHeap.Peek().similarity {
			heap.Pop(nnHeap)
			heap.Push(nnHeap, nearestNeighbor{i, similarity})
		}
	}

	return nnHeap
}

func CosineSimilarity(row []float32, query []float32) float32 {
	similarity := float32(0.0)
	for i := 0; i < len(row); i++ {
		similarity += (row[i] * query[i])
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
