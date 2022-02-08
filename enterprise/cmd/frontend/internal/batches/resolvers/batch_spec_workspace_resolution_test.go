package resolvers

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/search"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

func TestWorkspacesListArgsToDBOpts(t *testing.T) {
	tcs := []struct {
		name string
		args *graphqlbackend.ListWorkspacesArgs
		want store.ListBatchSpecWorkspacesOpts
	}{
		{
			name: "empty",
			args: &graphqlbackend.ListWorkspacesArgs{},
		},
		{
			name: "first set",
			args: &graphqlbackend.ListWorkspacesArgs{
				First: 1,
			},
			want: store.ListBatchSpecWorkspacesOpts{
				LimitOpts: store.LimitOpts{Limit: 1},
			},
		},
		{
			name: "after set",
			args: &graphqlbackend.ListWorkspacesArgs{
				After: strPtr("10"),
			},
			want: store.ListBatchSpecWorkspacesOpts{
				Cursor: 10,
			},
		},
		{
			name: "search set",
			args: &graphqlbackend.ListWorkspacesArgs{
				Search: strPtr("sourcegraph"),
			},
			want: store.ListBatchSpecWorkspacesOpts{
				TextSearch: []search.TextSearchTerm{{Term: "sourcegraph"}},
			},
		},
		{
			name: "state completed",
			args: &graphqlbackend.ListWorkspacesArgs{
				State: strPtr("COMPLETED"),
			},
			want: store.ListBatchSpecWorkspacesOpts{
				OnlyCachedOrCompleted: true,
			},
		},
		{
			name: "state pending",
			args: &graphqlbackend.ListWorkspacesArgs{
				State: strPtr("PENDING"),
			},
			want: store.ListBatchSpecWorkspacesOpts{
				OnlyWithoutExecution: true,
			},
		},
		{
			name: "state queued",
			args: &graphqlbackend.ListWorkspacesArgs{
				State: strPtr("QUEUED"),
			},
			want: store.ListBatchSpecWorkspacesOpts{
				State: types.BatchSpecWorkspaceExecutionJobStateQueued,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			have, err := workspacesListArgsToDBOpts(tc.args)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(have, tc.want); diff != "" {
				t.Fatal("invalid args returned" + diff)
			}
		})
	}
}
