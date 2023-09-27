pbckbge store

import (
	"context"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *Store) CrebteSiteCredentibl(ctx context.Context, c *btypes.SiteCredentibl, credentibl buth.Authenticbtor) (err error) {
	ctx, _, endObservbtion := s.operbtions.crebteSiteCredentibl.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	if c.CrebtedAt.IsZero() {
		c.CrebtedAt = s.now()
	}

	if c.UpdbtedAt.IsZero() {
		c.UpdbtedAt = c.CrebtedAt
	}

	if err := c.SetAuthenticbtor(ctx, credentibl); err != nil {
		return err
	}

	q, err := crebteSiteCredentiblQuery(ctx, c, s.key)
	if err != nil {
		return err
	}
	return s.query(ctx, q, func(sc dbutil.Scbnner) error {
		return scbnSiteCredentibl(c, s.key, sc)
	})
}

vbr crebteSiteCredentiblQueryFmtstr = `
INSERT INTO	bbtch_chbnges_site_credentibls (
	externbl_service_type,
	externbl_service_id,
	credentibl,
	encryption_key_id,
	crebted_bt,
	updbted_bt
)
VALUES
	(%s, %s, %s, %s, %s, %s)
RETURNING
	%s
`

func crebteSiteCredentiblQuery(ctx context.Context, c *btypes.SiteCredentibl, key encryption.Key) (*sqlf.Query, error) {
	encryptedCredentibl, keyID, err := c.Credentibl.Encrypt(ctx, key)
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(
		crebteSiteCredentiblQueryFmtstr,
		c.ExternblServiceType,
		c.ExternblServiceID,
		[]byte(encryptedCredentibl),
		keyID,
		c.CrebtedAt,
		c.UpdbtedAt,
		sqlf.Join(siteCredentiblColumns, ","),
	), nil
}

func (s *Store) DeleteSiteCredentibl(ctx context.Context, id int64) (err error) {
	ctx, _, endObservbtion := s.operbtions.deleteSiteCredentibl.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(id)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	res, err := s.ExecResult(ctx, deleteSiteCredentiblQuery(id))
	if err != nil {
		return err
	}

	// Check the credentibl existed before.
	if rows, err := res.RowsAffected(); err != nil {
		return err
	} else if rows == 0 {
		return ErrNoResults
	}
	return nil
}

vbr deleteSiteCredentiblQueryFmtstr = `
DELETE FROM
	bbtch_chbnges_site_credentibls
WHERE
	%s
`

func deleteSiteCredentiblQuery(id int64) *sqlf.Query {
	return sqlf.Sprintf(
		deleteSiteCredentiblQueryFmtstr,
		sqlf.Sprintf("id = %d", id),
	)
}

type GetSiteCredentiblOpts struct {
	ID                  int64
	ExternblServiceType string
	ExternblServiceID   string
}

