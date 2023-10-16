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
	"github.com/google/go-github/v55/github"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
)

var updateRecordings = flag.Bool("update-integration", false, "refresh integration test recordings")

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

	validateDefaultProtections := func(t *testing.T, protects *github.Protection) {
		// Require a pull request before merging
		assert.NotNil(t, protects.RequiredPullRequestReviews)
		assert.Equal(t, 1, protects.RequiredPullRequestReviews.RequiredApprovingReviewCount)
		// Require status checks to pass before merging
		assert.NotNil(t, protects.RequiredStatusChecks)
		assert.Empty(t, protects.RequiredStatusChecks.Contexts)
		assert.False(t, protects.RequiredStatusChecks.Strict)
		// Require linear history
		assert.NotNil(t, protects.RequireLinearHistory)
		assert.True(t, protects.RequireLinearHistory.Enabled)
	}

	t.Run("lock", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()
		locker := NewBranchLocker(ghc, "sourcegraph", "sourcegraph", testBranch)

		commits := []CommitInfo{
			//		{Commit: "be7f0f51b73b1966254db4aac65b656daa36e2fb"}, // @davejrt
			{Commit: "fac6d4973acad43fcd2f7579a3b496cd92619172"}, // @bobheadxi
			{Commit: "06a8636c2e0bea69944d8419aafa03ff3992527a"}, // @bobheadxi
			{Commit: "93971fa0b036b3e258cbb9a3eb7098e4032eefc4"}, // @jhchabran
		}
		lock, err := locker.Lock(ctx, commits, "dev-experience")
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, lock, "has callback")

		err = lock()
		if err != nil {
			t.Fatal(err)
		}

		// Validate live state
		validateLiveState := func() {
			protects, _, err := ghc.Repositories.GetBranchProtection(ctx, "sourcegraph", "sourcegraph", testBranch)
			if err != nil {
				t.Fatal(err)
			}
			validateDefaultProtections(t, protects)

			assert.NotNil(t, protects.Restrictions, "want push access restricted and granted")
			users := []string{}
			for _, u := range protects.Restrictions.Users {
				users = append(users, *u.Login)
			}
			sort.Strings(users)
			assert.Equal(t, []string{"bobheadxi", "jhchabran"}, users)

			teams := []string{}
			for _, t := range protects.Restrictions.Teams {
				teams = append(teams, *t.Slug)
			}
			assert.Equal(t, []string{"dev-experience"}, teams)
		}
		validateLiveState()

		// Repeated lock attempt shouldn't change anything
		lock, err = locker.Lock(ctx, []CommitInfo{}, "")
		if err != nil {
			t.Fatal(err)
		}
		assert.Nil(t, lock, "should not have callback")

		// should have same state as before
		validateLiveState()
	})

	t.Run("unlock", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()
		locker := NewBranchLocker(ghc, "sourcegraph", "sourcegraph", testBranch)

		unlock, err := locker.Unlock(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, unlock, "has callback")

		err = unlock()
		if err != nil {
			t.Fatal(err)
		}

		// Validate live state
		protects, _, err := ghc.Repositories.GetBranchProtection(ctx, "sourcegraph", "sourcegraph", testBranch)
		if err != nil {
			t.Fatal(err)
		}
		validateDefaultProtections(t, protects)
		assert.Nil(t, protects.Restrictions)

		// Repeat unlock
		unlock, err = locker.Unlock(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assert.Nil(t, unlock, "should not have callback")
	})
}
