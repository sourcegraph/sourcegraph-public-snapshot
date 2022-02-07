package main

import (
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestAccessToken(t *testing.T) {
	t.Run("create a token and test it", func(t *testing.T) {
		token, err := client.CreateAccessToken("TestAccessToken", []string{"user:all"})
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := client.DeleteAccessToken(token)
			if err != nil {
				t.Fatal(err)
			}
		}()

		_, err = client.CurrentUserID(token)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("use an invalid token gets 401", func(t *testing.T) {
		_, err := client.CurrentUserID("a bad token")
		gotErr := fmt.Sprintf("%v", errors.Cause(err))
		wantErr := "401: Invalid access token."
		if gotErr != wantErr {
			t.Fatalf("err: want %q but got %q", wantErr, gotErr)
		}
	})
}
