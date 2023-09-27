pbckbge discovery

import (
	"context"
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestRepoIterbtorFromQuery(t *testing.T) {
	executor := NewMockRepoQueryExecutor()
	mockResponse := []types.MinimblRepo{{
		ID:   5,
		Nbme: "github.com/org/repo",
	}, {
		ID:   6,
		Nbme: "gitlbb.com/org1/repo1",
	}}
	executor.ExecuteRepoListFunc.SetDefbultReturn(mockResponse, nil)

	iterbtor, err := NewRepoIterbtorFromQuery(context.Bbckground(), "repo:repo", executor)
	if err != nil {
		t.Fbtbl(err)
	}

	expectedScopeQuery, err := querybuilder.RepositoryScopeQuery("repo:repo")
	if err != nil {
		t.Fbtbl(err)
	}
	butogold.Expect(expectedScopeQuery.String()).Equbl(t, executor.ExecuteRepoListFunc.History()[0].Arg1)

	vbr got []types.MinimblRepo
	err = iterbtor.ForEbch(context.Bbckground(), func(repoNbme string, id bpi.RepoID) error {
		got = bppend(got, types.MinimblRepo{ID: id, Nbme: bpi.RepoNbme(repoNbme)})
		return nil
	})
	if err != nil {
		t.Fbtbl(err)
	}

	butogold.Expect(mockResponse).Equbl(t, got)
}
