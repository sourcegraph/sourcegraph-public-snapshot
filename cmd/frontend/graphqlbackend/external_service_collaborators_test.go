package graphqlbackend

import (
	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"

	"context"
	"sort"
	"testing"
)

func TestExternalServiceCollaborators_parallelRecentCommitters(t *testing.T) {
	ctx := context.Background()

	calls := []*github.RecentCommittersParams{}
	recentCommittersFunc := func(ctx context.Context, params *github.RecentCommittersParams) (*github.RecentCommittersResults, error) {
		calls = append(calls, params)

		var results github.RecentCommittersResults
		results.Nodes = append(results.Nodes, struct {
			Authors struct {
				Nodes []struct {
					Date      string
					Email     string
					Name      string
					AvatarURL string
				}
			}
		}{
			Authors: struct {
				Nodes []struct {
					Date      string
					Email     string
					Name      string
					AvatarURL string
				}
			}{
				Nodes: []struct {
					Date      string
					Email     string
					Name      string
					AvatarURL string
				}{
					{Name: params.Name + "-joe"},
					{Name: params.Name + "-jane"},
					{Name: params.Name + "-janet"},
				},
			},
		})

		return &results, nil
	}

	repos := []string{"gorilla/mux", "golang/go", "sourcegraph/sourcegraph"}
	recentCommitters, err := parallelRecentCommitters(ctx, repos, recentCommittersFunc)
	if err != nil {
		t.Fatal(err)
	}

	sort.Slice(calls, func(i, j int) bool {
		return calls[i].Name < calls[j].Name
	})

	autogold.Want("calls", []*github.RecentCommittersParams{
		{
			Name:  "go",
			Owner: "golang",
			First: 100,
		},
		{
			Name:  "mux",
			Owner: "gorilla",
			First: 100,
		},
		{
			Name:  "sourcegraph",
			Owner: "sourcegraph",
			First: 100,
		},
	}).Equal(t, calls)

	autogold.Want("recentCommitters", []*invitableCollaboratorResolver{
		{
			name: "sourcegraph-joe",
		},
		{name: "sourcegraph-jane"},
		{name: "sourcegraph-janet"},
		{name: "mux-joe"},
		{name: "mux-jane"},
		{name: "mux-janet"},
		{name: "go-joe"},
		{name: "go-jane"},
		{name: "go-janet"},
	}).Equal(t, recentCommitters)
}
