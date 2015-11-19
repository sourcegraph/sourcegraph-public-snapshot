// +build pgsqltest

package pgsql

import (
	"reflect"
	"testing"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestUsers_ListEmails(t *testing.T) {
	t.Parallel()

	var as Accounts
	var us Users
	ctx, done := testContext()
	defer done()

	userSpec := sourcegraph.UserSpec{UID: 1}
	insertEmails := []*sourcegraph.EmailAddr{
		{Email: "a@a.com", Primary: true, Verified: true},
	}
	if err := as.UpdateEmails(ctx, userSpec, insertEmails); err != nil {
		t.Fatal(err)
	}

	emailAddrs, err := us.ListEmails(ctx, userSpec)
	if err != nil {
		t.Fatal(err)
	}
	if want := insertEmails; !reflect.DeepEqual(emailAddrs, want) {
		t.Errorf("got emailAddrs %+v, want %+v", emailAddrs, want)
	}
}
