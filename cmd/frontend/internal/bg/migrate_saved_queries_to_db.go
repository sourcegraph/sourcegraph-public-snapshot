package bg

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

type SavedQuery struct {
	Key         string
	Description string
	Query       string
	Notify      bool
	NotifySlack bool
}

type SavedQueryField struct {
	SavedQueries []SavedQuery `json:"search.savedQueries"`
}

// MigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase migrates saved searches from the `search.SavedQueries` value in
// site settings into the `saved_searches` PostgreSQL table, and migrates organization Slack webhook URLs from the
// `notifications.slack` value in site settings to the `slack_webhook_url` column in the `saved_searches` table.
func MigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase(ctx context.Context) {
	if err := doMigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase(ctx); err != nil {
		log15.Error(`Warning: unable to migrate "search.savedQueries" settings to database. Please report this issue. The search.savedQueries setting has been deprecated in favor of a database table, and is no longer functional.`, "error", err)
	}
}

func doMigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase(ctx context.Context) error {
rerun:
	// List all settings that include a `search.savedQueries` field.
	settings, err := db.Settings.ListAll(ctx, "search.savedQueries")
	if err != nil {
		// Don't continue the migration if we can't list settings.
		return errors.WithMessagef(err, `Warning: unable to migrate saved queries to database. Error listing settings. Please report this issue`)
	}
	count := 0
	for _, s := range settings {
		migrated, err := migrateSavedQueryIntoDB(ctx, s)
		if err != nil {
			// Do we want to fail completely if one user's json is invalid, for example?
			return errors.WithMessagef(err, "for settings subject %s", s.Subject)
		}
		if migrated {
			count++
		}
	}

	// To reduce the (small) chance of a race condition whereby a new settings document can have an
	// "search.savedQueries" added without reporting a validation error (e.g., by an older deployed version while
	// this migration is running), rerun until we have nothing else to do.
	if count > 0 {
		log15.Info(`Migrated settings "search.savedQueries" to saved_searches table.`, "count", count)
		time.Sleep(60 * time.Second)
		goto rerun
	}

	// After migrating all saved searches into the DB, migrate organization Slack webhook URLs into the
	// slack_webhook_url column in the saved_searches table.
	MigrateSlackWebhookUrlsToSavedSearches(ctx)
	return nil
}

func migrateSavedQueryIntoDB(ctx context.Context, s *api.Settings) (bool, error) {
	var sq SavedQueryField
	err := jsonc.Unmarshal(s.Contents, &sq)
	if err != nil {
		return false, errors.WithMessagef(err, `Warning: unable to migrate saved queries from settings subject %s. Unable to unmarshal JSON value. Please report this issue.`, s.Subject)
	}
	err = insertSavedQueryIntoDB(ctx, s, &sq)
	if err != nil {
		return false, err
	}

	return true, nil
}

// getFirstSiteAdminID gets the database ID of the first site admin available in the users table.
func getFirstSiteAdminID(ctx context.Context) *int32 {
	var siteAdminID *int32
	err := dbconn.Global.QueryRowContext(ctx, "WITH site_admins AS (SELECT id, username FROM users WHERE site_admin=true) SELECT min(id) FROM site_admins;").Scan(&siteAdminID)
	if err != nil {
		log15.Error(`Warning: unable to migrate saved query into database. No site admin ID found.`, err.Error())
	}
	return siteAdminID
}

