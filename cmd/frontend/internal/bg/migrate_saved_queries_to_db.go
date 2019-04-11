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

func MigrateAllSavedQueriesFromSettingsToDatabase(ctx context.Context) {
	settings, err := db.Settings.ListAll(ctx, "search.savedQueries")
	if err != nil {
		log15.Error(`Warning: unable to migrate "saved queries" to database). Please report this issue`, err)
	}
	for _, s := range settings {
		var sq SavedQueryField
		err := jsonc.Unmarshal(s.Contents, &sq)
		if err != nil {
			log15.Error(`Unable to migrate saved query into database, unable to unmarshal JSON value`, err)
		}
		InsertSavedQueryIntoDB(ctx, s, &sq)
		// Remove "search.savedQueries" entry from settings.
		edits, _, err := jsonx.ComputePropertyRemoval(s.Contents, jsonx.MakePath("search.savedQueries"), conf.FormatOptions)
		if err != nil {
			log15.Error(`Unable to remove savedQuery from settings`, err)
		}
		text, err := jsonx.ApplyEdits(s.Contents, edits...)
		if err != nil {
			log15.Error(`Unable to remove savedQuery from settings`, err)
		}

		var lastID *int32
		lastID = &s.ID
		_, err = db.Settings.CreateIfUpToDate(ctx, s.Subject, lastID, s.AuthorUserID, text)
		if err != nil {
			log15.Error(`Unable to update settings`)
		}

	}

}

func InsertSavedQueryIntoDB(ctx context.Context, s *api.Settings, sq *SavedQueryField) {
	for _, query := range sq.SavedQueries {
		if s.Subject.User != nil {
			_, err := dbconn.Global.ExecContext(ctx, "INSERT INTO saved_searches(description, query, notify_owner, notify_slack, owner_kind, user_id) VALUES($1, $2, $3, $4, $5, $6)", query.Description, query.Query, query.Notify, query.NotifySlack, "user", s.AuthorUserID)
			if err != nil {
				log15.Error(`Warning: unable to migrate saved query into database.`, err.Error())
			}
		} else if s.Subject.Org != nil {
			_, err := dbconn.Global.ExecContext(ctx, "SELECT org_id FROM settings WHERE id=$1", s.ID)
			if err != nil {
				log15.Error(`Warning: unable to migrate saved query into database.`, err.Error())
			}
			_, err = dbconn.Global.ExecContext(ctx, "INSERT INTO saved_searches(description, query, notify_owner, notify_slack, owner_kind, user_id) VALUES($1, $2, $3, $4, $5, $6)", query.Description, query.Query, query.Notify, query.NotifySlack, "org", s.AuthorUserID)
			if err != nil {
				log15.Error(`Warning: unable to migrate saved query into database.`, err.Error())
			}
		}
	}

}

// List all saved queries from db using db.Settings.ListAll(ctx, "search.savedQueries")
