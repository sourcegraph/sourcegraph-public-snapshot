package bitbucketserver

import (
	"context"
	"database/sql"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
)

// Permissions of an external account to perform an action on the
// given bit set of object IDs of the defined type.
type Permissions struct {
	AccountID int32
	Perm      authz.Perm
	Type      string
	IDs       *roaring.Bitmap
	UpdatedAt time.Time
	ExpiredAt time.Time
}

type store struct {
	db     *sql.DB
	txOpts sql.TxOptions
}

func (s store) LoadPermissions(ctx context.Context, p *Permissions) error {
	q := s.loadQuery(p)
	row := s.db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)

	var ids []byte
	if err := row.Scan(&ids, &p.UpdatedAt, &p.ExpiredAt); err != nil {
		return err
	}

	if p.IDs == nil {
		p.IDs = roaring.NewBitmap()
	}

	return p.IDs.UnmarshalBinary(ids)
}

func (s store) loadQuery(p *Permissions) *sqlf.Query {
	return sqlf.Sprintf(
		loadQueryFmtStr,
		p.AccountID,
		p.Perm,
		p.Type,
	)
}

const loadQueryFmtStr = `
-- source: enterprise/cmd/frontend/internal/authz/bitbucketserver/store.go:postgresStore.LoadPermissions
SELECT
  object_ids,
  updated_at,
  expired_at
FROM external_permissions
WHERE account_id = %s AND permission = %s AND object_type = %s
`

func (s store) UpsertPermissions(ctx context.Context, p *Permissions) error {
	q, err := s.upsertQuery(p)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	return err
}

func (s store) upsertQuery(p *Permissions) (*sqlf.Query, error) {
	ids, err := p.IDs.ToBytes()
	if err != nil {
		return nil, err
	}

	if p.UpdatedAt.IsZero() {
		return nil, errors.New("UpdatedAt timestamp must be set")
	}

	if p.ExpiredAt.IsZero() {
		return nil, errors.New("ExpiredAt timestamp must be set")
	}

	return sqlf.Sprintf(
		upsertQueryFmtStr,
		p.AccountID,
		p.Perm,
		p.Type,
		ids,
		p.UpdatedAt,
		p.ExpiredAt,
	), nil
}

const upsertQueryFmtStr = `
-- source: enterprise/cmd/frontend/internal/authz/bitbucketserver/store.go:postgresStore.UpsertPermissions
INSERT INTO external_permissions
  (account_id, permission, object_type, object_ids, updated_at, expired_at)
VALUES
  (%s, %s, %s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  external_permissions_account_perm_object_unique
DO UPDATE SET
  object_ids = excluded.object_ids,
  updated_at = excluded.updated_at,
  expired_at = excluded.expired_at
`
