package commitgraph

import (
	"slices"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type Graph struct {
	commitGraphView *CommitGraphView
	graph           map[api.CommitID][]api.CommitID
	commits         []api.CommitID
	ancestorUploads map[api.CommitID]map[string]UploadMeta
}

type Envelope struct {
	Uploads *VisibilityRelationship
	Links   *LinkRelationship
}

type VisibilityRelationship struct {
	Commit  api.CommitID
	Uploads []UploadMeta
}

type LinkRelationship struct {
	Commit         api.CommitID
	AncestorCommit api.CommitID
	Distance       uint32
}

// NewGraph creates a commit graph decorated with the set of uploads visible from that commit
// based on the given commit graph and complete set of LSIF upload metadata.
func NewGraph(commitGraph *CommitGraph, commitGraphView *CommitGraphView) *Graph {
	graph := commitGraph.Graph()
	order := commitGraph.Order()

	ancestorUploads := populateUploadsByTraversal(graph, order, commitGraphView)
	slices.Sort(order)

	return &Graph{
		commitGraphView: commitGraphView,
		graph:           graph,
		commits:         order,
		ancestorUploads: ancestorUploads,
	}
}

// UploadsVisibleAtCommit returns the set of uploads that are visible from the given commit.
func (g *Graph) UploadsVisibleAtCommit(commit api.CommitID) []UploadMeta {
	ancestorUploads, ancestorDistance := traverseForUploads(g.graph, g.ancestorUploads, commit)
	return adjustVisibleUploads(ancestorUploads, ancestorDistance)
}

// Stream returns a channel of envelope values which indicate either the set of visible uploads
// at a particular commit, or the nearest neighbors at a particular commit, depending on the
// value within the envelope.
func (g *Graph) Stream() <-chan Envelope {
	ch := make(chan Envelope)

	go func() {
		defer close(ch)

		for _, commit := range g.commits {
			if ancestorCommit, ancestorDistance, found := traverseForCommit(g.graph, g.ancestorUploads, commit); found {
				if ancestorVisibleUploads := g.ancestorUploads[ancestorCommit]; ancestorDistance == 0 || len(ancestorVisibleUploads) == 1 {
					// We have either a single upload (which is cheap enough to store), or we have
					// multiple uploads but we were assigned a value in  ancestorVisibleUploads. The
					// later case means that the visible uploads for this commit is data required to
					// reconstruct the visible uploads of a descendant commit.

					ch <- Envelope{
						Uploads: &VisibilityRelationship{
							Commit:  commit,
							Uploads: adjustVisibleUploads(ancestorVisibleUploads, ancestorDistance),
						},
					}
				} else if len(ancestorVisibleUploads) > 1 {
					// We have more than a single upload. Because we also have a very cheap way of
					// reconstructing this particular commit's visible uploads from the ancestor,
					// we store that relationship which is much smaller when the number of distinct
					// LSIF roots becomes large.

					ch <- Envelope{
						Links: &LinkRelationship{
							Commit:         commit,
							AncestorCommit: ancestorCommit,
							Distance:       ancestorDistance,
						},
					}
				}
			}
		}
	}()

	return ch
}

// Gather reads the graph's stream to completion and returns a map of the values. This
// method is only used for convenience and testing and should not be used in a hot path.
// It can be VERY memory intensive in production to have a reference to each commit's
// upload metadata concurrently.
func (g *Graph) Gather() (uploads map[api.CommitID][]UploadMeta, links map[api.CommitID]LinkRelationship) {
	uploads = map[api.CommitID][]UploadMeta{}
	links = map[api.CommitID]LinkRelationship{}

	for v := range g.Stream() {
		if v.Uploads != nil {
			uploads[v.Uploads.Commit] = v.Uploads.Uploads
		}
		if v.Links != nil {
			links[v.Links.Commit] = *v.Links
		}
	}

	return uploads, links
}

// reverseGraph returns the reverse of the given graph by flipping all the edges.
func reverseGraph(graph map[api.CommitID][]api.CommitID) map[api.CommitID][]api.CommitID {
	reverse := make(map[api.CommitID][]api.CommitID, len(graph))
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
//  1. They define an upload,
//  2. They have multiple parents, or
//  3. They have a child with multiple parents.
//
// For all remaining commits, we can easily re-calculate the visible uploads without storing them.
// All such commits have a single, unambiguous path to an ancestor that does store data. These
// commits have the same visibility (the descendant is just farther away).
func populateUploadsByTraversal(graph map[api.CommitID][]api.CommitID, order []api.CommitID, commitGraphView *CommitGraphView) map[api.CommitID]map[string]UploadMeta {
	reverseGraph := reverseGraph(graph)

	uploads := make(map[api.CommitID]map[string]UploadMeta, len(order))
	for _, commit := range order {
		parents := graph[commit]

		if _, ok := commitGraphView.Meta[commit]; !ok && len(graph[commit]) <= 1 {
			dedicatedChildren := true
			for _, child := range reverseGraph[commit] {
				if len(graph[child]) > 1 {
					dedicatedChildren = false
				}
			}

			if dedicatedChildren {
				continue
			}
		}

		ancestors := parents
		distance := uint32(1)

		// Find nearest ancestors with data. If we end the loop with multiple ancestors, we
		// know that they are all the same distance from the starting commit, and all of them
		// have data as they've already been processed and all satisfy the properties above.
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
//  1. the set of uploads defined on that commit, and
//  2. the set of uploads visible from the ancestors with the minimum distance
//     for equivalent root and indexer values.
//
// If two ancestors have different uploads visible for the same root and indexer, the one with the
// smaller distance to the source commit will shadow the other. Similarly, If an ancestor and the
// child commit define uploads for the same root and indexer pair, the upload defined on the commit
// will shadow the upload defined on the ancestor.
//
// IMPORTANT: This logic should be kept in sync with store.makeVisibleUploadsQuery.
func populateUploadsForCommit(uploads map[api.CommitID]map[string]UploadMeta, ancestors []api.CommitID, distance uint32, commitGraphView *CommitGraphView, commit api.CommitID) map[string]UploadMeta {
	// The capacity chosen here is an underestimate, but seems to perform well in benchmarks using
	// live user data. We have attempted to make this value more precise to minimize the number of
	// re-hash operations, but any counting we do requires auxiliary space and takes additional CPU
	// to traverse the graph.
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
			upload.Distance += distance

			// Only update upload for this token if distance of new upload is less than current one
			if currentUpload, ok := uploadsByToken[token]; !ok || replaces(upload, currentUpload) {
				uploadsByToken[token] = upload
			}
		}
	}

	return uploadsByToken
}

// traverseForUploads returns the value in the given uploads map whose key matches the first ancestor
// in the graph with a value present in the map. The distance in the graph between the original commit
// and the ancestor is also returned.
func traverseForUploads(graph map[api.CommitID][]api.CommitID, uploads map[api.CommitID]map[string]UploadMeta, commit api.CommitID) (map[string]UploadMeta, uint32) {
	commit, distance, _ := traverseForCommit(graph, uploads, commit)
	return uploads[commit], distance
}

// traverseForCommit returns the commit in the given uploads map matching the first ancestor in
// the graph with a value present in the map. The distance in the graph between the original commit
// and the ancestor is also returned.
//
// NOTE: We assume that each commit with multiple parents have been assigned data while walking
// the graph in topological order. If that is not the case, one parent will be chosen arbitrarily.
func traverseForCommit(graph map[api.CommitID][]api.CommitID, uploads map[api.CommitID]map[string]UploadMeta, commit api.CommitID) (api.CommitID, uint32, bool) {
	for distance := uint32(0); ; distance++ {
		if _, ok := uploads[commit]; ok {
			return commit, distance, true
		}

		parents := graph[commit]
		if len(parents) == 0 {
			return "", 0, false
		}

		commit = parents[0]
	}
}

// adjustVisibleUploads returns a copy of the given uploads map with the distance adjusted by
// the given amount. This returns the uploads "inherited" from a the nearest ancestor with
// commit data.
func adjustVisibleUploads(ancestorVisibleUploads map[string]UploadMeta, ancestorDistance uint32) []UploadMeta {
	uploads := make([]UploadMeta, 0, len(ancestorVisibleUploads))
	for _, ancestorUpload := range ancestorVisibleUploads {
		ancestorUpload.Distance += ancestorDistance
		uploads = append(uploads, ancestorUpload)
	}

	return uploads
}

// replaces returns true if upload1 has a smaller distance than upload2.
//
// NOTE(id: upload-tie-breaking): Ties are broken by the maximum upload
// identifier to remain determinstic, and to prefer newer uploads over older ones.
func replaces(upload1, upload2 UploadMeta) bool {
	return upload1.Distance < upload2.Distance || (upload1.Distance == upload2.Distance && upload1.UploadID > upload2.UploadID)
}
