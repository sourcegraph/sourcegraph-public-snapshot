package authzchecked

import (
	"testing"

	"golang.org/x/net/context"

	"strings"

	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/mockstore"
)

func TestBuilds_Get(t *testing.T) {
	ctx, rc := mockRepoCheckerContext()

	var calledGet bool
	s := Builds(&mockstore.Builds{
		Get_: func(ctx context.Context, build sourcegraph.BuildSpec) (*sourcegraph.Build, error) {
			calledGet = true
			return &sourcegraph.Build{Repo: ""}, nil
		},
	})

	if _, err := s.Get(ctx, sourcegraph.BuildSpec{}); err != nil {
		t.Fatal(err)
	}
	if !calledGet {
		t.Error("!calledGet")
	}

	if !rc.calledCheckRepo {
		t.Error("!calledCheckRepo")
	}
}

func TestBuilds_List_all_siteAdmin(t *testing.T) {
	ctx := auth.WithActor(nil, auth.Actor{SiteAdmin_UNIMPLEMENTED: true})

	var calledList bool
	s := Builds(&mockstore.Builds{
		List_: func(ctx context.Context, opt *sourcegraph.BuildListOptions) ([]*sourcegraph.Build, error) {
			calledList = true
			return nil, nil
		},
	})

	if _, err := s.List(ctx, &sourcegraph.BuildListOptions{}); err != nil {
		t.Fatal(err)
	}
	if !calledList {
		t.Error("!calledList")
	}
}

func TestBuilds_List_all_notSiteAdmin(t *testing.T) {
	ctx := context.Background()

	var calledList bool
	s := Builds(&mockstore.Builds{
		List_: func(ctx context.Context, opt *sourcegraph.BuildListOptions) ([]*sourcegraph.Build, error) {
			calledList = true
			return nil, nil
		},
	})

	if _, err := s.List(ctx, &sourcegraph.BuildListOptions{}); err != ErrSiteAdminOnly {
		t.Errorf("got err == %v, want %v", err, ErrSiteAdminOnly)
	}
	if calledList {
		t.Error("calledList")
	}
}

func TestBuilds_List_Repo(t *testing.T) {
	ctx, rc := mockRepoCheckerContext()

	var calledList bool
	s := Builds(&mockstore.Builds{
		List_: func(ctx context.Context, opt *sourcegraph.BuildListOptions) ([]*sourcegraph.Build, error) {
			calledList = true
			return nil, nil
		},
	})

	if _, err := s.List(ctx, &sourcegraph.BuildListOptions{Repo: "r"}); err != nil {
		t.Fatal(err)
	}
	if !calledList {
		t.Error("!calledList")
	}
	if !rc.calledCheckRepo {
		t.Error("!calledCheckRepo")
	}
}

func TestBuilds_GetFirstInCommitOrder(t *testing.T) {
	ctx, rc := mockRepoCheckerContext()

	var calledGetFirstInCommitOrder bool
	s := Builds(&mockstore.Builds{
		GetFirstInCommitOrder_: func(ctx context.Context, repo string, commitIDs []string, successfulOnly bool) (*sourcegraph.Build, int, error) {
			calledGetFirstInCommitOrder = true
			return nil, 0, nil
		},
	})

	if _, _, err := s.GetFirstInCommitOrder(ctx, "r", nil, false); err != nil {
		t.Fatal(err)
	}
	if !calledGetFirstInCommitOrder {
		t.Error("!calledGetFirstInCommitOrder")
	}
	if !rc.calledCheckRepo {
		t.Error("!calledCheckRepo")
	}
}

func TestBuilds_Create(t *testing.T) {
	ctx, rc := mockRepoCheckerContext()

	var calledCreate bool
	s := Builds(&mockstore.Builds{
		Create_: func(ctx context.Context, build *sourcegraph.Build) (*sourcegraph.Build, error) {
			calledCreate = true
			return nil, nil
		},
	})

	if _, err := s.Create(ctx, &sourcegraph.Build{}); err != nil {
		t.Fatal(err)
	}
	if !calledCreate {
		t.Error("!calledCreate")
	}
	if !rc.calledCheckRepo {
		t.Error("!calledCheckRepo")
	}
}

func TestBuilds_Update(t *testing.T) {
	ctx, rc := mockRepoCheckerContext()

	var calledUpdate bool
	s := Builds(&mockstore.Builds{
		Update_: func(ctx context.Context, build sourcegraph.BuildSpec, info sourcegraph.BuildUpdate) error {
			calledUpdate = true
			return nil
		},
	})

	if err := s.Update(ctx, sourcegraph.BuildSpec{}, sourcegraph.BuildUpdate{}); err != nil {
		t.Fatal(err)
	}
	if !calledUpdate {
		t.Error("!calledUpdate")
	}
	if !rc.calledCheckRepo {
		t.Error("!calledCheckRepo")
	}
}

