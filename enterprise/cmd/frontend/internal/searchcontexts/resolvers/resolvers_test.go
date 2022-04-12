package resolvers

import (
	"context"
	"reflect"
	"strings"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestAutoDefinedSearchContexts(t *testing.T) {
	t.Run("Auto defined search contexts for user without organizations connected to repositories", func(t *testing.T) {
		key := int32(1)
		username := "alice"
		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: key})

		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig) // reset

		users := database.NewMockUserStore()
		users.GetByIDFunc.SetDefaultReturn(&types.User{Username: username}, nil)

		orgs := database.NewMockOrgStore()
		orgs.GetOrgsWithRepositoriesByUserIDFunc.SetDefaultReturn([]*types.Org{}, nil)

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.OrgsFunc.SetDefaultReturn(orgs)

		searchContexts, err := (&Resolver{db: db}).AutoDefinedSearchContexts(ctx)
		if err != nil {
			t.Fatal(err)
		}
		want := []graphqlbackend.SearchContextResolver{
			&searchContextResolver{sc: searchcontexts.GetGlobalSearchContext(), db: db},
			&searchContextResolver{sc: searchcontexts.GetUserSearchContext(key, username), db: db},
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

		mockrequire.Called(t, orgs.GetOrgsWithRepositoriesByUserIDFunc)
		mockrequire.Called(t, users.GetByIDFunc)
	})

	t.Run("Auto defined search contexts for user where 1 organization has repository connected", func(t *testing.T) {
		key := int32(1)
		username := "alice"
		orgID := int32(42)
		orgName := "acme"
		orgDisplayName := "ACME Company"
		ctx := context.Background()
		ctx = actor.WithActor(ctx, &actor.Actor{UID: key})

		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig) // reset

		users := database.NewMockUserStore()
		users.GetByIDFunc.SetDefaultReturn(&types.User{Username: username}, nil)
		orgs := database.NewMockOrgStore()
		orgs.GetOrgsWithRepositoriesByUserIDFunc.SetDefaultReturn([]*types.Org{{
			ID:          orgID,
			Name:        orgName,
			DisplayName: &orgDisplayName,
		}}, nil)

		db := database.NewMockDB()
		db.UsersFunc.SetDefaultReturn(users)
		db.OrgsFunc.SetDefaultReturn(orgs)

		searchContexts, err := (&Resolver{db: db}).AutoDefinedSearchContexts(ctx)
		if err != nil {
			t.Fatal(err)
		}
		want := []graphqlbackend.SearchContextResolver{
			&searchContextResolver{sc: searchcontexts.GetGlobalSearchContext(), db: db},
			&searchContextResolver{sc: searchcontexts.GetUserSearchContext(key, username), db: db},
			&searchContextResolver{sc: searchcontexts.GetOrganizationSearchContext(orgID, orgName, orgDisplayName), db: db},
		}
		if !reflect.DeepEqual(searchContexts, want) {
			t.Fatalf("got %+v, want %+v", searchContexts, want)
		}

		mockrequire.Called(t, users.GetByIDFunc)
		mockrequire.Called(t, orgs.GetOrgsWithRepositoriesByUserIDFunc)
	})
}

func TestSearchContexts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	userID := int32(1)
	graphqlUserID := graphqlbackend.MarshalUserID(userID)

	query := "ctx"
	tests := []struct {
		name     string
		args     *graphqlbackend.ListSearchContextsArgs
		wantErr  string
		wantOpts database.ListSearchContextsOptions
	}{
		{
			name:     "filtering by namespace",
			args:     &graphqlbackend.ListSearchContextsArgs{Query: &query, Namespaces: []*graphql.ID{&graphqlUserID}},
			wantOpts: database.ListSearchContextsOptions{Name: query, NamespaceUserIDs: []int32{userID}, NamespaceOrgIDs: []int32{}, OrderBy: database.SearchContextsOrderBySpec},
		},
		{
			name:     "filtering by instance",
			args:     &graphqlbackend.ListSearchContextsArgs{Query: &query, Namespaces: []*graphql.ID{nil}},
			wantOpts: database.ListSearchContextsOptions{Name: query, NamespaceUserIDs: []int32{}, NamespaceOrgIDs: []int32{}, NoNamespace: true, OrderBy: database.SearchContextsOrderBySpec},
		},
		{
			name:     "get all",
			args:     &graphqlbackend.ListSearchContextsArgs{Query: &query},
			wantOpts: database.ListSearchContextsOptions{Name: query, NamespaceUserIDs: []int32{}, NamespaceOrgIDs: []int32{}, OrderBy: database.SearchContextsOrderBySpec},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := database.NewMockSearchContextsStore()
			sc.CountSearchContextsFunc.SetDefaultReturn(0, nil)
			sc.ListSearchContextsFunc.SetDefaultHook(func(ctx context.Context, pageOpts database.ListSearchContextsPageOptions, opts database.ListSearchContextsOptions) ([]*types.SearchContext, error) {
				if diff := cmp.Diff(tt.wantOpts, opts); diff != "" {
					t.Fatalf("Mismatch (-want +got):\n%s", diff)
				}
				return []*types.SearchContext{}, nil
			})

			db := database.NewMockDB()
			db.SearchContextsFunc.SetDefaultReturn(sc)

			_, err := (&Resolver{db: db}).SearchContexts(ctx, tt.args)
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
			mockrequire.Called(t, sc.CountSearchContextsFunc)
			mockrequire.Called(t, sc.ListSearchContextsFunc)
		})
	}
}
