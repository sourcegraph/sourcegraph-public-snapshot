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
func splitRows(nRows int, nWorkers int) []partialRows {
	if nWorkers == 1 || nRows <= nWorkers {
		return []partialRows{{0, nRows}}
	}
	nRowsPerWorker := int(math.Ceil(float64(nRows) / float64(nWorkers)))

	rowsPerWorker := make([]partialRows, nWorkers)
	for i := 0; i < nWorkers; i++ {
		rowsPerWorker[i] = partialRows{
			start: min(i*nRowsPerWorker, nRows),
			end:   min((i+1)*nRowsPerWorker, nRows),
		}
	}
	return rowsPerWorker
}

// SimilaritySearch finds the `nResults` most similar rows to a query vector. It uses the cosine similarity metric.
// IMPORTANT: The vectors in the embedding index have to be normalized for similarity search to work correctly.
func (index *EmbeddingIndex[T]) SimilaritySearch(query []float32, nResults int, nWorkers int) []*T {
	if nResults == 0 {
		return []*T{}
	}

	nRows := len(index.RowMetadata)
	// Cannot request more results then there are rows.
	nResults = min(nRows, nResults)
	// We need at least 1 worker.
	nWorkers = max(1, nWorkers)

	// Split index rows among the workers. Each worker will run a partial similarity search on the assigned rows.
	rowsPerWorker := splitRows(nRows, nWorkers)
	heaps := make([]*nearestNeighborsHeap, len(rowsPerWorker))

	if len(rowsPerWorker) > 1 {
		var wg conc.WaitGroup
		for workerIdx := 0; workerIdx < len(rowsPerWorker); workerIdx++ {
			// Capture the loop variable value so we can use it in the closure below.
			workerIdx := workerIdx
			wg.Go(func() { heaps[workerIdx] = index.partialSimilaritySearch(query, nResults, rowsPerWorker[workerIdx]) })
		}
		wg.Wait()
	} else {
		// Run the similarity search directly when we have a single worker to eliminate the concurrency overhead.
		heaps[0] = index.partialSimilaritySearch(query, nResults, rowsPerWorker[0])
	}

	// Collect all heap neighbors from workers into a single array.
	neighbors := make([]nearestNeighbor, 0, len(rowsPerWorker)*nResults)
	for _, heap := range heaps {
		if heap != nil {
			neighbors = append(neighbors, heap.neighbors...)
		}
	}
	// And re-sort it according to the similarity (descending).
	sort.Slice(neighbors, func(i, j int) bool { return neighbors[i].similarity > neighbors[j].similarity })

	// Take top neighbors and return them as results.
	results := make([]*T, nResults)
	for idx := 0; idx < min(nResults, len(neighbors)); idx++ {
		results[idx] = &index.RowMetadata[neighbors[idx].index]
	}
	return results
}

func (index *EmbeddingIndex[T]) partialSimilaritySearch(query []float32, nResults int, partialRows partialRows) *nearestNeighborsHeap {
	nRows := partialRows.end - partialRows.start
	if nRows <= 0 {
		return nil
	}
	nResults = min(nRows, nResults)

	nnHeap := newNearestNeighborsHeap()
	for i := partialRows.start; i < partialRows.start+nResults; i++ {
		similarity := CosineSimilarity(
			index.Embeddings[i*index.ColumnDimension:(i+1)*index.ColumnDimension],
			query,
		)
		heap.Push(nnHeap, nearestNeighbor{i, similarity})
	}

	for i := partialRows.start + nResults; i < partialRows.end; i++ {
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

// TODO: Can potentially inline this for any performance benefits?
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
