package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ExecutorSecret represents a row in the `executor_secrets` table.
type ExecutorSecret struct {
	ID              int64
	Key             string
	Scope           string
	CreatorID       int32
	NamespaceUserID int32
	NamespaceOrgID  int32

	CreatedAt time.Time
	UpdatedAt time.Time

	EncryptedValue *encryption.Encryptable
}

type ExecutorSecretScope string

const (
	ExecutorSecretScopeBatches = "batches"
)

type ExecutorSecretsStore interface {
	basestore.ShareableStore
	With(basestore.ShareableStore) ExecutorSecretsStore
	Transact(context.Context) (ExecutorSecretsStore, error)

	Create(ctx context.Context, scope ExecutorSecretScope, secret *ExecutorSecret, value string) error
	Update(context.Context, ExecutorSecretScope, *ExecutorSecret) error
	Delete(ctx context.Context, scope ExecutorSecretScope, id int64) error
	GetByID(ctx context.Context, scope ExecutorSecretScope, id int64) (*ExecutorSecret, error)
	List(context.Context, ExecutorSecretScope, ExecutorSecretsListOpts) ([]*ExecutorSecret, int, error)
}

// ExecutorSecretsListOpts provide the options when listing secrets.
type ExecutorSecretsListOpts struct {
	*LimitOffset
	NamespaceUserID int32
	NamespaceOrgID  int32
}

// sql overrides LimitOffset.SQL() to give a LIMIT clause with one extra value
// so we can populate the next cursor.
func (opts *ExecutorSecretsListOpts) sql() *sqlf.Query {
	if opts.LimitOffset == nil || opts.Limit == 0 {
		return &sqlf.Query{}
	}

	return (&LimitOffset{Limit: opts.Limit + 1, Offset: opts.Offset}).SQL()
}

// executorSecretsStore provides access to the `executor_secrets` table.
type executorSecretsStore struct {
	logger log.Logger
	*basestore.Store
	key encryption.Key
}

// ExecutorSecretsWith instantiates and returns a new ExecutorSecretsStore using the other store handle.
func ExecutorSecretsWith(logger log.Logger, other basestore.ShareableStore, key encryption.Key) ExecutorSecretsStore {
	return &executorSecretsStore{
		logger: logger,
		Store:  basestore.NewWithHandle(other.Handle()),
		key:    key,
	}
}

func (s *executorSecretsStore) With(other basestore.ShareableStore) ExecutorSecretsStore {
	return &executorSecretsStore{
		logger: s.logger,
		Store:  s.Store.With(other),
		key:    s.key,
	}
}

func (s *executorSecretsStore) Transact(ctx context.Context) (ExecutorSecretsStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &executorSecretsStore{
		logger: s.logger,
		Store:  txBase,
		key:    s.key,
	}, err
}

