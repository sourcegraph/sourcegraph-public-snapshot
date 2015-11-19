package app_test

import (
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/util/httptestutil"
)

// mockRepoGet is like the go-sourcegraph mock.Repos.MockGet helper
// func, but it returns a *sourcegraph.Repo with the fields that the
// repo page requires (e.g., DefaultBranch).
func mockRepoGet(c *httptestutil.MockClients, wantRepo string) (called *bool) {
	called = new(bool)
	c.Repos.Get_ = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		*called = true
		if repo.URI != wantRepo {
			return nil, grpc.Errorf(codes.NotFound, "repo %s not found", wantRepo)
		}
		return &sourcegraph.Repo{
			URI:           repo.URI,
			DefaultBranch: "mybranch",
		}, nil
	}
	return called
}

func mockNoRepoBuild(c *httptestutil.MockClients) (called *bool) {
	called = new(bool)
	c.Builds.GetRepoBuildInfo_ = func(context.Context, *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error) {
		*called = true
		return nil, grpc.Errorf(codes.NotFound, "")
	}
	c.Builds.Create_ = func(context.Context, *sourcegraph.BuildsCreateOp) (*sourcegraph.Build, error) {
		return &sourcegraph.Build{}, nil
	}
	return called
}

func mockCurrentRepoBuild(c *httptestutil.MockClients) (called *bool) {
	called = new(bool)
	c.Builds.GetRepoBuildInfo_ = func(context.Context, *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error) {
		*called = true
		return &sourcegraph.RepoBuildInfo{
			Exact:                &sourcegraph.Build{},
			LastSuccessful:       &sourcegraph.Build{},
			LastSuccessfulCommit: &vcs.Commit{},
		}, nil
	}
	return called
}
func mockSpecificRepoBuild(c *httptestutil.MockClients, commitID string) (called *bool) {
	called = new(bool)
	c.Builds.GetRepoBuildInfo_ = func(context.Context, *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error) {
		*called = true
		return &sourcegraph.RepoBuildInfo{
			Exact:                &sourcegraph.Build{CommitID: commitID},
			LastSuccessful:       &sourcegraph.Build{CommitID: commitID},
			LastSuccessfulCommit: &vcs.Commit{ID: vcs.CommitID(commitID)},
		}, nil
	}
	return called
}

func mockNoRepoReadme(c *httptestutil.MockClients) (called *bool) {
	called = new(bool)
	c.Repos.GetReadme_ = func(context.Context, *sourcegraph.RepoRevSpec) (*sourcegraph.Readme, error) {
		*called = true
		return nil, grpc.Errorf(codes.NotFound, "")
	}
	return called
}

func mockEmptyTreeEntry(c *httptestutil.MockClients) (called *bool) {
	called = new(bool)
	c.RepoTree.Get_ = func(context.Context, *sourcegraph.RepoTreeGetOp) (*sourcegraph.TreeEntry, error) {
		*called = true
		return &sourcegraph.TreeEntry{TreeEntry: &vcsclient.TreeEntry{}}, nil
	}
	return called
}

func mockTreeEntryGet(c *httptestutil.MockClients, t *sourcegraph.TreeEntry) (called *bool) {
	called = new(bool)
	c.RepoTree.Get_ = func(context.Context, *sourcegraph.RepoTreeGetOp) (*sourcegraph.TreeEntry, error) {
		*called = true
		return t, nil
	}
	return called
}

func mockEmptyRepoConfig(c *httptestutil.MockClients) (called *bool) {
	called = new(bool)
	c.Repos.GetConfig_ = func(_ context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.RepoConfig, error) {
		*called = true
		return &sourcegraph.RepoConfig{}, nil
	}
	return called
}

func mockEnabledRepoConfig(c *httptestutil.MockClients) (called *bool) {
	called = new(bool)
	c.Repos.GetConfig_ = func(_ context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.RepoConfig, error) {
		*called = true
		return &sourcegraph.RepoConfig{Enabled: true}, nil
	}
	return called
}

// func mockRepoCommit(c *httptestutil.MockClients, commit *vcs.Commit) {
// 	c.Repos.GetCommit_ = func(_ context.Context,_ sourcegraph.RepoRevSpec, _ *sourcegraph.RepoGetCommitOptions) (*vcs.Commit, error) {
// 		return commit, nil, nil
// 	}
// }

func mockBasicRepoMainPage(c *httptestutil.MockClients) {
	mockEmptyTreeEntry(c)
	mockEnabledRepoConfig(c)
	mockNoRepoReadme(c)
	mockCurrentRepoBuild(c)
}

// func mockEmptyRepoList(c *httptestutil.MockClients) {
// 	c.Repos.List_ = func(_ context.Context,*sourcegraph.RepoListOptions) ([]*sourcegraph.Repo, error) {
// 		return nil, &fakeResponse{totalCount: 0}, nil
// 	}
// }

// func mockNoDefExamples(c *httptestutil.MockClients) {
// 	c.Defs.ListExamples_ = func(_ context.Context,sourcegraph.DefSpec, *sourcegraph.DefListExamplesOptions) ([]*sourcegraph.Example, error) {
// 		return nil, nil, nil
// 	}
// }

func commitID(c string) vcs.CommitID { return vcs.CommitID(strings.Repeat(c, 40)) }
