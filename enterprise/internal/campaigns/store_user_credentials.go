package campaigns

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/secret"
)

var userTokenColumns = []*sqlf.Query{
	sqlf.Sprintf("user_id"),
	sqlf.Sprintf("external_service_id"),
	sqlf.Sprintf("credential"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

func (s *Store) UpsertUserToken(ctx context.Context, cred *campaigns.UserCredential) error {
	if cred.CreatedAt.IsZero() {
		cred.CreatedAt = s.now()
	}

	if cred.UpdatedAt.IsZero() {
		cred.UpdatedAt = cred.CreatedAt
	}

	raw, err := marshalCredential(cred.Credential)
	if err != nil {
		return errors.Wrap(err, "marshalling credential")
	}

	columns := sqlf.Join(userTokenColumns, ", ")
	q := sqlf.Sprintf(
		upsertUserTokenQueryFmtstr,
		columns,
		cred.UserID,
		cred.ExternalServiceID,
		secret.StringValue{S: &raw},
		cred.CreatedAt,
		cred.UpdatedAt,
		columns,
	)

	return s.query(ctx, q, func(sc scanner) error {
		return scanUserCredential(cred, sc)
	})
}

const upsertUserTokenQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_user_tokens.go:CreateUserToken
INSERT INTO campaign_user_credentials (%s) VALUES (%s, %s, %s, %s, %s)
ON CONFLICT (user_id, external_service_id) DO UPDATE SET
	credential = excluded.credential,
	created_at = excluded.created_at,
	updated_at = excluded.updated_at
RETURNING %s
`

func (s *Store) DeleteUserToken(ctx context.Context, userID int32, externalServiceID int64) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(deleteUserTokenQueryFmtstr, userID, externalServiceID))
}

const deleteUserTokenQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_user_tokens.go:DeleteUserToken
DELETE FROM campaign_user_credentials WHERE user_id = %s AND external_service_id = %d
`

func (s *Store) GetUserToken(ctx context.Context, userID int32, externalServiceID int64) (*campaigns.UserCredential, error) {
	q := sqlf.Sprintf(
		getUserTokenQueryFmtstr,
		sqlf.Join(userTokenColumns, ", "),
		userID,
		externalServiceID,
	)

	var token campaigns.UserCredential
	if err := s.query(ctx, q, func(sc scanner) error {
		return scanUserCredential(&token, sc)
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
FROM campaign_user_credentials
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

func (s *Store) ListUserTokens(ctx context.Context, opts ListUserTokensOpts) (tokens []*campaigns.UserCredential, next int, err error) {
	q := listUserTokensQuery(&opts)

	tokens = make([]*campaigns.UserCredential, 0)
	err = s.query(ctx, q, func(sc scanner) error {
		var token campaigns.UserCredential
		if err := scanUserCredential(&token, sc); err != nil {
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
FROM campaign_user_credentials
WHERE %s
ORDER BY created_at ASC, user_id ASC, external_service_id ASC
%s
`

func scanUserCredential(token *campaigns.UserCredential, s scanner) error {
	raw := ""

	if err := s.Scan(
		&token.UserID,
		&token.ExternalServiceID,
		&secret.StringValue{S: &raw},
		&token.CreatedAt,
		&token.UpdatedAt,
	); err != nil {
		return err
	}

	var err error
	if token.Credential, err = unmarshalCredential(raw); err != nil {
		return errors.Wrap(err, "unmarshalling credential")
	}

	return nil
}

// Define credential type strings that we'll use when encoding credentials.
const (
	credTypeOAuthClient                        = "OAuthClient"
	credTypeBasicAuth                          = "BasicAuth"
	credTypeOAuthBearerToken                   = "OAuthBearerToken"
	credTypeBitbucketServerSudoableOAuthClient = "BitbucketSudoableOAuthClient"
	credTypeGitLabSudoableToken                = "GitLabSudoableToken"
)

func marshalCredential(a auth.Authenticator) (string, error) {
	var t string
	switch a.(type) {
	case *auth.OAuthClient:
		t = credTypeOAuthClient
	case *auth.BasicAuth:
		t = credTypeBasicAuth
	case *auth.OAuthBearerToken:
		t = credTypeOAuthBearerToken
	case *bitbucketserver.SudoableOAuthClient:
		t = credTypeBitbucketServerSudoableOAuthClient
	case *gitlab.SudoableToken:
		t = credTypeGitLabSudoableToken
	default:
		return "", errors.Errorf("unknown Authenticator implementation type: %T", a)
	}

	raw, err := json.Marshal(struct {
		Type string
		Auth auth.Authenticator
	}{
		Type: t,
		Auth: a,
	})
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func unmarshalCredential(raw string) (auth.Authenticator, error) {
	var partial struct {
		Type string
		Auth json.RawMessage
	}
	if err := json.Unmarshal([]byte(raw), &partial); err != nil {
		return nil, err
	}

	var a interface{}
	switch partial.Type {
	case credTypeOAuthClient:
		a = &auth.OAuthClient{}
	case credTypeBasicAuth:
		a = &auth.BasicAuth{}
	case credTypeOAuthBearerToken:
		a = &auth.OAuthBearerToken{}
	case credTypeBitbucketServerSudoableOAuthClient:
		a = &bitbucketserver.SudoableOAuthClient{}
	case credTypeGitLabSudoableToken:
		a = &gitlab.SudoableToken{}
	default:
		return nil, errors.Errorf("unknown credential type: %s", partial.Type)
	}

	log15.Info("a", "type", fmt.Sprintf("%T", a))
	if err := json.Unmarshal(partial.Auth, &a); err != nil {
		return nil, err
	}

	return a.(auth.Authenticator), nil
}
