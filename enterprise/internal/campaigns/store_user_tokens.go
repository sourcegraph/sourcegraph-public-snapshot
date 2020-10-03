package campaigns

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/secret"
)

var userTokenColumns = []*sqlf.Query{
	sqlf.Sprintf("user_id"),
	sqlf.Sprintf("external_service_id"),
	sqlf.Sprintf("token"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

func (s *Store) UpsertUserToken(ctx context.Context, token *campaigns.UserToken) error {
	if token.CreatedAt.IsZero() {
		token.CreatedAt = s.now()
	}

	if token.UpdatedAt.IsZero() {
		token.UpdatedAt = token.CreatedAt
	}

	columns := sqlf.Join(userTokenColumns, ", ")
	q := sqlf.Sprintf(
		upsertUserTokenQueryFmtstr,
		columns,
		token.UserID,
		token.ExternalServiceID,
		token.Token,
		token.CreatedAt,
		token.UpdatedAt,
		columns,
	)

	return s.query(ctx, q, func(sc scanner) error {
		return scanUserToken(token, sc)
	})
}

const upsertUserTokenQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_user_tokens.go:CreateUserToken
INSERT INTO campaign_user_tokens (%s) VALUES (%s, %s, %s, %s, %s)
ON CONFLICT (user_id, external_service_id) DO UPDATE SET
	token = excluded.token,
	created_at = excluded.created_at,
	updated_at = excluded.updated_at
RETURNING %s
`

func (s *Store) DeleteUserToken(ctx context.Context, userID int32, externalServiceID int64) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(deleteUserTokenQueryFmtstr, userID, externalServiceID))
}

const deleteUserTokenQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_user_tokens.go:DeleteUserToken
DELETE FROM campaign_user_tokens WHERE user_id = %s AND external_service_id = %d
`

func (s *Store) GetUserToken(ctx context.Context, userID int32, externalServiceID int64) (*campaigns.UserToken, error) {
	q := sqlf.Sprintf(
		getUserTokenQueryFmtstr,
		sqlf.Join(userTokenColumns, ", "),
		userID,
		externalServiceID,
	)

	var token campaigns.UserToken
	if err := s.query(ctx, q, func(sc scanner) error {
		return scanUserToken(&token, sc)
	}); err != nil {
		return nil, err
	}

	if token.UserID == 0 {
		return nil, ErrNoResults
	}

	return &token, nil
}

const getUserTokenQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_user_tokens.go:GetUserToken
SELECT %s
FROM campaign_user_tokens
WHERE user_id = %s AND external_service_id = %s
LIMIT 1
`

type ListUserTokensOpts struct {
	*db.LimitOffset
	UserID            int32
	ExternalServiceID int64
}

// sql overrides LimitOffset.SQL() to give a LIMIT clause with one extra value
// so we can populate the next cursor.
func (opts *ListUserTokensOpts) sql() *sqlf.Query {
	if opts.LimitOffset == nil || opts.Limit == 0 {
		return &sqlf.Query{}
	}

	return (&db.LimitOffset{Limit: opts.Limit + 1, Offset: opts.Offset}).SQL()
}

func (s *Store) ListUserTokens(ctx context.Context, opts ListUserTokensOpts) (tokens []*campaigns.UserToken, next int, err error) {
	q := listUserTokensQuery(&opts)

	tokens = make([]*campaigns.UserToken, 0)
	err = s.query(ctx, q, func(sc scanner) error {
		var token campaigns.UserToken
		if err := scanUserToken(&token, sc); err != nil {
			return err
		}
		tokens = append(tokens, &token)
		return nil
	})

	// There are more results, since the extra value is populated, so we need
	// to set the return cursor and lop off the last tokens entry.
	if opts.LimitOffset != nil && opts.Limit != 0 && len(tokens) == opts.Limit+1 {
		next = opts.Offset + opts.Limit
		tokens = tokens[:len(tokens)-1]
	}

	return tokens, next, err
}

func listUserTokensQuery(opts *ListUserTokensOpts) *sqlf.Query {
	preds := []*sqlf.Query{}

	if opts.UserID != 0 {
		preds = append(preds, sqlf.Sprintf("user_id = %s", opts.UserID))
	}
	if opts.ExternalServiceID != 0 {
		preds = append(preds, sqlf.Sprintf("external_service_id = %s", opts.ExternalServiceID))
	}

	return sqlf.Sprintf(
		listUserTokensQueryFmtstr,
		sqlf.Join(userTokenColumns, ", "),
		sqlf.Join(preds, "\n AND "),
		opts.sql(),
	)
}

const listUserTokensQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_user_tokens.go:ListUserTokens
SELECT %s
FROM campaign_user_tokens
WHERE %s
ORDER BY created_at ASC, user_id ASC, external_service_id ASC
%s
`

func scanUserToken(token *campaigns.UserToken, s scanner) error {
	// secret.StringValue can't Scan if it doesn't have a concrete string
	// within it, so we'll put something in place, knowing it'll get
	// overwritten by Scan() anyway.
	st := ""
	token.Token = secret.StringValue{S: &st}

	return s.Scan(
		&token.UserID,
		&token.ExternalServiceID,
		&token.Token,
		&token.CreatedAt,
		&token.UpdatedAt,
	)
}
