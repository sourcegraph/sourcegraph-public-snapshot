package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/google/go-github/v41/github"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
)

var updateRecordings = flag.Bool("update", false, "update integration test")

func newTestGitHubClient(ctx context.Context, t *testing.T) (ghc *github.Client, stop func() error) {
	recording := filepath.Join("testdata", strings.ReplaceAll(t.Name(), " ", "-"))
	recorder, err := httptestutil.NewRecorder(recording, *updateRecordings, func(i *cassette.Interaction) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if *updateRecordings {
		httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
		))
		recorder.SetTransport(httpClient.Transport)
	}
	return github.NewClient(&http.Client{Transport: recorder}), recorder.Stop
}

func TestRepoBranchLocker(t *testing.T) {
	ctx := context.Background()

	const testBranch = "test-buildsherrif-branch"

	t.Run("lock", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()
		locker := newBranchLocker(ghc, "sourcegraph", "sourcegraph", testBranch)

		commits := []commitInfo{
			{Commit: "be7f0f51b73b1966254db4aac65b656daa36e2fb"}, // @davejrt
			{Commit: "fac6d4973acad43fcd2f7579a3b496cd92619172"}, // @bobheadxi
			{Commit: "06a8636c2e0bea69944d8419aafa03ff3992527a"}, // @bobheadxi
			{Commit: "93971fa0b036b3e258cbb9a3eb7098e4032eefc4"}, // @jhchabran
		}
		modified, err := locker.Lock(ctx, commits, "dev-experience")
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, modified)

		// Validate live state
		validateLiveState := func() {
			protects, _, err := ghc.Repositories.GetBranchProtection(ctx, "sourcegraph", "sourcegraph", testBranch)
			if err != nil {
				t.Fatal(err)
			}
			assert.NotNil(t, protects.Restrictions)
			assert.NotNil(t, protects.RequiredPullRequestReviews)
			assert.Zero(t, protects.RequiredPullRequestReviews.RequiredApprovingReviewCount)

			users := []string{}
			for _, u := range protects.Restrictions.Users {
				users = append(users, *u.Login)
			}
			sort.Strings(users)
			assert.Equal(t, []string{"bobheadxi", "davejrt", "jhchabran"}, users)

			teams := []string{}
			for _, t := range protects.Restrictions.Teams {
				teams = append(teams, *t.Slug)
			}
			assert.Equal(t, []string{"dev-experience"}, teams)
		}
		validateLiveState()

		// Repeated lock attempt shouldn't change anything
		modified, err = locker.Lock(ctx, []commitInfo{}, "")
		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, modified)

		// should have same state as before
		validateLiveState()
	})

	t.Run("unlock", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()
		locker := newBranchLocker(ghc, "sourcegraph", "sourcegraph", testBranch)

		modified, err := locker.Unlock(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, modified)

		// Validate live state
		protects, _, err := ghc.Repositories.GetBranchProtection(ctx, "sourcegraph", "sourcegraph", testBranch)
		if err != nil {
			t.Fatal(err)
		}
		assert.Nil(t, protects.Restrictions)

		// Repeat unlock
		modified, err = locker.Unlock(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, modified)
	})
}
