package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

type SiteCredential struct {
	ID                  int64
	ExternalServiceType string
	ExternalServiceID   string
	Credential          auth.Authenticator
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (s *Store) CreateGlobalCredential(ctx context.Context, c *SiteCredential) error {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = s.now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	q := createGlobalCredentialQuery(c)
	return s.query(ctx, q, func(sc scanner) error {
		return scanGlobalCredential(c, sc)
	})
}

var createGlobalCredentialQueryFmtstr = `
-- source: enterprise/internal/batches/store/site_credentials.go:CreateGlobalCredential
INSERT INTO
	batch_changes_site_credentials (external_service_type, external_service_id, credential, created_at, updated_at)
VALUES
	(%s, %s, %s, %s, %s)
RETURNING
	id, external_service_type, external_service_id, credential, created_at, updated_at
`

func createGlobalCredentialQuery(c *SiteCredential) *sqlf.Query {
	return sqlf.Sprintf(
		createGlobalCredentialQueryFmtstr,
		c.ExternalServiceType,
		c.ExternalServiceID,
		&database.NullAuthenticator{A: &c.Credential},
		c.CreatedAt,
		c.UpdatedAt,
	)
}

func (s *Store) DeleteGlobalCredential(ctx context.Context, id int64) error {
	res, err := s.ExecResult(ctx, deleteGlobalCredentialQuery(id))
	if err != nil {
		return err
	}

	// Check the credential existed before.
	if rows, err := res.RowsAffected(); err != nil {
		return err
	} else if rows == 0 {
		return ErrNoResults
	}
	return nil
}

var deleteGlobalCredentialQueryFmtstr = `
-- source: enterprise/internal/batches/store/site_credentials.go:DeleteGlobalCredential
DELETE FROM
	batch_changes_site_credentials
WHERE
	%s
`

func deleteGlobalCredentialQuery(id int64) *sqlf.Query {
	preds := []*sqlf.Query{
		sqlf.Sprintf("id = %d", id),
	}
	return sqlf.Sprintf(
		deleteGlobalCredentialQueryFmtstr,
		sqlf.Join(preds, "AND"),
	)
}

type GetGlobalCredentialOpts struct {
	ID                  int64
	ExternalServiceType string
	ExternalServiceID   string
}

func (s *Store) GetGlobalCredential(ctx context.Context, opts GetGlobalCredentialOpts) (*SiteCredential, error) {
	q := getGlobalCredentialQuery(opts)

	var cred SiteCredential
	err := s.query(ctx, q, func(sc scanner) error { return scanGlobalCredential(&cred, sc) })
	if err != nil {
		return nil, err
	}

	if cred.ID == 0 {
		return nil, ErrNoResults
	}

	return &cred, nil
}

var getGlobalCredentialQueryFmtstr = `
-- source: enterprise/internal/batches/store/site_credentials.go:GetGlobalCredential
SELECT
	id, external_service_type, external_service_id, credential, created_at, updated_at
FROM batch_changes_site_credentials
WHERE
    %s
`

func getGlobalCredentialQuery(opts GetGlobalCredentialOpts) *sqlf.Query {
	preds := []*sqlf.Query{}
	if opts.ExternalServiceType != "" {
		preds = append(preds, sqlf.Sprintf("external_service_type = %s", opts.ExternalServiceType))
	}
	if opts.ExternalServiceID != "" {
		preds = append(preds, sqlf.Sprintf("external_service_id = %s", opts.ExternalServiceID))
	}
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("id = %d", opts.ID))
	}

	return sqlf.Sprintf(getGlobalCredentialQueryFmtstr, sqlf.Join(preds, "AND"))
}

type ListGlobalCredentialsOpts struct {
	LimitOpts
}

func (s *Store) ListGlobalCredentials(ctx context.Context, opts ListGlobalCredentialsOpts) (cs []*SiteCredential, next int64, err error) {
	q := listGlobalCredentialsQuery(opts)

	cs = make([]*SiteCredential, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc scanner) (err error) {
		var c SiteCredential
		if err := scanGlobalCredential(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})

	if opts.Limit != 0 && len(cs) == opts.DBLimit() {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

var listGlobalCredentialsQueryFmtstr = `
-- source: enterprise/internal/batches/store/site_credentials.go:ListGlobalCredentials
SELECT
	id, external_service_type, external_service_id, credential, created_at, updated_at
FROM batch_changes_site_credentials
ORDER BY external_service_type ASC, external_service_id ASC
`

func listGlobalCredentialsQuery(opts ListGlobalCredentialsOpts) *sqlf.Query {
	return sqlf.Sprintf(listGlobalCredentialsQueryFmtstr + opts.ToDB())
}

func scanGlobalCredential(c *SiteCredential, sc scanner) error {
	return sc.Scan(
		&c.ID,
		&c.ExternalServiceType,
		&c.ExternalServiceID,
		&database.NullAuthenticator{A: &c.Credential},
		&dbutil.NullTime{Time: &c.CreatedAt},
		&dbutil.NullTime{Time: &c.UpdatedAt},
	)
}
