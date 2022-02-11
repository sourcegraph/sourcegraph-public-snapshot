package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *Store) CreateSiteCredential(ctx context.Context, c *btypes.SiteCredential, credential auth.Authenticator) (err error) {
	ctx, endObservation := s.operations.createSiteCredential.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if c.CreatedAt.IsZero() {
		c.CreatedAt = s.now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	c.Key = s.key
	if err := c.SetAuthenticator(ctx, credential); err != nil {
		return err
	}

	q := createSiteCredentialQuery(c)
	return s.query(ctx, q, func(sc dbutil.Scanner) error {
		return scanSiteCredential(c, sc)
	})
}

var createSiteCredentialQueryFmtstr = `
-- source: enterprise/internal/batches/store/site_credentials.go:CreateSiteCredential
INSERT INTO	batch_changes_site_credentials (
	external_service_type,
	external_service_id,
	credential,
	encryption_key_id,
	created_at,
	updated_at
)
VALUES
	(%s, %s, %s, %s, %s, %s)
RETURNING
	%s
`

func createSiteCredentialQuery(c *btypes.SiteCredential) *sqlf.Query {
	return sqlf.Sprintf(
		createSiteCredentialQueryFmtstr,
		c.ExternalServiceType,
		c.ExternalServiceID,
		c.EncryptedCredential,
		c.EncryptionKeyID,
		c.CreatedAt,
		c.UpdatedAt,
		sqlf.Join(siteCredentialColumns, ","),
	)
}

func (s *Store) DeleteSiteCredential(ctx context.Context, id int64) (err error) {
	ctx, endObservation := s.operations.deleteSiteCredential.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(id)),
	}})
	defer endObservation(1, observation.Args{})

	res, err := s.ExecResult(ctx, deleteSiteCredentialQuery(id))
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

var deleteSiteCredentialQueryFmtstr = `
-- source: enterprise/internal/batches/store/site_credentials.go:DeleteSiteCredential
DELETE FROM
	batch_changes_site_credentials
WHERE
	%s
`

func deleteSiteCredentialQuery(id int64) *sqlf.Query {
	return sqlf.Sprintf(
		deleteSiteCredentialQueryFmtstr,
		sqlf.Sprintf("id = %d", id),
	)
}

type GetSiteCredentialOpts struct {
	ID                  int64
	ExternalServiceType string
	ExternalServiceID   string
}

func (s *Store) GetSiteCredential(ctx context.Context, opts GetSiteCredentialOpts) (sc *btypes.SiteCredential, err error) {
	ctx, endObservation := s.operations.getSiteCredential.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(opts.ID)),
	}})
	defer endObservation(1, observation.Args{})

	q := getSiteCredentialQuery(opts)

	cred := btypes.SiteCredential{Key: s.key}
	err = s.query(ctx, q, func(sc dbutil.Scanner) error { return scanSiteCredential(&cred, sc) })
	if err != nil {
		return nil, err
	}

	if cred.ID == 0 {
		return nil, ErrNoResults
	}

	return &cred, nil
}

var getSiteCredentialQueryFmtstr = `
-- source: enterprise/internal/batches/store/site_credentials.go:GetSiteCredential
SELECT
	%s
FROM batch_changes_site_credentials
WHERE
    %s
`

func getSiteCredentialQuery(opts GetSiteCredentialOpts) *sqlf.Query {
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

	return sqlf.Sprintf(
		getSiteCredentialQueryFmtstr,
		sqlf.Join(siteCredentialColumns, ","),
		sqlf.Join(preds, "AND"),
	)
}

type ListSiteCredentialsOpts struct {
	LimitOpts
	ForUpdate bool

	// TODO(batch-changes-site-credential-encryption): remove when no longer
	// needed.
	RequiresMigration bool

	// TODO(batch-changes-site-credential-encryption): remove when no longer
	// needed.
	OnlyEncrypted bool
}

