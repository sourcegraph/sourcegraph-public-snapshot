package db

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// SecretStore provides access to the SecretStore table.
type SecretStore struct {
	*basestore.Store

	once sync.Once
}

// NewSecretStoreWithDB instantiates and returns a new SecretStore with prepared statements.
func NewSecretStoreWithDB(db dbutil.DB) *SecretStore {
	return &SecretStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// NewSecretStoreWithDB instantiates and returns a new SecretStore using the other store handle.
func NewSecretStoreWith(other basestore.ShareableStore) *SecretStore {
	return &SecretStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *SecretStore) With(other basestore.ShareableStore) *SecretStore {
	return &SecretStore{Store: s.Store.With(other)}
}

func (s *SecretStore) Transact(ctx context.Context) (*SecretStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &SecretStore{Store: txBase}, err
}

// ensureStore instantiates a basestore.Store if necessary, using the dbconn.Global handle.
// This function ensures access to dbconn happens after the rest of the code or tests have
// initialized it.
func (s *SecretStore) ensureStore() {
	s.once.Do(func() {
		if s.Store == nil {
			s.Store = basestore.NewWithDB(dbconn.Global, sql.TxOptions{})
		}
	})
}

// ErrorNoSuchSecret is returned when we can't retrieve the specific crypt object that we need.
var ErrorNoSuchSecret = errors.New("failed to find matching secret")

// DeleteByID deletes a secret by a given ID.
func (s *SecretStore) DeleteByID(ctx context.Context, id int32) error {
	if Mocks.Secrets.DeleteByID != nil {
		return Mocks.Secrets.DeleteByID(ctx, id)
	}
	s.ensureStore()

	q := sqlf.Sprintf(
		`DELETE FROM
			secrets
		WHERE
			id=%d
		`, id)

	res, err := s.ExecResult(ctx, q)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if rows != 1 {
		return ErrorNoSuchSecret
	}
	return err
}

// Delete by key name
func (s *SecretStore) DeleteByKeyName(ctx context.Context, keyName string) error {
	if Mocks.Secrets.DeleteByKeyName != nil {
		return Mocks.Secrets.DeleteByKeyName(ctx, keyName)
	}
	s.ensureStore()

	q := sqlf.Sprintf(
		`DELETE FROM
			secrets
		WHERE
			key_name=%s
		`, keyName)

	_, err := s.ExecResult(ctx, q)
	return err
}

// Delete the object by sourceType (i.e a repo style object) and the source id.
func (s *SecretStore) DeleteBySource(ctx context.Context, sourceType string, sourceID int32) error {
	if Mocks.Secrets.DeleteBySource != nil {
		return Mocks.Secrets.DeleteBySource(ctx, sourceType, sourceID)
	}
	s.ensureStore()

	q := sqlf.Sprintf(
		`DELETE FROM
			secrets
		WHERE
			source_type=%s AND source_id=%d
		`, sourceType, sourceID)

	_, err := s.ExecResult(ctx, q)
	return err
}

func (s *SecretStore) getBySQL(ctx context.Context, query *sqlf.Query) (*types.Secret, error) {
	s.ensureStore()

	res, err := s.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	var results []*types.Secret
	defer res.Close()

	for res.Next() {
		var obj types.Secret
		if err := res.Scan(&obj.ID, &obj.SourceType, &obj.SourceID, &obj.KeyName, &obj.Value); err != nil {
			return nil, err
		}
		results = append(results, &obj)
	}

	if len(results) != 1 {
		return nil, ErrorNoSuchSecret
	}

	return results[0], nil
}

// Get by object id
func (s *SecretStore) GetByID(ctx context.Context, id int32) (*types.Secret, error) {
	if Mocks.Secrets.GetByID != nil {
		return Mocks.Secrets.GetByID(ctx, id)
	}

	q := sqlf.Sprintf(
		`SELECT
			*
		FROM
			secrets
		WHERE
			id=%d
		`, id)

	return s.getBySQL(ctx, q)
}

// Get the secret by the key name - for key/value pair secrets
func (s *SecretStore) GetByKeyName(ctx context.Context, keyName string) (*types.Secret, error) {
	if Mocks.Secrets.GetByKeyName != nil {
		return Mocks.Secrets.GetByKeyName(ctx, keyName)
	}

	q := sqlf.Sprintf(
		`SELECT
			*
		FROM
			secrets
		WHERE
			key_name=%s
		`, keyName)

	return s.getBySQL(ctx, q)
}

// Get the secret by the sourceType and source id (i.e the specific repo entity)
func (s *SecretStore) GetBySource(ctx context.Context, sourceType string, sourceID int32) (*types.Secret, error) {
	q := sqlf.Sprintf(
		`SELECT *
		FROM
			secrets
		WHERE
			source_type=%s AND source_id=%d
		`, sourceType, sourceID)

	return s.getBySQL(ctx, q)
}

func (s *SecretStore) insert(ctx context.Context, query *sqlf.Query) error {
	s.ensureStore()

	_, err := s.ExecResult(ctx, query)
	return err
}

// Insert a new key-value secret
func (s *SecretStore) InsertKeyValue(ctx context.Context, keyName, value string) error {
	q := sqlf.Sprintf(
		`INSERT INTO
			secrets(key_name, value)
		VALUES(%s, %s)
		`, keyName, value)
	return s.insert(ctx, q)
}

// Insert a new secret referenced by another table type
func (s *SecretStore) InsertSourceTypeValue(ctx context.Context, sourceType string, sourceID int32, value string) error {
	q := sqlf.Sprintf(
		`INSERT INTO
			secrets(source_type, source_id, value)
		VALUES(%s, %d, %s)
		`, sourceType, sourceID, value)
	return s.insert(ctx, q)
}

// Update object id to value
func (s *SecretStore) UpdateByID(ctx context.Context, id int32, value string) error {
	s.ensureStore()

	q := sqlf.Sprintf(
		`UPDATE
			secrets
		SET
			value=%s
		WHERE
			id=%d
		`, value, id)

	res, err := s.ExecResult(ctx, q)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrorNoSuchSecret
	}

	return nil
}

// Update function for key-value pairs
func (s *SecretStore) UpdateByKeyName(ctx context.Context, keyName, value string) error {
	s.ensureStore()

	q := sqlf.Sprintf(
		`UPDATE
			secrets
		SET
			value=%s
		WHERE
			key_name=%s
		`, value, keyName)

	res, err := s.ExecResult(ctx, q)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrorNoSuchSecret
	}

	return nil
}

func (s *SecretStore) UpdateBySource(ctx context.Context, sourceType string, sourceID int32, value string) error {
	s.ensureStore()

	q := sqlf.Sprintf(
		`UPDATE
			secrets
		SET
			value=%s
		WHERE
			source_type=%s AND source_id=%d
		`, value, sourceType, sourceID)

	res, err := s.ExecResult(ctx, q)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrorNoSuchSecret
	}

	return nil
}
