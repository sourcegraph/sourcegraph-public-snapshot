package db

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

// This object providers access to the encrypted secrets table
type cryptSecrets struct{}

// Returned when we can't retrieve the specific crypt object that we need
type cryptNotFoundError struct {
	args []interface{}
}

// We always try to find *one* object.
func (err cryptNotFoundError) Error() string {
	return "Failed to find matching token."
}

// Delete by object id
func (s *cryptSecrets) Delete(ctx context.Context, id int32) error {
	if Mocks.CryptSecrets.Delete != nil {
		return Mocks.CryptSecrets.Delete(ctx, id)
	}

	sqlQ := sqlf.Sprintf(
		`DELETE FROM
			crypt_secrets
		WHERE
			id=$1
		`, id)

	res, err := dbconn.Global.ExecContext(ctx, sqlQ.Query(sqlf.PostgresBindVar), sqlQ.Args()...)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil

}

// Delete the object by sourceType (i.e a repo style object) and the source id.
func (s *cryptSecrets) DeleteBySourceTypeAndId(ctx context.Context, sourceType string, sourceID int32) error {
	if Mocks.CryptSecrets.DeleteBySourceTypeAndID != nil {
		return Mocks.CryptSecrets.DeleteBySourceTypeAndID(ctx, sourceType, sourceID)
	}

	sqlQ :=
		`DELETE FROM
			crypt_secrets
		WHERE
			source_type=$1 AND source_id=$2
		`

	res, err := dbconn.Global.ExecContext(ctx, sqlQ, sourceType, sourceID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil

}

func (s *cryptSecrets) getBySQL(ctx context.Context, query *sqlf.Query) (*types.CryptSecret, error) {
	res, err := dbconn.Global.QueryContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return nil, err
	}

	var results []*types.CryptSecret
	defer res.Close()

	for res.Next() {
		var obj types.CryptSecret
		if err := res.Scan(&obj.ID, &obj.SourceType, &obj.SourceID, &obj.Value); err != nil {
			return nil, err
		}
		results = append(results, &obj)
	}

	if len(results) != 1 {
		return nil, cryptNotFoundError{}
	}

	return results[0], nil

}

// Get by object id
func (s *cryptSecrets) Get(ctx context.Context, id int32) (*types.CryptSecret, error) {
	if Mocks.CryptSecrets.Get != nil {
		return Mocks.CryptSecrets.Get(ctx, id)
	}

	sqlQ := sqlf.Sprintf(
		`SELECT *
		FROM
			crypt_secrets
		WHERE
			id=$1
		`, id)

	return s.getBySQL(ctx, sqlQ)
}

// Get the secret by the sourceType and source id (i.e the specific repo entity)
func (s *cryptSecrets) GetBySourceTypeAndID(ctx context.Context, sourceType string, sourceID int32) (*types.CryptSecret, error) {
	sqlQ := sqlf.Sprintf(
		`SELECT *
	FROM
		crypt_secrets
	WHERE
		source_type=$1 AND source_id=$2
	`, sourceType, sourceID)

	return s.getBySQL(ctx, sqlQ)

}

// Update object id to value
func (s *cryptSecrets) Update(ctx context.Context, value string, id int32) error {
	sqlQ := sqlf.Sprintf(
		`UPDATE
		crypt_secrets
	SET
		value=$1
	WHERE
		id=$2
	`, value, id)

	res, err := dbconn.Global.ExecContext(ctx, sqlQ.Query(sqlf.PostgresBindVar), sqlQ.Args()...)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *cryptSecrets) UpdateBySourceTypeAndId(ctx context.Context, sourceType string, sourceID, id int32) error {
	sqlQ := sqlf.Sprintf(`UPDATE
		crypt_secrets
	SET
		value=$1
	WHERE
		source_type=$1 AND source_id=$2
	`, sourceType, sourceID)

	res, err := dbconn.Global.ExecContext(ctx, sqlQ.Query(sqlf.PostgresBindVar), sqlQ.Args()...)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil

}
