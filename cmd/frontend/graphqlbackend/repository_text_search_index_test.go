package graphqlbackend

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type repoListerMock struct{}

func (r repoListerMock) List(ctx context.Context, q zoektquery.Q) (*zoekt.RepoList, error) {
	zoektRepo := []*zoekt.RepoListEntry{{
		Repository: zoekt.Repository{
			Name: string("alice/repo"),
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "deadbeef"},
				{Name: "main", Version: "deadbeef"},
				{Name: "1.0", Version: "deadbeef"},
			},
		},
	}}
	return &zoekt.RepoList{Repos: zoektRepo}, nil
}

func TestRetrievingAndDeduplicatingIndexedRefs(t *testing.T) {
	db := new(dbtesting.MockDB)
	defaultBranchRef := "refs/heads/main"
	git.Mocks.ResolveRevision = func(rev string, opt git.ResolveRevisionOptions) (api.CommitID, error) {
		if rev != defaultBranchRef && strings.HasSuffix(rev, defaultBranchRef) {
			return "", errors.New("x")
		}
		return api.CommitID("deadbeef"), nil
	}
	git.Mocks.ExecSafe = func(params []string) (stdout, stderr []byte, exitCode int, err error) {
		// Mock default branch lookup in (*RepsitoryResolver).DefaultBranch.
		return []byte(defaultBranchRef), nil, 0, nil
	}
	defer git.ResetMocks()

	repoIndexResolver := &repositoryTextSearchIndexResolver{
		repo:   NewRepositoryResolver(db, &types.Repo{Name: "alice/repo"}),
		client: &repoListerMock{},
	}
	refs, err := repoIndexResolver.Refs(context.Background())
	if err != nil {
		t.Fatal("Error retrieving refs:", err)
	}

	want := []string{"refs/heads/main", "refs/heads/1.0"}
	got := []string{}
	for _, ref := range refs {
		got = append(got, ref.ref.name)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