// insertSavedQueryIntoDB inserts an existing saved query from site settings into the saved_searches database table.
// Global saved queries will be associated with the first site admin's profile. It will be added with `owner_kind`
// set to "user", and with UserID set to the first site admin's user ID.
func insertSavedQueryIntoDB(ctx context.Context, s *api.Settings, sq *SavedQueryField) error {
	for _, query := range sq.SavedQueries {
		// Add case for global settings. It should make a site admin user the owner of that saved search.
		if s.Subject.User != nil {
			_, err := db.SavedSearches.Create(ctx, &types.SavedSearch{Description: query.Description, Query: query.Query, Notify: query.Notify, NotifySlack: query.NotifySlack, OwnerKind: "user", UserID: s.Subject.User})
			if err != nil {
				return errors.WithMessagef(err, `Warning: unable to insert user saved query into database.`)
			}
		} else if s.Subject.Org != nil {
			_, err := db.SavedSearches.Create(ctx, &types.SavedSearch{Description: query.Description, Query: query.Query, Notify: query.Notify, NotifySlack: query.NotifySlack, OwnerKind: "org", OrgID: s.Subject.Org})
			if err != nil {
				return errors.WithMessagef(err, `Warning: unable to migrate org saved query into database.`)
			}
		} else if s.Subject.Site || s.Subject.Default {
			siteAdminID := getFirstSiteAdminID(ctx)
			_, err := db.SavedSearches.Create(ctx, &types.SavedSearch{Description: query.Description, Query: query.Query, Notify: query.Notify, NotifySlack: query.NotifySlack, OwnerKind: "user", UserID: siteAdminID})
			if err != nil {
				return errors.WithMessagef(err, `Warning: unable to migrate global saved query into database.`)
			}
		}
	}

	return nil
}

type NotificationsSlackField struct {
	NotificationsSlack WebhookURL `json:"notifications.slack"`
}

type WebhookURL struct {
	WebhookURL string `json:"webhookURL"`
}

// MigrateSlackWebhookUrlsToSavedSearches migrates Slack webhook URLs from site settings into the
// slack_webhook_url column of the saved_searches database table. As a result, Slack webhook URLs
// will be associated with individual saved searches rather than user or organization profiles.
func MigrateSlackWebhookUrlsToSavedSearches(ctx context.Context) {
	settings, err := db.Settings.ListAll(ctx, "notifications.slack")
	if err != nil {
		log15.Error(`Warning: unable to migrate "saved queries" to database). Please report this issue`, err)
	}

	for _, s := range settings {
		var notifsField NotificationsSlackField
		err := jsonc.Unmarshal(s.Contents, &notifsField)
		if err != nil {
			log15.Error(`Unable to migrate Slack webhook URL to saved search table. Error unmarshaling notifications.slack setting.`, err)
			continue
		}
		if s.Subject.Org != nil {
			insertSlackWebhookURLIntoSavedSearchesTable(ctx, "org", s.Subject.Org, notifsField.NotificationsSlack.WebhookURL)
		} else if s.Subject.User != nil {
			insertSlackWebhookURLIntoSavedSearchesTable(ctx, "user", s.Subject.User, notifsField.NotificationsSlack.WebhookURL)
		} else if s.Subject.Site || s.Subject.Default {
			siteAdminID := getFirstSiteAdminID(ctx)
			insertSlackWebhookURLIntoSavedSearchesTable(ctx, "user", siteAdminID, notifsField.NotificationsSlack.WebhookURL)
		}
	}

}

// insertSlackWebhookURLIntoSavedSearchesTable inserts an existing Slack webhook URL into the slack_webhook_url column of the
// saved_searches table.
func insertSlackWebhookURLIntoSavedSearchesTable(ctx context.Context, location string, id *int32, webhookURL string) {
	if location == "org" {
		_, err := dbconn.Global.ExecContext(ctx, "UPDATE saved_searches SET slack_webhook_url=$1 WHERE org_id=$2 AND owner_kind=$3", webhookURL, *id, "org")
		if err != nil {
			log15.Error("`Warning: unable to migrate Slack webhook URL into saved search table. Error inserting webhook URL.", err)
		}
	} else if location == "user" {
		_, err := dbconn.Global.ExecContext(ctx, "UPDATE saved_searches SET slack_webhook_url=$1 WHERE user_id=$2 AND owner_kind=$3", webhookURL, *id, "user")
		if err != nil {
			log15.Error("`Warning: unable to migrate Slack webhook URL into saved search table. Error inserting webhook URL.", err)
		}
	}
}
