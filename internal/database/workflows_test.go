package database

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestWorkflowsCreate(t *testing.T) {
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

	input := types.Workflow{
		Name:         "n",
		Description:  "d",
		TemplateText: "q",
		Draft:        true,
		Owner:        types.NamespaceUser(user.ID),
	}
	got, err := db.Workflows().Create(ctx, &input, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	want := input
	want.ID = got.ID
	want.CreatedByUser = &user.ID
	want.UpdatedByUser = &user.ID
	normalizeWorkflow(got, &want)
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("got %+v, want %+v", *got, want)
	}
}

func TestWorkflowsUpdate(t *testing.T) {
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

	_, err = db.Workflows().Create(ctx, &types.Workflow{
		Name:         "n",
		TemplateText: "q",
		Owner:        types.NamespaceUser(user.ID),
	}, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	update := types.Workflow{
		ID:           1,
		Name:         "n2",
		TemplateText: "q2",
	}
	got, err := db.Workflows().Update(ctx, &update, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	want := update
	want.Owner = types.NamespaceUser(user.ID)
	want.CreatedByUser = &user.ID
	want.UpdatedByUser = &user.ID
	normalizeWorkflow(got, &want)
	if !reflect.DeepEqual(*got, want) {
		t.Errorf("got %+v, want %+v", *got, want)
	}
}

func TestWorkflowsUpdateOwner(t *testing.T) {
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

	fixture1, err := db.Workflows().Create(ctx, &types.Workflow{
		Name:         "n",
		TemplateText: "q",
		Owner:        types.NamespaceUser(user.ID),
	}, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	{
		// Transfer from user to org1.
		newOwner := types.NamespaceOrg(org1.ID)
		updated, err := db.Workflows().UpdateOwner(ctx, fixture1.ID, newOwner, user.ID)
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
		updated, err := db.Workflows().UpdateOwner(ctx, fixture1.ID, newOwner, user.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := updated.Owner, newOwner; !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	}
}

func TestWorkflowsDelete(t *testing.T) {
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

	fixture1, err := db.Workflows().Create(ctx, &types.Workflow{
		Name:         "n",
		TemplateText: "q",
		Owner:        types.NamespaceUser(user.ID),
	}, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("found", func(t *testing.T) {
		if err := db.Workflows().Delete(ctx, fixture1.ID); err != nil {
			t.Fatal(err)
		}
		if got, err := db.Workflows().Count(ctx, WorkflowListArgs{}); err != nil {
			t.Fatal(err)
		} else if got != 0 {
			t.Error()
		}
	})

	t.Run("not found", func(t *testing.T) {
		if err := db.Workflows().Delete(ctx, 123); err != errWorkflowNotFound {
			t.Fatal(err)
		}
	})
}

func TestWorkflowsGetByID(t *testing.T) {
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

	input := types.Workflow{
		Name:         "n",
		TemplateText: "q",
		Owner:        types.NamespaceUser(user.ID),
	}
	fixture1, err := db.Workflows().Create(ctx, &input, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("found", func(t *testing.T) {
		got, err := db.Workflows().GetByID(ctx, fixture1.ID)
		if err != nil {
			t.Fatal(err)
		}
		want := input
		want.ID = got.ID
		want.NameWithOwner = "u/n"
		want.CreatedByUser = &user.ID
		want.UpdatedByUser = &user.ID
		normalizeWorkflow(got, &want)
		if diff := cmp.Diff(want, *got); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("not found", func(t *testing.T) {
		if _, err := db.Workflows().GetByID(ctx, 123); err != errWorkflowNotFound {
			t.Fatal(err)
		}
	})
}

func TestWorkflows_ListCount(t *testing.T) {
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

	fixture1, err := db.Workflows().Create(ctx, &types.Workflow{
		Name:  "fixture1",
		Owner: types.NamespaceUser(user.ID),
	}, user.ID)
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

	fixture2, err := db.Workflows().Create(ctx, &types.Workflow{
		Name:  "fixture2",
		Owner: types.NamespaceOrg(org1.ID),
	}, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	fixture2.NameWithOwner = "org1/fixture2"

	fixture3, err := db.Workflows().Create(ctx, &types.Workflow{
		Name:  "fixture3",
		Owner: types.NamespaceOrg(org2.ID),
	}, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	fixture3.NameWithOwner = "org2/fixture3"

	if _, err = db.OrgMembers().Create(ctx, org1.ID, user.ID); err != nil {
		t.Fatal(err)
	}

	fixture4, err := db.Workflows().Create(ctx, &types.Workflow{
		Name:  "fixture4",
		Draft: true,
		Owner: types.NamespaceUser(user.ID),
	}, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	fixture4.NameWithOwner = "u/fixture4"

	testListCount := func(t *testing.T, args WorkflowListArgs, pgArgs *PaginationArgs, want []*types.Workflow) {
		t.Helper()

		if pgArgs == nil {
			pgArgs = &PaginationArgs{Ascending: true}
		}
		got, err := db.Workflows().List(ctx, args, pgArgs)
		if err != nil {
			t.Fatal(err)
		}
		normalizeWorkflow(got...)
		normalizeWorkflow(want...)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("Mismatch (-want +got):\n%s", diff)
		}

		gotCount, err := db.Workflows().Count(ctx, args)
		if err != nil {
			t.Fatal(err)
		}
		if wantCount := len(want); gotCount != wantCount {
			t.Errorf("got count %d, want %d", gotCount, wantCount)
		}
	}

	t.Run("list all", func(t *testing.T) {
		testListCount(t, WorkflowListArgs{}, nil, []*types.Workflow{fixture1, fixture2, fixture3, fixture4})
	})

	t.Run("query", func(t *testing.T) {
		testListCount(t, WorkflowListArgs{Query: "u/fiXTUre1"}, nil, []*types.Workflow{fixture1})
	})

	t.Run("empty result set", func(t *testing.T) {
		testListCount(t, WorkflowListArgs{Query: "doesntmatch"}, nil, nil)
	})

	t.Run("list owned by user", func(t *testing.T) {
		userNS := types.NamespaceUser(user.ID)
		testListCount(t, WorkflowListArgs{Owner: &userNS}, nil, []*types.Workflow{fixture1, fixture4})
	})

	t.Run("list owned by nonexistent user", func(t *testing.T) {
		userNS := types.NamespaceUser(1234999 /* user doesn't exist */)
		testListCount(t, WorkflowListArgs{Owner: &userNS}, nil, nil)
	})

	t.Run("list owned by org1", func(t *testing.T) {
		orgNS := types.NamespaceOrg(org1.ID)
		testListCount(t, WorkflowListArgs{Owner: &orgNS}, nil, []*types.Workflow{fixture2})
	})

	t.Run("affiliated with user", func(t *testing.T) {
		testListCount(t, WorkflowListArgs{AffiliatedUser: &user.ID}, nil, []*types.Workflow{fixture1, fixture2, fixture4})
	})

	t.Run("hide drafts", func(t *testing.T) {
		userNS := types.NamespaceUser(user.ID)
		testListCount(t, WorkflowListArgs{Owner: &userNS, HideDrafts: true}, nil, []*types.Workflow{fixture1})
	})

	t.Run("order by", func(t *testing.T) {
		orderBy, ascending := WorkflowsOrderByUpdatedAt.ToOptions()
		testListCount(t, WorkflowListArgs{}, &PaginationArgs{OrderBy: orderBy, Ascending: ascending}, []*types.Workflow{fixture4, fixture3, fixture2, fixture1})
	})
}

func normalizeWorkflow(workflows ...*types.Workflow) {
	for _, workflow := range workflows {
		workflow.CreatedAt = time.Time{}
		workflow.UpdatedAt = time.Time{}
	}
}
