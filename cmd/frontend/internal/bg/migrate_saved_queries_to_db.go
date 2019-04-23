package bg

import (
	"context"

	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
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
// site settings into the `saved_searches` table in the database, and migrates organization Slack webhook URLs from the
// `notifications.slack` value in site settings to the `slack_webhook_url` column in the `saved_searches` table.
func MigrateSavedQueriesAndSlackWebhookURLsFromSettingsToDatabase(ctx context.Context) {
	settings, err := db.Settings.ListAll(ctx, "search.savedQueries")
	if err != nil {
		log15.Error(`Warning: unable to migrate saved queries to database. Error listing settings. Please report this issue`, err)
	}
	for _, s := range settings {
		var sq SavedQueryField
		err := jsonc.Unmarshal(s.Contents, &sq)
		if err != nil {
			log15.Error(`Unable to migrate saved query into database, unable to unmarshal JSON value. Please report this issue.`, err)
		}
		InsertSavedQueryIntoDB(ctx, s, &sq)
		// Remove "search.savedQueries" entry from settings.
		edits, _, err := jsonx.ComputePropertyRemoval(s.Contents, jsonx.MakePath("search.savedQueries"), conf.FormatOptions)
		if err != nil {
			log15.Error(`Unable to remove saved query from settings. Please report this issue.`, err)
		}
		text, err := jsonx.ApplyEdits(s.Contents, edits...)
		if err != nil {
			log15.Error(`Unable to remove saved query from settings. Please report this issue.`, err)
		}

		lastID := &s.ID
		_, err = db.Settings.CreateIfUpToDate(ctx, s.Subject, lastID, s.AuthorUserID, text)
		if err != nil {
			log15.Error(`Unable to update settings with saved queries removed. Please report this issue.`)
		}
	}
	// After migrating all saved searches into the DB, migrate organization Slack webhook URLs into the
	// slack_webhook_url column in the saved_searches table.
	MigrateSlackWebhookUrlsToSavedSearches(ctx)

}

// InsertSavedQueryIntoDB inserts an existing saved query from site settings into the saved_searches database table.
// Global saved queries will be associated with the first site admin's profile. It will be added with `owner_kind`
// set to "user", and with UserID set to the first site_admin's user ID.
func InsertSavedQueryIntoDB(ctx context.Context, s *api.Settings, sq *SavedQueryField) {
	for _, query := range sq.SavedQueries {
		// Add case for global settings. It should make a site admin user the owner of that saved search.
		if s.Subject.User != nil {
			_, err := dbconn.Global.ExecContext(ctx, "INSERT INTO saved_searches(description, query, notify_owner, notify_slack, owner_kind, user_id) VALUES($1, $2, $3, $4, $5, $6)", query.Description, query.Query, query.Notify, query.NotifySlack, "user", *s.Subject.User)
			if err != nil {
				log15.Error(`Warning: unable to migrate user saved query into database.`, err.Error())
			}
		} else if s.Subject.Org != nil {
			_, err := dbconn.Global.ExecContext(ctx, "INSERT INTO saved_searches(description, query, notify_owner, notify_slack, owner_kind, org_id) VALUES($1, $2, $3, $4, $5, $6)", query.Description, query.Query, query.Notify, query.NotifySlack, "org", *s.Subject.Org)
			if err != nil {
				log15.Error(`Warning: unable to migrate org saved query into database.`, err.Error())
			}
		} else if s.Subject.Site || s.Subject.Default {
			var siteAdminID *int32
			err := dbconn.Global.QueryRowContext(ctx, "SELECT id FROM users WHERE site_admin=true LIMIT 1").Scan(&siteAdminID)
			if err != nil {
				log15.Error(`Warning: unable to migrate saved query into database. No site admin ID found.`, err.Error())
			}

			// AuthorUserID is the UserID of the person who last wrote to the settings.
			_, err = dbconn.Global.ExecContext(ctx, "INSERT INTO saved_searches(description, query, notify_owner, notify_slack, owner_kind, user_id) VALUES($1, $2, $3, $4, $5, $6)", query.Description, query.Query, query.Notify, query.NotifySlack, "user", *siteAdminID)
			if err != nil {
				log15.Error(`Warning: unable to migrate global saved query into database.`, err.Error())
			}
		}
	}
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
		}
		if s.Subject.Org != nil {
			InsertSlackWebhookURLIntoSavedSearchesTable(ctx, "org", s.Subject.Org, notifsField.NotificationsSlack.WebhookURL)
		} else if s.Subject.User != nil {
			InsertSlackWebhookURLIntoSavedSearchesTable(ctx, "user", s.Subject.User, notifsField.NotificationsSlack.WebhookURL)
		}

		edits, _, err := jsonx.ComputePropertyRemoval(s.Contents, jsonx.MakePath("notifications.slack"), conf.FormatOptions)
		if err != nil {
			log15.Error(`Unable to remove Slack webhook URL from settings. Please report this issue.`, err)
		}
		text, err := jsonx.ApplyEdits(s.Contents, edits...)
		if err != nil {
			log15.Error(`Unable to apply settings with Slack webhook URL removed from settings. Please report this issue.`, err)
		}

		lastID := &s.ID
		_, err = db.Settings.CreateIfUpToDate(ctx, s.Subject, lastID, s.AuthorUserID, text)
		if err != nil {
			log15.Error(`Unable to create new settings with Slack webhook URL removed. Please report this issue.`)
		}
	}

}

// InsertSlackWebhookURLIntoSavedSearchesTable inserts an existing Slack webhook URL into the slack_webhook_url column of the
// saved_searches table.
func InsertSlackWebhookURLIntoSavedSearchesTable(ctx context.Context, location string, id *int32, webhookURL string) {
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
