package searchcontexts

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestResolvingValidSearchContextSpecs(t *testing.T) {
	tests := []struct {
		name                  string
		searchContextSpec     string
		wantSearchContextName string
	}{
		{name: "resolve user search context", searchContextSpec: "@user", wantSearchContextName: "user"},
		{name: "resolve global search context", searchContextSpec: "global", wantSearchContextName: "global"},
		{name: "resolve empty search context as global", searchContextSpec: "", wantSearchContextName: "global"},
	}

	db := new(dbtesting.MockDB)
	database.Mocks.Namespaces.GetByName = func(ctx context.Context, name string) (*database.Namespace, error) {
		return &database.Namespace{Name: name, User: 1}, nil
	}
	defer func() { database.Mocks.Namespaces.GetByName = nil }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchContext, err := ResolveSearchContextSpec(context.Background(), db, tt.searchContextSpec)
			if err != nil {
				t.Fatal(err)
			}
			if searchContext.Name != tt.wantSearchContextName {
				t.Fatalf("got %q, expected %q", searchContext.Name, tt.wantSearchContextName)
			}
		})
	}
}

func TestResolvingInvalidSearchContextSpecs(t *testing.T) {
	tests := []struct {
		name              string
		searchContextSpec string
		wantErr           string
	}{
		{name: "invalid format", searchContextSpec: "+user", wantErr: "search context not found"},
		{name: "user not found", searchContextSpec: "@user", wantErr: "search context '@user' not found"},
		{name: "empty user not found", searchContextSpec: "@", wantErr: "search context '@' not found"},
	}

	db := new(dbtesting.MockDB)
	database.Mocks.Namespaces.GetByName = func(ctx context.Context, name string) (*database.Namespace, error) {
		return &database.Namespace{}, nil
	}
	database.Mocks.SearchContexts.GetSearchContext = func(ctx context.Context, opts database.GetSearchContextOptions) (*types.SearchContext, error) {
		return nil, errors.New("search context not found")
	}
	defer func() {
		database.Mocks.Namespaces.GetByName = nil
		database.Mocks.SearchContexts.GetSearchContext = nil
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolveSearchContextSpec(context.Background(), db, tt.searchContextSpec)
			if err == nil {
				t.Error("Expected error, but there was none")
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("err: got %q, expected %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestConstructingSearchContextSpecs(t *testing.T) {
	tests := []struct {
		name                  string
		searchContext         *types.SearchContext
		wantSearchContextSpec string
	}{
		{name: "global search context", searchContext: GetGlobalSearchContext(), wantSearchContextSpec: "global"},
		{name: "user search context", searchContext: &types.SearchContext{Name: "user", NamespaceUserID: 1}, wantSearchContextSpec: "@user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchContextSpec := GetSearchContextSpec(tt.searchContext)
			if searchContextSpec != tt.wantSearchContextSpec {
				t.Fatalf("got %q, expected %q", searchContextSpec, tt.wantSearchContextSpec)
			}
		})
	}
}

func TestGettingSearchContextFromVersionContext(t *testing.T) {
	tests := []struct {
		name              string
		versionContext    *schema.VersionContext
		wantSearchContext *types.SearchContext
	}{
		{
			name:              "simple version context",
			versionContext:    &schema.VersionContext{Name: "vc1", Description: "vc1 description"},
			wantSearchContext: &types.SearchContext{Name: "vc1", Description: "vc1 description", Public: true},
		},
		{
			name:              "version context with spaces in name",
			versionContext:    &schema.VersionContext{Name: "Version Context  2", Description: "Version Context 2 description"},
			wantSearchContext: &types.SearchContext{Name: "Version_Context_2", Description: "Version Context 2 description", Public: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSearchContext := getSearchContextFromVersionContext(tt.versionContext)
			if !reflect.DeepEqual(tt.wantSearchContext, gotSearchContext) {
				t.Fatalf("want %+v, got %+v", tt.wantSearchContext, gotSearchContext)
			}
		})
	}
}

func TestConvertingVersionContextToSearchContext(t *testing.T) {
	db := dbtesting.GetDB(t)
	ctx := context.Background()
	r := database.Repos(db)

	err := r.Create(ctx, &types.Repo{Name: "github.com/example/a"}, &types.Repo{Name: "github.com/example/b"})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	repoA, err := r.GetByName(ctx, "github.com/example/a")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	repoB, err := r.GetByName(ctx, "github.com/example/b")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	repoAName := &types.RepoName{ID: repoA.ID, Name: repoA.Name}
	repoBName := &types.RepoName{ID: repoB.ID, Name: repoB.Name}

	versionContext := &schema.VersionContext{
		Name:        "vc1",
		Description: "vc1 description",
		Revisions: []*schema.VersionContextRevision{
			{Repo: "github.com/example/a", Rev: "branch-1"},
			{Repo: "github.com/example/a", Rev: "branch-3"},
			{Repo: "github.com/example/b", Rev: "branch-2"},
		},
	}

	wantRepositoryRevisions := []*search.RepositoryRevisions{
		{Repo: repoAName, Revs: []search.RevisionSpecifier{{RevSpec: "branch-1"}, {RevSpec: "branch-3"}}},
		{Repo: repoBName, Revs: []search.RevisionSpecifier{{RevSpec: "branch-2"}}},
	}
	wantSearchContext := &types.SearchContext{ID: 1, Name: "vc1", Description: "vc1 description", Public: true}

	gotSearchContext, err := ConvertVersionContextToSearchContext(ctx, db, versionContext)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if !reflect.DeepEqual(wantSearchContext, gotSearchContext) {
		t.Fatalf("want search context %+v, got %+v", wantSearchContext, gotSearchContext)
	}

	gotRepositoryRevisions, err := database.SearchContexts(db).GetSearchContextRepositoryRevisions(ctx, gotSearchContext.ID)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if !reflect.DeepEqual(wantRepositoryRevisions, gotRepositoryRevisions) {
		t.Errorf("wanted %v repository revisions, got %v", wantRepositoryRevisions, gotRepositoryRevisions)
	}
}
