package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
)

type savedSearches struct{}

func (s *savedSearches) IsEmpty(ctx context.Context) (bool, error) {
	q := `SELECT true FROM saved_searches LIMIT 1`
	var isNotEmpty bool
	err := dbconn.Global.QueryRow(q).Scan(&isNotEmpty)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		return false, err
	}

	return false, nil

}

// ListAll lists all the saved searches on an instance.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the user is an admin.
// It is the caller's responsibility to make sure this response never makes it to a user.
func (s *savedSearches) ListAll(ctx context.Context) (_ []api.SavedQuerySpecAndConfig, err error) {
	if Mocks.SavedSearches.ListAll != nil {
		return Mocks.SavedSearches.ListAll(ctx)
	}

	tr, ctx := trace.New(ctx, "db.SavedSearches.ListAll", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	q := sqlf.Sprintf(`SELECT id,
		description,
		query,
		notify_owner,
		notify_slack,
		user_id,
		org_id,
		slack_webhook_url FROM saved_searches
	`)
	rows, err := dbconn.Global.QueryContext(ctx, q.Query(sqlf.PostgresBindVar))
	if err != nil {
		return nil, err
	}
	var savedSearches []api.SavedQuerySpecAndConfig
	for rows.Next() {
		var sq api.SavedQuerySpecAndConfig
		if err := rows.Scan(
			&sq.Config.Key,
			&sq.Config.Description,
			&sq.Config.Query,
			&sq.Config.Notify,
			&sq.Config.NotifySlack,
			&sq.Config.UserID,
			&sq.Config.OrgID,
			&sq.Config.SlackWebhookURL); err != nil {
			return nil, err
		}
		sq.Spec.Key = sq.Config.Key
		if sq.Config.UserID != nil {
			sq.Spec.Subject.User = sq.Config.UserID
		} else if sq.Config.OrgID != nil {
			sq.Spec.Subject.Org = sq.Config.OrgID
		}
		savedSearches = append(savedSearches, sq)
	}
	return savedSearches, nil
}

func (s *savedSearches) GetSavedSearchByID(ctx context.Context, id int32) (*api.SavedQuerySpecAndConfig, error) {
	if Mocks.SavedSearches.GetSavedSearchByID != nil {
		return Mocks.SavedSearches.GetSavedSearchByID(ctx, id)
	}
	var savedSearch api.SavedQuerySpecAndConfig

	err := dbconn.Global.QueryRowContext(ctx, `SELECT
		id,
		description,
		query,
		notify_owner,
		notify_slack,
		user_id,
		org_id,
		slack_webhook_url
		FROM saved_searches WHERE id=$1`, id).Scan(
		&savedSearch.Config.Key,
		&savedSearch.Config.Description,
		&savedSearch.Config.Query,
		&savedSearch.Config.Notify,
		&savedSearch.Config.NotifySlack,
		&savedSearch.Config.UserID,
		&savedSearch.Config.OrgID,
		&savedSearch.Config.SlackWebhookURL)
	savedSearch.Spec.Key = savedSearch.Config.Key
	if savedSearch.Config.UserID != nil {
		savedSearch.Spec.Subject = api.SettingsSubject{User: savedSearch.Config.UserID}
	} else if savedSearch.Config.OrgID != nil {
		savedSearch.Spec.Subject = api.SettingsSubject{Org: savedSearch.Config.OrgID}
	}

	if err != nil {
		return nil, err
	}
	return &savedSearch, err
}

// ListSavedSearchesByUserID lists all the saved searches associated with a user,
// including saved searches in organizations the user is a member of.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the user is an admin.
// It is the caller's responsibility to make sure this method returns only saved searches associated with
// the current user.
func (s *savedSearches) ListSavedSearchesByUserID(ctx context.Context, userID int32) ([]*types.SavedSearch, error) {
	if Mocks.SavedSearches.ListSavedSearchesByUserID != nil {
		return Mocks.SavedSearches.ListSavedSearchesByUserID(ctx, userID)
	}
	var savedSearches []*types.SavedSearch
	orgIDRows, err := dbconn.Global.QueryContext(ctx, `SELECT org_id FROM org_members WHERE user_id=$1`, userID)
	var orgIDs []int32
	for orgIDRows.Next() {
		var orgID int32
		err := orgIDRows.Scan(&orgID)
		if err != nil {
			return nil, err
		}
		orgIDs = append(orgIDs, orgID)
	}
	var orgConditions []*sqlf.Query
	for _, orgID := range orgIDs {
		orgConditions = append(orgConditions, sqlf.Sprintf("org_id=%d", orgID))
	}
	conds := sqlf.Sprintf("WHERE user_id=%d", userID)
	if err != nil {
		return nil, err
	}

	if len(orgConditions) > 0 {
		conds = sqlf.Sprintf("%v OR %v", conds, sqlf.Join(orgConditions, " ) OR ("))
	}

	query := sqlf.Sprintf(`SELECT
		id,
		description,
		query,
		notify_owner,
		notify_slack,
		user_id,
		org_id,
		slack_webhook_url
		FROM saved_searches %v`, conds)

	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var ss types.SavedSearch
		if err := rows.Scan(&ss.ID, &ss.Description, &ss.Query, &ss.Notify, &ss.NotifySlack, &ss.UserID, &ss.OrgID, &ss.SlackWebhookURL); err != nil {
			return nil, err
		}
		savedSearches = append(savedSearches, &ss)
	}
	return savedSearches, err
}

