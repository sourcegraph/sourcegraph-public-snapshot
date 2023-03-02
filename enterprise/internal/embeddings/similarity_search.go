package embeddings

import (
	"container/heap"
	"sort"
)

type nearestNeighbor struct {
	index      int
	similarity float32
}

type nearestNeighborsHeap struct {
	heap []nearestNeighbor
}

func (nn *nearestNeighborsHeap) Len() int { return len(nn.heap) }

func (nn *nearestNeighborsHeap) Less(i, j int) bool {
	return nn.heap[i].similarity < nn.heap[j].similarity
}

func (nn *nearestNeighborsHeap) Swap(i, j int) { nn.heap[i], nn.heap[j] = nn.heap[j], nn.heap[i] }

func (nn *nearestNeighborsHeap) Push(x any) {
	nn.heap = append(nn.heap, x.(nearestNeighbor))
}

func (nn *nearestNeighborsHeap) Pop() any {
	old := nn.heap
	n := len(old)
	x := old[n-1]
	nn.heap = old[0 : n-1]
	return x
}

func (nn *nearestNeighborsHeap) Peek() nearestNeighbor {
	return nn.heap[0]
}

func newNearestNeighborsHeap() *nearestNeighborsHeap {
	nn := &nearestNeighborsHeap{heap: make([]nearestNeighbor, 0)}
	heap.Init(nn)
	return nn
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SimilaritySearch finds the `nResults` most similar rows to a query vector. It uses the cosine similarity metric.
// IMPORTANT: The vectors in the embedding index have to be normalized for similarity search to work correctly.
func (index *EmbeddingIndex[T]) SimilaritySearch(query []float32, nResults int) []*T {
	// TODO: Parallelize. Split the rows among N threads, each finds `nResults` within its chunk, combine the heaps, sort, return top `nResults`.
	if nResults == 0 {
		return []*T{}
	}

	nRows := len(index.RowMetadata)
	nResults = min(nRows, nResults)

	nnHeap := newNearestNeighborsHeap()
	for i := 0; i < nResults; i++ {
		similarity := CosineSimilarity(
			index.Embeddings[i*index.ColumnDimension:(i+1)*index.ColumnDimension],
			query,
		)
		heap.Push(nnHeap, nearestNeighbor{i, similarity})
	}

	for i := nResults; i < nRows; i++ {
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

	sort.Slice(nnHeap.heap, func(i, j int) bool {
		return nnHeap.heap[i].similarity > nnHeap.heap[j].similarity
	})

	results := make([]*T, len(nnHeap.heap))
	for idx, neighbor := range nnHeap.heap {
		results[idx] = &index.RowMetadata[neighbor.index]
	}
	return results
}

// TODO: Can potentially inline this for any performance benefits?
func CosineSimilarity(row []float32, query []float32) float32 {
	similarity := float32(0.0)
	for i := 0; i < len(row); i++ {
		similarity += (row[i] * query[i])
	}
	return similarity
}
