package dbstore

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
)

// UploadMeta contains the subset of fields from the lsif_uploads table that are used
// to determine visibility of an upload from a particular commit.
type UploadMeta struct {
	UploadID int
	Root     string
	Indexer  string

	// Distance is the number of commits between the definition and reference commits.
	Distance int

	// AncestorVisible indicates whether or not this upload was visible to the commit by
	// looking for older uploads defined in an ancestor commit. False indicates that the
	// upload is only visible via another upload defined in a descendant commit.
	AncestorVisible bool

	// Overwritten indicates that this upload defined on an ancestor commit that is farther
	// away than an upload defined on a descendant commit that has the same root and indexer.
	//
	// If overwritten, this upload is not considered in nearest commit operations, but is kept
	// in the database so that we can reconstruct the set of all ancestor-visible uploads of a
	// commit, which is useful when determining the closest uploads with only a partial commit
	// graph.
	Overwritten bool
}

// calculateVisibleUploads transforms the given commit graph and the set of LSIF uploads
// defined on each commit with LSIF upload into a map from a commit to the set of uploads
// which are visible from that commit.
func calculateVisibleUploads(commitGraph *gitserver.CommitGraph, uploads map[string][]UploadMeta) (map[string][]UploadMeta, error) {
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
			for _, upload := range uploads[commit] {
				upload.AncestorVisible = true
				ancestorVisibleUploads[commit] = append(ancestorVisibleUploads[commit], upload)
			}
		}

		for _, commit := range order {
			uploads := ancestorVisibleUploads[commit]
			for _, parent := range graph[commit] {
				for _, upload := range ancestorVisibleUploads[parent] {
					upload.Distance++
					uploads = addUploadMeta(uploads, upload, true)
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
			for _, upload := range uploads[commit] {
				upload.AncestorVisible = false
				descendantVisibleUploads[commit] = append(descendantVisibleUploads[commit], upload)
			}
		}

		// Calculate mapping from commits to their children
		reverseGraph := reverseGraph(graph)

		for i := len(order) - 1; i >= 0; i-- {
			commit := order[i]
			uploads := descendantVisibleUploads[commit]
			for _, child := range reverseGraph[commit] {
				for _, upload := range descendantVisibleUploads[child] {
					upload.Distance++
					uploads = addUploadMeta(uploads, upload, true)
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
			uploads = addUploadMeta(uploads, upload, false)
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
func addUploadMeta(uploads []UploadMeta, upload UploadMeta, replace bool) []UploadMeta {
	for i, candidate := range uploads {
		if upload.Root == candidate.Root && upload.Indexer == candidate.Indexer {
			if upload.Distance < candidate.Distance || (upload.Distance == candidate.Distance && upload.UploadID < candidate.UploadID) {
				u := uploads
				if !replace {
					u = append(uploads, UploadMeta{
						UploadID:        uploads[i].UploadID,
						Root:            uploads[i].Root,
						Indexer:         uploads[i].Indexer,
						Distance:        uploads[i].Distance,
						AncestorVisible: uploads[i].AncestorVisible,
						Overwritten:     true,
					})
				}
				u[i] = upload
				return u
			}

			return uploads
		}
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