// Create inserts the given ExecutorSecret into the database.
func (s *executorSecretsStore) Create(ctx context.Context, scope ExecutorSecretScope, secret *ExecutorSecret, value string) error {
	// SECURITY: check that the current user is authorized to create a secret for the given namespace.
	if err := ensureActorHasNamespaceAccess(ctx, NewDBWith(s.logger, s), secret); err != nil {
		return err
	}

	// Set the current actor as the secret creator.
	if secret.CreatorID == 0 {
		secret.CreatorID = actor.FromContext(ctx).UID
	}

	encryptedValue, keyID, err := encryptExecutorSecret(ctx, s.key, value)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		executorSecretCreateQueryFmtstr,
		scope,
		secret.Key,
		encryptedValue, // N.B.: is already a []byte
		keyID,
		&dbutil.NullInt32{N: &secret.NamespaceUserID},
		&dbutil.NullInt32{N: &secret.NamespaceOrgID},
		secret.CreatorID,
		sqlf.Join(executorSecretsColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	if err := scanExecutorSecret(secret, s.key, row); err != nil {
		return err
	}

	return nil
}

// Update updates a secret in the database. If the secret cannot be found,
// an error is returned.
func (s *executorSecretsStore) Update(ctx context.Context, scope ExecutorSecretScope, secret *ExecutorSecret) error {
	// SECURITY: check that the current user is authorized to create a secret for the given namespace.
	if err := ensureActorHasNamespaceAccess(ctx, NewDBWith(s.logger, s), secret); err != nil {
		return err
	}

	secret.UpdatedAt = timeutil.Now()
	encryptedValue, keyID, err := secret.EncryptedValue.Encrypt(ctx, s.key)
	if err != nil {
		return err
	}

	authz, err := executorSecretsAuthzQueryConds(ctx)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(
		executorSecretUpdateQueryFmtstr,
		[]byte(encryptedValue),
		keyID,
		secret.UpdatedAt,
		secret.ID,
		scope,
		authz,
		sqlf.Join(executorSecretsColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	if err := scanExecutorSecret(secret, s.key, row); err != nil {
		return err
	}

	return nil
}

// Delete deletes the given user credential. Note that there is no concept of a
// soft delete with user credentials: once deleted, the relevant records are
// _gone_, so that we don't hold any sensitive data unexpectedly. 💀
func (s *executorSecretsStore) Delete(ctx context.Context, scope ExecutorSecretScope, id int64) error {
	authz, err := executorSecretsAuthzQueryConds(ctx)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf("DELETE FROM executor_secrets WHERE id = %s AND scope = %s AND %s", id, scope, authz)
	res, err := s.ExecResult(ctx, q)
	if err != nil {
		return err
	}

	if rows, err := res.RowsAffected(); err != nil {
		return err
	} else if rows == 0 {
		return UserCredentialNotFoundErr{args: []any{id}}
	}

	return nil
}

// GetByID returns the user credential matching the given ID, or
// UserCredentialNotFoundErr if no such credential exists.
func (s *executorSecretsStore) GetByID(ctx context.Context, scope ExecutorSecretScope, id int64) (*ExecutorSecret, error) {
	authz, err := executorSecretsAuthzQueryConds(ctx)
	if err != nil {
		return nil, err
	}

	q := sqlf.Sprintf(
		"SELECT %s FROM user_credentials WHERE id = %s AND %s",
		sqlf.Join(executorSecretsColumns, ", "),
		id,
		authz,
	)

	secret := ExecutorSecret{}
	row := s.QueryRow(ctx, q)
	if err := scanExecutorSecret(&secret, s.key, row); err == sql.ErrNoRows {
		return nil, UserCredentialNotFoundErr{args: []any{id}}
	} else if err != nil {
		return nil, err
	}

	return &secret, nil
}

// List returns all secrets matching the given options.
func (s *executorSecretsStore) List(ctx context.Context, scope ExecutorSecretScope, opts ExecutorSecretsListOpts) ([]*ExecutorSecret, int, error) {
	authz, err := executorSecretsAuthzQueryConds(ctx)
	if err != nil {
		return nil, 0, err
	}

	preds := []*sqlf.Query{authz}
	if opts.NamespaceOrgID != 0 {
		preds = append(preds, sqlf.Sprintf("namespace_org_id = %s", opts.NamespaceOrgID))
	} else if opts.NamespaceUserID != 0 {
		preds = append(preds, sqlf.Sprintf("namespace_user_id = %s", opts.NamespaceUserID))
	} else {
		preds = append(preds, sqlf.Sprintf("namespace_user_id IS NULL AND namespace_org_id IS NULL"))
	}

	q := sqlf.Sprintf(
		executorSecretsListQueryFmtstr,
		sqlf.Join(executorSecretsColumns, ", "),
		sqlf.Join(preds, "\n AND "),
		opts.sql(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var secrets []*ExecutorSecret
	for rows.Next() {
		secret := ExecutorSecret{}
		if err := scanExecutorSecret(&secret, s.key, rows); err != nil {
			return nil, 0, err
		}
		secrets = append(secrets, &secret)
	}

	// Check if there were more results than the limit: if so, then we need to
	// set the return cursor and lop off the extra secret that we retrieved.
	next := 0
	if opts.LimitOffset != nil && opts.Limit != 0 && len(secrets) == opts.Limit+1 {
		next = opts.Offset + opts.Limit
		secrets = secrets[:len(secrets)-1]
	}

	return secrets, next, nil
}

// executorSecretsColumns are the columns that must be selected by
// executor_secrets queries in order to use scanExecutorSecret().
var executorSecretsColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("key"),
	sqlf.Sprintf("value"),
	sqlf.Sprintf("scope"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("namespace_user_id"),
	sqlf.Sprintf("namespace_org_id"),
	sqlf.Sprintf("creator_id"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

const executorSecretsGetByScopeQueryFmtstr = `
-- source: internal/database/executor_secrets.go:GetByScope
SELECT %s
FROM executor_secrets
WHERE
	scope = %s AND
	id = %s AND
	%s -- authz query conds
`

const executorSecretsListQueryFmtstr = `
-- source: internal/database/executor_secrets.go:List
SELECT %s
FROM executor_secrets
WHERE %s
ORDER BY key ASC
%s  -- LIMIT clause
`

const executorSecretCreateQueryFmtstr = `
INSERT INTO
	executor_secrets (
		scope,
		key,
		value,
		encryption_key_id,
		namespace_user_id,
		namespace_org_id,
		creator_id,
		created_at,
		updated_at
	)
	VALUES (
		%s,
		%s,
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

const executorSecretUpdateQueryFmtstr = `
UPDATE executor_secrets
SET
	value = %s,
	encryption_key_id = %s,
	updated_at = %s
WHERE
	id = %s AND
	scope = %s AND
	%s -- authz query conds
RETURNING %s
`

// scanExecutorSecret scans a secret from the given scanner into the given
// ExecutorSecret.
func scanExecutorSecret(secret *ExecutorSecret, key encryption.Key, s interface {
	Scan(...any) error
}) error {
	var (
		value []byte
		keyID string
	)

	if err := s.Scan(
		&secret.ID,
		&secret.Key,
		&value,
		&secret.Scope,
		&keyID,
		&secret.NamespaceUserID,
		&secret.NamespaceOrgID,
		&secret.CreatorID,
		&secret.CreatedAt,
		&secret.UpdatedAt,
	); err != nil {
		return err
	}

	secret.EncryptedValue = NewEncryptedCredential(string(value), keyID, key)
	return nil
}

func ensureActorHasNamespaceAccess(ctx context.Context, db DB, secret *ExecutorSecret) error {
	a := actor.FromContext(ctx)
	if a.IsInternal() {
		return nil
	}
	if !a.IsAuthenticated() {
		return errors.New("not logged in")
	}

	// TODO: This should use the helpers from the auth package, but it depends on this package.
	if secret.NamespaceOrgID != 0 {
		// Check if the current user is org member.
		resp, err := db.OrgMembers().GetByOrgIDAndUserID(ctx, secret.NamespaceOrgID, a.UID)
		if err != nil {
			if !errcode.IsNotFound(err) {
				return err
			}
			// Not found case: Fall through and eventually end up down at the site-admin
			// check.
		}
		// If membership is found, the user may pass.
		if resp != nil {
			return nil
		}
	} else if secret.NamespaceUserID != 0 {
		// If the actor is the same user as the namespace user, pass. Otherwise
		// fall through and check if they're site-admin.
		if a.UID == secret.NamespaceUserID {
			return nil
		}
	}

	// Check user is site admin.
	user, err := db.Users().GetByID(ctx, a.UID)
	if err != nil {
		return err
	}
	if user == nil || !user.SiteAdmin {
		return errors.New("not site-admin")
	}
	return nil
}

// executorSecretsAuthzQueryConds generates authz query conditions for checking
// access to the secret at the database level.
// Internal actors will always pass.
func executorSecretsAuthzQueryConds(ctx context.Context) (*sqlf.Query, error) {
	a := actor.FromContext(ctx)
	if a.IsInternal() {
		return sqlf.Sprintf("(TRUE)"), nil
	}

	return sqlf.Sprintf(
		executorSecretsAuthzQueryCondsFmtstr,
		a.UID,
		a.UID,
		a.UID,
	), nil
}

const executorSecretsAuthzQueryCondsFmtstr = `
(
	(
		-- user is the same as the actor
		executor_secrets.namespace_user_id = %s
	)
	OR
	(
		-- actor is part of the org
		executor_secrets.namespace_org_id IS NOT NULL
		AND
		EXISTS (
			SELECT 1
			FROM orgs
			JOIN org_members ON org_members.org_id = orgs.id
			WHERE org_members.user_id = %s
		)
	)
	OR
	(
		-- actor is site admin
		EXISTS (
			SELECT 1
			FROM users
			WHERE site_admin = TRUE AND id = %s  -- actor user ID
		)
	)
)
`

func encryptExecutorSecret(ctx context.Context, key encryption.Key, raw string) ([]byte, string, error) {
	data, keyID, err := encryption.MaybeEncrypt(ctx, key, raw)
	return []byte(data), keyID, err
}
