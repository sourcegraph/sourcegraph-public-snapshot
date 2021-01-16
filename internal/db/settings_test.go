package db

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestSettings_ListAll(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	user1, err := Users.Create(ctx, NewUser{Username: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	user2, err := Users.Create(ctx, NewUser{Username: "u2"})
	if err != nil {
		t.Fatal(err)
	}

	// Try creating both with non-nil author and nil author.
	if _, err := Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &user1.ID}, nil, &user1.ID, `{"abc": 1}`); err != nil {
		t.Fatal(err)
	}
	if _, err := Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &user2.ID}, nil, nil, `{"xyz": 2}`); err != nil {
		t.Fatal(err)
	}

	t.Run("all", func(t *testing.T) {
		settings, err := Settings.ListAll(ctx, "")
		if err != nil {
			t.Fatal(err)
		}
		if want := 2; len(settings) != want {
			t.Errorf("got %d settings, want %d", len(settings), want)
		}
	})

	t.Run("impreciseSubstring", func(t *testing.T) {
		settings, err := Settings.ListAll(ctx, "xyz")
		if err != nil {
			t.Fatal(err)
		}
		if want := 1; len(settings) != want {
			t.Errorf("got %d settings, want %d", len(settings), want)
		}
		if want := `{"xyz": 2}`; settings[0].Contents != want {
			t.Errorf("got contents %q, want %q", settings[0].Contents, want)
		}
	})
}

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
