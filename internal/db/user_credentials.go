package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

// UserCredential represents a row in the `user_credentials` table.
type UserCredential struct {
	ID                  int64
	Domain              string
	UserID              int32
	ExternalServiceType string
	ExternalServiceID   string
	Credential          auth.Authenticator
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// This const block contains the valid domain values for user credentials.
const (
	UserCredentialDomainCampaigns = "campaigns"
)

// UserCredentialNotFoundErr is returned when a credential cannot be found from
// its ID or scope.
type UserCredentialNotFoundErr struct{ args []interface{} }

func (err UserCredentialNotFoundErr) Error() string {
	return fmt.Sprintf("user credential not found: %v", err.args)
}

func (UserCredentialNotFoundErr) NotFound() bool {
	return true
}

// userCredentials provides access to the `user_credentials` table.
type userCredentials struct{}

// UserCredentialScope represents the unique scope for a credential. Only one
// credential may exist within a scope.
type UserCredentialScope struct {
	Domain              string
	UserID              int32
	ExternalServiceType string
	ExternalServiceID   string
}

// Create creates a new user credential based on the given scope and
// authenticator. If the scope already has a credential, an error will be
// returned.
func (*userCredentials) Create(ctx context.Context, scope UserCredentialScope, credential auth.Authenticator) (*UserCredential, error) {
	if Mocks.UserCredentials.Create != nil {
		return Mocks.UserCredentials.Create(ctx, scope, credential)
	}

	raw, err := marshalCredential(credential)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling credential")
	}

	q := sqlf.Sprintf(
		userCredentialsCreateQueryFmtstr,
		scope.Domain,
		scope.UserID,
		scope.ExternalServiceType,
		scope.ExternalServiceID,
		raw,
		sqlf.Join(userCredentialsColumns, ", "),
	)

	cred := UserCredential{}
	row := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err := scanUserCredential(&cred, row); err != nil {
		return nil, err
	}

	return &cred, nil
}

// Delete deletes the given user credential. Note that there is no concept of a
// soft delete with user credentials: once deleted, the relevant records are
// _gone_, so that we don't hold any sensitive data unexpectedly. 💀
func (*userCredentials) Delete(ctx context.Context, id int64) error {
	if Mocks.UserCredentials.Delete != nil {
		return Mocks.UserCredentials.Delete(ctx, id)
	}

	q := sqlf.Sprintf("DELETE FROM user_credentials WHERE id = %s", id)
	res, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	if rows, err := res.RowsAffected(); err != nil {
		return err
	} else if rows == 0 {
		return UserCredentialNotFoundErr{args: []interface{}{id}}
	}

	return nil
}

// GetByID returns the user credential matching the given ID, or
// UserCredentialNotFoundErr if no such credential exists.
func (*userCredentials) GetByID(ctx context.Context, id int64) (*UserCredential, error) {
	if Mocks.UserCredentials.GetByID != nil {
		return Mocks.UserCredentials.GetByID(ctx, id)
	}

	q := sqlf.Sprintf(
		"SELECT %s FROM user_credentials WHERE id = %s",
		sqlf.Join(userCredentialsColumns, ", "),
		id,
	)

	cred := UserCredential{}
	row := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err := scanUserCredential(&cred, row); err == sql.ErrNoRows {
		return nil, UserCredentialNotFoundErr{args: []interface{}{id}}
	} else if err != nil {
		return nil, err
	}

	return &cred, nil
}

// GetByScope returns the user credential matching the given scope, or
// UserCredentialNotFoundErr if no such credential exists.
func (*userCredentials) GetByScope(ctx context.Context, scope UserCredentialScope) (*UserCredential, error) {
	if Mocks.UserCredentials.GetByScope != nil {
		return Mocks.UserCredentials.GetByScope(ctx, scope)
	}

	q := sqlf.Sprintf(
		userCredentialsGetByScopeQueryFmtstr,
		sqlf.Join(userCredentialsColumns, ", "),
		scope.Domain,
		scope.UserID,
		scope.ExternalServiceType,
		scope.ExternalServiceID,
	)

	cred := UserCredential{}
	row := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err := scanUserCredential(&cred, row); err == sql.ErrNoRows {
		return nil, UserCredentialNotFoundErr{args: []interface{}{scope}}
	} else if err != nil {
		return nil, err
	}

	return &cred, nil
}

// UserCredentialsListOpts provide the options when listing credentials. At
// least one field in Scope must be set.
type UserCredentialsListOpts struct {
	*LimitOffset
	Scope UserCredentialScope
}

// sql overrides LimitOffset.SQL() to give a LIMIT clause with one extra value
// so we can populate the next cursor.
func (opts *UserCredentialsListOpts) sql() *sqlf.Query {
	if opts.LimitOffset == nil || opts.Limit == 0 {
		return &sqlf.Query{}
	}

	return (&LimitOffset{Limit: opts.Limit + 1, Offset: opts.Offset}).SQL()
}

