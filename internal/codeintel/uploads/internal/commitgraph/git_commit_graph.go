package commitgraph

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type CommitGraph struct {
	graph map[api.CommitID][]api.CommitID
	order []api.CommitID
}

func (c *CommitGraph) Graph() map[api.CommitID][]api.CommitID { return c.graph }
func (c *CommitGraph) Order() []api.CommitID                  { return c.order }

// ParseCommitGraph converts the output of git log into a map from commits to
// parent commits, and a topological ordering of commits such that parents come
// before children. If a commit is listed but has no ancestors then its parent
// slice is empty, but is still present in the map and the ordering. If the
// ordering is to be correct, the given commits must be ordered with
// gitserver.CommitsOrderTopoDate.
func ParseCommitGraph(commits []*gitdomain.Commit) *CommitGraph {
	// Process lines backwards so that we see all parents before children. We get a
	// topological ordering by simply scraping the keys off in this order.

	n := len(commits) - 1
	for i := range len(commits) / 2 {
		commits[i], commits[n-i] = commits[n-i], commits[i]
	}

	graph := make(map[api.CommitID][]api.CommitID, len(commits))
	order := make([]api.CommitID, 0, len(commits))

	var prefix []api.CommitID
	for _, commit := range commits {
		if len(commit.Parents) == 0 {
			graph[commit.ID] = []api.CommitID{}
		} else {
			graph[commit.ID] = commit.Parents
		}

		order = append(order, commit.ID)

		for _, parent := range commit.Parents {
			if _, ok := graph[parent]; !ok {
				graph[parent] = []api.CommitID{}
				prefix = append(prefix, parent)
			}
		}
	}

	return &CommitGraph{
		graph: graph,
		order: append(prefix, order...),
	}
}
