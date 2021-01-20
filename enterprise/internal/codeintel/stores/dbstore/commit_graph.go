package dbstore

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
)

// calculateVisibleUploads transforms the given commit graph and the set of LSIF uploads
// defined on each commit with LSIF upload into a map from a commit to the set of uploads
// which are visible from that commit.
func calculateVisibleUploads(commitGraph *gitserver.CommitGraph, commitGraphView *CommitGraphView) (map[string][]UploadMeta, error) {
	graph := commitGraph.Graph()
	order := commitGraph.Order()

	var wg sync.WaitGroup
	wg.Add(2)

	// ancestorVisibleUploads maps commits to the set of uploads visible by looking up the ancestor
	// paths of that commit. This map is populated by first inserting each upload at the commit
	// where it is defined. Then we populate the remaining data by walking the graph in topological
	// order (parents before children), and "push" each visible upload down descendant paths. At each
	// commit, if multiple uploads with the same root and indexer are visible, the one with the minmum
	// distance from the source commit will be used.
	ancestorVisibleUploads := map[string][]UploadMeta{}

	go func() {
		defer wg.Done()

		for commit := range graph {
			for _, upload := range commitGraphView.Meta[commit] {
				ancestorVisibleUploads[commit] = append(ancestorVisibleUploads[commit], UploadMeta{
					UploadID: upload.UploadID,
					Flags:    upload.Flags | FlagAncestorVisible,
				})
			}
		}

		for _, commit := range order {
			uploads := ancestorVisibleUploads[commit]
			for _, parent := range graph[commit] {
				for _, upload := range ancestorVisibleUploads[parent] {
					uploads = addUploadMeta(uploads, commitGraphView, UploadMeta{
						UploadID: upload.UploadID,
						Flags:    upload.Flags + 1,
					}, true)
				}
			}
			ancestorVisibleUploads[commit] = uploads
		}
	}()

	// descendantVisibleUploads maps commits to the set of uploads visible by looking down the
	// descendant paths of that commit. This map is populated by first inserting each upload at the
	// commit where it is defined. Then we populate the remaining data by walking the graph in reverse
	// topological order (children before parents), and "push" each visible upload up ancestor paths.
	// At each  commit, if multiple uploads with the same root and indexer are visible, the one with
	// the minmum  distance from the source commit will be used.
	descendantVisibleUploads := map[string][]UploadMeta{}

	go func() {
		defer wg.Done()

		for commit := range graph {
			for _, upload := range commitGraphView.Meta[commit] {
				descendantVisibleUploads[commit] = append(descendantVisibleUploads[commit], UploadMeta{
					UploadID: upload.UploadID,
					Flags:    upload.Flags &^ FlagAncestorVisible,
				})
			}
		}

		// Calculate mapping from commits to their children
		reverseGraph := reverseGraph(graph)

		for i := len(order) - 1; i >= 0; i-- {
			commit := order[i]
			uploads := descendantVisibleUploads[commit]
			for _, child := range reverseGraph[commit] {
				for _, upload := range descendantVisibleUploads[child] {
					uploads = addUploadMeta(uploads, commitGraphView, UploadMeta{
						UploadID: upload.UploadID,
						Flags:    upload.Flags + 1,
					}, true)
				}
			}
			descendantVisibleUploads[commit] = uploads
		}
	}()

	wg.Wait()

	// For each commit, merge the set of uploads visible by looking in each direction from that
	// commit. We do this by merging the descendant-visible uploads on top of the ancestor-visible
	// uploads. If we find a ancestor-visible upload and a descendant-visible upload with the same
	// root and  indexer, where the descendant-visible upload has a smaller distance, then we keep
	// both upload entries, but mark the ancestor-visible upload as overwritten.
	//
	// This produces a list of visible uploads with the properties we need for several features:
	//   - We can ask for only the non-overwritten uploads, which will give us the set of
	//     visible uploads with minimal distance we need to determine the (actual) nearest
	//     upload for a given commit.
	//   - We can ask for only the ancestor-visible uploads, which gives us the partial
	//     graph we need to determine the nearest upload for a commit without needing to
	//     operate over the entire commit graph. See the method code intel store method
	//     FindClosestDumpsFromGraphFragment for additional details.
	for commit, uploads := range ancestorVisibleUploads {
		for _, upload := range descendantVisibleUploads[commit] {
			uploads = addUploadMeta(uploads, commitGraphView, upload, false)
		}

		ancestorVisibleUploads[commit] = uploads
	}

	return ancestorVisibleUploads, nil
}

// addUploadMeta merges the given upload metadata into the given list, resolving conflicts.
//
// If there already exists an upload with the same root and indexer but a larger distance, then
// that upload will take the place of the existing upload (if replace is true), or the existing
// upload will be marked as overwritten and the upload will be appended to the end of the list
// (if replace is false). If there is no such upload with the same root and indexer, then the
// given upload will be appended to the end of the list.
func addUploadMeta(uploads []UploadMeta, commitGraphView *CommitGraphView, upload UploadMeta, replace bool) []UploadMeta {
	sharedFieldToken := commitGraphView.Tokens[upload.UploadID]

	for i, candidate := range uploads {
		candidateSharedFieldToken := commitGraphView.Tokens[candidate.UploadID]
		if sharedFieldToken != candidateSharedFieldToken {
			continue
		}

		uploadDistance := upload.Flags & MaxDistance
		candidateDistance := candidate.Flags & MaxDistance

		if uploadDistance < candidateDistance || (uploadDistance == candidateDistance && upload.UploadID < candidate.UploadID) {
			if !replace {
				uploads = append(uploads, UploadMeta{
					UploadID: candidate.UploadID,
					Flags:    candidate.Flags | FlagOverwritten,
				})
			}
			uploads[i] = upload
			return uploads
		}

		return uploads
	}

	return append(uploads, upload)
}

// reverseGraph returns the reverse of the given graph by flipping all the edges.
func reverseGraph(graph map[string][]string) map[string][]string {
	reverse := map[string][]string{}
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
