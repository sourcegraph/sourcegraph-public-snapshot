package graphqlbackend

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSearchFilterSuggestions(t *testing.T) {
	db := new(dbtesting.MockDB)

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
			repos: []string{"^bar-repo$", `^github\.com/foo/repo$`}},
			globbing: false,
		},
		{want: &searchFilterSuggestions{
			repos: []string{"bar-repo", `github.com/foo/repo`}},
			globbing: true,
		},
	}

	mockDecodedViewerFinalSettings = &schema.Settings{}
	defer func() { mockDecodedViewerFinalSettings = nil }()

	for _, tt := range tests {
		mockDecodedViewerFinalSettings.SearchGlobbing = &tt.globbing

		r, err := (&schemaResolver{db: database.NewDB(db)}).SearchFilterSuggestions(context.Background())
		if err != nil {
			t.Fatal("SearchFilterSuggestions:", err)
		}

		sort.Strings(r.repos)
		if !reflect.DeepEqual(r, tt.want) {
			t.Errorf("got != want\ngot:  %v\nwant: %v", r, tt.want)
		}
	}
}