func TestBuilds_ListBuildTasks(t *testing.T) {
	ctx, rc := mockRepoCheckerContext()

	var calledListBuildTasks bool
	s := Builds(&mockstore.Builds{
		ListBuildTasks_: func(ctx context.Context, build sourcegraph.BuildSpec, opt *sourcegraph.BuildTaskListOptions) ([]*sourcegraph.BuildTask, error) {
			calledListBuildTasks = true
			return nil, nil
		},
	})

	if _, err := s.ListBuildTasks(ctx, sourcegraph.BuildSpec{}, &sourcegraph.BuildTaskListOptions{}); err != nil {
		t.Fatal(err)
	}
	if !calledListBuildTasks {
		t.Error("!calledListBuildTasks")
	}
	if !rc.calledCheckRepo {
		t.Error("!calledCheckRepo")
	}
}

func TestBuilds_CreateTasks(t *testing.T) {
	ctx, rc := mockRepoCheckerContext()

	var calledCreateTasks bool
	s := Builds(&mockstore.Builds{
		CreateTasks_: func(ctx context.Context, tasks []*sourcegraph.BuildTask) ([]*sourcegraph.BuildTask, error) {
			calledCreateTasks = true
			return nil, nil
		},
	})

	if _, err := s.CreateTasks(ctx, []*sourcegraph.BuildTask{{Repo: "r0", Attempt: 0, CommitID: strings.Repeat("A", 40)}, {Repo: "r1", Attempt: 1, CommitID: strings.Repeat("A", 40)}}); err != nil {
		t.Fatal(err)
	}
	if !calledCreateTasks {
		t.Error("!calledCreateTasks")
	}
	if !rc.calledCheckRepo {
		t.Error("!calledCheckRepo")
	}
	if err := rc.calledWithRepoArgs("r0", "r1"); err != nil {
		t.Error(err)
	}
}

func TestBuilds_UpdateTask(t *testing.T) {
	ctx, rc := mockRepoCheckerContext()

	var calledUpdateTask bool
	s := Builds(&mockstore.Builds{
		UpdateTask_: func(ctx context.Context, task sourcegraph.TaskSpec, info sourcegraph.TaskUpdate) error {
			calledUpdateTask = true
			return nil
		},
	})

	if err := s.UpdateTask(ctx, sourcegraph.TaskSpec{}, sourcegraph.TaskUpdate{}); err != nil {
		t.Fatal(err)
	}
	if !calledUpdateTask {
		t.Error("!calledUpdateTask")
	}
	if !rc.calledCheckRepo {
		t.Error("!calledCheckRepo")
	}
}

func TestBuilds_DequeueNext_siteAdmin(t *testing.T) {
	ctx := auth.WithActor(nil, auth.Actor{SiteAdmin_UNIMPLEMENTED: true})

	var calledDequeueNext bool
	s := Builds(&mockstore.Builds{
		DequeueNext_: func(ctx context.Context) (*sourcegraph.Build, error) {
			calledDequeueNext = true
			return nil, nil
		},
	})

	if _, err := s.DequeueNext(ctx); err != nil {
		t.Fatal(err)
	}
	if !calledDequeueNext {
		t.Error("!calledDequeueNext")
	}
}

func TestBuilds_DequeueNext_notSiteAdmin(t *testing.T) {
	ctx := context.Background()

	var calledDequeueNext bool
	s := Builds(&mockstore.Builds{
		DequeueNext_: func(ctx context.Context) (*sourcegraph.Build, error) {
			calledDequeueNext = true
			return nil, nil
		},
	})

	if _, err := s.DequeueNext(ctx); err != ErrSiteAdminOnly {
		t.Errorf("got err == %v, want %v", err, ErrSiteAdminOnly)
	}
	if calledDequeueNext {
		t.Error("calledDequeueNext")
	}
}

func TestBuilds_GetTask(t *testing.T) {
	ctx, rc := mockRepoCheckerContext()

	var calledGetTask bool
	s := Builds(&mockstore.Builds{
		GetTask_: func(ctx context.Context, task sourcegraph.TaskSpec) (*sourcegraph.BuildTask, error) {
			calledGetTask = true
			return &sourcegraph.BuildTask{Attempt: task.Attempt, CommitID: strings.Repeat("A", 40), Repo: "r/r"}, nil
		},
	})

	if _, err := s.GetTask(ctx, sourcegraph.TaskSpec{}); err != nil {
		t.Fatal(err)
	}
	if !calledGetTask {
		t.Error("!calledGetTask")
	}
	if !rc.calledCheckRepo {
		t.Error("!calledCheckRepo")
	}
}
