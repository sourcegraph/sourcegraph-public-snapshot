package sources

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/inconshreveable/log15" //nolint:logging // TODO move all logging to sourcegraph/log
	"github.com/stretchr/testify/assert"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestBitbucketServerSource_LoadChangeset(t *testing.T) {
	ratelimit.SetupForTest(t)

	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		// The test fixtures and golden files were generated with
		// this config pointed to bitbucket.sgdev.org
		instanceURL = "https://bitbucket.sgdev.org"
	}

	repo := &types.Repo{
		Metadata: &bitbucketserver.Repo{
			Slug:    "vegeta",
			Project: &bitbucketserver.Project{Key: "SOUR"},
		},
	}

	changesets := []*Changeset{
		{RemoteRepo: repo, TargetRepo: repo, Changeset: &btypes.Changeset{ExternalID: "2"}},
		{RemoteRepo: repo, TargetRepo: repo, Changeset: &btypes.Changeset{ExternalID: "999"}},
	}

	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "found",
			cs:   changesets[0],
		},
		{
			name: "not-found",
			cs:   changesets[1],
			err:  `Changeset with external ID 999 not found`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "BitbucketServerSource_LoadChangeset_" + tc.name

		t.Run(tc.name, func(t *testing.T) {
			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.BitbucketServerConnection{
					Url:   instanceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Background()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplaceAll(tc.err, "${INSTANCEURL}", instanceURL)

			err = bbsSrc.LoadChangeset(ctx, tc.cs)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			testutil.AssertGolden(
				t,
				"testdata/golden/"+tc.name,
				update(tc.name),
				tc.cs.Changeset.Metadata.(*bitbucketserver.PullRequest),
			)
		})
	}
}

func TestBitbucketServerSource_CreateChangeset(t *testing.T) {
	ratelimit.SetupForTest(t)

	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		// The test fixtures and golden files were generated with
		// this config pointed to bitbucket.sgdev.org
		instanceURL = "https://bitbucket.sgdev.org"
	}

	repo := &types.Repo{
		Metadata: &bitbucketserver.Repo{
			ID:      10070,
			Slug:    "automation-testing",
			Project: &bitbucketserver.Project{Key: "SOUR"},
		},
	}

	testCases := []struct {
		name   string
		cs     *Changeset
		err    string
		exists bool
	}{
		{
			name: "abbreviated_refs",
			cs: &Changeset{
				Title:      "This is a test PR",
				Body:       "This is the body of a test PR",
				BaseRef:    "master",
				HeadRef:    "test-pr-bbs-11",
				RemoteRepo: repo,
				TargetRepo: repo,
				Changeset:  &btypes.Changeset{},
			},
		},
		{
			name: "success",
			cs: &Changeset{
				Title:      "This is a test PR",
				Body:       "This is the body of a test PR",
				BaseRef:    "refs/heads/master",
				HeadRef:    "refs/heads/test-pr-bbs-12",
				RemoteRepo: repo,
				TargetRepo: repo,
				Changeset:  &btypes.Changeset{},
			},
		},
		{
			name: "already_exists",
			cs: &Changeset{
				Title:      "This is a test PR",
				Body:       "This is the body of a test PR",
				BaseRef:    "refs/heads/master",
				HeadRef:    "refs/heads/always-open-pr-bbs",
				RemoteRepo: repo,
				TargetRepo: repo,
				Changeset:  &btypes.Changeset{},
			},
			// CreateChangeset is idempotent so if the PR already exists
			// it is not an error
			err:    "",
			exists: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "BitbucketServerSource_CreateChangeset_" + tc.name

		t.Run(tc.name, func(t *testing.T) {
			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.BitbucketServerConnection{
					Url:   instanceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Background()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplaceAll(tc.err, "${INSTANCEURL}", instanceURL)

			exists, err := bbsSrc.CreateChangeset(ctx, tc.cs)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			if have, want := exists, tc.exists; have != want {
				t.Errorf("exists:\nhave: %t\nwant: %t", have, want)
			}

			pr := tc.cs.Changeset.Metadata.(*bitbucketserver.PullRequest)
			testutil.AssertGolden(t, "testdata/golden/"+tc.name, update(tc.name), pr)
		})
	}
}

func TestBitbucketServerSource_CloseChangeset(t *testing.T) {
	ratelimit.SetupForTest(t)

	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		// The test fixtures and golden files were generated with
		// this config pointed to bitbucket.sgdev.org
		instanceURL = "https://bitbucket.sgdev.org"
	}

	pr := &bitbucketserver.PullRequest{ID: 59, Version: 4}
	pr.ToRef.Repository.Slug = "automation-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"

	// Version is too low
	outdatedPR := &bitbucketserver.PullRequest{ID: 156, Version: 1}
	outdatedPR.ToRef.Repository.Slug = "automation-testing"
	outdatedPR.ToRef.Repository.Project.Key = "SOUR"

	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "success",
			cs:   &Changeset{Changeset: &btypes.Changeset{Metadata: pr}},
		},
		{
			name: "outdated",
			cs:   &Changeset{Changeset: &btypes.Changeset{Metadata: outdatedPR}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "BitbucketServerSource_CloseChangeset_" + strings.ReplaceAll(tc.name, " ", "_")

		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Updating fixtures: %t", update(tc.name))

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.BitbucketServerConnection{
					Url:   instanceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Background()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplaceAll(tc.err, "${INSTANCEURL}", instanceURL)

			err = bbsSrc.CloseChangeset(ctx, tc.cs)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Changeset.Metadata.(*bitbucketserver.PullRequest)
			testutil.AssertGolden(t, "testdata/golden/"+tc.name, update(tc.name), pr)
		})
	}
}

