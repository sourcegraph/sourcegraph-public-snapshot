package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"go.opentelemetry.io/otel/attribute"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *Store) CreateSiteCredential(ctx context.Context, c *btypes.SiteCredential, credential auth.Authenticator) (err error) {
	ctx, _, endObservation := s.operations.createSiteCredential.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	if c.CreatedAt.IsZero() {
		c.CreatedAt = s.now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	if err := c.SetAuthenticator(ctx, credential); err != nil {
		return err
	}

	q, err := createSiteCredentialQuery(ctx, c, s.key)
	if err != nil {
		return err
	}
	return s.query(ctx, q, func(sc dbutil.Scanner) error {
		return scanSiteCredential(c, s.key, sc)
	})
}

var createSiteCredentialQueryFmtstr = `
INSERT INTO	batch_changes_site_credentials (
	external_service_type,
	external_service_id,
	credential,
	encryption_key_id,
	github_app_id,
	created_at,
	updated_at
)
VALUES
	(%s, %s, %s, %s, %s, %s, %s)
RETURNING
	%s
`

func createSiteCredentialQuery(ctx context.Context, c *btypes.SiteCredential, key encryption.Key) (*sqlf.Query, error) {
	encryptedCredential, keyID, err := c.Credential.Encrypt(ctx, key)
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(
		createSiteCredentialQueryFmtstr,
		c.ExternalServiceType,
		c.ExternalServiceID,
		[]byte(encryptedCredential),
		keyID,
		dbutil.NewNullInt(c.GitHubAppID),
		c.CreatedAt,
		c.UpdatedAt,
		sqlf.Join(siteCredentialColumns, ","),
	), nil
}

func (s *Store) DeleteSiteCredential(ctx context.Context, id int64) (err error) {
	ctx, _, endObservation := s.operations.deleteSiteCredential.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(id)),
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
	ctx, _, endObservation := s.operations.getSiteCredential.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(opts.ID)),
	}})
	defer endObservation(1, observation.Args{})

	q := getSiteCredentialQuery(opts)

	cred := btypes.SiteCredential{}
	err = s.query(ctx, q, func(sc dbutil.Scanner) error { return scanSiteCredential(&cred, s.key, sc) })
	if err != nil {
		return nil, err
	}

	if cred.ID == 0 {
		return nil, ErrNoResults
	}

	return &cred, nil
}

var getSiteCredentialQueryFmtstr = `
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
}

func (s *Store) ListSiteCredentials(ctx context.Context, opts ListSiteCredentialsOpts) (cs []*btypes.SiteCredential, next int64, err error) {
	ctx, _, endObservation := s.operations.listSiteCredentials.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listSiteCredentialsQuery(opts)

	cs = make([]*btypes.SiteCredential, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scanner) (err error) {
		c := btypes.SiteCredential{}
		if err := scanSiteCredential(&c, s.key, sc); err != nil {
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
SELECT
	%s
FROM batch_changes_site_credentials
WHERE %s
ORDER BY external_service_type ASC, external_service_id ASC
%s  -- optional FOR UPDATE
`

func listSiteCredentialsQuery(opts ListSiteCredentialsOpts) *sqlf.Query {
	preds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
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
	ctx, _, endObservation := s.operations.updateSiteCredential.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("ID", int(c.ID)),
	}})
	defer endObservation(1, observation.Args{})

	c.UpdatedAt = s.now()

	updated := &btypes.SiteCredential{}
	q, err := s.updateSiteCredentialQuery(ctx, c, s.key)
	if err != nil {
		return err
	}
	if err := s.query(ctx, q, func(sc dbutil.Scanner) error {
		return scanSiteCredential(updated, s.key, sc)
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
UPDATE
	batch_changes_site_credentials
SET
	external_service_type = %s,
	external_service_id = %s,
	credential = %s,
	encryption_key_id = %s,
	github_app_id = %s,
	created_at = %s,
	updated_at = %s
WHERE
	id = %s
RETURNING
	%s
`

func (s *Store) updateSiteCredentialQuery(ctx context.Context, c *btypes.SiteCredential, key encryption.Key) (*sqlf.Query, error) {
	encryptedCredential, keyID, err := c.Credential.Encrypt(ctx, key)
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(
		updateSiteCredentialQueryFmtstr,
		c.ExternalServiceType,
		c.ExternalServiceID,
		[]byte(encryptedCredential),
		keyID,
		dbutil.NewNullInt(c.GitHubAppID),
		c.CreatedAt,
		c.UpdatedAt,
		c.ID,
		sqlf.Join(siteCredentialColumns, ","),
	), nil
}

var siteCredentialColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("external_service_type"),
	sqlf.Sprintf("external_service_id"),
	sqlf.Sprintf("credential"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("github_app_id"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

func scanSiteCredential(c *btypes.SiteCredential, key encryption.Key, sc dbutil.Scanner) error {
	var (
		encryptedCredential []byte
		keyID               string
	)
	if err := sc.Scan(
		&c.ID,
		&c.ExternalServiceType,
		&c.ExternalServiceID,
		&encryptedCredential,
		&keyID,
		&dbutil.NullInt{N: &c.GitHubAppID},
		&dbutil.NullTime{Time: &c.CreatedAt},
		&dbutil.NullTime{Time: &c.UpdatedAt},
	); err != nil {
		return err
	}

	c.Credential = database.NewEncryptedCredential(string(encryptedCredential), keyID, key)
	return nil
}
