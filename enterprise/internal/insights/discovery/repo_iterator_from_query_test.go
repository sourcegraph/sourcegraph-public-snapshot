package discovery

import (
	"context"
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/querybuilder"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRepoIteratorFromQuery(t *testing.T) {
	executor := NewMockRepoQueryExecutor()
	mockResponse := []types.MinimalRepo{{
		ID:   5,
		Name: "github.com/org/repo",
	}, {
		ID:   6,
		Name: "gitlab.com/org1/repo1",
	}}
	executor.ExecuteRepoListFunc.SetDefaultReturn(mockResponse, nil)

	var names []string
	for _, repo := range mockResponse {
		names = append(names, string(repo.Name))
	}

	iterator, err := NewRepoIteratorFromQuery(context.Background(), "repo:repo", executor)
	if err != nil {
		t.Fatal(err)
	}

	expectedScopeQuery, err := querybuilder.RepositoryScopeQuery("repo:repo")
	if err != nil {
		t.Fatal(err)
	}
	autogold.Want("expect_equal_repo_scope_query", expectedScopeQuery.String()).Equal(t, executor.ExecuteRepoListFunc.History()[0].Arg1)

	var got []types.MinimalRepo
	err = iterator.ForEach(context.Background(), func(repoName string, id api.RepoID) error {
		got = append(got, types.MinimalRepo{ID: id, Name: api.RepoName(repoName)})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	autogold.Want("expect_equal_repo_names", mockResponse).Equal(t, got)
}
