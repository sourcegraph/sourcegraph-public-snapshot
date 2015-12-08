package testsuite

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/randstring"
)

func Invites_test(ctx context.Context, t *testing.T, s store.Invites) {
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
}