func TestBitbucketServerSource_CloseChangeset_DeleteSourceBranch(t *testing.T) {
	ratelimit.SetupForTest(t)

	// Repository used: https://bitbucket.sgdev.org/projects/SOUR/repos/automation-testing
	//
	// This test can be updated with `-update BitbucketServerSource_CloseChangeset_DeleteSourceBranch`,
	// provided this PR is open: https://bitbucket.sgdev.org/projects/SOUR/repos/automation-testing/pull-requests/168/overview

	pr := &bitbucketserver.PullRequest{ID: 168, Version: 1}
	pr.ToRef.Repository.Slug = "automation-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"
	pr.FromRef.ID = "refs/heads/delete-me"

	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "success",
			cs:   &Changeset{Changeset: &btypes.Changeset{Metadata: pr}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "BitbucketServerSource_CloseChangeset_DeleteSourceBranch_" + strings.ReplaceAll(tc.name, " ", "_")

		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Updating fixtures: %t", update(tc.name))

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					BatchChangesAutoDeleteBranch: true,
				},
			})
			defer conf.Mock(nil)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.BitbucketServerConnection{
					Url:   "https://bitbucket.sgdev.org",
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Background()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplaceAll(tc.err, "${INSTANCEURL}", "https://bitbucket.sgdev.org")

			err = bbsSrc.CloseChangeset(ctx, tc.cs)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Changeset.Metadata.(*bitbucketserver.PullRequest)
			testutil.AssertGolden(t, "testdata/golden/"+tc.name, update(tc.name), pr)
		})
	}
}

