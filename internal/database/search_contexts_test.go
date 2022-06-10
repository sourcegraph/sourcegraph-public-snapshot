package database

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func createSearchContexts(ctx context.Context, store SearchContextsStore, searchContexts []*types.SearchContext) ([]*types.SearchContext, error) {
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
	db := NewDB(dbtest.NewDB(t))
	t.Parallel()
	ctx := actor.WithInternalActor(context.Background())
	u := db.Users()
	o := db.Orgs()
	sc := db.SearchContexts()

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

func TestSearchContexts_Update(t *testing.T) {
	db := NewDB(dbtest.NewDB(t))
	t.Parallel()
	ctx := actor.WithInternalActor(context.Background())
	u := db.Users()
	o := db.Orgs()
	sc := db.SearchContexts()

	user, err := u.Create(ctx, NewUser{Username: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	displayName := "My Org"
	org, err := o.Create(ctx, "myorg", &displayName)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	created, err := createSearchContexts(ctx, sc, []*types.SearchContext{
		{Name: "instance", Description: "instance level", Public: true},
		{Name: "user", Description: "user level", Public: true, NamespaceUserID: user.ID},
		{Name: "org", Description: "org level", Public: true, NamespaceOrgID: org.ID},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	instanceSC := created[0]
	userSC := created[1]
	orgSC := created[2]

	set := func(sc *types.SearchContext, f func(*types.SearchContext)) *types.SearchContext {
		copied := *sc
		f(&copied)
		return &copied
	}

	tests := []struct {
		name    string
		updated *types.SearchContext
		revs    []*types.SearchContextRepositoryRevisions
	}{
		{
			name:    "update public",
			updated: set(instanceSC, func(sc *types.SearchContext) { sc.Public = false }),
		},
		{
			name:    "update description",
			updated: set(userSC, func(sc *types.SearchContext) { sc.Description = "testdescription" }),
		},
		{
			name:    "update name",
			updated: set(orgSC, func(sc *types.SearchContext) { sc.Name = "testname" }),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := sc.UpdateSearchContextWithRepositoryRevisions(ctx, tt.updated, nil)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// Ignore updatedAt change
			updated.UpdatedAt = tt.updated.UpdatedAt
			if diff := cmp.Diff(tt.updated, updated); diff != "" {
				t.Fatalf("unexpected result: %s", diff)
			}
		})
	}
}

func TestSearchContexts_List(t *testing.T) {
	db := NewDB(dbtest.NewDB(t))
	t.Parallel()
	ctx := actor.WithInternalActor(context.Background())
	u := db.Users()
	sc := db.SearchContexts()

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
		ListSearchContextsOptions{NamespaceUserIDs: []int32{user.ID}},
	)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	if !reflect.DeepEqual(wantUserSearchContexts, gotUserSearchContexts) {
		t.Fatalf("wanted %v search contexts, got %v", wantUserSearchContexts, gotUserSearchContexts)
	}
}

func TestSearchContexts_PaginationAndCount(t *testing.T) {
	db := NewDB(dbtest.NewDB(t))
	t.Parallel()
	ctx := actor.WithInternalActor(context.Background())
	u := db.Users()
	o := db.Orgs()
	sc := db.SearchContexts()

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
			pageOptions:        ListSearchContextsPageOptions{First: 2, After: 1},
			totalCount:         4,
		},
		{
			name:               "user-level contexts",
			wantSearchContexts: createdSearchContexts[6:7],
			options:            ListSearchContextsOptions{NamespaceUserIDs: []int32{user.ID}},
			pageOptions:        ListSearchContextsPageOptions{First: 1, After: 2},
			totalCount:         3,
		},
		{
			name:               "org-level contexts",
			wantSearchContexts: createdSearchContexts[7:9],
			options:            ListSearchContextsOptions{NamespaceOrgIDs: []int32{org.ID}},
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
		{
			name:               "by namespace name only",
			wantSearchContexts: []*types.SearchContext{createdSearchContexts[4], createdSearchContexts[5], createdSearchContexts[6]},
			options:            ListSearchContextsOptions{NamespaceName: "u"},
			pageOptions:        ListSearchContextsPageOptions{First: 3},
			totalCount:         3,
		},
		{
			name:               "by namespace name and search context name",
			wantSearchContexts: []*types.SearchContext{createdSearchContexts[8]},
			options:            ListSearchContextsOptions{NamespaceName: "org", Name: "v2"},
			pageOptions:        ListSearchContextsPageOptions{First: 1},
			totalCount:         1,
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
	db := NewDB(dbtest.NewDB(t))
	t.Parallel()
	ctx := actor.WithInternalActor(context.Background())
	u := db.Users()
	o := db.Orgs()
	sc := db.SearchContexts()

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
	db := NewDB(dbtest.NewDB(t))
	t.Parallel()
	ctx := actor.WithInternalActor(context.Background())
	sc := db.SearchContexts()
	r := db.Repos()

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

	repoAName := types.MinimalRepo{ID: repoA.ID, Name: repoA.Name}
	repoBName := types.MinimalRepo{ID: repoB.ID, Name: repoB.Name}

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
	db := NewDB(dbtest.NewDB(t))
	t.Parallel()
	internalCtx := actor.WithInternalActor(context.Background())
	u := db.Users()
	o := db.Orgs()
	om := db.OrgMembers()
	sc := db.SearchContexts()

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
			name:               "site-admin user2 has access to all public contexts and private instance-level contexts",
			userID:             user2.ID,
			wantSearchContexts: []*types.SearchContext{searchContexts[0], searchContexts[1], searchContexts[2], searchContexts[4]},
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
			name:          "authenticated site-admin user2 does not have access to private user1 context",
			userID:        user2.ID,
			searchContext: searchContexts[3],
			siteAdmin:     true,
			wantErr:       "search context not found",
		},
		{
			name:          "authenticated user1 does not have access to private instance-level context",
			userID:        user1.ID,
			searchContext: searchContexts[1],
			wantErr:       "search context not found",
		},
		{
			name:          "site-admin user2 has access to private instance-level context",
			userID:        user2.ID,
			siteAdmin:     true,
			searchContext: searchContexts[1],
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
	db := NewDB(dbtest.NewDB(t))
	t.Parallel()
	ctx := context.Background()
	sc := db.SearchContexts()

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

func reverseSearchContextsSlice(s []*types.SearchContext) []*types.SearchContext {
	copySlice := make([]*types.SearchContext, len(s))
	copy(copySlice, s)
	for i, j := 0, len(copySlice)-1; i < j; i, j = i+1, j-1 {
		copySlice[i], copySlice[j] = copySlice[j], copySlice[i]
	}
	return copySlice
}

func getSearchContextNames(s []*types.SearchContext) []string {
	names := make([]string, 0, len(s))
	for _, sc := range s {
		names = append(names, sc.Name)
	}
	return names
}

func TestSearchContexts_OrderBy(t *testing.T) {
	db := NewDB(dbtest.NewDB(t))
	t.Parallel()
	internalCtx := actor.WithInternalActor(context.Background())
	u := db.Users()
	o := db.Orgs()
	om := db.OrgMembers()
	sc := db.SearchContexts()

	user1, err := u.Create(internalCtx, NewUser{Username: "u1", Password: "p"})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	err = u.SetIsSiteAdmin(internalCtx, user1.ID, false)
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
		{Name: "A-instance-level", Public: true},
		{Name: "B-instance-level", Public: false},
		{Name: "A-user-level", Public: true, NamespaceUserID: user1.ID},
		{Name: "B-user-level", Public: false, NamespaceUserID: user1.ID},
		{Name: "A-org-level", Public: true, NamespaceOrgID: org.ID},
		{Name: "B-org-level", Public: false, NamespaceOrgID: org.ID},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = sc.UpdateSearchContextWithRepositoryRevisions(internalCtx, searchContexts[1], nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = sc.UpdateSearchContextWithRepositoryRevisions(internalCtx, searchContexts[3], nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = sc.UpdateSearchContextWithRepositoryRevisions(internalCtx, searchContexts[5], nil)
	if err != nil {
		t.Fatal(err)
	}

	searchContextsOrderedBySpec := []*types.SearchContext{searchContexts[4], searchContexts[5], searchContexts[2], searchContexts[3], searchContexts[0], searchContexts[1]}
	searchContextsOrderedByUpdatedAt := []*types.SearchContext{searchContexts[0], searchContexts[2], searchContexts[4], searchContexts[1], searchContexts[3], searchContexts[5]}

	tests := []struct {
		name                   string
		orderBy                SearchContextsOrderByOption
		descending             bool
		wantSearchContextNames []string
	}{
		{
			name:                   "order by id",
			orderBy:                SearchContextsOrderByID,
			wantSearchContextNames: getSearchContextNames(searchContexts),
		},
		{
			name:                   "order by spec",
			orderBy:                SearchContextsOrderBySpec,
			wantSearchContextNames: getSearchContextNames(searchContextsOrderedBySpec),
		},
		{
			name:                   "order by updated at",
			orderBy:                SearchContextsOrderByUpdatedAt,
			wantSearchContextNames: getSearchContextNames(searchContextsOrderedByUpdatedAt),
		},
		{
			name:                   "order by id descending",
			orderBy:                SearchContextsOrderByID,
			descending:             true,
			wantSearchContextNames: getSearchContextNames(reverseSearchContextsSlice(searchContexts)),
		},
		{
			name:                   "order by spec descending",
			orderBy:                SearchContextsOrderBySpec,
			descending:             true,
			wantSearchContextNames: getSearchContextNames(reverseSearchContextsSlice(searchContextsOrderedBySpec)),
		},
		{
			name:                   "order by updated at descending",
			orderBy:                SearchContextsOrderByUpdatedAt,
			descending:             true,
			wantSearchContextNames: getSearchContextNames(reverseSearchContextsSlice(searchContextsOrderedByUpdatedAt)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSearchContexts, err := sc.ListSearchContexts(internalCtx, ListSearchContextsPageOptions{First: 6}, ListSearchContextsOptions{OrderBy: tt.orderBy, OrderByDescending: tt.descending})
			if err != nil {
				t.Fatal(err)
			}
			gotSearchContextNames := getSearchContextNames(gotSearchContexts)
			if !reflect.DeepEqual(tt.wantSearchContextNames, gotSearchContextNames) {
				t.Fatalf("wanted %+v search contexts, got %+v", tt.wantSearchContextNames, gotSearchContextNames)
			}
		})
	}
}

func TestSearchContexts_GetAllRevisionsForRepos(t *testing.T) {
	db := NewDB(dbtest.NewDB(t))
	t.Parallel()
	// Required for this DB query.
	internalCtx := actor.WithInternalActor(context.Background())
	sc := db.SearchContexts()
	r := db.Repos()

	repos := []*types.Repo{
		{Name: "testA", URI: "https://example.com/a"},
		{Name: "testB", URI: "https://example.com/b"},
		{Name: "testC", URI: "https://example.com/c"},
	}
	err := r.Create(internalCtx, repos...)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	testRevision := "asdf"
	searchContexts := []*types.SearchContext{
		{Name: "public-instance-level", Public: true},
		{Name: "private-instance-level", Public: false},
		{Name: "deleted", Public: true},
	}
	for idx, searchContext := range searchContexts {
		searchContexts[idx], err = sc.CreateSearchContextWithRepositoryRevisions(
			internalCtx,
			searchContext,
			[]*types.SearchContextRepositoryRevisions{{Repo: types.MinimalRepo{ID: repos[idx].ID, Name: repos[idx].Name}, Revisions: []string{testRevision}}},
		)
		if err != nil {
			t.Fatalf("Expected no error, got %s", err)
		}
	}

	if err := sc.DeleteSearchContext(internalCtx, searchContexts[2].ID); err != nil {
		t.Fatalf("Failed to delete search context %s", err)
	}

	listSearchContextsTests := []struct {
		name    string
		repoIDs []api.RepoID
		want    map[api.RepoID][]string
	}{
		{
			name:    "all contexts, deleted ones excluded",
			repoIDs: []api.RepoID{repos[0].ID, repos[1].ID, repos[2].ID},
			want: map[api.RepoID][]string{
				repos[0].ID: {testRevision},
				repos[1].ID: {testRevision},
			},
		},
		{
			name:    "subset of repos",
			repoIDs: []api.RepoID{repos[0].ID},
			want: map[api.RepoID][]string{
				repos[0].ID: {testRevision},
			},
		},
	}

	for _, tt := range listSearchContextsTests {
		t.Run(tt.name, func(t *testing.T) {
			gotSearchContexts, err := sc.GetAllRevisionsForRepos(internalCtx, tt.repoIDs)
			if err != nil {
				t.Fatalf("Expected no error, got %s", err)
			}
			if !reflect.DeepEqual(tt.want, gotSearchContexts) {
				t.Fatalf("wanted %v search contexts, got %v", tt.want, gotSearchContexts)
			}
		})
	}
}