// ListSavedSearchesByUserID lists all the saved searches associated with an organization.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the user is an admin.
// It is the caller's responsibility to make sure this method returns only saved searches from organizations the user is a part of.
func (s *savedSearches) ListSavedSearchesByOrgID(ctx context.Context, orgID int32) ([]*types.SavedSearch, error) {
	var savedSearches []*types.SavedSearch
	conds := sqlf.Sprintf("WHERE org_id=%d", orgID)
	query := sqlf.Sprintf(`SELECT
		id,
		description,
		query,
		notify_owner,
		notify_slack,
		user_id,
		org_id,
		slack_webhook_url
		FROM saved_searches %v`, conds)

	rows, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var ss types.SavedSearch
		if err := rows.Scan(&ss.ID, &ss.Description, &ss.Query, &ss.Notify, &ss.NotifySlack, &ss.UserID, &ss.OrgID, &ss.SlackWebhookURL); err != nil {
			return nil, err
		}
		savedSearches = append(savedSearches, &ss)
	}
	return savedSearches, err
}

func (s *savedSearches) Create(ctx context.Context, newSavedSearch *types.SavedSearch) (savedQuery *types.SavedSearch, err error) {
	if Mocks.SavedSearches.Create != nil {
		return Mocks.SavedSearches.Create(ctx, newSavedSearch)
	}

	if newSavedSearch.ID != 0 {
		return nil, errors.New("newSavedSearch.ID must be zero")
	}

	tr, ctx := trace.New(ctx, "db.SavedSearches.Create", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	savedQuery = &types.SavedSearch{
		Description: newSavedSearch.Description,
		Query:       newSavedSearch.Query,
		Notify:      newSavedSearch.Notify,
		NotifySlack: newSavedSearch.NotifySlack,
		UserID:      newSavedSearch.UserID,
		OrgID:       newSavedSearch.OrgID,
	}

	err = dbconn.Global.QueryRowContext(ctx, `INSERT INTO saved_searches(
			description,
			query,
			notify_owner,
			notify_slack,
			user_id,
			org_id
		) VALUES($1, $2, $3, $4, $5, $6) RETURNING id`,
		newSavedSearch.Description,
		newSavedSearch.Query,
		newSavedSearch.Notify,
		newSavedSearch.NotifySlack,
		newSavedSearch.UserID,
		newSavedSearch.OrgID,
	).Scan(&savedQuery.ID)
	if err != nil {
		return nil, err
	}
	return savedQuery, nil
}

func (s *savedSearches) Update(ctx context.Context, savedSearch *types.SavedSearch) (savedQuery *types.SavedSearch, err error) {
	if Mocks.SavedSearches.Update != nil {
		return Mocks.SavedSearches.Update(ctx, savedSearch)
	}

	tr, ctx := trace.New(ctx, "db.SavedSearches.Update", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	savedQuery = &types.SavedSearch{
		Description: savedSearch.Description,
		Query:       savedSearch.Query,
		Notify:      savedSearch.Notify,
		NotifySlack: savedSearch.NotifySlack,
		UserID:      savedSearch.UserID,
		OrgID:       savedSearch.OrgID,
	}

	fieldUpdates := []*sqlf.Query{
		sqlf.Sprintf("updated_at=now()"),
		sqlf.Sprintf("description=%s", savedSearch.Description),
		sqlf.Sprintf("query=%s", savedSearch.Query),
		sqlf.Sprintf("notify_owner=%t", savedSearch.Notify),
		sqlf.Sprintf("notify_slack=%t", savedSearch.NotifySlack),
		sqlf.Sprintf("user_id=%v", savedSearch.UserID),
		sqlf.Sprintf("org_id=%v", savedSearch.OrgID),
	}

	updateQuery := sqlf.Sprintf(`UPDATE saved_searches SET %s WHERE ID=%v RETURNING id`, sqlf.Join(fieldUpdates, ", "), savedSearch.ID)
	if err := dbconn.Global.QueryRowContext(ctx, updateQuery.Query(sqlf.PostgresBindVar), updateQuery.Args()...).Scan(&savedQuery.ID); err != nil {
		return nil, err
	}
	return savedQuery, nil
}

func (s *savedSearches) Delete(ctx context.Context, id int32) (err error) {
	if Mocks.SavedSearches.Delete != nil {
		return Mocks.SavedSearches.Delete(ctx, id)
	}

	tr, ctx := trace.New(ctx, "db.SavedSearches.Delete", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	_, err = dbconn.Global.ExecContext(ctx, `DELETE FROM saved_searches WHERE ID=$1`, id)
	if err != nil {
		return err
	}
	return nil
}
