package database

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSavedSearchesCreate(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	user, err := db.Users().Create(ctx, NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	ctx = actor.WithActor(ctx, &actor.Actor{UID: user.ID})

	input := types.SavedSearch{
		Description:      "d",
		Query:            "q",
		Draft:            true,
		Owner:            types.NamespaceUser(user.ID),
		VisibilitySecret: true,
	}
	got, err := db.SavedSearches().Create(ctx, &input)
	if err != nil {
		t.Fatal(err)
	}
	want := input
	want.ID = got.ID
	want.CreatedByUser = &user.ID
	want.UpdatedByUser = &user.ID
	normalizeSavedSearch(got, &want)
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("got %+v, want %+v", *got, want)
	}
}

func TestSavedSearchesUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	user, err := db.Users().Create(ctx, NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	ctx = actor.WithActor(ctx, &actor.Actor{UID: user.ID})

	_, err = db.SavedSearches().Create(ctx, &types.SavedSearch{
		Description: "d",
		Query:       "q",
		Owner:       types.NamespaceUser(user.ID),
	})
	if err != nil {
		t.Fatal(err)
	}

	update := types.SavedSearch{
		ID:          1,
		Description: "test2",
		Query:       "test2",
	}
	got, err := db.SavedSearches().Update(ctx, &update)
	if err != nil {
		t.Fatal(err)
	}
	want := update
	want.Owner = types.NamespaceUser(user.ID)
	want.CreatedByUser = &user.ID
	want.UpdatedByUser = &user.ID
	normalizeSavedSearch(got, &want)
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("got %+v, want %+v", *got, want)
	}
}

func TestSavedSearchesUpdateOwner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	user, err := db.Users().Create(ctx, NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	org1, err := db.Orgs().Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx = actor.WithActor(ctx, &actor.Actor{UID: user.ID})
	fixture1, err := db.SavedSearches().Create(ctx, &types.SavedSearch{
		Description: "d",
		Query:       "q",
		Owner:       types.NamespaceUser(user.ID),
	})
	if err != nil {
		t.Fatal(err)
	}

	{
		// Transfer from user to org1.
		newOwner := types.NamespaceOrg(org1.ID)
		updated, err := db.SavedSearches().UpdateOwner(ctx, fixture1.ID, newOwner)
		if err != nil {
			t.Fatal(err)
		}
		got := updated.Owner
		want := newOwner
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	}

	{
		// Transfer back from org1 to user.
		newOwner := types.NamespaceUser(user.ID)
		updated, err := db.SavedSearches().UpdateOwner(ctx, fixture1.ID, newOwner)
		if err != nil {
			t.Fatal(err)
		}
		got := updated.Owner
		want := newOwner
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	}
}

