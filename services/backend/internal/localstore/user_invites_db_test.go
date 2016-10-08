package localstore

import (
	"reflect"
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
)

/*
 * Helpers
 */

func userInviteURIs(invites []*sourcegraph.UserInvite) []string {
	var uris []string
	for _, invite := range invites {
		uris = append(uris, invite.URI)
	}
	return uris
}

func (s *userInvites) mustCreate(ctx context.Context, t *testing.T, invites ...*sourcegraph.UserInvite) []*sourcegraph.UserInvite {
	var createdInvites []*sourcegraph.UserInvite
	for _, invite := range invites {
		if err := s.Create(ctx, invite); err != nil {
			t.Fatal(err)
		}
		invite, err := s.GetByURI(ctx, invite.URI)
		if err != nil {
			t.Fatal(err)
		}
		createdInvites = append(createdInvites, invite)
	}
	return createdInvites
}

/*
 * Tests
 */

func TestUserInvites_Get(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx, done := testContext()
	defer done()

	s := userInvites{}

	want := s.mustCreate(ctx, t, &sourcegraph.UserInvite{URI: "i", UserID: "test", OrgID: "123", OrgName: "sgtest123"})

	invite, err := s.GetByURI(ctx, want[0].URI)
	if err != nil {
		t.Fatal(err)
	}
	if !jsonEqual(t, invite, want[0]) {
		t.Errorf("got %v, want %v", invite, want[0])
	}
}

func TestUserInvites_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx, done := testContext()
	defer done()

	ctx = github.WithMockHasAuthedUser(ctx, false)

	s := userInvites{}

	want := s.mustCreate(ctx, t, &sourcegraph.UserInvite{URI: "j", UserID: "test", OrgID: "1234", OrgName: "sgtest1234"})

	invites, err := s.List(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !jsonEqual(t, invites, want) {
		t.Errorf("got %v, want %v", invites, want)
	}
}

// TestUserInvites_List_URIs tests the behavior of UserInvites.List when called with
// URIs.
func TestUserInvites_List_URIs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx, done := testContext()
	defer done()

	ctx = github.WithMockHasAuthedUser(ctx, false)

	s := userInvites{}

	// Add some invites.
	if err := s.Create(ctx, &sourcegraph.UserInvite{URI: "k", UserID: "test", OrgID: "12345", OrgName: "sgtest12345"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(ctx, &sourcegraph.UserInvite{URI: "l", UserID: "test", OrgID: "123456", OrgName: "sgtest123456"}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		uris []string
		want []string
	}{
		{[]string{"k"}, []string{"k"}},
		{[]string{"xy"}, nil},
		{[]string{"k", "l"}, []string{"k", "l"}},
		{[]string{"k", "xy", "l"}, []string{"k", "l"}},
	}
	for _, test := range tests {
		invites, err := s.List(ctx, &UserInviteListOp{URIs: test.uris})
		if err != nil {
			t.Fatal(err)
		}
		if got := userInviteURIs(invites); !reflect.DeepEqual(got, test.want) {
			t.Errorf("%v: got invites %q, want %q", test.uris, got, test.want)
		}
	}
}