func (s *Store) ListSiteCredentials(ctx context.Context, opts ListSiteCredentialsOpts) (cs []*btypes.SiteCredential, next int64, err error) {
	ctx, endObservation := s.operations.listSiteCredentials.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listSiteCredentialsQuery(opts)

	cs = make([]*btypes.SiteCredential, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		c := btypes.SiteCredential{Key: s.key}
		if err := scanSiteCredential(&c, sc); err != nil {
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

var listSiteCredentialsQueryFmtstr = `
-- source: enterprise/internal/batches/store/site_credentials.go:ListSiteCredentials
SELECT
	%s
FROM batch_changes_site_credentials
WHERE %s
ORDER BY external_service_type ASC, external_service_id ASC
%s  -- optional FOR UPDATE
`

func listSiteCredentialsQuery(opts ListSiteCredentialsOpts) *sqlf.Query {
	preds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opts.RequiresMigration {
		preds = append(preds, sqlf.Sprintf(
			"encryption_key_id IN (%s, %s)",
			btypes.SiteCredentialPlaceholderEncryptionKeyID,
			btypes.SiteCredentialUnmigratedEncryptionKeyID,
		))
	}
	if opts.OnlyEncrypted {
		preds = append(preds, sqlf.Sprintf(
			"encryption_key_id NOT IN ('', %s)",
			btypes.SiteCredentialUnmigratedEncryptionKeyID,
		))
	}

	forUpdate := &sqlf.Query{}
	if opts.ForUpdate {
		forUpdate = sqlf.Sprintf("FOR UPDATE")
	}

	return sqlf.Sprintf(
		listSiteCredentialsQueryFmtstr+opts.ToDB(),
		sqlf.Join(siteCredentialColumns, ","),
		sqlf.Join(preds, "AND"),
		forUpdate,
	)
}

func (s *Store) UpdateSiteCredential(ctx context.Context, c *btypes.SiteCredential) (err error) {
	ctx, endObservation := s.operations.updateSiteCredential.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(c.ID)),
	}})
	defer endObservation(1, observation.Args{})

	c.UpdatedAt = s.now()

	updated := &btypes.SiteCredential{Key: s.key}
	q := s.updateSiteCredentialQuery(c)
	if err := s.query(ctx, q, func(sc dbutil.Scanner) error {
		return scanSiteCredential(updated, sc)
	}); err != nil {
		return err
	}

	if updated.ID == 0 {
		return ErrNoResults
	}
	*c = *updated
	return nil
}

const updateSiteCredentialQueryFmtstr = `
-- source: enterprise/internal/batches/store/site_credentials.go:UpdateSiteCredential
UPDATE
	batch_changes_site_credentials
SET
	external_service_type = %s,
	external_service_id = %s,
	credential = %s,
	encryption_key_id = %s,
	created_at = %s,
	updated_at = %s
WHERE
	id = %s
RETURNING
	%s
`

func (s *Store) updateSiteCredentialQuery(c *btypes.SiteCredential) *sqlf.Query {
	return sqlf.Sprintf(
		updateSiteCredentialQueryFmtstr,
		c.ExternalServiceType,
		c.ExternalServiceID,
		c.EncryptedCredential,
		c.EncryptionKeyID,
		c.CreatedAt,
		c.UpdatedAt,
		c.ID,
		sqlf.Join(siteCredentialColumns, ","),
	)
}

var siteCredentialColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("external_service_type"),
	sqlf.Sprintf("external_service_id"),
	sqlf.Sprintf("credential"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

func scanSiteCredential(c *btypes.SiteCredential, sc dbutil.Scanner) error {
	return sc.Scan(
		&c.ID,
		&c.ExternalServiceType,
		&c.ExternalServiceID,
		&c.EncryptedCredential,
		&c.EncryptionKeyID,
		&dbutil.NullTime{Time: &c.CreatedAt},
		&dbutil.NullTime{Time: &c.UpdatedAt},
	)
}
