package authzchecked

import (
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/mockstore"
)

func TestRepoConfigs_Get(t *testing.T) {
	ctx, rc := mockRepoCheckerContext()

	var calledGet bool
	s := RepoConfigs(&mockstore.RepoConfigs{
		Get_: func(ctx context.Context, repo string) (*sourcegraph.RepoConfig, error) {
			calledGet = true
			return nil, nil
		},
	})

	if _, err := s.Get(ctx, ""); err != nil {
		t.Fatal(err)
	}
	if !calledGet {
		t.Error("!calledGet")
	}
	if !rc.calledCheckRepo {
		t.Error("!calledCheckRepo")
	}
}

func TestRepoConfigs_Update(t *testing.T) {
	ctx, rc := mockRepoCheckerContext()

	var calledUpdate bool
	s := RepoConfigs(&mockstore.RepoConfigs{
		Update_: func(ctx context.Context, repo string, settings sourcegraph.RepoConfig) error {
			calledUpdate = true
			return nil
		},
	})

	if err := s.Update(ctx, "", sourcegraph.RepoConfig{}); err != nil {
		t.Fatal(err)
	}
	if !calledUpdate {
		t.Error("!calledUpdate")
	}
	if !rc.calledCheckRepo {
		t.Error("!calledCheckRepo")
	}
}
