package commitgraph

import "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"

type VisibilityRelationship struct {
	Commit  string
	Uploads []UploadMeta
}

// CalculateVisibleUploads returns a channel returning values indicating that a given
// set of uploads are visible from a particular commit, based on the given commit graph
// and complete set of LSIF upload metadata.
func CalculateVisibleUploads(commitGraph *gitserver.CommitGraph, commitGraphView *CommitGraphView) <-chan VisibilityRelationship {
	graph := commitGraph.Graph()
	reverseGraph := reverseGraph(graph)
	order := commitGraph.Order()

	ch := make(chan VisibilityRelationship, len(order))

	go func() {
		defer close(ch)

		combineVisibleUploads(
			ch,
			graph,
			reverseGraph,
			order,
			populateUploadsByTraversal(graph, reverseGraph, order, commitGraphView, false),
			populateUploadsByTraversal(reverseGraph, graph, order, commitGraphView, true),
		)
	}()

	return ch
}

// CalculateVisibleUploadsForCommit returns the set of uploads that are visible from the
// given commit, based on the given commit graph and complete set of LSIF upload metadata.
func CalculateVisibleUploadsForCommit(commitGraph *gitserver.CommitGraph, commitGraphView *CommitGraphView, commit string) []UploadMeta {
	graph := commitGraph.Graph()
	reverseGraph := reverseGraph(graph)
	order := commitGraph.Order()

	ancestorUploads := populateUploadsByTraversal(graph, reverseGraph, order, commitGraphView, false)
	descendantUploads := populateUploadsByTraversal(reverseGraph, graph, order, commitGraphView, true)

	ancestorUploadByToken, ancestorDistance := traverse(graph, ancestorUploads, commit)
	descendantUploadsByToken, descendantDistance := traverse(reverseGraph, descendantUploads, commit)
	return combineVisibleUploadsForCommit(ancestorUploadByToken, descendantUploadsByToken, ancestorDistance, descendantDistance)
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

// populateUploadsByTraversal populates a map from select commits (see below) to another map from
// tokens to upload meta value. Select commits are any commits that satisfy one of the following
// properties:
//
//   1. They define an upload
//   2. They have multiple children
//   3. They have multiple parents
//   4. Their parent has multiple children
//   5. Their child has multiple parents
//
// For all remaining commits, we can easily re-calculate the visible uploads without storing them.
// All such commits have a parent whose only child is the commit (or has no parent), and a single
// child whose only parent is the commit (or has no children). This means that there is a single
// unambiguous path to an ancestor with calculated data, and symmetrically in the other direction.
// See combineVisibleUploads for additional details.
func populateUploadsByTraversal(graph, reverseGraph map[string][]string, order []string, commitGraphView *CommitGraphView, reverse bool) map[string]map[string]UploadMeta {
	uploads := make(map[string]map[string]UploadMeta, len(order))
	for i, commit := range order {
		if reverse {
			commit = order[len(order)-i-1]
		}

		parents := graph[commit]
		children := reverseGraph[commit]

		_, ok := commitGraphView.Meta[commit]
		if !ok && // ¬Property 1
			len(children) <= 1 && // ¬Property 2
			len(parents) <= 1 && // ¬Property 3
			(len(parents) == 0 || len(reverseGraph[parents[0]]) == 1) && // ¬Property 4
			(len(children) == 0 || len(reverseGraph[children[0]]) == 1) { // ¬Property 5
			continue
		}

		ancestors := parents
		distance := uint32(1)

		// Find nearest ancestors with data. If we end the loop with multiple ancestors, we know
		// that they are all the same distance from the starting commit, and all of them have
		// data as they've already been processed and all satisfy Property 5 above.
		for len(ancestors) == 1 {
			if _, ok := uploads[ancestors[0]]; ok {
				break
			}

			distance++
			ancestors = graph[ancestors[0]]
		}

		uploads[commit] = populateUploadsForCommit(uploads, ancestors, distance, commitGraphView, commit)
	}

	return uploads
}

// populateUploadsForCommit populates the items stored in the given mapping for the given commit.
// The uploads considered visible for a commit include:
//
//   1. the set of uploads defined on that commit, and
//   2. the set of  uploads visible from the ancestors (or descendants) with the minimum distance
//      for equivalent root and indexer values.
//
// If two ancestors have different uploads visible for the same root and indexer, the one with the
// smaller distance to the source commit will shadow the other. Similarly, If an ancestor and the
// child commit define uploads for the same root and indexer pair, the upload defined on the commit
// will shadow the upload defined on the ancestor.
func populateUploadsForCommit(uploads map[string]map[string]UploadMeta, ancestors []string, distance uint32, commitGraphView *CommitGraphView, commit string) map[string]UploadMeta {
	// The capacity chosen here is an underestimate, but seems to perform well in
	// benchmarks using live user data. We have attempted to make this value more
	// precise to minimize the number of re-hash operations, but any counting we
	// do requires auxiliary space and takes additional CPU to traverse the graph.
	capacity := len(commitGraphView.Meta[commit])
	for _, ancestor := range ancestors {
		if temp := len(uploads[ancestor]); temp > capacity {
			capacity = temp
		}
	}
	uploadsByToken := make(map[string]UploadMeta, capacity)

	// Populate uploads defined here
	for _, upload := range commitGraphView.Meta[commit] {
		token := commitGraphView.Tokens[upload.UploadID]
		uploadsByToken[token] = upload
	}

	// Combine with uploads visible from the nearest ancestors
	for _, ancestor := range ancestors {
		for _, upload := range uploads[ancestor] {
			token := commitGraphView.Tokens[upload.UploadID]

			// Increase distance from source before comparison
			upload.Flags += distance

			// Only update upload for this token if distance of new upload is less than current one
			if currentUpload, ok := uploadsByToken[token]; !ok || replaces(upload, currentUpload) {
				uploadsByToken[token] = upload
			}
		}
	}

	return uploadsByToken
}

// combineVisibleUploads writes values indicating the set of uploads visible at each commit in the
// given graph. This is done by combining the maps produced by traversing the graph in both directions
// and pre-computing the values for select commits (see populateUploadsByTraversal).
//
// For each commit, we determine the closet ancestor and descendant commits with pre-computed data.
// There is at most one of each commit. We then merge the uploads visible from these commits, with
// the distances of each upload meta value adjusted proportionally to the traversal distance.
//
// See combineVisibleUploadsForCommit for more details about the merge logic.
func combineVisibleUploads(ch chan<- VisibilityRelationship, graph, reverseGraph map[string][]string, order []string, ancestorUploads, descendantUploads map[string]map[string]UploadMeta) {
	for _, commit := range order {
		ancestorUploads, ancestorDistance := traverse(graph, ancestorUploads, commit)
		descendantUploads, descendantDistance := traverse(reverseGraph, descendantUploads, commit)
		uploads := combineVisibleUploadsForCommit(ancestorUploads, descendantUploads, ancestorDistance, descendantDistance)
		ch <- VisibilityRelationship{Commit: commit, Uploads: uploads}
	}
}

// traverse returns the value in the given uploads map whose key matches the first ancestor in the
// graph with a value present in the map. The distance in the graph between the original commit and
// the ancestor is also returned.
//
// NOTE: Explicitly assigning nil/empty map into the uploads map will cause the traversal to stop
// at that commit, returning a map with zero uploads. It is actually advised to assign nil maps
// at certain points in the graph so that traversals towards the edges of the graph are kept short.
//
// NOTE: We assume that each commit with multiple parents have been assigned data while walking
// the graph in topological order. If that is not the case, one parent will be chosen arbitrarily.
func traverse(graph map[string][]string, uploads map[string]map[string]UploadMeta, commit string) (uploadsByToken map[string]UploadMeta, distance uint32) {
	for distance := uint32(0); ; distance++ {
		if uploadsByToken, ok := uploads[commit]; ok {
			return uploadsByToken, distance
		}

		parents := graph[commit]
		if len(parents) == 0 {
			return nil, 0
		}

		commit = parents[0]
	}
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
func combineVisibleUploadsForCommit(ancestorVisibleUploads, descendantVisibleUploads map[string]UploadMeta, ancestorDistance, descendantDistance uint32) []UploadMeta {
	// The capacity chosen here is an underestimate, but seems to perform well in
	// benchmarks using live user data. See populateUploadsForCommit for a simlar
	// implementation note.
	capacity := len(ancestorVisibleUploads)
	if temp := len(descendantVisibleUploads); temp > capacity {
		capacity = temp
	}
	uploads := make([]UploadMeta, 0, capacity)

	for token, ancestorUpload := range ancestorVisibleUploads {
		ancestorUpload.Flags += ancestorDistance
		ancestorUpload.Flags |= FlagAncestorVisible

		if descendantUpload, ok := descendantVisibleUploads[token]; ok {
			descendantUpload.Flags += descendantDistance
			descendantUpload.Flags &^= FlagAncestorVisible

			// If the matching descendant upload has a smaller distance
			if replaces(descendantUpload, ancestorUpload) {
				// Mark the ancestor upload as overwritten
				ancestorUpload.Flags |= FlagOverwritten
				// Add the descendant upload
				uploads = append(uploads, descendantUpload)
			}
		}

		// Add all ancestor uploads
		uploads = append(uploads, ancestorUpload)
	}

	for token, descendantUpload := range descendantVisibleUploads {
		descendantUpload.Flags += descendantDistance
		descendantUpload.Flags &^= FlagAncestorVisible

		// Add all descendant uploads that have no ancestor matches
		if _, ok := ancestorVisibleUploads[token]; !ok {
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
