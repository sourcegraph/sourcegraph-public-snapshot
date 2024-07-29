package main

import (
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/dev/gqltest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestAccessToken(t *testing.T) {
	t.Run("create a token and test it", func(t *testing.T) {
		token, err := gqltest.Client.CreateAccessToken("TestAccessToken", []string{"user:all"}, pointers.Ptr(3600))
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := gqltest.Client.DeleteAccessToken(token)
			if err != nil {
				t.Fatal(err)
			}
		}()

		_, err = gqltest.Client.CurrentUserID(token)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("use an invalid token gets 401", func(t *testing.T) {
		_, err := gqltest.Client.CurrentUserID("a bad token")
		gotErr := fmt.Sprintf("%v", errors.Cause(err))
		wantErr := "401: Invalid access token."
		if gotErr != wantErr {
			t.Fatalf("err: want %q but got %q", wantErr, gotErr)
		}
	})
}
