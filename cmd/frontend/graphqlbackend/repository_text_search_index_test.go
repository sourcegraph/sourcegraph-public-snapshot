package graphqlbackend

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type repoListerMock struct{}

func (r repoListerMock) List(ctx context.Context, q zoektquery.Q, opts *zoekt.ListOptions) (*zoekt.RepoList, error) {
	zoektRepo := []*zoekt.RepoListEntry{{
		Repository: zoekt.Repository{
			Name: string("alice/repo"),
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "deadbeef"},
				{Name: "main", Version: "deadbeef"},
				{Name: "1.0", Version: "deadbeef"},
			},
		},
		IndexMetadata: zoekt.IndexMetadata{
			IndexTime: time.Now(),
		},
	}}
	return &zoekt.RepoList{Repos: zoektRepo}, nil
}

func TestRetrievingAndDeduplicatingIndexedRefs(t *testing.T) {
	db := database.NewDB(nil)
	defaultBranchRef := "refs/heads/main"
	gitserver.Mocks.ResolveRevision = func(rev string, opt gitserver.ResolveRevisionOptions) (api.CommitID, error) {
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
		repo:   NewRepositoryResolver(database.NewDB(db), &types.Repo{Name: "alice/repo"}),
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
