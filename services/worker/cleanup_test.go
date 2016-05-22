package worker

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/mock"

	"golang.org/x/net/context"
)

func TestBuildCleanup(t *testing.T) {
	var (
		updateOp         *sourcegraph.BuildsUpdateOp
		listBuildTasksOp *sourcegraph.BuildsListBuildTasksOp
	)
	ctx := sourcegraph.WithMockClient(context.Background(), &sourcegraph.Client{
		Builds: &mock.BuildsClient{
			Update_: func(_ context.Context, op *sourcegraph.BuildsUpdateOp) (*sourcegraph.Build, error) {
				updateOp = op
				return nil, nil
			},
			ListBuildTasks_: func(_ context.Context, op *sourcegraph.BuildsListBuildTasksOp) (*sourcegraph.BuildTaskList, error) {
				listBuildTasksOp = op
				return &sourcegraph.BuildTaskList{BuildTasks: []*sourcegraph.BuildTask{}}, nil
			},
		},
	})

	build := &sourcegraph.BuildJob{
		Spec: sourcegraph.BuildSpec{
			Repo: sourcegraph.RepoSpec{
				URI: "test",
			},
			ID: 1,
		},
	}
	activeBuilds := newActiveBuilds()
	activeBuilds.Add(build)
	buildCleanup(ctx, activeBuilds)
	activeBuilds.Remove(build)

	expectedUpdate := &sourcegraph.BuildsUpdateOp{
		Build: build.Spec,
		Info: sourcegraph.BuildUpdate{
			EndedAt: updateOp.Info.EndedAt,
			Killed:  true,
		},
	}
	if !reflect.DeepEqual(updateOp, expectedUpdate) {
		t.Errorf("Unexpected updateOp %#v", updateOp)
	}
	if listBuildTasksOp == nil || !reflect.DeepEqual(listBuildTasksOp.Build, build.Spec) {
		t.Errorf("Unexpected updateOp %#v", listBuildTasksOp)
	}
}
