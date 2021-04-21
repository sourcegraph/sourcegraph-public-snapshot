package graphqlbackend

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestAutoDefinedSearchContexts(t *testing.T) {
	key := int32(1)
	username := "alice"
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: key})
	db := new(dbtesting.MockDB)

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

	namespaceFilter := namespaceFilterType("NAMESPACE")
	instanceFilter := namespaceFilterType("INSTANCE")
	query := "ctx"
	tests := []struct {
		name     string
		args     *listSearchContextsArgs
		wantErr  string
		wantOpts database.ListSearchContextsOptions
	}{
		{
			name:     "filtering by namespace",
			args:     &listSearchContextsArgs{Query: &query, Namespace: &graphqlUserID, NamespaceFilterType: &namespaceFilter},
			wantOpts: database.ListSearchContextsOptions{Name: query, NamespaceUserID: userID},
		},
		{
			name:     "filtering by instance",
			args:     &listSearchContextsArgs{Query: &query, NamespaceFilterType: &instanceFilter},
			wantOpts: database.ListSearchContextsOptions{Name: query, NoNamespace: true},
		},
		{
			name:     "get all",
			args:     &listSearchContextsArgs{Query: &query},
			wantOpts: database.ListSearchContextsOptions{Name: query},
		},
		{
			name:    "cannot filter by namespace with nil namespace",
			args:    &listSearchContextsArgs{Query: &query, NamespaceFilterType: &namespaceFilter},
			wantErr: "namespace has to be non-nil if namespaceFilterType is NAMESPACE",
		},
		{
			name:    "cannot filter by instance with non-nil namespace",
			args:    &listSearchContextsArgs{Query: &query, Namespace: &graphqlUserID, NamespaceFilterType: &instanceFilter},
			wantErr: "namespace can only be used if namespaceFilterType is NAMESPACE",
		},
		{
			name:    "cannot use non-nil namespace if no filter is present",
			args:    &listSearchContextsArgs{Query: &query, Namespace: &graphqlUserID},
			wantErr: "namespace can only be used if namespaceFilterType is NAMESPACE",
		},
	}

	database.Mocks.SearchContexts.CountSearchContexts = func(ctx context.Context, opts database.ListSearchContextsOptions) (int32, error) {
		return 0, nil
	}
	defer func() { database.Mocks.SearchContexts.CountSearchContexts = nil }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database.Mocks.SearchContexts.ListSearchContexts = func(ctx context.Context, pageOpts database.ListSearchContextsPageOptions, opts database.ListSearchContextsOptions) ([]*types.SearchContext, error) {
				if !reflect.DeepEqual(tt.wantOpts, opts) {
					t.Fatalf("wanted %+v opts, got %+v", tt.wantOpts, opts)
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
