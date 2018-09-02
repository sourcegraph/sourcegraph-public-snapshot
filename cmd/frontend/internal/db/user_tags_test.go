package db

import (
	"reflect"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

func TestUsers_SetTag(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := testContext()

	// Create user.
	u, err := Users.Create(ctx, NewUser{Username: "u"})
	if err != nil {
		t.Fatal(err)
	}

	checkTags := func(t *testing.T, userID int32, wantTags []string) {
		t.Helper()
		u, err := Users.GetByID(ctx, userID)
		if err != nil {
			t.Fatal(err)
		}
		sort.Strings(u.Tags)
		sort.Strings(wantTags)
		if !reflect.DeepEqual(u.Tags, wantTags) {
			t.Errorf("got tags %v, want %v", u.Tags, wantTags)
		}
	}

	t.Run("fails on nonexistent user", func(t *testing.T) {
		if err := Users.SetTag(ctx, 1234 /* doesn't exist */, "t", true); !errcode.IsNotFound(err) {
			t.Errorf("got err %v, want errcode.IsNotFound", err)
		}
		if err := Users.SetTag(ctx, 1234 /* doesn't exist */, "t", false); !errcode.IsNotFound(err) {
			t.Errorf("got err %v, want errcode.IsNotFound", err)
		}
	})

	t.Run("tags begins empty", func(t *testing.T) {
		checkTags(t, u.ID, []string{})
	})

	t.Run("adds and removes tag", func(t *testing.T) {
		if err := Users.SetTag(ctx, u.ID, "t1", true); err != nil {
			t.Fatal(err)
		}
		checkTags(t, u.ID, []string{"t1"})

		t.Run("deduplicates", func(t *testing.T) {
			if err := Users.SetTag(ctx, u.ID, "t1", true); err != nil {
				t.Fatal(err)
			}
			checkTags(t, u.ID, []string{"t1"})
		})

		if err := Users.SetTag(ctx, u.ID, "t2", true); err != nil {
			t.Fatal(err)
		}
		checkTags(t, u.ID, []string{"t1", "t2"})

		if err := Users.SetTag(ctx, u.ID, "t1", false); err != nil {
			t.Fatal(err)
		}
		checkTags(t, u.ID, []string{"t2"})

		t.Run("removing nonexistent tag is noop", func(t *testing.T) {
			if err := Users.SetTag(ctx, u.ID, "t1", false); err != nil {
				t.Fatal(err)
			}
			checkTags(t, u.ID, []string{"t2"})
		})

		if err := Users.SetTag(ctx, u.ID, "t2", false); err != nil {
			t.Fatal(err)
		}
		checkTags(t, u.ID, []string{})
	})
}
