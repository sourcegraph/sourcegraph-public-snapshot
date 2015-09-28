//+build off

// these are old tests from package app that will be useful for the GetRepo* funcs

package handlerutil

import (
	"testing"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app"
)

// Tests that when the URL contains no revision and a build exists, the newest build's commit ID is used.
func TestGetRepoFromSpec_NoRevAndBuildExists(t *testing.T) {
	t.Skip("SKIP_BUILD_BRANCH (see TODOs.x.txt)")

	setup()
	defer teardown()

	wantCommitID := "c"
	var fetchedRepo, fetchedBuild bool
	mockAPISourcegraph.Repos.(*sourcegraph.MockReposService).Get_ = func(repo sourcegraph.RepoSpec, opt *sourcegraph.RepoGetOptions) (*sourcegraph.Repo, sourcegraph.Response, error) {
		if opt.ResolveRevision {
			t.Error("want !opt.ResolveRevision (because we get the revision from the build)")
		}
		if repo.CommitID != wantCommitID {
			t.Errorf("got repoSpec.CommitID == %q, want %q", repo.CommitID, wantCommitID)
		}
		fetchedRepo = true
		return &sourcegraph.Repo{CommitID: wantCommitID}, nil, nil
	}
	mockAPISourcegraph.Builds.(*sourcegraph.MockBuildsService).ListByRepo_ = func(repo sourcegraph.RepoSpec, opt *sourcegraph.BuildListByRepoOptions) ([]*sourcegraph.Build, sourcegraph.Response, error) {
		if fetchedRepo {
			t.Error("repo should be fetched after the build is fetched")
		}
		if repo.CommitID != "" {
			t.Errorf("got repoSpec.CommitID == %q, want empty", repo.CommitID)
		}
		fetchedBuild = true
		return []*sourcegraph.Build{{CommitID: wantCommitID}}, nil, nil
	}

	repo, build, repoSpec, err := app.GetRepoFromSpec(nil, &sourcegraph.RepoSpec{URI: "r"})
	if err != nil {
		t.Error(err)
	}

	if build.CommitID != wantCommitID {
		t.Errorf("got build.CommitID == %q, want %q", build.CommitID, wantCommitID)
	}
	if repo.CommitID != wantCommitID {
		t.Errorf("got repo.CommitID == %q, want %q", repo.CommitID, wantCommitID)
	}
	if repoSpec.CommitID != wantCommitID {
		t.Errorf("got repoSpec.CommitID == %q, want %q", repoSpec.CommitID, wantCommitID)
	}

	if !fetchedRepo {
		t.Error("!fetchedRepo")
	}
	if !fetchedBuild {
		t.Error("!fetchedBuild")
	}
}

// Tests that when the URL contains no revision and no build exists, the repo's
// default resolved commit (which is defined in the service to be the default
// branch's tip commit) is used.
func TestGetRepoFromSpec_NoRevAndNoBuild(t *testing.T) {
	t.Skip("SKIP_BUILD_BRANCH (see TODOs.x.txt)")
	setup()
	defer teardown()

	wantCommitID := "c"
	var fetchedRepo, fetchedBuild bool
	mockAPISourcegraph.Repos.(*sourcegraph.MockReposService).Get_ = func(repo sourcegraph.RepoSpec, opt *sourcegraph.RepoGetOptions) (*sourcegraph.Repo, sourcegraph.Response, error) {
		if !fetchedBuild {
			t.Error("build should be attempted to be fetched before the repo is fetched (but in this test case, there is no build)")
		}
		if !opt.ResolveRevision {
			t.Error("want opt.ResolveRevision==true (because we get the revision from the repository)")
		}
		if repo.CommitID != "" {
			t.Errorf("got repoSpec.CommitID == %q, want empty", repo.CommitID)
		}
		fetchedRepo = true
		return &sourcegraph.Repo{CommitID: wantCommitID}, nil, nil
	}
	mockAPISourcegraph.Builds.(*sourcegraph.MockBuildsService).ListByRepo_ = func(repo sourcegraph.RepoSpec, opt *sourcegraph.BuildListByRepoOptions) ([]*sourcegraph.Build, sourcegraph.Response, error) {
		if fetchedRepo {
			t.Error("repo should be fetched after the build is fetched (in this test, there is no build)")
		}
		if fetchedBuild {
			t.Error("build fetched twice")
		}
		if repo.CommitID != "" {
			t.Errorf("got repoSpec.CommitID == %q, want empty", repo.CommitID)
		}
		fetchedBuild = true
		return []*sourcegraph.Build{}, nil, nil
	}

	repo, build, repoSpec, err := app.GetRepoFromSpec(nil, &sourcegraph.RepoSpec{URI: "r"})
	if err != nil {
		t.Error(err)
	}

	if build != nil {
		t.Errorf("got build == %v, want nil", build)
	}
	if repo.CommitID != wantCommitID {
		t.Errorf("got repo.CommitID == %q, want %q", repo.CommitID, wantCommitID)
	}
	if repoSpec.CommitID != wantCommitID {
		t.Errorf("got repoSpec.CommitID == %q, want %q", repoSpec.CommitID, wantCommitID)
	}

	if !fetchedRepo {
		t.Error("!fetchedRepo")
	}
	if !fetchedBuild {
		t.Error("!fetchedBuild")
	}
}