func TestBitbucketServerSource_ReopenChangeset(t *testing.T) {
	ratelimit.SetupForTest(t)

	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		// The test fixtures and golden files were generated with
		// this config pointed to bitbucket.sgdev.org
		instanceURL = "https://bitbucket.sgdev.org"
	}

	pr := &bitbucketserver.PullRequest{ID: 95, Version: 1}
	pr.ToRef.Repository.Slug = "automation-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"

	// Version is far too low
	outdatedPR := &bitbucketserver.PullRequest{ID: 160, Version: 1}
	outdatedPR.ToRef.Repository.Slug = "automation-testing"
	outdatedPR.ToRef.Repository.Project.Key = "SOUR"

	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "success",
			cs:   &Changeset{Changeset: &btypes.Changeset{Metadata: pr}},
		},
		{
			name: "outdated",
			cs:   &Changeset{Changeset: &btypes.Changeset{Metadata: outdatedPR}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "BitbucketServerSource_ReopenChangeset_" + strings.ReplaceAll(tc.name, " ", "_")

		t.Run(tc.name, func(t *testing.T) {
			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.BitbucketServerConnection{
					Url:   instanceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Background()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplaceAll(tc.err, "${INSTANCEURL}", instanceURL)

			err = bbsSrc.ReopenChangeset(ctx, tc.cs)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Changeset.Metadata.(*bitbucketserver.PullRequest)
			testutil.AssertGolden(t, "testdata/golden/"+tc.name, update(tc.name), pr)
		})
	}
}

func TestBitbucketServerSource_UpdateChangeset(t *testing.T) {
	ratelimit.SetupForTest(t)

	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		// The test fixtures and golden files were generated with
		// this config pointed to bitbucket.sgdev.org
		instanceURL = "https://bitbucket.sgdev.org"
	}

	reviewers := []bitbucketserver.Reviewer{
		{
			Role:               "REVIEWER",
			LastReviewedCommit: "7549846524f8aed2bd1c0249993ae1bf9d3c9998",
			Approved:           false,
			Status:             "UNAPPROVED",
			User: &bitbucketserver.User{
				Name: "batch-change-buddy",
				Slug: "batch-change-buddy",
				ID:   403,
			},
		},
	}

	successPR := &bitbucketserver.PullRequest{ID: 154, Version: 22, Reviewers: reviewers}
	successPR.ToRef.Repository.Slug = "automation-testing"
	successPR.ToRef.Repository.Project.Key = "SOUR"

	// This version is too low
	outdatedPR := &bitbucketserver.PullRequest{ID: 155, Version: 13, Reviewers: reviewers}
	outdatedPR.ToRef.Repository.Slug = "automation-testing"
	outdatedPR.ToRef.Repository.Project.Key = "SOUR"

	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "success",
			cs: &Changeset{
				Title:     "This is a new title",
				Body:      "This is a new body",
				BaseRef:   "refs/heads/master",
				Changeset: &btypes.Changeset{Metadata: successPR},
			},
		},
		{
			name: "outdated",
			cs: &Changeset{
				Title:     "This is a new title",
				Body:      "This is a new body",
				BaseRef:   "refs/heads/master",
				Changeset: &btypes.Changeset{Metadata: outdatedPR},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "BitbucketServerSource_UpdateChangeset_" + strings.ReplaceAll(tc.name, " ", "_")

		t.Run(tc.name, func(t *testing.T) {
			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.BitbucketServerConnection{
					Url:   instanceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Background()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplaceAll(tc.err, "${INSTANCEURL}", instanceURL)

			err = bbsSrc.UpdateChangeset(ctx, tc.cs)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Changeset.Metadata.(*bitbucketserver.PullRequest)
			testutil.AssertGolden(t, "testdata/golden/"+tc.name, update(tc.name), pr)
		})
	}
}

func TestBitbucketServerSource_CreateComment(t *testing.T) {
	ratelimit.SetupForTest(t)

	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		// The test fixtures and golden files were generated with
		// this config pointed to bitbucket.sgdev.org
		instanceURL = "https://bitbucket.sgdev.org"
	}

	pr := &bitbucketserver.PullRequest{ID: 59, Version: 4}
	pr.ToRef.Repository.Slug = "automation-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"

	// This version is too low
	outdatedPR := &bitbucketserver.PullRequest{ID: 154, Version: 1}
	outdatedPR.ToRef.Repository.Slug = "automation-testing"
	outdatedPR.ToRef.Repository.Project.Key = "SOUR"

	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "success",
			cs:   &Changeset{Changeset: &btypes.Changeset{Metadata: pr}},
		},
		{
			name: "outdated",
			cs:   &Changeset{Changeset: &btypes.Changeset{Metadata: outdatedPR}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "BitbucketServerSource_CreateComment_" + strings.ReplaceAll(tc.name, " ", "_")

		t.Run(tc.name, func(t *testing.T) {
			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.BitbucketServerConnection{
					Url:   instanceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Background()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplaceAll(tc.err, "${INSTANCEURL}", instanceURL)

			err = bbsSrc.CreateComment(ctx, tc.cs, "test-comment")
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}
		})
	}
}

