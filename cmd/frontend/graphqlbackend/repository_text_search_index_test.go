package graphqlbackend

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/zoekt"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestRetrievingAndDeduplicatingIndexedRefs(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, nil)
	defaultBranchRef := "refs/heads/main"
	gitserver.Mocks.ResolveRevision = func(rev string, opt gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		if rev != defaultBranchRef && strings.HasSuffix(rev, defaultBranchRef) {
			return "", errors.New("x")
		}
		return api.CommitID("deadbeef"), nil
	}
	gitserver.Mocks.GetDefaultBranch = func(repo api.RepoName) (refName string, commit api.CommitID, err error) {
		// Mock default branch lookup in (*RepsitoryResolver).DefaultBranch.
		return defaultBranchRef, "", nil
	}
	defer gitserver.ResetMocks()

	repoIndexResolver := &repositoryTextSearchIndexResolver{
		repo: NewRepositoryResolver(db, &types.Repo{Name: "alice/repo"}),
		client: &backend.FakeSearcher{Repos: []*zoekt.RepoListEntry{{
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
		}}},
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
