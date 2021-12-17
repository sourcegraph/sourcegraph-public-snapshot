package graphqlbackend

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSearchFilterSuggestions(t *testing.T) {
	repos := database.NewMockRepoStore()
	repos.ListFunc.SetDefaultReturn(
		[]*types.Repo{
			{Name: "github.com/foo/repo"},
			{Name: "bar-repo"},
		},
		nil,
	)

	db := database.NewMockDB()
	db.ReposFunc.SetDefaultReturn(repos)

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

		r, err := newSchemaResolver(db).SearchFilterSuggestions(context.Background())
		if err != nil {
			t.Fatal("SearchFilterSuggestions:", err)
		}

		sort.Strings(r.repos)
		assert.Equal(t, tt.want, r)
	}
}
