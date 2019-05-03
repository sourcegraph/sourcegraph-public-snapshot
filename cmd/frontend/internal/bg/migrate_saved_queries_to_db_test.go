package bg

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func TestMigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)

	t.Run("valid", func(t *testing.T) {
		u, err := db.Users.Create(ctx, db.NewUser{Username: "u1"})
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
		u, err := db.Users.Create(ctx, db.NewUser{Username: "u2"})
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

	t.Run("do not throw error on an invalid settings JSON", func(t *testing.T) {
		u, err := db.Users.Create(ctx, db.NewUser{Username: "u3"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &u.ID}, nil, nil, `{"search.savedQueries": [1]}`); err != nil {
			t.Fatal(err)
		}
		err = doMigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase(ctx)
		if err != nil {
			t.Error(err, "there should not be an error for an invalid settings json")
		}
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
		u, err := db.Users.Create(ctx, db.NewUser{Username: "u1"})
		if err != nil {
			t.Fatal(err)
		}
		settings, err := db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &u.ID}, nil, nil, `{"search.savedQueries": [{"key": "1a2b3c", "description": "test query", "query": "test type:diff"}]}`)
		if err != nil {
			t.Fatal(err)
		}
		check(t, settings, true)
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

func TestInsertSavedQueryIntoDB(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)

	t.Run("user saved search inserted into db table", func(t *testing.T) {
		u, err := db.Users.Create(ctx, db.NewUser{Username: "u1"})
		if err != nil {
			t.Fatal(err)
		}
		settings, err := db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &u.ID}, nil, nil, `{"search.savedQueries": [{"key": "1a2b3c", "description": "test query", "query": "test type:diff"}]}`)
		if err := insertSavedQueryIntoDB(ctx, settings, &SavedQueryField{SavedQueries: []SavedQuery{SavedQuery{Key: "1a2b3c", Description: "test query", Query: "test type:diff"}}}); err != nil {
			t.Fatal(err)
		}
		ss, err := db.SavedSearches.ListSavedSearchesByUserID(ctx, u.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(ss) == 0 {
			t.Fatal("Error: user saved query not created")
		}
		got := ss[0]
		want := &types.SavedSearch{ID: "1", Description: "test query", Query: "test type:diff", Notify: false, NotifySlack: false, OwnerKind: "user", UserID: &u.ID, OrgID: nil, SlackWebhookURL: nil}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("org saved search inserted into db table", func(t *testing.T) {
		org, err := db.Orgs.Create(ctx, "myorg", nil)
		if err != nil {
			t.Fatal(err)
		}
		settings, err := db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{Org: &org.ID}, nil, nil, `{"search.savedQueries": [{"key": "1a2b3c", "description": "test query", "query": "test type:diff"}]}`)
		if err := insertSavedQueryIntoDB(ctx, settings, &SavedQueryField{SavedQueries: []SavedQuery{SavedQuery{Key: "1a2b3c", Description: "test query", Query: "test type:diff"}}}); err != nil {
			t.Fatal(err)
		}
		ss, err := db.SavedSearches.ListSavedSearchesByOrgID(ctx, org.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(ss) == 0 {
			t.Fatal("Error: organization saved query not created")
		}
		got := ss[0]
		want := &types.SavedSearch{ID: "2", Description: "test query", Query: "test type:diff", Notify: false, NotifySlack: false, OwnerKind: "org", UserID: nil, OrgID: &org.ID, SlackWebhookURL: nil}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})

	t.Run("global saved search inserted into db table under site admin user", func(t *testing.T) {
		// The site-admin is the first user we created in this DB, and will have an ID of 1.
		user, err := db.Users.GetByID(ctx, 1)
		settings, err := db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &user.ID}, nil, nil, `{"search.savedQueries": [{"key": "1a2b3c", "description": "test query", "query": "test type:diff"}]}`)
		if err := insertSavedQueryIntoDB(ctx, settings, &SavedQueryField{SavedQueries: []SavedQuery{SavedQuery{Key: "1a2b3c", Description: "test query", Query: "test type:diff"}}}); err != nil {
			t.Fatal(err)
		}
		ss, err := db.SavedSearches.ListSavedSearchesByUserID(ctx, user.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(ss) == 0 {
			t.Fatal("Error: global saved query not created")
		}
		got := ss[0]
		want := &types.SavedSearch{ID: "1", Description: "test query", Query: "test type:diff", Notify: false, NotifySlack: false, OwnerKind: "user", UserID: &user.ID, OrgID: nil, SlackWebhookURL: nil}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		t.Log(db.SavedSearches.ListSavedSearchesByUserID(ctx, user.ID))
	})
}

func TestMigrateSlackWebhookURL(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := dbtesting.TestContext(t)

	t.Run("migrate user slack webhook URL", func(t *testing.T) {
		u, err := db.Users.Create(ctx, db.NewUser{Username: "u1"})
		if err != nil {
			t.Fatal(err)
		}

		// Migrate saved query without Slack webhook URL into DB.
		latest, err := db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &u.ID}, nil, nil, `{"search.savedQueries": [{"key": "1a2b3c", "description": "test query", "query": "test type:diff"}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if err := doMigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase(ctx); err != nil {
			t.Fatal(err)
		}

		// Create settings with notifications.slack for the same user.
		_, err = db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &u.ID}, &latest.ID, nil, `{"notifications.slack": {"webhookURL": "https://test.slackwebhook.com"}}`)
		if err != nil {
			t.Fatal(err)
		}

		err = MigrateSlackWebhookUrlsToSavedSearches(ctx)
		if err != nil {
			t.Fatal(err)
		}

		ss, err := db.SavedSearches.ListSavedSearchesByUserID(ctx, u.ID)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range ss {
			t.Logf("%v+", s)
		}

		got := ss[0]
		webhookURL := "https://test.slackwebhook.com"
		want := &types.SavedSearch{ID: "1", Description: "test query", Query: "test type:diff", Notify: false, NotifySlack: false, OwnerKind: "user", UserID: &u.ID, OrgID: nil, SlackWebhookURL: &webhookURL}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("migrate org slack webhook URL", func(t *testing.T) {
		org, err := db.Orgs.Create(ctx, "myorg", nil)
		if err != nil {
			t.Fatal(err)
		}

		// Migrate saved query without Slack webhook URL into DB.
		latest, err := db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{Org: &org.ID}, nil, nil, `{"search.savedQueries": [{"key": "1a2b3c", "description": "test query", "query": "test type:diff"}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if err := doMigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase(ctx); err != nil {
			t.Fatal(err)
		}

		// Create settings with notifications.slack for the same org.
		_, err = db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{Org: &org.ID}, &latest.ID, nil, `{"notifications.slack": {"webhookURL": "https://test.slackwebhook.com"}}`)
		if err != nil {
			t.Fatal(err)
		}

		err = MigrateSlackWebhookUrlsToSavedSearches(ctx)
		if err != nil {
			t.Fatal(err)
		}

		ss, err := db.SavedSearches.ListSavedSearchesByOrgID(ctx, org.ID)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range ss {
			t.Logf("%v+", s)
		}

		got := ss[0]
		webhookURL := "https://test.slackwebhook.com"
		want := &types.SavedSearch{ID: "2", Description: "test query", Query: "test type:diff", Notify: false, NotifySlack: false, OwnerKind: "org", UserID: nil, OrgID: &org.ID, SlackWebhookURL: &webhookURL}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("migrate global slack webhook URL", func(t *testing.T) {
		// Get the site admin user.
		user, err := db.Users.GetByUsername(ctx, "u1")
		if err != nil {
			t.Fatal(err)
		}

		// Create global settings with notifications.slack.
		_, err = db.Settings.CreateIfUpToDate(ctx, api.SettingsSubject{}, nil, nil, `{"notifications.slack": {"webhookURL": "https://test.slackwebhook.com"}}`)
		if err != nil {
			t.Fatal(err)
		}

		err = MigrateSlackWebhookUrlsToSavedSearches(ctx)
		if err != nil {
			t.Fatal(err)
		}

		ss, err := db.SavedSearches.ListSavedSearchesByUserID(ctx, user.ID)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range ss {
			t.Logf("%v+", s)
		}

		got := ss[0]
		webhookURL := "https://test.slackwebhook.com"
		want := &types.SavedSearch{ID: "1", Description: "test query", Query: "test type:diff", Notify: false, NotifySlack: false, OwnerKind: "user", UserID: &user.ID, OrgID: nil, SlackWebhookURL: &webhookURL}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

}
