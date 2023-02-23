package discovery

import (
	"context"
	"reflect"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestScopedRepoIteratorForEach(t *testing.T) {
	repoStore := NewMockRepoStore()
	mockResponse := []*types.Repo{{
		ID:   5,
		Name: "github.com/org/repo",
	}, {
		ID:   6,
		Name: "gitlab.com/org1/repo1",
	}}
	repoStore.ListFunc.SetDefaultReturn(mockResponse, nil)

	var names []string
	for _, repo := range mockResponse {
		names = append(names, string(repo.Name))
	}

	iterator, err := NewScopedRepoIterator(context.Background(), names, repoStore)
	if err != nil {
		t.Fatal(err)
	}

	// verify the names argument actually matches what is expected and we arent just trusting a mock blindly
	if !reflect.DeepEqual(repoStore.ListFunc.History()[0].Arg1.Names, names) {
		t.Error("argument mismatch on repo names")
	}

	var gotNames []string
	var gotIds []api.RepoID
	err = iterator.ForEach(context.Background(), func(repoName string, id api.RepoID) error {
		gotNames = append(gotNames, repoName)
		gotIds = append(gotIds, id)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("expect_equal_repo_names", func(t *testing.T) {
		autogold.Expect([]string{"github.com/org/repo", "gitlab.com/org1/repo1"}).Equal(t, gotNames)
	})
	t.Run("expect_equal_repo_ids", func(t *testing.T) {
		autogold.Expect([]api.RepoID{api.RepoID(5), api.RepoID(6)}).Equal(t, gotIds)
	})
}
