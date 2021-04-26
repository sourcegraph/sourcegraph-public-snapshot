package database

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func createSearchContexts(ctx context.Context, store *SearchContextsStore, searchContexts []*types.SearchContext) ([]*types.SearchContext, error) {
	emptyRepositoryRevisions := []*types.SearchContextRepositoryRevisions{}
	createdSearchContexts := make([]*types.SearchContext, len(searchContexts))
	for idx, searchContext := range searchContexts {
		createdSearchContext, err := store.CreateSearchContextWithRepositoryRevisions(ctx, searchContext, emptyRepositoryRevisions)
		if err != nil {
			return nil, err
		}
		createdSearchContexts[idx] = createdSearchContext
	}
	return createdSearchContexts, nil
}

func TestSearchContexts_Get(t *testing.T) {
	db := dbtesting.GetDB(t)
	ctx := actor.WithInternalActor(context.Background())
	u := Users(db)
	o := Orgs(db)
	sc := SearchContexts(db)

	user, err := u.Create(ctx, NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	displayName := "My Org"
	org, err := o.Create(ctx, "myorg", &displayName)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	createdSearchContexts, err := createSearchContexts(ctx, sc, []*types.SearchContext{
		{Name: "instance", Description: "instance level", Public: true},
		{Name: "user", Description: "user level", Public: true, NamespaceUserID: user.ID},
		{Name: "org", Description: "org level", Public: true, NamespaceOrgID: org.ID},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	tests := []struct {
		name    string
		opts    GetSearchContextOptions
		want    *types.SearchContext
		wantErr string
	}{
		{name: "get instance-level search context", opts: GetSearchContextOptions{Name: "instance"}, want: createdSearchContexts[0]},
		{name: "get user search context", opts: GetSearchContextOptions{Name: "user", NamespaceUserID: user.ID}, want: createdSearchContexts[1]},
		{name: "get org search context", opts: GetSearchContextOptions{Name: "org", NamespaceOrgID: org.ID}, want: createdSearchContexts[2]},
		{name: "get user and org context", opts: GetSearchContextOptions{NamespaceUserID: 1, NamespaceOrgID: 2}, wantErr: "options NamespaceUserID and NamespaceOrgID are mutually exclusive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchContext, err := sc.GetSearchContext(ctx, tt.opts)
			if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("got error %v, want it to contain %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.want, searchContext) {
				t.Fatalf("wanted %v search contexts, got %v", tt.want, searchContext)
			}
		})
	}
}

func TestSearchContexts_List(t *testing.T) {
	db := dbtesting.GetDB(t)
	ctx := actor.WithInternalActor(context.Background())
	u := Users(db)
	sc := SearchContexts(db)

	user, err := u.Create(ctx, NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	createdSearchContexts, err := createSearchContexts(ctx, sc, []*types.SearchContext{
		{Name: "instance", Description: "instance level", Public: true},
		{Name: "user", Description: "user level", Public: true, NamespaceUserID: user.ID},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	wantInstanceLevelSearchContexts := createdSearchContexts[:1]
	gotInstanceLevelSearchContexts, err := sc.ListSearchContexts(
		ctx,
		ListSearchContextsPageOptions{First: 1},
		ListSearchContextsOptions{NoNamespace: true},
	)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	if !reflect.DeepEqual(wantInstanceLevelSearchContexts, gotInstanceLevelSearchContexts) {
		t.Fatalf("wanted %v search contexts, got %v", wantInstanceLevelSearchContexts, gotInstanceLevelSearchContexts)
	}

	wantUserSearchContexts := createdSearchContexts[1:]
	gotUserSearchContexts, err := sc.ListSearchContexts(
		ctx,
		ListSearchContextsPageOptions{First: 1},
		ListSearchContextsOptions{NamespaceUserID: user.ID},
	)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	if !reflect.DeepEqual(wantUserSearchContexts, gotUserSearchContexts) {
		t.Fatalf("wanted %v search contexts, got %v", wantUserSearchContexts, gotUserSearchContexts)
	}
}

func TestSearchContexts_PaginationAndCount(t *testing.T) {
	db := dbtesting.GetDB(t)
	ctx := actor.WithInternalActor(context.Background())
	u := Users(db)
	o := Orgs(db)
	sc := SearchContexts(db)

	user, err := u.Create(ctx, NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	displayName := "My Org"
	org, err := o.Create(ctx, "myorg", &displayName)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	createdSearchContexts, err := createSearchContexts(ctx, sc, []*types.SearchContext{
		{Name: "instance-v1", Public: true},
		{Name: "instance-v2", Public: true},
		{Name: "instance-v3", Public: true},
		{Name: "instance-v4", Public: true},
		{Name: "user-v1", Public: true, NamespaceUserID: user.ID},
		{Name: "user-v2", Public: true, NamespaceUserID: user.ID},
		{Name: "user-v3", Public: true, NamespaceUserID: user.ID},
		{Name: "org-v1", Public: true, NamespaceOrgID: org.ID},
		{Name: "org-v2", Public: true, NamespaceOrgID: org.ID},
		{Name: "org-v3", Public: true, NamespaceOrgID: org.ID},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	tests := []struct {
		name               string
		wantSearchContexts []*types.SearchContext
		options            ListSearchContextsOptions
		pageOptions        ListSearchContextsPageOptions
		totalCount         int32
	}{
		{
			name:               "instance-level contexts",
			wantSearchContexts: createdSearchContexts[1:3],
			options:            ListSearchContextsOptions{Name: "instance-v", NoNamespace: true},
			pageOptions:        ListSearchContextsPageOptions{First: 2, AfterID: createdSearchContexts[0].ID},
			totalCount:         4,
		},
		{
			name:               "user-level contexts",
			wantSearchContexts: createdSearchContexts[6:7],
			options:            ListSearchContextsOptions{NamespaceUserID: user.ID},
			pageOptions:        ListSearchContextsPageOptions{First: 1, AfterID: createdSearchContexts[5].ID},
			totalCount:         3,
		},
		{
			name:               "org-level contexts",
			wantSearchContexts: createdSearchContexts[7:9],
			options:            ListSearchContextsOptions{NamespaceOrgID: org.ID},
			pageOptions:        ListSearchContextsPageOptions{First: 2},
			totalCount:         3,
		},
		{
			name:               "by name only",
			wantSearchContexts: []*types.SearchContext{createdSearchContexts[0], createdSearchContexts[4]},
			options:            ListSearchContextsOptions{Name: "v1"},
			pageOptions:        ListSearchContextsPageOptions{First: 2},
			totalCount:         3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSearchContexts, err := sc.ListSearchContexts(ctx, tt.pageOptions, tt.options)
			if err != nil {
				t.Fatalf("Expected no error, got %s", err)
			}
			if !reflect.DeepEqual(tt.wantSearchContexts, gotSearchContexts) {
				t.Fatalf("wanted %+v search contexts, got %+v", tt.wantSearchContexts, gotSearchContexts)
			}
		})
	}
}

func TestSearchContexts_CaseInsensitiveNames(t *testing.T) {
	db := dbtesting.GetDB(t)
	ctx := actor.WithInternalActor(context.Background())
	u := Users(db)
	o := Orgs(db)
	sc := SearchContexts(db)

	user, err := u.Create(ctx, NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	displayName := "My Org"
	org, err := o.Create(ctx, "myorg", &displayName)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	tests := []struct {
		name           string
		searchContexts []*types.SearchContext
		wantErr        string
	}{
		{
			name:           "contexts with same case-insensitive name and different namespaces",
			searchContexts: []*types.SearchContext{{Name: "ctx"}, {Name: "Ctx", NamespaceUserID: user.ID}, {Name: "CTX", NamespaceOrgID: org.ID}},
		},
		{
			name:           "same case-insensitive name, same instance-level namespace",
			searchContexts: []*types.SearchContext{{Name: "instance"}, {Name: "InStanCe"}},
			wantErr:        `violates unique constraint "search_contexts_name_without_namespace_unique"`,
		},
		{
			name:           "same case-insensitive name, same user namespace",
			searchContexts: []*types.SearchContext{{Name: "user", NamespaceUserID: user.ID}, {Name: "UsEr", NamespaceUserID: user.ID}},
			wantErr:        `violates unique constraint "search_contexts_name_namespace_user_id_unique"`,
		},
		{
			name:           "same case-insensitive name, same org namespace",
			searchContexts: []*types.SearchContext{{Name: "org", NamespaceOrgID: org.ID}, {Name: "OrG", NamespaceOrgID: org.ID}},
			wantErr:        `violates unique constraint "search_contexts_name_namespace_org_id_unique"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := createSearchContexts(ctx, sc, tt.searchContexts)
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

func TestSearchContexts_CreateAndSetRepositoryRevisions(t *testing.T) {
	db := dbtesting.GetDB(t)
	ctx := actor.WithInternalActor(context.Background())
	sc := SearchContexts(db)
	r := Repos(db)

	err := r.Create(ctx, &types.Repo{Name: "testA", URI: "https://example.com/a"}, &types.Repo{Name: "testB", URI: "https://example.com/b"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	repoA, err := r.GetByName(ctx, "testA")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	repoB, err := r.GetByName(ctx, "testB")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	repoAName := types.RepoName{ID: repoA.ID, Name: repoA.Name}
	repoBName := types.RepoName{ID: repoB.ID, Name: repoB.Name}

	// Create a search context with initial repository revisions
	initialRepositoryRevisions := []*types.SearchContextRepositoryRevisions{
		{Repo: repoAName, Revisions: []string{"branch-1", "branch-6"}},
		{Repo: repoBName, Revisions: []string{"branch-2"}},
	}
	searchContext, err := sc.CreateSearchContextWithRepositoryRevisions(
		ctx,
		&types.SearchContext{Name: "sc", Description: "sc", Public: true},
		initialRepositoryRevisions,
	)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	gotRepositoryRevisions, err := sc.GetSearchContextRepositoryRevisions(ctx, searchContext.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	if !reflect.DeepEqual(initialRepositoryRevisions, gotRepositoryRevisions) {
		t.Fatalf("wanted %v repository revisions, got %v", initialRepositoryRevisions, gotRepositoryRevisions)
	}

	// Modify the repository revisions for the search context
	modifiedRepositoryRevisions := []*types.SearchContextRepositoryRevisions{
		{Repo: repoAName, Revisions: []string{"branch-1", "branch-3"}},
		{Repo: repoBName, Revisions: []string{"branch-0", "branch-2", "branch-4"}},
	}
	err = sc.SetSearchContextRepositoryRevisions(ctx, searchContext.ID, modifiedRepositoryRevisions)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	gotRepositoryRevisions, err = sc.GetSearchContextRepositoryRevisions(ctx, searchContext.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	if !reflect.DeepEqual(modifiedRepositoryRevisions, gotRepositoryRevisions) {
		t.Fatalf("wanted %v repository revisions, got %v", modifiedRepositoryRevisions, gotRepositoryRevisions)
	}
}

func TestSearchContexts_Permissions(t *testing.T) {
	db := dbtesting.GetDB(t)
	internalCtx := actor.WithInternalActor(context.Background())
	u := Users(db)
	o := Orgs(db)
	om := OrgMembers(db)
	sc := SearchContexts(db)

	user1, err := u.Create(internalCtx, NewUser{Username: "u1", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	err = u.SetIsSiteAdmin(internalCtx, user1.ID, false)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	user2, err := u.Create(internalCtx, NewUser{Username: "u2", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	displayName := "My Org"
	org, err := o.Create(internalCtx, "myorg", &displayName)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	_, err = om.Create(internalCtx, org.ID, user1.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	searchContexts, err := createSearchContexts(internalCtx, sc, []*types.SearchContext{
		{Name: "public-instance-level", Public: true},
		{Name: "private-instance-level", Public: false},
		{Name: "public-user-level", Public: true, NamespaceUserID: user1.ID},
		{Name: "private-user-level", Public: false, NamespaceUserID: user1.ID},
		{Name: "public-org-level", Public: true, NamespaceOrgID: org.ID},
		{Name: "private-org-level", Public: false, NamespaceOrgID: org.ID},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	listSearchContextsTests := []struct {
		name               string
		userID             int32
		wantSearchContexts []*types.SearchContext
		siteAdmin          bool
	}{
		{
			name:               "unauthenticated user only has access to public contexts",
			userID:             int32(0),
			wantSearchContexts: []*types.SearchContext{searchContexts[0], searchContexts[2], searchContexts[4]},
		},
		{
			name:               "authenticated user1 has access to his private context, his orgs private context, and all public contexts",
			userID:             user1.ID,
			wantSearchContexts: []*types.SearchContext{searchContexts[0], searchContexts[2], searchContexts[3], searchContexts[4], searchContexts[5]},
		},
		{
			name:               "authenticated user2 has access to all public contexts and no private contexts",
			userID:             user2.ID,
			wantSearchContexts: []*types.SearchContext{searchContexts[0], searchContexts[2], searchContexts[4]},
		},
		{
			name:               "site-admin user2 has access to all contexts",
			userID:             user2.ID,
			wantSearchContexts: searchContexts,
			siteAdmin:          true,
		},
	}

	for _, tt := range listSearchContextsTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.siteAdmin {
				err = u.SetIsSiteAdmin(internalCtx, tt.userID, true)
				if err != nil {
					t.Fatalf("Expected no error, got %s", err)
				}
			}

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: tt.userID})
			gotSearchContexts, err := sc.ListSearchContexts(ctx,
				ListSearchContextsPageOptions{First: int32(len(searchContexts))},
				ListSearchContextsOptions{},
			)
			if err != nil {
				t.Fatalf("Expected no error, got %s", err)
			}
			if !reflect.DeepEqual(tt.wantSearchContexts, gotSearchContexts) {
				t.Fatalf("wanted %v search contexts, got %v", tt.wantSearchContexts, gotSearchContexts)
			}

			if tt.siteAdmin {
				err = u.SetIsSiteAdmin(internalCtx, tt.userID, false)
				if err != nil {
					t.Fatalf("Expected no error, got %s", err)
				}
			}
		})
	}

	getSearchContextTests := []struct {
		name          string
		userID        int32
		searchContext *types.SearchContext
		siteAdmin     bool
		wantErr       string
	}{
		{
			name:          "unauthenticated user does not have access to private context",
			userID:        int32(0),
			searchContext: searchContexts[3],
			wantErr:       "search context not found",
		},
		{
			name:          "authenticated user2 does not have access to private user1 context",
			userID:        user2.ID,
			searchContext: searchContexts[3],
			wantErr:       "search context not found",
		},
		{
			name:          "authenticated user2 does not have access to private org context",
			userID:        user2.ID,
			searchContext: searchContexts[5],
			wantErr:       "search context not found",
		},
		{
			name:          "authenticated site-admin user2 has access to private user1 context",
			userID:        user2.ID,
			searchContext: searchContexts[3],
			siteAdmin:     true,
		},
		{
			name:          "authenticated user1 does not have access to private instance-level context",
			userID:        user1.ID,
			searchContext: searchContexts[1],
			wantErr:       "search context not found",
		},
		{
			name:          "authenticated user1 has access to his private context",
			userID:        user1.ID,
			searchContext: searchContexts[3],
		},
		{
			name:          "authenticated user1 has access to his orgs private context",
			userID:        user1.ID,
			searchContext: searchContexts[5],
		},
	}

	for _, tt := range getSearchContextTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.siteAdmin {
				err = u.SetIsSiteAdmin(internalCtx, tt.userID, true)
				if err != nil {
					t.Fatalf("Expected no error, got %s", err)
				}
			}

			ctx := actor.WithActor(context.Background(), &actor.Actor{UID: tt.userID})
			gotSearchContext, err := sc.GetSearchContext(ctx,
				GetSearchContextOptions{
					Name:            tt.searchContext.Name,
					NamespaceUserID: tt.searchContext.NamespaceUserID,
					NamespaceOrgID:  tt.searchContext.NamespaceOrgID,
				},
			)

			expectErr := tt.wantErr != ""
			if !expectErr && err != nil {
				t.Fatalf("expected no error, got %s", err)
			}
			if !expectErr && !reflect.DeepEqual(tt.searchContext, gotSearchContext) {
				t.Fatalf("wanted %v search context, got %v", tt.searchContext, gotSearchContext)
			}
			if expectErr && err == nil {
				t.Fatalf("wanted error, got none")
			}
			if expectErr && err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("wanted error containing %s, got %s", tt.wantErr, err)
			}

			if tt.siteAdmin {
				err = u.SetIsSiteAdmin(internalCtx, tt.userID, false)
				if err != nil {
					t.Fatalf("Expected no error, got %s", err)
				}
			}
		})
	}
}

func TestSearchContexts_Delete(t *testing.T) {
	db := dbtesting.GetDB(t)
	ctx := context.Background()
	sc := SearchContexts(db)

	initialSearchContexts, err := createSearchContexts(ctx, sc, []*types.SearchContext{
		{Name: "ctx", Public: true},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	err = sc.DeleteSearchContext(ctx, initialSearchContexts[0].ID)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	// We should not be able to find the search context
	_, err = sc.GetSearchContext(ctx, GetSearchContextOptions{Name: initialSearchContexts[0].Name})
	if err != ErrSearchContextNotFound {
		t.Fatal("Expected not to find the search context")
	}

	// We should be able to create a search context with the same name
	_, err = createSearchContexts(ctx, sc, []*types.SearchContext{
		{Name: "ctx", Public: true},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
}
