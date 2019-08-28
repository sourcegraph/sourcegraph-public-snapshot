package graphqlbackend

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestUpdateUserPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)

	_, err := db.Users.Create(ctx, db.NewUser{Username: "john"})
	if err != nil {
		t.Fatal(err)
	}

	_, err := db.Users.Create(ctx, db.NewUser{Username: "john"})
	if err != nil {
		t.Fatal(err)
	}

	_, err := db.Users.Create(ctx, db.NewUser{Username: "john"})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name  string
		input *UpdateUserPermissionsInput
		resp  *EmptyResponse
		err   string
	}{
		{
			name:  "username not found",
			input: &UpdateUserPermissionsInput{},
			err:   "user not found: []",
		},
		{
			name:  "us",
			input: &UpdateUserPermissionsInput{Username: "},
			err:   "user not found: []",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := (&schemaResolver{}).UpdateUserPermissions(
				ctx,
				&UpdateUserPermissionsArgs{Input: tc.input},
			)

			if have, want := resp, tc.resp; !reflect.DeepEqual(have, want) {
				t.Errorf("resp:\nhave %+v\nwant %+v", have, want)
			}

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave %q\nwant %q", have, want)
			}
		})
	}
	// t.Run("not found", nil)
	// t.Run("unauthorized", nil)
	// t.Run("no providers", nil)
	// t.Run("updated", nil)
}

type fakeProvider struct {
	updateError error
}

func (p fakeProvider) UpdatePermissions(_ context.Context, u *types.User) error {
	return p.updateError
}
