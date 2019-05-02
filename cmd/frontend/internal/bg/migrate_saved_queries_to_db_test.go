package bg

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestMigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)

	t.Run("valid", func(t *testing.T) {
		u, err := db.Users.Create(ctx, db.NewUser{Username: "u4"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &u.ID}, nil, nil, `{"search.savedQueries": [{"key": "1a2b3c", "description": "test query", "query": "test type:diff"}]}`); err != nil {
			t.Fatal(err)
		}
		if err := doMigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase(ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("no saved queries", func(t *testing.T) {
		u, err := db.Users.Create(ctx, db.NewUser{Username: "u5"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &u.ID}, nil, nil, `{}`); err != nil {
			t.Fatal(err)
		}
		if err := doMigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase(ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		u, err := db.Users.Create(ctx, db.NewUser{Username: "u6"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &u.ID}, nil, nil, `{"search.savedQueries": [1]}`); err != nil {
			t.Fatal(err)
		}
		err = doMigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase(ctx)
		if err == nil {
			t.Error("want error due to invalid settings json")
		}
		t.Logf(err.Error())
	})
}
func TestMigrateSavedQueryIntoDB(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)

	check := func(t *testing.T, settings *api.Settings, wantChanged bool) {
		t.Helper()
		if changed, err := migrateSavedQueryIntoDB(ctx, settings); err != nil {
			t.Fatal(err)
		} else if changed != wantChanged {
			t.Errorf("got changed %v, want %v", changed, wantChanged)
		}
	}

	t.Run("user needing migration", func(t *testing.T) {
		u, err := db.Users.Create(ctx, db.NewUser{Username: "u4"})
		if err != nil {
			t.Fatal(err)
		}
		settings, err := db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &u.ID}, nil, nil, `{"search.savedQueries": [{"key": "1a2b3c", "description": "test query", "query": "test type:diff"}]}`)
		if err != nil {
			t.Fatal(err)
		}
		check(t, settings, true)
	})

	t.Run("user not needing migration", func(t *testing.T) {
		u, err := db.Users.Create(ctx, db.NewUser{Username: "u5"})
		if err != nil {
			t.Fatal(err)
		}
		settings, err := db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &u.ID}, nil, nil, `{"other": []}`)
		if err != nil {
			t.Fatal(err)
		}
		check(t, settings, false)
	})

	t.Run("org needing migration", func(t *testing.T) {
		org, err := db.Orgs.Create(ctx, "myorg", nil)
		if err != nil {
			t.Fatal(err)
		}
		settings, err := db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{Org: &org.ID}, nil, nil, `{"search.savedQueries": [{"key": "1a2b3c", "description": "test query", "query": "test type:diff"}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if err := doMigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase(ctx); err != nil {
			t.Fatal(err)
		}
		check(t, settings, true)
	})

	t.Run("global settings needing migration", func(t *testing.T) {
		settings, err := db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{}, nil, nil, `{"search.savedQueries": [{"key": "1a2b3c", "description": "test query", "query": "test type:diff"}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if err := doMigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase(ctx); err != nil {
			t.Fatal(err)
		}
		check(t, settings, true)
	})
}
