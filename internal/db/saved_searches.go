package db

import (
	"bytes"
	"context"
	"database/sql"

	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	secretPkg "github.com/sourcegraph/sourcegraph/internal/secrets"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type savedSearches struct{}

// IsEmpty tells if there are no saved searches (at all) on this Sourcegraph
// instance.
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
// 🚨 SECURITY: This method does NOT verify the user's identity or that the
// user is an admin. It is the callers responsibility to ensure that only users
// with the proper permissions can access the returned saved searches.
func (s *savedSearches) ListAll(ctx context.Context) (savedSearches []api.SavedQuerySpecAndConfig, err error) {
	if Mocks.SavedSearches.ListAll != nil {
		return Mocks.SavedSearches.ListAll(ctx)
	}

	tr, ctx := trace.New(ctx, "db.SavedSearches.ListAll", "")
	defer func() {
		tr.SetError(err)
		tr.LogFields(otlog.Int("count", len(savedSearches)))
		tr.Finish()
	}()

	q := sqlf.Sprintf(`SELECT
		id,
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
		return nil, errors.Wrap(err, "QueryContext")
	}

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
			return nil, errors.Wrap(err, "Scan")
		}
		sq.Spec.Key = sq.Config.Key
		if sq.Config.UserID != nil {
			sq.Spec.Subject.User = sq.Config.UserID
		} else if sq.Config.OrgID != nil {
			sq.Spec.Subject.Org = sq.Config.OrgID
		}
		if secretPkg.ConfiguredToEncrypt() {
			plaintext, err := secretPkg.DecodeAndDecryptBytes([]byte(sq.Config.Query))
			if err != nil {
				return nil, errors.Wrap(err, "saved_searches: unable to decrypt")
			}
			sq.Config.Query = string(plaintext)
		}

		savedSearches = append(savedSearches, sq)
	}
	return savedSearches, nil
}

// GetByID returns the saved search with the given ID.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the
// user is an admin. It is the callers responsibility to ensure this response
// only makes it to users with proper permissions to access the saved search.
func (s *savedSearches) GetByID(ctx context.Context, id int32) (*api.SavedQuerySpecAndConfig, error) {
	if Mocks.SavedSearches.GetByID != nil {
		return Mocks.SavedSearches.GetByID(ctx, id)
	}
	var sq api.SavedQuerySpecAndConfig

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
		&sq.Config.Key,
		&sq.Config.Description,
		&sq.Config.Query,
		&sq.Config.Notify,
		&sq.Config.NotifySlack,
		&sq.Config.UserID,
		&sq.Config.OrgID,
		&sq.Config.SlackWebhookURL)
	if err != nil {
		return nil, err
	}
	sq.Spec.Key = sq.Config.Key
	if sq.Config.UserID != nil {
		sq.Spec.Subject.User = sq.Config.UserID
	} else if sq.Config.OrgID != nil {
		sq.Spec.Subject.Org = sq.Config.OrgID
	}
	if secretPkg.ConfiguredToEncrypt() {
		plaintext, err := secretPkg.DecodeAndDecryptBytes([]byte(sq.Config.Query))
		if err != nil {
			return nil, errors.Wrap(err, "saved_searches: unable to decrypt")
		}
		sq.Config.Query = string(plaintext)
	}

	return &sq, err
}

// ListSavedSearchesByUserID lists all the saved searches associated with a
// user, including saved searches in organizations the user is a member of.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the
// user is an admin. It is the callers responsibility to ensure that only the
// specified user or users with proper permissions can access the returned
// saved searches.
func (s *savedSearches) ListSavedSearchesByUserID(ctx context.Context, userID int32) ([]*types.SavedSearch, error) {
	if Mocks.SavedSearches.ListSavedSearchesByUserID != nil {
		return Mocks.SavedSearches.ListSavedSearchesByUserID(ctx, userID)
	}
	var savedSearches []*types.SavedSearch
	orgs, err := Orgs.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var orgIDs []int32
	for _, org := range orgs {
		orgIDs = append(orgIDs, org.ID)
	}
	var orgConditions []*sqlf.Query
	for _, orgID := range orgIDs {
		orgConditions = append(orgConditions, sqlf.Sprintf("org_id=%d", orgID))
	}
	conds := sqlf.Sprintf("WHERE user_id=%d", userID)

	if len(orgConditions) > 0 {
		conds = sqlf.Sprintf("%v OR %v", conds, sqlf.Join(orgConditions, " OR "))
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
		return nil, errors.Wrap(err, "QueryContext(2)")
	}
	for rows.Next() {
		var ss types.SavedSearch
		if err := rows.Scan(&ss.ID, &ss.Description, &ss.Query, &ss.Notify, &ss.NotifySlack, &ss.UserID, &ss.OrgID, &ss.SlackWebhookURL); err != nil {
			return nil, errors.Wrap(err, "Scan(2)")
		}
		if secretPkg.ConfiguredToEncrypt() {
			// TODO base64 decode
			plaintext, err := secretPkg.DecodeAndDecryptBytes([]byte(ss.Query))
			if err != nil {
				return nil, errors.Wrap(err, "saved_searches: unable to decrypt")
			}
			ss.Query = string(plaintext)
		}
		savedSearches = append(savedSearches, &ss)
	}
	return savedSearches, nil
}

// ListSavedSearchesByUserID lists all the saved searches associated with an
// organization.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the
// user is an admin. It is the callers responsibility to ensure only admins or
// members of the specified organization can access the returned saved
// searches.
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
		return nil, errors.Wrap(err, "QueryContext")
	}
	for rows.Next() {
		var ss types.SavedSearch
		if err := rows.Scan(&ss.ID, &ss.Description, &ss.Query, &ss.Notify, &ss.NotifySlack, &ss.UserID, &ss.OrgID, &ss.SlackWebhookURL); err != nil {
			return nil, errors.Wrap(err, "Scan")
		}

		if secretPkg.ConfiguredToEncrypt() {
			plaintext, err := secretPkg.DecodeAndDecryptBytes([]byte(ss.Query))
			if err != nil {
				return nil, err
			}
			ss.Query = string(plaintext)
		}

		savedSearches = append(savedSearches, &ss)
	}
	return savedSearches, nil
}

// Create creates a new saved search with the specified parameters. The ID
// field must be zero, or an error will be returned.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the
// user is an admin. It is the callers responsibility to ensure the user has
// proper permissions to create the saved search.
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

	if secretPkg.ConfiguredToEncrypt() {
		ciphertext, err := secretPkg.EncodeAndEncryptBytes([]byte(newSavedSearch.Query))
		if err != nil {
			return nil, errors.Wrap(err, "saved search: failed to encrypt")
		}
		newSavedSearch.Query = string(ciphertext)
	}

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

// Update updates an existing saved search.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the
// user is an admin. It is the callers responsibility to ensure the user has
// proper permissions to perform the update.
func (s *savedSearches) Update(ctx context.Context, savedSearch *types.SavedSearch) (savedQuery *types.SavedSearch, err error) {
	if Mocks.SavedSearches.Update != nil {
		return Mocks.SavedSearches.Update(ctx, savedSearch)
	}

	tr, ctx := trace.New(ctx, "db.SavedSearches.Update", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if secretPkg.ConfiguredToEncrypt() {
		ciphertext, err := secretPkg.EncodeAndEncryptBytes([]byte(savedSearch.Query))
		if err != nil {
			return nil, errors.Wrap(err, "saved search: failed to encrypt")
		}
		savedSearch.Query = string(ciphertext)
	}

	savedQuery = &types.SavedSearch{
		Description:     savedSearch.Description,
		Query:           savedSearch.Query,
		Notify:          savedSearch.Notify,
		NotifySlack:     savedSearch.NotifySlack,
		UserID:          savedSearch.UserID,
		OrgID:           savedSearch.OrgID,
		SlackWebhookURL: savedSearch.SlackWebhookURL,
	}

	fieldUpdates := []*sqlf.Query{
		sqlf.Sprintf("updated_at=now()"),
		sqlf.Sprintf("description=%s", savedSearch.Description),
		sqlf.Sprintf("query=%s", savedSearch.Query),
		sqlf.Sprintf("notify_owner=%t", savedSearch.Notify),
		sqlf.Sprintf("notify_slack=%t", savedSearch.NotifySlack),
		sqlf.Sprintf("user_id=%v", savedSearch.UserID),
		sqlf.Sprintf("org_id=%v", savedSearch.OrgID),
		sqlf.Sprintf("slack_webhook_url=%v", savedSearch.SlackWebhookURL),
	}

	updateQuery := sqlf.Sprintf(`UPDATE saved_searches SET %s WHERE ID=%v RETURNING id`, sqlf.Join(fieldUpdates, ", "), savedSearch.ID)
	if err := dbconn.Global.QueryRowContext(ctx, updateQuery.Query(sqlf.PostgresBindVar), updateQuery.Args()...).Scan(&savedQuery.ID); err != nil {
		return nil, err
	}
	return savedQuery, nil
}

// Delete hard-deletes an existing saved search.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the
// user is an admin. It is the callers responsibility to ensure the user has
// proper permissions to perform the delete.
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

func (*savedSearches) EncryptTable(ctx context.Context) error {
	if !secretPkg.ConfiguredToEncrypt() {
		return nil
	}

	q := sqlf.Sprintf("SELECT id, query from saved_searches where query like %s%%", secretPkg.KeyHash())
	tx, err := dbconn.Global.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				err = multierror.Append(err, rollErr)
			}
			return
		}
		err = tx.Commit()
	}()

	rows, err := tx.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args())
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sq types.SavedSearch
		err := rows.Scan(&sq.ID, &sq.Query)
		if err != nil {
			return err
		}

		var cryptBytes []byte
		byteQuery := []byte(sq.Query)
		if secretPkg.ConfiguredToRotate() {
			cryptBytes, err = secretPkg.RotateEncryption(byteQuery)
			if err != nil {
				return err
			}
			if bytes.Equal(byteQuery, cryptBytes) {
				continue
			}
		} else {
			cryptBytes, err = secretPkg.EncodeAndEncryptBytes(byteQuery)
			if err != nil {
				return err
			}
			if bytes.Equal(byteQuery, cryptBytes) {
				continue
			}
		}

		updateQ := sqlf.Sprintf("UPDATE saved_searches SET query = %s WHERE id=%d", cryptBytes, sq.ID)
		_, err = tx.ExecContext(ctx, updateQ.Query(sqlf.PostgresBindVar), updateQ.Args())
		return err
	}
	return rows.Err()
}
