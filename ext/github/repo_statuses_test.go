package github

import (
	"reflect"
	"testing"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"github.com/sourcegraph/go-github/github"
)

func TestRepoStatuses_GetCombinedStatus(t *testing.T) {
	var calledGetCombinedStatus bool
	ctx := testContext(&minimalClient{
		repos: mockGitHubRepos{
			GetCombinedStatus_: func(owner, repo, commit string, opt *github.ListOptions) (*github.CombinedStatus, *github.Response, error) {
				calledGetCombinedStatus = true
				return &github.CombinedStatus{
					State: github.String("a"),
					SHA:   github.String("b"),
					Statuses: []github.RepoStatus{
						{
							State:     github.String("c"),
							Context:   github.String("d"),
							TargetURL: github.String("e"),
						},
					},
				}, nil, nil
			},
		},
	})

	want := &sourcegraph.CombinedStatus{
		State:    "a",
		CommitID: "b",
		Statuses: []*sourcegraph.RepoStatus{
			{
				State:     "c",
				Context:   "d",
				TargetURL: "e",
			},
		},
	}

	status, err := (&RepoStatuses{}).GetCombined(ctx, sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "github.com/o/r"},
		Rev:      "c",
		CommitID: "c",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(status, want) {
		t.Errorf("got %+v, want %+v", status, want)
	}
	if !calledGetCombinedStatus {
		t.Error("!calledGetCombinedStatus")
	}
}

func TestRepoStatuses_CreateStatus(t *testing.T) {
	var calledCreateStatus bool
	ctx := testContext(&minimalClient{
		repos: mockGitHubRepos{
			CreateStatus_: func(owner, repo, commit string, status *github.RepoStatus) (*github.RepoStatus, *github.Response, error) {
				calledCreateStatus = true
				return &github.RepoStatus{
					State:     github.String("a"),
					Context:   github.String("b"),
					TargetURL: github.String("c"),
				}, nil, nil
			},
		},
	})

	repoRev := sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: "github.com/o/r"},
		Rev:      "c",
		CommitID: "c",
	}
	err := (&RepoStatuses{}).Create(ctx, repoRev, &sourcegraph.RepoStatus{
		State:     "a",
		Context:   "b",
		TargetURL: "c",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !calledCreateStatus {
		t.Error("!calledCreateStatus")
	}
}
