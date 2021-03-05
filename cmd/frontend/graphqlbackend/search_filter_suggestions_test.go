package graphqlbackend

import (
	"context"
	"reflect"
	"sort"
	"testing"

	searchrepos "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSearchFilterSuggestions(t *testing.T) {
	db := new(dbtesting.MockDB)

	searchrepos.MockResolveRepoGroups = func() (map[string][]searchrepos.RepoGroupValue, error) {
		return map[string][]searchrepos.RepoGroupValue{
			"repogroup1": {},
			"repogroup2": {},
		}, nil
	}
	defer func() { searchrepos.MockResolveRepoGroups = nil }()

	database.Mocks.Repos.List = func(_ context.Context, _ database.ReposListOptions) ([]*types.Repo, error) {
		return []*types.Repo{
			{Name: "github.com/foo/repo"},
			{Name: "bar-repo"},
		}, nil
	}
	defer func() { database.Mocks.Repos.List = nil }()

	tests := []struct {
		want     *searchFilterSuggestions
		globbing bool
	}{
		{want: &searchFilterSuggestions{
			repogroups: []string{"repogroup1", "repogroup2"},
			repos:      []string{"^bar-repo$", `^github\.com/foo/repo$`}},
			globbing: false,
		},
		{want: &searchFilterSuggestions{
			repogroups: []string{"repogroup1", "repogroup2"},
			repos:      []string{"bar-repo", `github.com/foo/repo`}},
			globbing: true,
		},
	}

	mockDecodedViewerFinalSettings = &schema.Settings{}
	defer func() { mockDecodedViewerFinalSettings = nil }()

	for _, tt := range tests {
		mockDecodedViewerFinalSettings.SearchGlobbing = &tt.globbing

		r, err := (&schemaResolver{db: db}).SearchFilterSuggestions(context.Background())
		if err != nil {
			t.Fatal("SearchFilterSuggestions:", err)
		}

		sort.Strings(r.repogroups)
		sort.Strings(r.repos)
		if !reflect.DeepEqual(r, tt.want) {
			t.Errorf("got != want\ngot:  %v\nwant: %v", r, tt.want)
		}
	}
}
