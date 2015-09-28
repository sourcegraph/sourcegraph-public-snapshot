package local

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/sourcegraph/store/mockstore"
)

// Tests that mapEmailsToUIDs performs case-insensitive lookups but
// the map keys in the returned map are the same case as those in the
// slice argument. This may not be the actual desired behavior (we may
// want to canonicalize the case to the case that's stored in the
// user_email table), but callers of mapEmailsToUIDs expect that the
// returned map keys are the same as those emails in the slice right
// now (so any different behavior could cause a nil pointer deref).
func TestMapEmailsToUIDs_caseInsensitiveAndCasePreserving(t *testing.T) {
	userByEmail := map[string]*sourcegraph.UserSpec{
		"a@A.com": &sourcegraph.UserSpec{UID: 1},
		"b@B.com": &sourcegraph.UserSpec{UID: 2},
	}
	ctx := store.WithDirectory(context.Background(), &mockstore.Directory{
		GetUserByEmail_: func(ctx context.Context, email string) (*sourcegraph.UserSpec, error) {
			if userSpec, present := userByEmail[email]; present {
				return userSpec, nil
			}
			return nil, &store.UserNotFoundError{Login: "email=" + email}
		},
	})

	email2uid, err := mapEmailsToUIDs(ctx, []string{"a@A.com", "b@B.com"})
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]int{"a@A.com": 1, "b@B.com": 2}
	if !reflect.DeepEqual(email2uid, want) {
		t.Errorf("got %v, want %v", email2uid, want)
	}
}
