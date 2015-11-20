package authzchecked

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// Builds wraps base's methods with authorization checks.
func Builds(base store.Builds) store.Builds { return &builds{base} }

// builds adds authorization checks to an underlying Builds.
type builds struct {
	noauthz store.Builds
}

func (s *builds) Get(ctx context.Context, build sourcegraph.BuildSpec) (*sourcegraph.Build, error) {
	if err := checkBuild(ctx, build, auth.Read); err != nil {
		return nil, err
	}
	return s.noauthz.Get(ctx, build)
}

func (s *builds) List(ctx context.Context, opt *sourcegraph.BuildListOptions) ([]*sourcegraph.Build, error) {
	if opt != nil && opt.Repo != "" {
		if err := auth.CheckRepo(ctx, opt.Repo, auth.Read); err != nil {
			return nil, err
		}
	} else {
		if err := checkSiteAdmin(ctx); err != nil {
			return nil, err
		}
	}
	return s.noauthz.List(ctx, opt)
}

func (s *builds) GetFirstInCommitOrder(ctx context.Context, repo string, commitIDs []string, successfulOnly bool) (build *sourcegraph.Build, nth int, err error) {
	if err := auth.CheckRepo(ctx, repo, auth.Read); err != nil {
		return nil, 0, err
	}
	return s.noauthz.GetFirstInCommitOrder(ctx, repo, commitIDs, successfulOnly)
}

func (s *builds) Create(ctx context.Context, b *sourcegraph.Build) (*sourcegraph.Build, error) {
	if err := auth.CheckRepo(ctx, b.Repo, auth.Write); err != nil {
		return nil, err
	}
	return s.noauthz.Create(ctx, b)
}

func (s *builds) Update(ctx context.Context, build sourcegraph.BuildSpec, info sourcegraph.BuildUpdate) error {
	if err := checkBuild(ctx, build, auth.Write); err != nil {
		return err
	}
	return s.noauthz.Update(ctx, build, info)
}

func (s *builds) ListBuildTasks(ctx context.Context, build sourcegraph.BuildSpec, opt *sourcegraph.BuildTaskListOptions) ([]*sourcegraph.BuildTask, error) {
	if err := checkBuild(ctx, build, auth.Read); err != nil {
		return nil, err
	}
	return s.noauthz.ListBuildTasks(ctx, build, opt)
}

func (s *builds) CreateTasks(ctx context.Context, tasks []*sourcegraph.BuildTask) ([]*sourcegraph.BuildTask, error) {
	for _, task := range tasks {
		if err := checkBuild(ctx, task.Spec().BuildSpec, auth.Write); err != nil {
			return nil, err
		}
	}
	return s.noauthz.CreateTasks(ctx, tasks)
}

func (s *builds) UpdateTask(ctx context.Context, task sourcegraph.TaskSpec, info sourcegraph.TaskUpdate) error {
	if err := checkTask(ctx, task, auth.Write); err != nil {
		return err
	}
	return s.noauthz.UpdateTask(ctx, task, info)
}

func (s *builds) DequeueNext(ctx context.Context) (*sourcegraph.Build, error) {
	if err := checkSiteAdmin(ctx); err != nil {
		return nil, err
	}
	return s.noauthz.DequeueNext(ctx)
}

func (s *builds) GetTask(ctx context.Context, task sourcegraph.TaskSpec) (*sourcegraph.BuildTask, error) {
	if err := checkTask(ctx, task, auth.Read); err != nil {
		return nil, err
	}
	return s.noauthz.GetTask(ctx, task)
}

func checkBuild(ctx context.Context, build sourcegraph.BuildSpec, what auth.PermType) error {
	return auth.CheckRepo(ctx, build.Repo.URI, what)
}

func checkTask(ctx context.Context, task sourcegraph.TaskSpec, perm auth.PermType) error {
	return checkBuild(ctx, task.BuildSpec, perm)
}
