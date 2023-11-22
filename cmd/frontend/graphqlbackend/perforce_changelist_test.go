package graphqlbackend

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestToPerforceChangelistResolver(t *testing.T) {
	repo := &types.Repo{
		ID:           2,
		Name:         "perforce.sgdev.org/foo/bar",
		ExternalRepo: api.ExternalRepoSpec{ServiceType: extsvc.TypePerforce},
	}

	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefaultReturn(repo, nil)

	db := dbmocks.NewMockDB()
	db.ReposFunc.SetDefaultReturn(repos)

	repoResolver := NewRepositoryResolver(db, nil, repo)

	testCases := []struct {
		name              string
		inputCommit       *gitdomain.Commit
		inputChangelistID string
		expectedResolver  *PerforceChangelistResolver
		expectedErr       error
	}{
		{
			name: "p4-fusion",
			inputCommit: &gitdomain.Commit{
				ID: exampleCommitSHA1,
				Message: `test change
[p4-fusion: depot-paths = "//test-perms/": change = 80972]`,
			},
			inputChangelistID: "80972",
			expectedResolver: &PerforceChangelistResolver{
				cid:          "80972",
				canonicalURL: "/perforce.sgdev.org/foo/bar/-/changelist/80972",
			},
		},
		{
			name: "git-p4",
			inputCommit: &gitdomain.Commit{
				ID: exampleCommitSHA1,
				Message: `test change
[git-p4: depot-paths = "//test-perms/": change = 80999]`,
			},
			inputChangelistID: "80999",
			expectedResolver: &PerforceChangelistResolver{
				cid:          "80999",
				canonicalURL: "/perforce.sgdev.org/foo/bar/-/changelist/80999",
			},
		},
		{
			name: "error",
			inputCommit: &gitdomain.Commit{
				ID: exampleCommitSHA1,
				Message: `test change
foo bar`,
			},
			expectedResolver: nil,
			expectedErr: errors.Wrap(
				errors.New(`failed to retrieve changelist ID from commit body: "foo bar"`), "failed to generate perforceChangelistID",
			),
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gcr := NewGitCommitResolver(db, nil, repoResolver, tc.inputCommit.ID, tc.inputCommit)
			gotResolver, gotErr := toPerforceChangelistResolver(ctx, gcr)

			if !errors.Is(gotErr, tc.expectedErr) {
				t.Fatalf("mismatched errors, \nwant: %v\n got: %v", tc.expectedErr, gotErr)
				return
			}

			// Checks after this point are for non-nil expectedResolver test cases.
			if tc.expectedResolver == nil {
				return
			}

			// Note: We cannot compare the struct directly because we have unexported fields. It is
			// simpler to compare the two fields instead of implmenting a custom comparer to use
			// with cmp.Diff.
			//
			// If the resolver evolves to have more fields, then it might make more sense to
			// implement the custom comparison func in the future.
			if gotResolver.cid != tc.expectedResolver.cid {
				t.Errorf("mismatched cid, \nwant: %v\n got: %v", tc.expectedResolver.cid, gotResolver.cid)
			}

			if gotResolver.canonicalURL != tc.expectedResolver.canonicalURL {
				t.Errorf("mismatched canonicalURL, \nwant: %v\n got: %v", tc.expectedResolver.canonicalURL, gotResolver.canonicalURL)
			}

			// Now test the exported methods of the resolver too.
			if value := gotResolver.CID(); value != tc.expectedResolver.cid {
				t.Errorf("mismatched value from method CID(), \nwant: %v\n got: %v", tc.expectedResolver.cid, value)
			}

			if value := gotResolver.CanonicalURL(); value != tc.expectedResolver.canonicalURL {
				t.Errorf("mismatched value from method CanonicalURL(), \nwant: %v\n got: %v", tc.expectedResolver.canonicalURL, value)
			}
		})
	}
}

func TestParseP4FusionCommitSubject(t *testing.T) {
	testCases := []struct {
		input           string
		expectedSubject string
		expectedErr     string
	}{
		{
			input:           "83732 - adding sourcegraph repos",
			expectedSubject: "adding sourcegraph repos",
		},
		{
			input:           "abc1234 - updating config",
			expectedSubject: "",
			expectedErr:     `failed to parse commit subject "abc1234 - updating config" for commit converted by p4-fusion`,
		},
		{
			input:           "- fixing bug",
			expectedSubject: "",
			expectedErr:     `failed to parse commit subject "- fixing bug" for commit converted by p4-fusion`,
		},
		{
			input:           "fixing bug",
			expectedSubject: "",
			expectedErr:     `failed to parse commit subject "fixing bug" for commit converted by p4-fusion`,
		},
	}

	for _, tc := range testCases {
		subject, err := parseP4FusionCommitSubject(tc.input)
		if err != nil && err.Error() != tc.expectedErr {
			t.Errorf("Expected error %q, got %q", err.Error(), tc.expectedErr)
		}

		if subject != tc.expectedSubject {
			t.Errorf("Expected subject %q, got %q", tc.expectedSubject, subject)
		}
	}
}
