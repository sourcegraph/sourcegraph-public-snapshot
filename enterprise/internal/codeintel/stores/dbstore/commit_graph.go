package dbstore

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
)

// calculateVisibleUploads transforms the given commit graph and the set of LSIF uploads
// defined on each commit with LSIF upload into a map from a commit to the set of uploads
// which are visible from that commit.
func calculateVisibleUploads(commitGraph *gitserver.CommitGraph, commitGraphView *CommitGraphView) map[string][]UploadMeta {
	graph := commitGraph.Graph()
	order := commitGraph.Order()
	reverseGraph := reverseGraph(graph)

	var wg sync.WaitGroup
	wg.Add(2)

	ancestorVisibleUploads := make(map[string]map[string]UploadMeta, len(order))
	go func() {
		defer wg.Done()
		populateUploadsByTraversal(ancestorVisibleUploads, graph, order, commitGraphView, false)
	}()

	descendantVisibleUploads := make(map[string]map[string]UploadMeta, len(order))
	go func() {
		defer wg.Done()
		populateUploadsByTraversal(descendantVisibleUploads, reverseGraph, order, commitGraphView, true)
	}()

	wg.Wait()

	return combineVisibleUploads(ancestorVisibleUploads, descendantVisibleUploads, order)
}

// reverseGraph returns the reverse of the given graph by flipping all the edges.
func reverseGraph(graph map[string][]string) map[string][]string {
	reverse := make(map[string][]string, len(graph))
	for child := range graph {
		reverse[child] = nil
	}

	for child, parents := range graph {
		for _, parent := range parents {
			reverse[parent] = append(reverse[parent], child)
		}
	}

	return reverse
}

// populateUploadsByTraversal populates the given mapping by traversing the graph in one direction,
// either up ancestor paths or down descendant paths, depending on the encoding of the given graph
// and the reverse parameter.
func populateUploadsByTraversal(uploads map[string]map[string]UploadMeta, graph map[string][]string, order []string, commitGraphView *CommitGraphView, reverse bool) {
	for i, commit := range order {
		if reverse {
			commit = order[len(order)-i-1]
		}

		uploads[commit] = populateUploadsForCommit(uploads, graph, commitGraphView, commit)
	}
}

// populateUploadsForCommit populates the items stored in the given mapping for the given commit.
// The uploads considered visible for a commit include:
//
//   1. the set of uploads defined on that commit, and
//   2. the set of  uploads visible from the parents (or children) with the minimum distance
//      for equivalent root and indexer values.
//
// If two parents have different uploads visible for the same root and indexer, the one with the
// smaller distance to the source commit will shadow the other. Similarly, If a parent and the child
// commit define uploads for the same root and indexer pair, the upload defined on the commit will
// shadow the upload defined on the parent.
func populateUploadsForCommit(uploads map[string]map[string]UploadMeta, graph map[string][]string, commitGraphView *CommitGraphView, commit string) map[string]UploadMeta {
	capacity := len(commitGraphView.Meta[commit])
	for _, parent := range graph[commit] {
		if temp := len(uploads[parent]); temp > capacity {
			capacity = temp
		}
	}
	uploadsByToken := make(map[string]UploadMeta, capacity)

	for _, upload := range commitGraphView.Meta[commit] {
		token := commitGraphView.Tokens[upload.UploadID]
		uploadsByToken[token] = upload
	}

	for _, parent := range graph[commit] {
		for _, upload := range uploads[parent] {
			token := commitGraphView.Tokens[upload.UploadID]

			// Increase distance from source before comparison
			upload.Flags++

			// Only update upload for this token if distance of new upload is less than current one
			if currentUpload, ok := uploadsByToken[token]; !ok || replaces(upload, currentUpload) {
				uploadsByToken[token] = upload
			}
		}
	}

	return uploadsByToken
}

// combineVisibleUploads combines the set of uploads visible by traversing the ancestor and descendant
// paths for ever commit. See combineVisibleUploadsForCommit for more details about the merge logic.
func combineVisibleUploads(ancestorVisibleUploads, descendantVisibleUploads map[string]map[string]UploadMeta, order []string) map[string][]UploadMeta {
	combined := make(map[string][]UploadMeta, len(order))
	for _, commit := range order {
		combined[commit] = combineVisibleUploadsForCommit(ancestorVisibleUploads, descendantVisibleUploads, commit)
	}

	return combined
}

// combineVisibleUploadsForCommit combines sets of uploads visible by looking in opposite directions
// in the graph. This will produce a flat list of upload meta objects for each commit that consists of:
//
//   1. the set of ancestor-visible uploads,
//   2. the set of descendant-visible uploads where there does not exist an ancestor-visible upload
//      with an equivalent root an indexer value, and
//   3. the set of descendant-visible uploads where there exists an ancestor-visible upload with an
//      equivalent root and indexer value but a greater distance. In this case, the ancestor-visible
//      upload is also present in the list, but is flagged as overwritten.
func combineVisibleUploadsForCommit(ancestorVisibleUploads, descendantVisibleUploads map[string]map[string]UploadMeta, commit string) []UploadMeta {
	capacity := len(ancestorVisibleUploads[commit])
	if temp := len(descendantVisibleUploads[commit]); temp > capacity {
		capacity = temp
	}
	uploads := make([]UploadMeta, 0, capacity)

	for token, ancestorUpload := range ancestorVisibleUploads[commit] {
		if descendantUpload, ok := descendantVisibleUploads[commit][token]; ok {
			if replaces(descendantUpload, ancestorUpload) {
				// Clear direction flag
				descendantUpload.Flags &^= FlagAncestorVisible
				uploads = append(uploads, descendantUpload)

				// Mark upload as overwritten by descendant-visible upload
				ancestorUpload.Flags |= FlagOverwritten
			}
		}

		// Set direction flag
		ancestorUpload.Flags |= FlagAncestorVisible
		uploads = append(uploads, ancestorUpload)
	}

	for token, descendantUpload := range descendantVisibleUploads[commit] {
		if _, ok := ancestorVisibleUploads[commit][token]; !ok {
			// Clear direction flag
			descendantUpload.Flags &^= FlagAncestorVisible
			uploads = append(uploads, descendantUpload)
		}
	}

	return uploads
}

// replaces returns true if upload1 has a smaller distance than upload2.
// Ties are broken by the minimum upload identifier to remain determinstic.
func replaces(upload1, upload2 UploadMeta) bool {
	distance1 := upload1.Flags & MaxDistance
	distance2 := upload2.Flags & MaxDistance

	return distance1 < distance2 || (distance1 == distance2 && upload1.UploadID < upload2.UploadID)
}