// Tests that when the URL contains a specific revision and no build exists for
// that revision, the specific revision is resolved and used.
func TestGetRepoFromSpec_RevSpecifiedAndNoBuild(t *testing.T) {
	setup()
	defer teardown()

	wantCommitID := "c"
	var fetchedRepo, fetchedBuild bool
	mockAPISourcegraph.Repos.(*sourcegraph.MockReposService).Get_ = func(repo sourcegraph.RepoSpec, opt *sourcegraph.RepoGetOptions) (*sourcegraph.Repo, sourcegraph.Response, error) {
		if !opt.ResolveRevision {
			t.Error("want opt.ResolveRevision==true (because we get the revision from the repository)")
		}
		if repo.CommitID != wantCommitID {
			t.Errorf("got repoSpec.CommitID == %q, want %q", repo.CommitID, wantCommitID)
		}
		fetchedRepo = true
		return &sourcegraph.Repo{CommitID: wantCommitID}, nil, nil
	}
	mockAPISourcegraph.Builds.(*sourcegraph.MockBuildsService).ListByRepo_ = func(repo sourcegraph.RepoSpec, opt *sourcegraph.BuildListByRepoOptions) ([]*sourcegraph.Build, sourcegraph.Response, error) {
		if !fetchedRepo {
			t.Error("repo should be fetched before the build is attempted to be fetched (because the URL has a specific revision)")
		}
		if fetchedBuild {
			t.Error("build fetched twice")
		}
		if repo.CommitID != wantCommitID {
			t.Errorf("got repoSpec.CommitID == %q, want %q", repo.CommitID, wantCommitID)
		}
		fetchedBuild = true
		return []*sourcegraph.Build{}, nil, nil
	}

	repo, build, repoSpec, err := app.GetRepoFromSpec(nil, &sourcegraph.RepoSpec{URI: "r", CommitID: wantCommitID})
	if err != nil {
		t.Error(err)
	}

	if build != nil {
		t.Errorf("got build == %v, want nil", build)
	}
	if repo.CommitID != wantCommitID {
		t.Errorf("got repo.CommitID == %q, want %q", repo.CommitID, wantCommitID)
	}
	if repoSpec.CommitID != wantCommitID {
		t.Errorf("got repoSpec.CommitID == %q, want %q", repoSpec.CommitID, wantCommitID)
	}

	if !fetchedRepo {
		t.Error("!fetchedRepo")
	}
	if !fetchedBuild {
		t.Error("!fetchedBuild")
	}
}

// Tests that when the URL contains a specific revision and a build exists for
// that revision, the specific revision is resolved and used, and the build is returned.
func TestGetRepoFromSpec_RevSpecifiedAndBuildExists(t *testing.T) {
	setup()
	defer teardown()

	wantCommitID := "c"
	var fetchedRepo, fetchedBuild bool
	mockAPISourcegraph.Repos.(*sourcegraph.MockReposService).Get_ = func(repo sourcegraph.RepoSpec, opt *sourcegraph.RepoGetOptions) (*sourcegraph.Repo, sourcegraph.Response, error) {
		if !opt.ResolveRevision {
			t.Error("want opt.ResolveRevision==true (because we get the revision from the repository)")
		}
		if repo.CommitID != wantCommitID {
			t.Errorf("got repoSpec.CommitID == %q, want %q", repo.CommitID, wantCommitID)
		}
		fetchedRepo = true
		return &sourcegraph.Repo{CommitID: wantCommitID}, nil, nil
	}
	mockAPISourcegraph.Builds.(*sourcegraph.MockBuildsService).ListByRepo_ = func(repo sourcegraph.RepoSpec, opt *sourcegraph.BuildListByRepoOptions) ([]*sourcegraph.Build, sourcegraph.Response, error) {
		if !fetchedRepo {
			t.Error("repo should be fetched before the build is attempted to be fetched (because the URL has a specific revision)")
		}
		if fetchedBuild {
			t.Error("build fetched twice")
		}
		if repo.CommitID != wantCommitID {
			t.Errorf("got repoSpec.CommitID == %q, want %q", repo.CommitID, wantCommitID)
		}
		fetchedBuild = true
		return []*sourcegraph.Build{{CommitID: wantCommitID}}, nil, nil
	}

	repo, build, repoSpec, err := app.GetRepoFromSpec(nil, &sourcegraph.RepoSpec{URI: "r", CommitID: wantCommitID})
	if err != nil {
		t.Error(err)
	}

	if build == nil || build.CommitID != wantCommitID {
		t.Errorf("got build == %v, want build.CommitID == %q", build, wantCommitID)
	}
	if repo.CommitID != wantCommitID {
		t.Errorf("got repo.CommitID == %q, want %q", repo.CommitID, wantCommitID)
	}
	if repoSpec.CommitID != wantCommitID {
		t.Errorf("got repoSpec.CommitID == %q, want %q", repoSpec.CommitID, wantCommitID)
	}

	if !fetchedRepo {
		t.Error("!fetchedRepo")
	}
	if !fetchedBuild {
		t.Error("!fetchedBuild")
	}
}
