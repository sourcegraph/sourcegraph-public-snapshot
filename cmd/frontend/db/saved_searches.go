package db

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
)

type savedSearches struct{}

func (s *savedSearches) ListAll(ctx context.Context) (_ []api.SavedQuerySpecAndConfig, err error) {
	tr, ctx := trace.New(ctx, "db.SavedSearches.ListAll", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	q := sqlf.Sprintf(`SELECT id, description, query, notify_owner, notify_slack, owner_kind, user_id, org_id FROM saved_searches`)
	rows, err := dbconn.Global.QueryContext(ctx, q.Query(sqlf.PostgresBindVar))
	if err != nil {
		return nil, err
	}
	var savedQueries []api.SavedQuerySpecAndConfig
	for rows.Next() {
		var sq api.SavedQuerySpecAndConfig
		if err := rows.Scan(&sq.Config.ID, &sq.Config.Description, &sq.Config.Query, &sq.Config.Notify, &sq.Config.NotifySlack, &sq.Config.UserOrOrg, &sq.Config.UserID, &sq.Config.OrgID); err != nil {
			return nil, err
		}
		savedQueries = append(savedQueries, sq)
	}
	return savedQueries, nil
}

type SavedQueryFields struct {
	Description string `json:"description"`
	Query       string `json:"query"`
	Notify      bool   `json:"notify,omitempty"`
	NotifySlack bool   `json:"notifySlack,omitempty"`
	UserOrOrg   string
	UserID      *int32
	OrgID       *int32
}

func (s *savedSearches) Create(ctx context.Context, description string, query string, notify bool, notifySlack bool, userOrOrg string, userID *int32, orgID *int32) (err error) {
	tr, ctx := trace.New(ctx, "db.SavedSearches.ListAll", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	_, err = dbconn.Global.ExecContext(ctx, `INSERT INTO saved_searches(description, query, notify_owner, notify_slack, owner_kind, user_id, org_id) VALUES($1, $2, $3, $4, $5, $6, $7)`, description, query, notify, notifySlack, strings.ToLower(userOrOrg), userID, orgID)
	if err != nil {
		return err
	}
	return nil
}
