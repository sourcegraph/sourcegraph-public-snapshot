package store

import (
	"github.com/pkg/errors"
)

// UploadMeta contains the subset of fields from the lsif_uploads table that are used
// to determine visibility of an upload from a particular commit.
type UploadMeta struct {
	UploadID int
	Root     string
	Indexer  string

	// Distance is the number of commits between the reference to definition commits.
	Distance int
}

// calculateVisibleUploads transforms the given commit graph and the set of LSIF uploads
// defined on each commit with LSIF upload into a map from a commit to the set of uploads
// which are visible from that commit.
func calculateVisibleUploads(graph map[string][]string, uploads map[string][]UploadMeta) (map[string][]UploadMeta, error) {
	// Calculate an ordering of vertices so that all children come before parents.
	// Iterating this order will walk you "up" the commit graph.
	ordering, err := topologicalSort(graph)
	if err != nil {
		return nil, err
	}

	// Calculate the reverse graph so we can efficiently look up the set of children
	// for each vertex.
	reverseGraph := reverseGraph(graph)

	// Create two distinct mappings commits to the set of visible uploads populated at
	// first with only the uploads that the commit defines. We will then "push" these
	// uploads up to ancestors and down to descendants by traversing twice.
	forwardUploads := map[string][]UploadMeta{}
	reverseUploads := map[string][]UploadMeta{}
	for commit := range graph {
		for _, uploadMeta := range uploads[commit] {
			forwardUploads[commit] = append(forwardUploads[commit], uploadMeta)
			reverseUploads[commit] = append(reverseUploads[commit], uploadMeta)
		}
	}

	// Forward direction:
	// Iterate the vertices in topological order (children before parents) and push the
	// set of visible uploads "down" the tree. Each upload a vertex can see is the set of
	// uploads its parents can see with an increased distance. If two parents can see an
	// upload for the same root and indexer, we keep only the one with the minimum distance.
	for _, commit := range ordering {
		uploads := forwardUploads[commit]
		for _, parent := range graph[commit] {
			for _, upload := range forwardUploads[parent] {
				uploads = addUploadMeta(uploads, upload.UploadID, upload.Root, upload.Indexer, upload.Distance+1)
			}
		}

		forwardUploads[commit] = uploads
	}

	// Reverse direction:
	// Iterate the vertices in reverse topological order (parents before children) and push
	// the set of visible uploads "up" the tree. Each upload a vertex can see is the set of
	// uploads its children can see with an increased distance. If two children can see an
	// upload for the same root and indexer, we keep only the one with the minimum distance.
	for i := len(ordering) - 1; i >= 0; i-- {
		commit := ordering[i]
		uploads := reverseUploads[commit]
		for _, child := range reverseGraph[commit] {
			for _, upload := range reverseUploads[child] {
				uploads = addUploadMeta(uploads, upload.UploadID, upload.Root, upload.Indexer, upload.Distance+1)
			}
		}

		reverseUploads[commit] = uploads
	}

	// Combine directions:
	// Merge the set of visible upload for each commit with the same rules as above: for each
	// pair of uploads with the same root and indexer, keep only the one with the minimum distance.
	// After this merge, each commit can see uploads defined by any direct ancestor or descendant,
	// but cannot see uploads which are defined by relatives which require you to reverse traversal
	// direction to find.
	for commit, uploads := range forwardUploads {
		for _, upload := range reverseUploads[commit] {
			uploads = addUploadMeta(uploads, upload.UploadID, upload.Root, upload.Indexer, upload.Distance)
		}

		forwardUploads[commit] = uploads
	}

	return forwardUploads, nil
}

// addUploadMeta adds the given upload metadata to the given list. If there already exists an upload
// with the same root and indexer, then that upload will be replaced if it has a greater distance.
// The list is returned unmodified if such an upload has a smaller distance.
func addUploadMeta(uploads []UploadMeta, uploadID int, root, indexer string, distance int) []UploadMeta {
	for i, x := range uploads {
		if root == x.Root && indexer == x.Indexer {
			if distance < x.Distance || (distance == x.Distance && uploadID < x.UploadID) {
				uploads[i].UploadID = uploadID
				uploads[i].Distance = distance
			}

			return uploads
		}
	}

	return append(uploads, UploadMeta{
		UploadID: uploadID,
		Root:     root,
		Indexer:  indexer,
		Distance: distance,
	})
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

type sortMarker int

const (
	markTemp sortMarker = iota
	markPermenant
)

var ErrCyclicCommitGraph = errors.New("commit graph contains cycles")

// topologicalSort returns an ordering of the vertices of the given graph such that
// each vertex occurs before all of its children. The input graph is expected to be
// represented as a mapping from a vertex to its parents (a la git log) rather than
// a mapping of children. If the graph is not acyclic an error is returned as no
// such ordering can exist.
func topologicalSort(graph map[string][]string) (_ []string, err error) {
	commits := make([]string, 0, len(graph))
	visited := make(map[string]sortMarker, len(graph))

	for commit := range graph {
		if commits, err = visitForTopologicalSort(graph, commits, visited, commit); err != nil {
			return nil, err
		}
	}

	return commits, nil
}

// visitForTopologicalSort performs a post-order DFS traversal of the given graph
// from the given target commit. Commits are uniquely appended to the commits slice
// only after all of their descendants have been added.
func visitForTopologicalSort(graph map[string][]string, commits []string, visited map[string]sortMarker, commit string) (_ []string, err error) {
	if mark, ok := visited[commit]; ok {
		if mark == markTemp {
			return nil, ErrCyclicCommitGraph
		}

		return commits, nil
	}

	parents, ok := graph[commit]
	if !ok {
		return commits, nil
	}

	visited[commit] = markTemp

	for _, parent := range parents {
		if commits, err = visitForTopologicalSort(graph, commits, visited, parent); err != nil {
			return nil, err
		}
	}

	visited[commit] = markPermenant
	commits = append(commits, commit)
	return commits, nil
}
