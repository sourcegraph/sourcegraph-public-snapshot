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

func TestPromptsCreate(t *testing.T) {
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

	input := types.Prompt{
		Name:             "n",
		Description:      "d",
		DefinitionText:   "q",
		Draft:            true,
		Owner:            types.NamespaceUser(user.ID),
		VisibilitySecret: true,
	}
	got, err := db.Prompts().Create(ctx, &input)
	if err != nil {
		t.Fatal(err)
	}
	want := input
	want.ID = got.ID
	want.CreatedByUser = &user.ID
	want.UpdatedByUser = &user.ID
	normalizePrompt(got, &want)
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("got %+v, want %+v", *got, want)
	}
}

func TestPromptsUpdate(t *testing.T) {
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

	_, err = db.Prompts().Create(ctx, &types.Prompt{
		Name:           "n",
		DefinitionText: "q",
		Owner:          types.NamespaceUser(user.ID),
	})
	if err != nil {
		t.Fatal(err)
	}

	update := types.Prompt{
		ID:             1,
		Name:           "n2",
		DefinitionText: "q2",
	}
	got, err := db.Prompts().Update(ctx, &update)
	if err != nil {
		t.Fatal(err)
	}
	want := update
	want.Owner = types.NamespaceUser(user.ID)
	want.CreatedByUser = &user.ID
	want.UpdatedByUser = &user.ID
	normalizePrompt(got, &want)
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("got %+v, want %+v", *got, want)
	}
}

func TestPromptsUpdateOwner(t *testing.T) {
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
	org1, err := db.Orgs().Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}

	fixture1, err := db.Prompts().Create(ctx, &types.Prompt{
		Name:           "n",
		DefinitionText: "q",
		Owner:          types.NamespaceUser(user.ID),
	})
	if err != nil {
		t.Fatal(err)
	}

	{
		// Transfer from user to org1.
		newOwner := types.NamespaceOrg(org1.ID)
		updated, err := db.Prompts().UpdateOwner(ctx, fixture1.ID, newOwner)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := updated.Owner, newOwner; !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	}

	{
		// Transfer back from org1 to user.
		newOwner := types.NamespaceUser(user.ID)
		updated, err := db.Prompts().UpdateOwner(ctx, fixture1.ID, newOwner)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := updated.Owner, newOwner; !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	}
}

