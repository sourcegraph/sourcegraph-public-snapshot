package repos

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestRevisionValidation(t *testing.T) {

	// mocks a repo repoFoo with revisions revBar and revBas
	git.Mocks.ResolveRevision = func(spec string, opt git.ResolveRevisionOptions) (api.CommitID, error) {
		// trigger errors
		if spec == "bad_commit" {
			return "", git.BadCommitError{}
		}
		if spec == "deadline_exceeded" {
			return "", context.DeadlineExceeded
		}

		// known revisions
		m := map[string]struct{}{
			"revBar": {},
			"revBas": {},
		}
		if _, ok := m[spec]; ok {
			return "", nil
		}
		return "", &gitserver.RevisionNotFoundError{Repo: "repoFoo", Spec: spec}
	}
	defer func() { git.Mocks.ResolveRevision = nil }()

	db.Mocks.Repos.ListRepoNames = func(ctx context.Context, opts db.ReposListOptions) ([]*types.RepoName, error) {
		return []*types.RepoName{{Name: "repoFoo"}}, nil
	}
	defer func() { db.Mocks.Repos.List = nil }()

	tests := []struct {
		repoFilters              []string
		wantRepoRevs             []*search.RepositoryRevisions
		wantMissingRepoRevisions []*search.RepositoryRevisions
		wantErr                  error
	}{
		{
			repoFilters: []string{"repoFoo@revBar:^revBas"},
			wantRepoRevs: []*search.RepositoryRevisions{{
				Repo: &types.RepoName{Name: "repoFoo"},
				Revs: []search.RevisionSpecifier{
					{
						RevSpec:        "revBar",
						RefGlob:        "",
						ExcludeRefGlob: "",
					},
					{
						RevSpec:        "^revBas",
						RefGlob:        "",
						ExcludeRefGlob: "",
					},
				},
			}},
			wantMissingRepoRevisions: nil,
		},
		{
			repoFilters: []string{"repoFoo@*revBar:*!revBas"},
			wantRepoRevs: []*search.RepositoryRevisions{{
				Repo: &types.RepoName{Name: "repoFoo"},
				Revs: []search.RevisionSpecifier{
					{
						RevSpec:        "",
						RefGlob:        "revBar",
						ExcludeRefGlob: "",
					},
					{
						RevSpec:        "",
						RefGlob:        "",
						ExcludeRefGlob: "revBas",
					},
				},
			}},
			wantMissingRepoRevisions: nil,
		},
		{
			repoFilters: []string{"repoFoo@revBar:^revQux"},
			wantRepoRevs: []*search.RepositoryRevisions{{
				Repo: &types.RepoName{Name: "repoFoo"},
				Revs: []search.RevisionSpecifier{
					{
						RevSpec:        "revBar",
						RefGlob:        "",
						ExcludeRefGlob: "",
					},
				},
				ListRefs: nil,
			}},
			wantMissingRepoRevisions: []*search.RepositoryRevisions{{
				Repo: &types.RepoName{Name: "repoFoo"},
				Revs: []search.RevisionSpecifier{
					{
						RevSpec:        "^revQux",
						RefGlob:        "",
						ExcludeRefGlob: "",
					},
				},
			}},
		},
		{
			repoFilters:              []string{"repoFoo@revBar:bad_commit"},
			wantRepoRevs:             nil,
			wantMissingRepoRevisions: nil,
			wantErr:                  git.BadCommitError{},
		},
		{
			repoFilters:              []string{"repoFoo@revBar:^bad_commit"},
			wantRepoRevs:             nil,
			wantMissingRepoRevisions: nil,
			wantErr:                  git.BadCommitError{},
		},
		{
			repoFilters:              []string{"repoFoo@revBar:deadline_exceeded"},
			wantRepoRevs:             nil,
			wantMissingRepoRevisions: nil,
			wantErr:                  context.DeadlineExceeded,
		},
		{
			repoFilters: []string{"repoFoo"},
			wantRepoRevs: []*search.RepositoryRevisions{{
				Repo: &types.RepoName{Name: "repoFoo"},
				Revs: []search.RevisionSpecifier{
					{
						RevSpec:        "",
						RefGlob:        "",
						ExcludeRefGlob: "",
					},
				},
			}},
			wantMissingRepoRevisions: nil,
			wantErr:                  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.repoFilters[0], func(t *testing.T) {

			op := Options{RepoFilters: tt.repoFilters}
			resolved, err := ResolveRepositories(context.Background(), op)

			if diff := cmp.Diff(tt.wantRepoRevs, resolved.RepoRevs); diff != "" {
				t.Error(diff)
			}
			if diff := cmp.Diff(tt.wantMissingRepoRevisions, resolved.MissingRepoRevs); diff != "" {
				t.Error(diff)
			}
			if tt.wantErr != err {
				t.Errorf("got: %v, expected: %v", err, tt.wantErr)
			}
		})
	}
}
