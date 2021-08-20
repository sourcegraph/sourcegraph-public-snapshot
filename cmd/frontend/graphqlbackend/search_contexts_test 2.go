package graphqlbackend

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestAutoDefinedSearchContexts(t *testing.T) {
	key := int32(1)
	username := "alice"
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: key})
	db := new(dbtesting.MockDB)

	orig := envvar.SourcegraphDotComMode()
	envvar.MockSourcegraphDotComMode(true)
	defer envvar.MockSourcegraphDotComMode(orig) // reset

	database.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{Username: username}, nil
	}
	defer resetMocks()

	searchContexts, err := (&schemaResolver{db: db}).AutoDefinedSearchContexts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []*searchContextResolver{
		{sc: searchcontexts.GetGlobalSearchContext(), db: db},
		{sc: searchcontexts.GetUserSearchContext(username, key), db: db},
	}
	if !reflect.DeepEqual(searchContexts, want) {
		t.Fatalf("got %+v, want %+v", searchContexts, want)
	}

	for _, resolver := range searchContexts {
		repositories, err := resolver.Repositories(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(repositories) != 0 {
			t.Fatal("auto-defined search contexts should not return repositories")
		}
	}
}

func TestSearchContexts(t *testing.T) {
	db := new(dbtesting.MockDB)
	ctx := context.Background()

	userID := int32(1)
	graphqlUserID := MarshalUserID(userID)

	query := "ctx"
	tests := []struct {
		name     string
		args     *listSearchContextsArgs
		wantErr  string
		wantOpts database.ListSearchContextsOptions
	}{
		{
			name:     "filtering by namespace",
			args:     &listSearchContextsArgs{Query: &query, Namespaces: []*graphql.ID{&graphqlUserID}},
			wantOpts: database.ListSearchContextsOptions{Name: query, NamespaceUserIDs: []int32{userID}, NamespaceOrgIDs: []int32{}, OrderBy: database.SearchContextsOrderBySpec},
		},
		{
			name:     "filtering by instance",
			args:     &listSearchContextsArgs{Query: &query, Namespaces: []*graphql.ID{nil}},
			wantOpts: database.ListSearchContextsOptions{Name: query, NamespaceUserIDs: []int32{}, NamespaceOrgIDs: []int32{}, NoNamespace: true, OrderBy: database.SearchContextsOrderBySpec},
		},
		{
			name:     "get all",
			args:     &listSearchContextsArgs{Query: &query},
			wantOpts: database.ListSearchContextsOptions{Name: query, NamespaceUserIDs: []int32{}, NamespaceOrgIDs: []int32{}, OrderBy: database.SearchContextsOrderBySpec},
		},
	}

	database.Mocks.SearchContexts.CountSearchContexts = func(ctx context.Context, opts database.ListSearchContextsOptions) (int32, error) {
		return 0, nil
	}
	defer func() { database.Mocks.SearchContexts.CountSearchContexts = nil }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database.Mocks.SearchContexts.ListSearchContexts = func(ctx context.Context, pageOpts database.ListSearchContextsPageOptions, opts database.ListSearchContextsOptions) ([]*types.SearchContext, error) {
				if diff := cmp.Diff(tt.wantOpts, opts); diff != "" {
					t.Fatalf("Mismatch (-want +got):\n%s", diff)
				}
				return []*types.SearchContext{}, nil
			}
			defer func() { database.Mocks.SearchContexts.ListSearchContexts = nil }()

			_, err := (&schemaResolver{db: db}).SearchContexts(ctx, tt.args)
			expectErr := tt.wantErr != ""
			if !expectErr && err != nil {
				t.Fatalf("expected no error, got %s", err)
			}
			if expectErr && err == nil {
				t.Fatalf("wanted error, got none")
			}
			if expectErr && err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("wanted error containing %s, got %s", tt.wantErr, err)
			}
		})
	}
}