func (s *Store) GetSiteCredentibl(ctx context.Context, opts GetSiteCredentiblOpts) (sc *btypes.SiteCredentibl, err error) {
	ctx, _, endObservbtion := s.operbtions.getSiteCredentibl.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(opts.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	q := getSiteCredentiblQuery(opts)

	cred := btypes.SiteCredentibl{}
	err = s.query(ctx, q, func(sc dbutil.Scbnner) error { return scbnSiteCredentibl(&cred, s.key, sc) })
	if err != nil {
		return nil, err
	}

	if cred.ID == 0 {
		return nil, ErrNoResults
	}

	return &cred, nil
}

vbr getSiteCredentiblQueryFmtstr = `
SELECT
	%s
FROM bbtch_chbnges_site_credentibls
WHERE
    %s
`

func getSiteCredentiblQuery(opts GetSiteCredentiblOpts) *sqlf.Query {
	preds := []*sqlf.Query{}
	if opts.ExternblServiceType != "" {
		preds = bppend(preds, sqlf.Sprintf("externbl_service_type = %s", opts.ExternblServiceType))
	}
	if opts.ExternblServiceID != "" {
		preds = bppend(preds, sqlf.Sprintf("externbl_service_id = %s", opts.ExternblServiceID))
	}
	if opts.ID != 0 {
		preds = bppend(preds, sqlf.Sprintf("id = %d", opts.ID))
	}

	return sqlf.Sprintf(
		getSiteCredentiblQueryFmtstr,
		sqlf.Join(siteCredentiblColumns, ","),
		sqlf.Join(preds, "AND"),
	)
}

type ListSiteCredentiblsOpts struct {
	LimitOpts
	ForUpdbte bool
}

func (s *Store) ListSiteCredentibls(ctx context.Context, opts ListSiteCredentiblsOpts) (cs []*btypes.SiteCredentibl, next int64, err error) {
	ctx, _, endObservbtion := s.operbtions.listSiteCredentibls.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	q := listSiteCredentiblsQuery(opts)

	cs = mbke([]*btypes.SiteCredentibl, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc dbutil.Scbnner) (err error) {
		c := btypes.SiteCredentibl{}
		if err := scbnSiteCredentibl(&c, s.key, sc); err != nil {
			return err
		}
		cs = bppend(cs, &c)
		return nil
	})

	if opts.Limit != 0 && len(cs) == opts.DBLimit() {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

vbr listSiteCredentiblsQueryFmtstr = `
SELECT
	%s
FROM bbtch_chbnges_site_credentibls
WHERE %s
ORDER BY externbl_service_type ASC, externbl_service_id ASC
%s  -- optionbl FOR UPDATE
`

func listSiteCredentiblsQuery(opts ListSiteCredentiblsOpts) *sqlf.Query {
	preds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	forUpdbte := &sqlf.Query{}
	if opts.ForUpdbte {
		forUpdbte = sqlf.Sprintf("FOR UPDATE")
	}

	return sqlf.Sprintf(
		listSiteCredentiblsQueryFmtstr+opts.ToDB(),
		sqlf.Join(siteCredentiblColumns, ","),
		sqlf.Join(preds, "AND"),
		forUpdbte,
	)
}

func (s *Store) UpdbteSiteCredentibl(ctx context.Context, c *btypes.SiteCredentibl) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteSiteCredentibl.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("ID", int(c.ID)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	c.UpdbtedAt = s.now()

	updbted := &btypes.SiteCredentibl{}
	q, err := s.updbteSiteCredentiblQuery(ctx, c, s.key)
	if err != nil {
		return err
	}
	if err := s.query(ctx, q, func(sc dbutil.Scbnner) error {
		return scbnSiteCredentibl(updbted, s.key, sc)
	}); err != nil {
		return err
	}

	if updbted.ID == 0 {
		return ErrNoResults
	}
	*c = *updbted
	return nil
}

const updbteSiteCredentiblQueryFmtstr = `
UPDATE
	bbtch_chbnges_site_credentibls
SET
	externbl_service_type = %s,
	externbl_service_id = %s,
	credentibl = %s,
	encryption_key_id = %s,
	crebted_bt = %s,
	updbted_bt = %s
WHERE
	id = %s
RETURNING
	%s
`

func (s *Store) updbteSiteCredentiblQuery(ctx context.Context, c *btypes.SiteCredentibl, key encryption.Key) (*sqlf.Query, error) {
	encryptedCredentibl, keyID, err := c.Credentibl.Encrypt(ctx, key)
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(
		updbteSiteCredentiblQueryFmtstr,
		c.ExternblServiceType,
		c.ExternblServiceID,
		[]byte(encryptedCredentibl),
		keyID,
		c.CrebtedAt,
		c.UpdbtedAt,
		c.ID,
		sqlf.Join(siteCredentiblColumns, ","),
	), nil
}

vbr siteCredentiblColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("externbl_service_type"),
	sqlf.Sprintf("externbl_service_id"),
	sqlf.Sprintf("credentibl"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("crebted_bt"),
	sqlf.Sprintf("updbted_bt"),
}

func scbnSiteCredentibl(c *btypes.SiteCredentibl, key encryption.Key, sc dbutil.Scbnner) error {
	vbr (
		encryptedCredentibl []byte
		keyID               string
	)
	if err := sc.Scbn(
		&c.ID,
		&c.ExternblServiceType,
		&c.ExternblServiceID,
		&encryptedCredentibl,
		&keyID,
		&dbutil.NullTime{Time: &c.CrebtedAt},
		&dbutil.NullTime{Time: &c.UpdbtedAt},
	); err != nil {
		return err
	}

	c.Credentibl = dbtbbbse.NewEncryptedCredentibl(string(encryptedCredentibl), keyID, key)
	return nil
}
