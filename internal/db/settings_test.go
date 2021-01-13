package db

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestCreateIfUpToDate(t *testing.T) {
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()
	u, err := Users.Create(ctx, NewUser{Username: "test"})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("quicklink with safe link", func(t *testing.T) {
		contents := "{\"quicklinks\": [{\"name\": \"malicious link test\",      \"url\": \"https://example.com\"}]}"

		_, err := Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &u.ID}, nil, nil, contents)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("quicklink with javascript link", func(t *testing.T) {
		contents := "{\"quicklinks\": [{\"name\": \"malicious link test\",      \"url\": \"javascript:alert(1)\"}]}"

		want := "invalid settings: quicklinks.0.url: Does not match pattern '^(https?://|/)'"

		_, err := Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &u.ID}, nil, nil, contents)
		if err == nil {
			t.Log("Expected an error")
			t.Fail()
		} else {
			got := err.Error()
			if got != want {
				t.Errorf("err: want %q but got %q", want, got)
			}
		}
	})
}
