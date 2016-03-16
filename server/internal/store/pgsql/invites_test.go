// +build pgsqltest

package pgsql

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/randstring"
)

func TestInvites(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	s := &invites{}
	invite := &sourcegraph.AccountInvite{Email: "u@d.com", Write: true}

	token, err := s.CreateOrUpdate(ctx, invite)
	if err != nil {
		t.Fatal(err)
	}

	// Valid token must succeed.
	got, err := s.Retrieve(ctx, token)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, invite) {
		t.Errorf("Invite: got %+v, want %+v", got, invite)
	}

	// Invalid token must fail.
	_, err = s.Retrieve(ctx, randstring.NewLen(20))
	if err == nil {
		t.Errorf("expected error with invalid token, got nil")
	}

	// Second access for same token must fail.
	_, err = s.Retrieve(ctx, token)
	if err == nil {
		t.Errorf("expected error with duplicate access, got nil")
	}

	// MarkUnused must succeed.
	err = s.MarkUnused(ctx, token)
	if err != nil {
		t.Fatal(err)
	}

	// Unused token fetch must succeed.
	got, err = s.Retrieve(ctx, token)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, invite) {
		t.Errorf("Invite: got %+v, want %+v", got, invite)
	}

	// List must succeed.
	list, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []*sourcegraph.AccountInvite{invite}
	if !reflect.DeepEqual(list, want) {
		t.Errorf("InviteList: got %+v, want %+v", got, want)
	}

	// Delete must succeed.
	err = s.Delete(ctx, token)
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve must fail after delete.
	_, err = s.Retrieve(ctx, token)
	if err == nil {
		t.Errorf("expected error with invalid access, got nil")
	}

	// Recreate to test DeleteByEmail.
	token, err = s.CreateOrUpdate(ctx, &sourcegraph.AccountInvite{Email: "u@d.com", Write: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteByEmail(ctx, "u@d.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Retrieve(ctx, token); err == nil {
		t.Errorf("expected error retrieving deleted token, got nil")
	}
	if err := s.DeleteByEmail(ctx, "u@d.com"); err == nil {
		t.Errorf("expected error deleting already deleted token, got nil")
	}

	// Test that inviting the same person twice sends the same token both times
	token, err = s.CreateOrUpdate(ctx, &sourcegraph.AccountInvite{Email: "n@b.com", Write: true, Admin: true})
	if err != nil {
		t.Fatal(err)
	}

	token1, err := s.CreateOrUpdate(ctx, &sourcegraph.AccountInvite{Email: "n@b.com", Write: false, Admin: false})
	if err != nil {
		t.Fatal(err)
	}

	if token != token1 {
		t.Errorf("expected same token to be returned when re-inviting the same email")
	}

	// Test that the second invite updates the invite permissions
	got1, err := s.get(ctx, token1)
	if got1.Admin != false || got1.Write != false {
		t.Errorf("expected second update to update write/admin permissions")
	}
}