// List returns all user credentials matching the given options.
func (*userCredentials) List(ctx context.Context, opts UserCredentialsListOpts) ([]*UserCredential, int, error) {
	if Mocks.UserCredentials.List != nil {
		return Mocks.UserCredentials.List(ctx, opts)
	}

	preds := []*sqlf.Query{}
	if opts.Scope.Domain != "" {
		preds = append(preds, sqlf.Sprintf("domain = %s", opts.Scope.Domain))
	}
	if opts.Scope.UserID != 0 {
		preds = append(preds, sqlf.Sprintf("user_id = %s", opts.Scope.UserID))
	}
	if opts.Scope.ExternalServiceType != "" {
		preds = append(preds, sqlf.Sprintf("external_service_type = %s", opts.Scope.ExternalServiceType))
	}
	if opts.Scope.ExternalServiceID != "" {
		preds = append(preds, sqlf.Sprintf("external_service_id = %s", opts.Scope.ExternalServiceID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	q := sqlf.Sprintf(
		userCredentialsListQueryFmtstr,
		sqlf.Join(userCredentialsColumns, ", "),
		sqlf.Join(preds, "\n AND "),
		opts.sql(),
	)

	rows, err := dbconn.Global.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var creds []*UserCredential
	for rows.Next() {
		cred := UserCredential{}
		if err := scanUserCredential(&cred, rows); err != nil {
			return nil, 0, err
		}
		creds = append(creds, &cred)
	}

	// Check if there were more results than the limit: if so, then we need to
	// set the return cursor and lop off the extra credential that we retrieved.
	next := 0
	if opts.LimitOffset != nil && opts.Limit != 0 && len(creds) == opts.Limit+1 {
		next = opts.Offset + opts.Limit
		creds = creds[:len(creds)-1]
	}

	return creds, next, nil
}

// 🐉 This marks the end of the public API. Beyond here are dragons.

// userCredentialsColumns are the columns that must be selected by
// user_credentials queries in order to use scanUserCredential().
var userCredentialsColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("domain"),
	sqlf.Sprintf("user_id"),
	sqlf.Sprintf("external_service_type"),
	sqlf.Sprintf("external_service_id"),
	sqlf.Sprintf("credential"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

// The more unwieldy queries are below rather than inline in the above methods
// in a vain attempt to improve their readability.

const userCredentialsGetByScopeQueryFmtstr = `
-- source: internal/db/user_credentials.go:GetByScope
SELECT %s
FROM user_credentials
WHERE
	domain = %s AND
	user_id = %s AND
	external_service_type = %s AND
	external_service_id = %s
`

const userCredentialsListQueryFmtstr = `
-- source: internal/db/user_credentials.go:List
SELECT %s
FROM user_credentials
WHERE %s
ORDER BY created_at ASC, domain ASC, user_id ASC, external_service_id ASC
%s  -- LIMIT clause
`

const userCredentialsCreateQueryFmtstr = `
-- source: internal/db/user_credentials.go:Upsert
INSERT INTO
	user_credentials (
		domain,
		user_id,
		external_service_type,
		external_service_id,
		credential,
		created_at,
		updated_at
	)
	VALUES (
		%s,
		%s,
		%s,
		%s,
		%s,
		NOW(),
		NOW()
	)
	RETURNING %s
`

// scanUserCredential scans a credential from the given scanner into the given
// credential.
//
// s is inspired by the campaigns scanner type, but also matches sql.Row, which
// is generally used directly in this module.
func scanUserCredential(cred *UserCredential, s interface {
	Scan(...interface{}) error
}) error {
	// Set up a string for the credential to be decrypted into.
	raw := ""

	if err := s.Scan(
		&cred.ID,
		&cred.Domain,
		&cred.UserID,
		&cred.ExternalServiceType,
		&cred.ExternalServiceID,
		&raw,
		&cred.CreatedAt,
		&cred.UpdatedAt,
	); err != nil {
		return err
	}

	// Now we have the credential, we need to unmarshal it into the right Go
	// type.
	var err error
	if cred.Credential, err = unmarshalCredential(raw); err != nil {
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

// marshalCredential encodes an Authenticator into a JSON string.
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

// unmarshalCredential decodes a JSON string into an Authenticator.
func unmarshalCredential(raw string) (auth.Authenticator, error) {
	// We do two unmarshals: the first just to get the type, and then the second
	// to actually unmarshal the authenticator itself.
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

	if err := json.Unmarshal(partial.Auth, &a); err != nil {
		return nil, err
	}

	return a.(auth.Authenticator), nil
}