func TestBitbucketServerSource_MergeChangeset(t *testing.T) {
	ratelimit.SetupForTest(t)

	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		// The test fixtures and golden files were generated with
		// this config pointed to bitbucket.sgdev.org
		instanceURL = "https://bitbucket.sgdev.org"
	}

	pr := &bitbucketserver.PullRequest{ID: 159, Version: 0}
	pr.ToRef.Repository.Slug = "automation-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"

	// Version is too low
	outdatedPR := &bitbucketserver.PullRequest{ID: 157, Version: 1}
	outdatedPR.ToRef.Repository.Slug = "automation-testing"
	outdatedPR.ToRef.Repository.Project.Key = "SOUR"

	// Version is also too low, but PR has a conflict too, we want err
	conflictPR := &bitbucketserver.PullRequest{ID: 154, Version: 8}
	conflictPR.ToRef.Repository.Slug = "automation-testing"
	conflictPR.ToRef.Repository.Project.Key = "SOUR"

	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "success",
			cs:   &Changeset{Changeset: &btypes.Changeset{Metadata: pr}},
		},
		{
			name: "outdated",
			cs:   &Changeset{Changeset: &btypes.Changeset{Metadata: outdatedPR}},
		},
		{
			name: "conflict",
			cs:   &Changeset{Changeset: &btypes.Changeset{Metadata: conflictPR}},
			err:  "changeset cannot be merged:\nBitbucket API HTTP error: code=409 url=\"${INSTANCEURL}/rest/api/1.0/projects/SOUR/repos/automation-testing/pull-requests/154/merge?version=10\" body=\"{\\\"errors\\\":[{\\\"context\\\":null,\\\"message\\\":\\\"The pull request has conflicts and cannot be merged.\\\",\\\"exceptionName\\\":\\\"com.atlassian.bitbucket.pull.PullRequestMergeVetoedException\\\",\\\"conflicted\\\":true,\\\"vetoes\\\":[]}]}\"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "BitbucketServerSource_MergeChangeset_" + strings.ReplaceAll(tc.name, " ", "_")

		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Updating fixtures: %t", update(tc.name))

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.BitbucketServerConnection{
					Url:   instanceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Background()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplaceAll(tc.err, "${INSTANCEURL}", instanceURL)

			err = bbsSrc.MergeChangeset(ctx, tc.cs, false)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Changeset.Metadata.(*bitbucketserver.PullRequest)
			testutil.AssertGolden(t, "testdata/golden/"+tc.name, update(tc.name), pr)
		})
	}
}

func TestBitbucketServerSource_WithAuthenticator(t *testing.T) {
	ratelimit.SetupForTest(t)

	svc := &types.ExternalService{
		Kind: extsvc.KindBitbucketServer,
		Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.BitbucketServerConnection{
			Url:   "https://bitbucket.sgdev.org",
			Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
		})),
	}

	ctx := context.Background()
	bbsSrc, err := NewBitbucketServerSource(ctx, svc, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("supported", func(t *testing.T) {
		for name, tc := range map[string]auth.Authenticator{
			"BasicAuth":           &auth.BasicAuth{},
			"OAuthBearerToken":    &auth.OAuthBearerToken{},
			"SudoableOAuthClient": &bitbucketserver.SudoableOAuthClient{},
		} {
			t.Run(name, func(t *testing.T) {
				src, err := bbsSrc.WithAuthenticator(tc)
				if err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}

				if gs, ok := src.(*BitbucketServerSource); !ok {
					t.Error("cannot coerce Source into bbsSource")
				} else if gs == nil {
					t.Error("unexpected nil Source")
				} else if gs.au != tc {
					t.Errorf("incorrect authenticator: have=%v want=%v", gs.au, tc)
				}
			})
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		for name, tc := range map[string]auth.Authenticator{
			"nil":         nil,
			"OAuthClient": &auth.OAuthClient{},
		} {
			t.Run(name, func(t *testing.T) {
				src, err := bbsSrc.WithAuthenticator(tc)
				if err == nil {
					t.Error("unexpected nil error")
				} else if !errors.HasType[UnsupportedAuthenticatorError](err) {
					t.Errorf("unexpected error of type %T: %v", err, err)
				}
				if src != nil {
					t.Errorf("expected non-nil Source: %v", src)
				}
			})
		}
	})
}