func TestPromptsUpdateVisibility(t *testing.T) {
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

	fixture1, err := db.Prompts().Create(ctx, &types.Prompt{
		Name:             "n",
		DefinitionText:   "q",
		Owner:            types.NamespaceUser(user.ID),
		VisibilitySecret: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Make public then secret again.
	for _, secret := range []bool{false, true} {
		updated, err := db.Prompts().UpdateVisibility(ctx, fixture1.ID, secret)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := updated.VisibilitySecret, secret; !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	}
}

func TestPromptsDelete(t *testing.T) {
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

	fixture1, err := db.Prompts().Create(ctx, &types.Prompt{
		Name:           "n",
		DefinitionText: "q",
		Owner:          types.NamespaceUser(user.ID),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("found", func(t *testing.T) {
		if err := db.Prompts().Delete(ctx, fixture1.ID); err != nil {
			t.Fatal(err)
		}
		if got, err := db.Prompts().Count(ctx, PromptListArgs{}); err != nil {
			t.Fatal(err)
		} else if got != 0 {
			t.Error()
		}
	})

	t.Run("not found", func(t *testing.T) {
		if err := db.Prompts().Delete(ctx, 123); err != errPromptNotFound {
			t.Fatal(err)
		}
	})
}

func TestPromptsGetByID(t *testing.T) {
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

	input := types.Prompt{
		Name:           "n",
		DefinitionText: "q",
		Owner:          types.NamespaceUser(user.ID),
	}
	fixture1, err := db.Prompts().Create(ctx, &input)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("found", func(t *testing.T) {
		got, err := db.Prompts().GetByID(ctx, fixture1.ID)
		if err != nil {
			t.Fatal(err)
		}
		want := input
		want.ID = got.ID
		want.NameWithOwner = "u/n"
		want.CreatedByUser = &user.ID
		want.UpdatedByUser = &user.ID
		normalizePrompt(got, &want)
		if diff := cmp.Diff(want, *got); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("not found", func(t *testing.T) {
		if _, err := db.Prompts().GetByID(ctx, 123); err != errPromptNotFound {
			t.Fatal(err)
		}
	})
}

func TestPrompts_ListCount(t *testing.T) {
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

	fixture1, err := db.Prompts().Create(ctx, &types.Prompt{
		Name:             "fixture1",
		Owner:            types.NamespaceUser(user.ID),
		VisibilitySecret: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	fixture1.NameWithOwner = "u/fixture1"

	org1, err := db.Orgs().Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}
	org2, err := db.Orgs().Create(ctx, "org2", nil)
	if err != nil {
		t.Fatal(err)
	}

	fixture2, err := db.Prompts().Create(ctx, &types.Prompt{
		Name:             "fixture2",
		Owner:            types.NamespaceOrg(org1.ID),
		VisibilitySecret: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	fixture2.NameWithOwner = "org1/fixture2"

	fixture3, err := db.Prompts().Create(ctx, &types.Prompt{
		Name:             "fixture3",
		Owner:            types.NamespaceOrg(org2.ID),
		VisibilitySecret: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	fixture3.NameWithOwner = "org2/fixture3"

	if _, err = db.OrgMembers().Create(ctx, org1.ID, user.ID); err != nil {
		t.Fatal(err)
	}

	fixture4, err := db.Prompts().Create(ctx, &types.Prompt{
		Name:             "fixture4",
		Draft:            true,
		Owner:            types.NamespaceUser(user.ID),
		VisibilitySecret: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	fixture4.NameWithOwner = "u/fixture4"

	user2, err := db.Users().Create(ctx, NewUser{Username: "u2"})
	if err != nil {
		t.Fatal(err)
	}
	fixture5, err := db.Prompts().Create(ctx, &types.Prompt{
		Name:             "fixture5",
		Owner:            types.NamespaceUser(user2.ID),
		VisibilitySecret: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	fixture5.NameWithOwner = "u2/fixture5"

	testListCount := func(t *testing.T, args PromptListArgs, pgArgs *PaginationArgs, want []*types.Prompt) {
		t.Helper()

		if pgArgs == nil {
			pgArgs = &PaginationArgs{Ascending: true}
		}
		got, err := db.Prompts().List(ctx, args, pgArgs)
		if err != nil {
			t.Fatal(err)
		}
		normalizePrompt(got...)
		normalizePrompt(want...)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}

		gotCount, err := db.Prompts().Count(ctx, args)
		if err != nil {
			t.Fatal(err)
		}
		if wantCount := len(want); gotCount != wantCount {
			t.Errorf("got count %d, want %d", gotCount, wantCount)
		}
	}

	t.Run("list all", func(t *testing.T) {
		testListCount(t, PromptListArgs{}, nil, []*types.Prompt{fixture1, fixture2, fixture3, fixture4, fixture5})
	})

	t.Run("query", func(t *testing.T) {
		testListCount(t, PromptListArgs{Query: "u/fiXTUre1"}, nil, []*types.Prompt{fixture1})
	})

	t.Run("empty result set", func(t *testing.T) {
		testListCount(t, PromptListArgs{Query: "doesntmatch"}, nil, nil)
	})

	t.Run("list owned by user", func(t *testing.T) {
		userNS := types.NamespaceUser(user.ID)
		testListCount(t, PromptListArgs{Owner: &userNS}, nil, []*types.Prompt{fixture1, fixture4})
	})

	t.Run("list owned by nonexistent user", func(t *testing.T) {
		userNS := types.NamespaceUser(1234999 /* user doesn't exist */)
		testListCount(t, PromptListArgs{Owner: &userNS}, nil, nil)
	})

	t.Run("list owned by org1", func(t *testing.T) {
		orgNS := types.NamespaceOrg(org1.ID)
		testListCount(t, PromptListArgs{Owner: &orgNS}, nil, []*types.Prompt{fixture2})
	})

	t.Run("affiliated with user", func(t *testing.T) {
		testListCount(t, PromptListArgs{AffiliatedUser: &user.ID}, nil, []*types.Prompt{fixture1, fixture2, fixture4, fixture5})
	})

	t.Run("public only", func(t *testing.T) {
		testListCount(t, PromptListArgs{PublicOnly: true}, nil, []*types.Prompt{fixture5})
	})

	t.Run("hide drafts", func(t *testing.T) {
		userNS := types.NamespaceUser(user.ID)
		testListCount(t, PromptListArgs{Owner: &userNS, HideDrafts: true}, nil, []*types.Prompt{fixture1})
	})

	t.Run("order by", func(t *testing.T) {
		orderBy, ascending := PromptsOrderByUpdatedAt.ToOptions()
		testListCount(t, PromptListArgs{}, &PaginationArgs{OrderBy: orderBy, Ascending: ascending}, []*types.Prompt{fixture5, fixture4, fixture3, fixture2, fixture1})
	})
}

func normalizePrompt(prompts ...*types.Prompt) {
	for _, prompt := range prompts {
		prompt.CreatedAt = time.Time{}
		prompt.UpdatedAt = time.Time{}
	}
}
