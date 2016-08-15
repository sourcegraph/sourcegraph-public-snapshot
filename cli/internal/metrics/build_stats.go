package metrics

import (
	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func ComputeBuildStats(cl *sourcegraph.Client, ctx context.Context) (map[string]int32, error) {
	buildsList, err := cl.Builds.List(ctx, &sourcegraph.BuildListOptions{
		Sort:        "created_at",
		Direction:   "desc",
		ListOptions: sourcegraph.ListOptions{PerPage: 10000},
	})
	if err != nil {
		return nil, err
	}

	numBuilds := map[string]int32{
		"queued":    0,
		"active":    0,
		"ended":     0,
		"succeeded": 0,
		"failed":    0,
		"purged":    0,
		"total":     0,
	}
	for _, b := range buildsList.Builds {
		if b.Queue && b.StartedAt == nil {
			numBuilds["queued"] += 1
		}
		if b.StartedAt != nil && b.EndedAt == nil {
			numBuilds["active"] += 1
		}
		if b.EndedAt != nil {
			numBuilds["ended"] += 1
		}
		if b.Success {
			numBuilds["succeeded"] += 1
		}
		if b.Failure {
			numBuilds["failed"] += 1
		}
		if b.Purged {
			numBuilds["purged"] += 1
		}
	}
	numBuilds["total"] = int32(len(buildsList.Builds))
	return numBuilds, nil
}
