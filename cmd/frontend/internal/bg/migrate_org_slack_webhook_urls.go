package bg

import (
	"context"
	"time"

	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// MigrateOrgSlackWebhookURLs copies org webhooks from the DB (orgs.slack_webhook_url DB
// column) to the org's JSON settings ("notifications.slack" "webhookURL").
//
// Storing org webhooks in the DB is no longer supported. The DB column will eventually be removed.
func MigrateOrgSlackWebhookURLs(ctx context.Context) {
rerun:
	orgIDsToWebhookURL, err := db.Orgs.TmpListAllOrgsWithSlackWebhookURL(ctx)
	if err != nil {
		log15.Error("Error migrating org Slack webhook URLs: unable to list orgs needing migration.", "error", err)
		return
	}
	for orgID, webhookURL := range orgIDsToWebhookURL {
		if err := migrateOrgSlackWebhookURL(ctx, orgID, webhookURL); err != nil {
			log15.Error("Error migrating org Slack webhook URLs.", "org", orgID, "error", err)
			continue
		}
		if err := db.Orgs.TmpRemoveOrgSlackWebhookURL(ctx, orgID); err != nil {
			log15.Error("Error migrating org Slack webhook URLs: unable to clear org Slack webhook URL.", "org", orgID, "error", err)
			continue
		}
	}

	// To reduce the (already very small) chance of a race condition whereby a new org can
	// have a webhook URL added on an older deployed version while this migration is
	// running, rerun until we have nothing else to do.
	if len(orgIDsToWebhookURL) > 0 {
		log15.Info("Migrated org Slack webhook URLs to org JSON settings.", "data", orgIDsToWebhookURL)
		time.Sleep(60 * time.Second)
		goto rerun
	}
}

func migrateOrgSlackWebhookURL(ctx context.Context, orgID int32, webhookURL string) error {
	subject := api.ConfigurationSubject{Org: &orgID}
	settings, err := db.Settings.GetLatest(ctx, subject)
	if err != nil {
		return err
	}

	contents := "{}"
	if settings != nil {
		contents = settings.Contents
	}

	// Don't clobber existing Slack webhook URL (unlikely case to occur, would need to be
	// set in a very brief window of time).
	var parsed schema.Settings
	if err := jsonc.Unmarshal(contents, &parsed); err != nil {
		return err
	}
	if parsed.NotificationsSlack != nil && parsed.NotificationsSlack.WebhookURL != "" {
		return nil
	}

	// Insert into JSON settings.
	keyPath := jsonx.MakePath("notifications.slack", "webhookURL")
	edits, _, err := jsonx.ComputePropertyEdit(contents, keyPath, webhookURL, nil, jsonx.FormatOptions{InsertSpaces: true, TabSize: 2, EOL: "\n"})
	if err != nil {
		return err
	}
	newConfig, err := jsonx.ApplyEdits(contents, edits...)
	if err != nil {
		return err
	}

	// HACK: Settings must all have an author user ID (enforced by foreign key). Fake that
	// an org member made the update. It would be nice if we could encode that a
	// background process made the update, but it's not important (the author isn't shown
	// anywhere, anyway).
	var fakeAuthorUserID int32
	members, err := db.OrgMembers.GetByOrgID(ctx, orgID)
	if err != nil {
		return err
	}
	if len(members) > 0 {
		fakeAuthorUserID = members[0].UserID
	} else {
		// HACK: The org has no members, but we still need a valid author user ID. Just
		// take any user's ID.
		users, err := db.Users.List(ctx, &db.UsersListOptions{
			LimitOffset: &db.LimitOffset{Limit: 1},
		})
		if err != nil {
			return err
		}
		if len(users) > 0 {
			fakeAuthorUserID = users[0].ID
		} else {
			// Crazy edge case: there's an org, but no users in the entire site. It's
			// probably safe to just delete this org's Slack webhook URL.
			return nil
		}
	}

	// Save new settings to DB.
	var lastID *int32
	if settings != nil {
		lastID = &settings.ID
	}
	_, err = db.Settings.CreateIfUpToDate(ctx, subject, lastID, fakeAuthorUserID, newConfig)
	return err
}