func TestSavedSearchesUpdateVisibility(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	user, err := db.Users().Create(ctx, NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	ctx = actor.WithActor(ctx, &actor.Actor{UID: user.ID})

	fixture1, err := db.SavedSearches().Create(ctx, &types.SavedSearch{
		Description:      "d",
		Query:            "q",
		Owner:            types.NamespaceUser(user.ID),
		VisibilitySecret: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Make public then secret again.
	for _, secret := range []bool{false, true} {
		updated, err := db.SavedSearches().UpdateVisibility(ctx, fixture1.ID, secret)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := updated.VisibilitySecret, secret; !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	}
}

func TestSavedSearchesDelete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	user, err := db.Users().Create(ctx, NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	ctx = actor.WithActor(ctx, &actor.Actor{UID: user.ID})

	fixture1, err := db.SavedSearches().Create(ctx, &types.SavedSearch{
		Description: "d",
		Query:       "q",
		Owner:       types.NamespaceUser(user.ID),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("found", func(t *testing.T) {
		if err := db.SavedSearches().Delete(ctx, fixture1.ID); err != nil {
			t.Fatal(err)
		}
		if got, err := db.SavedSearches().Count(ctx, SavedSearchListArgs{}); err != nil {
			t.Fatal(err)
		} else if got != 0 {
			t.Error()
		}
	})

	t.Run("not found", func(t *testing.T) {
		if err := db.SavedSearches().Delete(ctx, 123); err != errSavedSearchNotFound {
			t.Fatal(err)
		}
	})
}

func TestSavedSearchesGetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	user, err := db.Users().Create(ctx, NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	ctx = actor.WithActor(ctx, &actor.Actor{UID: user.ID})

	input := types.SavedSearch{
		Description: "d",
		Query:       "q",
		Owner:       types.NamespaceUser(user.ID),
	}
	fixture1, err := db.SavedSearches().Create(ctx, &input)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("found", func(t *testing.T) {
		got, err := db.SavedSearches().GetByID(ctx, fixture1.ID)
		if err != nil {
			t.Fatal(err)
		}
		want := input
		want.ID = got.ID
		want.CreatedByUser = &user.ID
		want.UpdatedByUser = &user.ID
		normalizeSavedSearch(got, &want)
		if diff := cmp.Diff(want, *got); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("not found", func(t *testing.T) {
		if _, err := db.SavedSearches().GetByID(ctx, 123); err != errSavedSearchNotFound {
			t.Fatal(err)
		}
	})
}

func TestSavedSearches_ListCount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	user, err := db.Users().Create(ctx, NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}
	ctx = actor.WithActor(ctx, &actor.Actor{UID: user.ID})

	fixture1, err := db.SavedSearches().Create(ctx, &types.SavedSearch{
		Description:      "fixture1",
		Query:            "fixture1",
		Owner:            types.NamespaceUser(user.ID),
		VisibilitySecret: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	org1, err := db.Orgs().Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}
	org2, err := db.Orgs().Create(ctx, "org2", nil)
	if err != nil {
		t.Fatal(err)
	}
	fixture2, err := db.SavedSearches().Create(ctx, &types.SavedSearch{
		Description:      "fixture2",
		Query:            "fixture2",
		Owner:            types.NamespaceOrg(org1.ID),
		VisibilitySecret: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	fixture3, err := db.SavedSearches().Create(ctx, &types.SavedSearch{
		Description:      "fixture3",
		Query:            "fixture3",
		Owner:            types.NamespaceOrg(org2.ID),
		VisibilitySecret: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err = db.OrgMembers().Create(ctx, org1.ID, user.ID); err != nil {
		t.Fatal(err)
	}

	fixture4, err := db.SavedSearches().Create(ctx, &types.SavedSearch{
		Description:      "fixture4",
		Query:            "fixture4",
		Draft:            true,
		Owner:            types.NamespaceUser(user.ID),
		VisibilitySecret: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	user2, err := db.Users().Create(ctx, NewUser{Username: "u2"})
	if err != nil {
		t.Fatal(err)
	}
	fixture5, err := db.SavedSearches().Create(ctx, &types.SavedSearch{
		Description:      "fixture5",
		Query:            "fixture5",
		Owner:            types.NamespaceUser(user2.ID),
		VisibilitySecret: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	testListCount := func(t *testing.T, args SavedSearchListArgs, pgArgs *PaginationArgs, want []*types.SavedSearch) {
		t.Helper()

		if pgArgs == nil {
			pgArgs = &PaginationArgs{Ascending: true}
		}
		got, err := db.SavedSearches().List(ctx, args, pgArgs)
		if err != nil {
			t.Fatal(err)
		}
		normalizeSavedSearch(got...)
		normalizeSavedSearch(want...)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}

		gotCount, err := db.SavedSearches().Count(ctx, args)
		if err != nil {
			t.Fatal(err)
		}
		if wantCount := len(want); gotCount != wantCount {
			t.Errorf("got count %d, want %d", gotCount, wantCount)
		}
	}

	t.Run("list all", func(t *testing.T) {
		testListCount(t, SavedSearchListArgs{}, nil, []*types.SavedSearch{fixture1, fixture2, fixture3, fixture4, fixture5})
	})

	t.Run("query", func(t *testing.T) {
		testListCount(t, SavedSearchListArgs{Query: "Ure3"}, nil, []*types.SavedSearch{fixture3})
	})

	t.Run("empty result set", func(t *testing.T) {
		testListCount(t, SavedSearchListArgs{Query: "doesntmatch"}, nil, nil)
	})

	t.Run("list owned by user", func(t *testing.T) {
		userNS := types.NamespaceUser(user.ID)
		testListCount(t, SavedSearchListArgs{Owner: &userNS}, nil, []*types.SavedSearch{fixture1, fixture4})
	})

	t.Run("list owned by nonexistent user", func(t *testing.T) {
		userNS := types.NamespaceUser(1234999 /* user doesn't exist */)
		testListCount(t, SavedSearchListArgs{Owner: &userNS}, nil, nil)
	})

	t.Run("list owned by org1", func(t *testing.T) {
		orgNS := types.NamespaceOrg(org1.ID)
		testListCount(t, SavedSearchListArgs{Owner: &orgNS}, nil, []*types.SavedSearch{fixture2})
	})

	t.Run("affiliated with user", func(t *testing.T) {
		testListCount(t, SavedSearchListArgs{AffiliatedUser: &user.ID}, nil, []*types.SavedSearch{fixture1, fixture2, fixture4, fixture5})
	})

	t.Run("public only", func(t *testing.T) {
		testListCount(t, SavedSearchListArgs{PublicOnly: true}, nil, []*types.SavedSearch{fixture5})
	})

	t.Run("hide drafts", func(t *testing.T) {
		userNS := types.NamespaceUser(user.ID)
		testListCount(t, SavedSearchListArgs{Owner: &userNS, HideDrafts: true}, nil, []*types.SavedSearch{fixture1})
	})

	t.Run("order by", func(t *testing.T) {
		orderBy, ascending := SavedSearchesOrderByUpdatedAt.ToOptions()
		testListCount(t, SavedSearchListArgs{}, &PaginationArgs{OrderBy: orderBy, Ascending: ascending}, []*types.SavedSearch{fixture5, fixture4, fixture3, fixture2, fixture1})
	})
}

func normalizeSavedSearch(savedSearches ...*types.SavedSearch) {
	for _, savedSearch := range savedSearches {
		savedSearch.CreatedAt = time.Time{}
		savedSearch.UpdatedAt = time.Time{}
	}
}