func TestBitbucketServerSource_GetFork(t *testing.T) {
	ratelimit.SetupForTest(t)

	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		// The test fixtures and golden files were generated with
		// this config pointed to bitbucket.sgdev.org
		instanceURL = "https://bitbucket.sgdev.org"
	}

	newBitbucketServerRepo := func(urn, key, slug string, id int) *types.Repo {
		return &types.Repo{
			Metadata: &bitbucketserver.Repo{
				ID:      id,
				Slug:    slug,
				Project: &bitbucketserver.Project{Key: key},
			},
			Sources: map[string]*types.SourceInfo{
				urn: {
					ID:       urn,
					CloneURL: "https://bitbucket.sgdev.org/" + key + "/" + slug,
				},
			},
		}
	}

	newExternalService := func(t *testing.T, token *string) *types.ExternalService {
		var actualToken string
		if token == nil {
			actualToken = os.Getenv("BITBUCKET_SERVER_TOKEN")
		} else {
			actualToken = *token
		}

		return &types.ExternalService{
			Kind: extsvc.KindBitbucketServer,
			Config: extsvc.NewUnencryptedConfig(marshalJSON(t, &schema.BitbucketServerConnection{
				Url:   instanceURL,
				Token: actualToken,
			})),
		}
	}

	testName := func(t *testing.T) string {
		return strings.ReplaceAll(t.Name(), "/", "_")
	}

	lg := log15.New()
	lg.SetHandler(log15.DiscardHandler())
	urn := extsvc.URN(extsvc.KindBitbucketCloud, 1)

	t.Run("bad username", func(t *testing.T) {
		cf, save := newClientFactory(t, testName(t))
		defer save(t)

		svc := newExternalService(t, pointers.Ptr("invalid"))

		ctx := context.Background()
		bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
		assert.Nil(t, err)

		fork, err := bbsSrc.GetFork(ctx, newBitbucketServerRepo(urn, "SOUR", "read-only", 10103), nil, nil)
		assert.Nil(t, fork)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "getting username")
	})

	// This test validates the behavior when `GetFork` is called but the response from the
	// API indicates the destination we would like to fork the repo into is already used
	// for a different repo. `GetFork` should return `nil` and an error.
	t.Run("not a fork", func(t *testing.T) {
		// This test expects that:
		// - The repo BAT/vcr-fork-test-repo exists and is not a fork.
		// - The repo ~MILTON/vcr-fork-test-repo exists and is not a fork.
		// Use credentials in 1Password for "milton" to access or update this test.
		name := testName(t)
		cf, save := newClientFactory(t, name)
		defer save(t)

		svc := newExternalService(t, nil)
		target := newBitbucketServerRepo(urn, "BAT", "vcr-fork-test-repo", 0)

		ctx := context.Background()
		bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
		assert.Nil(t, err)

		fork, err := bbsSrc.GetFork(ctx, target, pointers.Ptr("~milton"), pointers.Ptr("vcr-fork-test-repo"))
		assert.Nil(t, fork)
		assert.ErrorContains(t, err, "repo is not a fork")
	})

	// This test validates the behavior when `GetFork` is called but the response from the
	// API indicates the destination we would like to fork the repo into is a fork of a
	// different repo. `GetFork` should return `nil` and an error.
	t.Run("not forked from parent", func(t *testing.T) {
		// This test expects that:
		// - The repo BAT/vcr-fork-test-repo-already-forked exists and is not a fork.
		// - The repo ~MILTON/BAT-vcr-fork-test-repo-already-forked exists and is a fork of it.
		// Use credentials in 1Password for "milton" to access or update this test.
		name := testName(t)
		cf, save := newClientFactory(t, name)
		defer save(t)

		svc := newExternalService(t, nil)
		// We'll give the target repo the incorrect ID, which will result in the
		// origin ID check in checkAndCopy() failing.
		target := newBitbucketServerRepo(urn, "BAT", "vcr-fork-test-repo-already-forked", 0)

		ctx := context.Background()
		bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
		assert.Nil(t, err)

		fork, err := bbsSrc.GetFork(ctx, target, pointers.Ptr("~milton"), pointers.Ptr("BAT-vcr-fork-test-repo-already-forked"))
		assert.Nil(t, fork)
		assert.ErrorContains(t, err, "repo was not forked from the given parent")
	})

	// This test validates the behavior when `GetFork` is called without a namespace or
	// name set, but a fork of the repo already exists in the user's namespace with the
	// default fork name. `GetFork` should return the existing fork.
	t.Run("success with new changeset and existing fork", func(t *testing.T) {
		// This test expects that:
		// - The repo BAT/vcr-fork-test-repo-already-forked exists and is not a fork.
		// - The repo ~MILTON/BAT-vcr-fork-test-repo-already-forked exists and is a fork of it.
		// - The current user is ~MILTON and the default fork naming convention would produce
		//   the fork name "BAT-vcr-fork-test-repo-already-forked".
		// Use credentials in 1Password for "milton" to access or update this test.
		name := testName(t)
		cf, save := newClientFactory(t, name)
		defer save(t)

		svc := newExternalService(t, nil)
		// Code host ID for this repo can be found in the VCR cassette or by inspecting
		// the response body at GET
		// https://bitbucket.sgdev.org/rest/api/1.0/projects/BAT/repos/vcr-fork-test-repo-already-forked
		target := newBitbucketServerRepo(urn, "BAT", "vcr-fork-test-repo-already-forked", 24378)

		ctx := context.Background()
		bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
		assert.Nil(t, err)

		username, err := bbsSrc.client.AuthenticatedUsername(ctx)
		assert.Nil(t, err)

		fork, err := bbsSrc.GetFork(ctx, target, nil, nil)
		assert.Nil(t, err)
		assert.NotNil(t, fork)
		assert.NotEqual(t, fork, target)
		assert.Equal(t, "~"+strings.ToUpper(username), fork.Metadata.(*bitbucketserver.Repo).Project.Key)
		assert.Equal(t, fork.Sources[urn].CloneURL, "https://bitbucket.sgdev.org/~"+username+"/bat-vcr-fork-test-repo-already-forked")

		testutil.AssertGolden(t, "testdata/golden/"+name, update(name), fork)
	})

	// This test validates the behavior when `GetFork` is called without a namespace or
	// name set and no fork of the repo exists in the user's namespace with the default
	// fork name. `GetFork` should return the newly-created fork.
	//
	// NOTE: It is not possible to update this test and "success with existing changeset
	// and new fork" at the same time.
	t.Run("success with new changeset and new fork", func(t *testing.T) {
		t.Skip()
		// This test expects that:
		// - The repo BAT/vcr-fork-test-repo-not-forked exists and is not a fork.
		// - The repo ~MILTON/BAT-vcr-fork-test-repo-not-forked does not exist.
		// - The current user is ~MILTON and the default fork naming convention would produce
		//   the fork name "BAT-vcr-fork-test-repo-not-forked".
		// Use credentials in 1Password for "milton" to access or update this test.
		name := testName(t)
		cf, save := newClientFactory(t, name)
		defer save(t)

		svc := newExternalService(t, nil)
		// Code host ID for this repo can be found in the VCR cassette or by inspecting
		// the response body at GET
		// https://bitbucket.sgdev.org/rest/api/1.0/projects/BAT/repos/vcr-fork-test-repo-not-forked
		target := newBitbucketServerRepo(urn, "BAT", "vcr-fork-test-repo-not-forked", 216974)

		ctx := context.Background()
		bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
		assert.Nil(t, err)

		username, err := bbsSrc.client.AuthenticatedUsername(ctx)
		assert.Nil(t, err)

		fork, err := bbsSrc.GetFork(ctx, target, nil, nil)
		assert.Nil(t, err)
		assert.NotNil(t, fork)
		assert.NotEqual(t, fork, target)
		assert.Equal(t, "~"+strings.ToUpper(username), fork.Metadata.(*bitbucketserver.Repo).Project.Key)
		assert.Equal(t, fork.Sources[urn].CloneURL, "https://bitbucket.sgdev.org/~"+username+"/bat-vcr-fork-test-repo-not-forked")

		testutil.AssertGolden(t, "testdata/golden/"+name, update(name), fork)
	})

	// This test validates the behavior when `GetFork` is called with a namespace and name
	// both already set, and a fork of the repo already exists at that destination.
	// `GetFork` should return the existing fork.
	t.Run("success with existing changeset and existing fork", func(t *testing.T) {
		// This test expects that:
		// - The repo BAT/vcr-fork-test-repo-already-forked exists and is not a fork.
		// - The repo ~MILTON/BAT-vcr-fork-test-repo-already-forked exists and is a fork of it.
		// Use credentials in 1Password for "milton" to access or update this test.
		name := testName(t)
		cf, save := newClientFactory(t, name)
		defer save(t)

		svc := newExternalService(t, nil)
		// Code host ID for this repo can be found in the VCR cassette or by inspecting
		// the response body at GET
		// https://bitbucket.sgdev.org/rest/api/1.0/projects/BAT/repos/vcr-fork-test-repo-already-forked
		target := newBitbucketServerRepo(urn, "BAT", "vcr-fork-test-repo-already-forked", 24378)

		ctx := context.Background()
		bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
		assert.Nil(t, err)

		username, err := bbsSrc.client.AuthenticatedUsername(ctx)
		assert.Nil(t, err)

		fork, err := bbsSrc.GetFork(ctx, target, pointers.Ptr("~milton"), pointers.Ptr("BAT-vcr-fork-test-repo-already-forked"))
		assert.Nil(t, err)
		assert.NotNil(t, fork)
		assert.NotEqual(t, fork, target)
		assert.Equal(t, "~"+strings.ToUpper(username), fork.Metadata.(*bitbucketserver.Repo).Project.Key)
		assert.Equal(t, fork.Sources[urn].CloneURL, "https://bitbucket.sgdev.org/~"+username+"/bat-vcr-fork-test-repo-already-forked")

		testutil.AssertGolden(t, "testdata/golden/"+name, update(name), fork)
	})

	// This test validates the behavior when `GetFork` is called with a namespace and name
	// both already set, but no fork of the repo already exists at that destination. This
	// situation is only possible if the changeset and fork repo have been deleted on the
	// code host since the changeset was created. `GetFork` should return the
	// newly-created fork.
	//
	// NOTE: It is not possible to update this test and "success with new changeset and
	// new fork" at the same time.
	t.Run("success with existing changeset and new fork", func(t *testing.T) {
		t.Skip()
		// This test expects that:
		// - The repo BAT/vcr-fork-test-repo-not-forked exists and is not a fork.
		// - The repo ~MILTON/BAT-vcr-fork-test-repo-not-forked does not exist.
		// Use credentials in 1Password for "milton" to access or update this test.
		name := testName(t)
		cf, save := newClientFactory(t, name)
		defer save(t)

		svc := newExternalService(t, nil)
		// Code host ID for this repo can be found in the VCR cassette or by inspecting
		// the response body at GET
		// https://bitbucket.sgdev.org/rest/api/1.0/projects/BAT/repos/vcr-fork-test-repo-not-forked
		target := newBitbucketServerRepo(urn, "BAT", "vcr-fork-test-repo-not-forked", 216974)

		ctx := context.Background()
		bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
		assert.Nil(t, err)

		username, err := bbsSrc.client.AuthenticatedUsername(ctx)
		assert.Nil(t, err)

		fork, err := bbsSrc.GetFork(ctx, target, pointers.Ptr("~milton"), pointers.Ptr("BAT-vcr-fork-test-repo-not-forked"))
		assert.Nil(t, err)
		assert.NotNil(t, fork)
		assert.NotEqual(t, fork, target)
		assert.Equal(t, "~"+strings.ToUpper(username), fork.Metadata.(*bitbucketserver.Repo).Project.Key)
		assert.Equal(t, fork.Sources[urn].CloneURL, "https://bitbucket.sgdev.org/~"+username+"/bat-vcr-fork-test-repo-not-forked")

		testutil.AssertGolden(t, "testdata/golden/"+name, update(name), fork)
	})
}
