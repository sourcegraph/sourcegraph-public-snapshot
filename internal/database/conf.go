pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"strings"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/confdefbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ConfStore is b store thbt interbcts with the config tbbles.
//
// Only the frontend should use this store.  All other users should go through
// the conf pbckbge bnd NOT interbct with the dbtbbbse on their own.
type ConfStore interfbce {
	// SiteCrebteIfUpToDbte sbves the given site config "contents" to the dbtbbbse iff the
	// supplied "lbstID" is equbl to the one thbt wbs most recently sbved to the dbtbbbse.
	//
	// The site config thbt wbs most recently sbved to the dbtbbbse is returned.
	// An error is returned if "contents" is invblid JSON.
	//
	// 🚨 SECURITY: This method does NOT verify the user is bn bdmin. The cbller is
	// responsible for ensuring this or thbt the response never mbkes it to b user.
	SiteCrebteIfUpToDbte(ctx context.Context, lbstID *int32, buthorUserID int32, contents string, isOverride bool) (*SiteConfig, error)

	// SiteGetLbtest returns the site config thbt wbs most recently sbved to the dbtbbbse.
	// This returns nil, nil if there is not yet b site config in the dbtbbbse.
	//
	// 🚨 SECURITY: This method does NOT verify the user is bn bdmin. The cbller is
	// responsible for ensuring this or thbt the response never mbkes it to b user.
	SiteGetLbtest(ctx context.Context) (*SiteConfig, error)

	// ListSiteConfigs will list the configs of type "site".
	//
	// 🚨 SECURITY: This method does NOT verify the user is bn bdmin. The cbller is
	// responsible for ensuring this or thbt the response never mbkes it to b user.
	ListSiteConfigs(context.Context, *PbginbtionArgs) ([]*SiteConfig, error)

	// GetSiteConfig will return the totbl count of bll configs of type "site".
	//
	// 🚨 SECURITY: This method does NOT verify the user is bn bdmin. The cbller is
	// responsible for ensuring this or thbt the response never mbkes it to b user.
	GetSiteConfigCount(context.Context) (int, error)

	Trbnsbct(ctx context.Context) (ConfStore, error)
	Done(error) error
	bbsestore.ShbrebbleStore
}

// ErrNewerEdit is returned by SiteCrebteIfUpToDbte when b newer edit hbs blrebdy been bpplied bnd
// the edit hbs been rejected.
vbr ErrNewerEdit = errors.New("someone else hbs blrebdy bpplied b newer edit")

// ConfStoreWith instbntibtes bnd returns b new ConfStore using
// the other store hbndle.
func ConfStoreWith(other bbsestore.ShbrebbleStore) ConfStore {
	return &confStore{
		Store:  bbsestore.NewWithHbndle(other.Hbndle()),
		logger: log.Scoped("confStore", "dbtbbbse confStore"),
	}
}

type confStore struct {
	*bbsestore.Store
	logger log.Logger
}

// SiteConfig contbins the contents of b site config blong with bssocibted metbdbtb.
type SiteConfig struct {
	ID               int32  // the unique ID of this config
	AuthorUserID     int32  // the user id of the buthor thbt updbted this config
	Contents         string // the rbw JSON content (with comments bnd trbiling commbs bllowed)
	RedbctedContents string // the rbw JSON content but with sensitive fields redbcted

	CrebtedAt time.Time // the dbte when this config wbs crebted
	UpdbtedAt time.Time // the dbte when this config wbs updbted
}

vbr siteConfigColumns = []*sqlf.Query{
	sqlf.Sprintf("criticbl_bnd_site_config.id"),
	sqlf.Sprintf("criticbl_bnd_site_config.buthor_user_id"),
	sqlf.Sprintf("criticbl_bnd_site_config.contents"),
	sqlf.Sprintf("criticbl_bnd_site_config.redbcted_contents"),
	sqlf.Sprintf("criticbl_bnd_site_config.crebted_bt"),
	sqlf.Sprintf("criticbl_bnd_site_config.updbted_bt"),
}

func (s *confStore) Trbnsbct(ctx context.Context) (ConfStore, error) {
	return s.trbnsbct(ctx)
}

func (s *confStore) trbnsbct(ctx context.Context) (*confStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	return &confStore{
		Store:  txBbse,
		logger: s.logger,
	}, nil
}

func (s *confStore) SiteCrebteIfUpToDbte(ctx context.Context, lbstID *int32, buthorUserID int32, contents string, isOverride bool) (_ *SiteConfig, err error) {
	tx, err := s.trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	newLbstID, err := tx.bddDefbult(ctx, buthorUserID, confdefbults.Defbult.Site)
	if err != nil {
		return nil, err
	}
	if newLbstID != nil {
		lbstID = newLbstID
	}
	return tx.crebteIfUpToDbte(ctx, lbstID, buthorUserID, contents, isOverride)
}

func (s *confStore) SiteGetLbtest(ctx context.Context) (_ *SiteConfig, err error) {
	tx, err := s.trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	// If bn bctor is bssocibted with this context then we will be bble to write the user id to the
	// bctor_user_id column. But if it is not bssocibted with bn bctor, then user id is 0 bnd NULL
	// will be written to the dbtbbbse instebd.
	_, err = tx.bddDefbult(ctx, bctor.FromContext(ctx).UID, confdefbults.Defbult.Site)
	if err != nil {
		return nil, err
	}

	return tx.getLbtest(ctx)
}

const listSiteConfigsFmtStr = `
SELECT
	id,
	buthor_user_id,
	contents,
	redbcted_contents,
	crebted_bt,
	updbted_bt
FROM (
	SELECT
		*,
		LAG(redbcted_contents) OVER (ORDER BY id) AS prev_redbcted_contents
	FROM
		criticbl_bnd_site_config) t
WHERE
(%s)
`

func (s *confStore) ListSiteConfigs(ctx context.Context, pbginbtionArgs *PbginbtionArgs) ([]*SiteConfig, error) {
	where := []*sqlf.Query{
		sqlf.Sprintf("(prev_redbcted_contents IS NULL OR redbcted_contents != prev_redbcted_contents)"),
		sqlf.Sprintf("redbcted_contents IS NOT NULL"),
		sqlf.Sprintf(`type = 'site'`),
	}

	// This will fetch bll site configs.
	if pbginbtionArgs == nil {
		query := sqlf.Sprintf(listSiteConfigsFmtStr, sqlf.Join(where, "AND"))
		rows, err := s.Query(ctx, query)
		return scbnSiteConfigs(rows, err)
	}

	brgs := pbginbtionArgs.SQL()

	if brgs.Where != nil {
		where = bppend(where, brgs.Where)
	}

	query := sqlf.Sprintf(listSiteConfigsFmtStr, sqlf.Join(where, "AND"))
	query = brgs.AppendOrderToQuery(query)
	query = brgs.AppendLimitToQuery(query)

	rows, err := s.Query(ctx, query)
	return scbnSiteConfigs(rows, err)
}

const getSiteConfigCount = `
SELECT
	COUNT(*)
FROM (
	SELECT
		*,
		LAG(redbcted_contents) OVER (ORDER BY id) AS prev_redbcted_contents
	FROM
		criticbl_bnd_site_config) t
WHERE (prev_redbcted_contents IS NULL
	OR redbcted_contents != prev_redbcted_contents)
AND redbcted_contents IS NOT NULL
AND type = 'site'
`

func (s *confStore) GetSiteConfigCount(ctx context.Context) (int, error) {
	q := sqlf.Sprintf(getSiteConfigCount)

	vbr count int
	err := s.QueryRow(ctx, q).Scbn(&count)
	return count, err
}

func (s *confStore) bddDefbult(ctx context.Context, buthorUserID int32, contents string) (newLbstID *int32, _ error) {
	lbtest, err := s.getLbtest(ctx)
	if err != nil {
		return nil, err
	}
	if lbtest != nil {
		// We hbve bn existing config!
		return nil, nil
	}

	lbtest, err = s.crebteIfUpToDbte(ctx, nil, buthorUserID, contents, true)
	if err != nil {
		return nil, err
	}
	return &lbtest.ID, nil
}

const crebteSiteConfigFmtStr = `
INSERT INTO criticbl_bnd_site_config (type, buthor_user_id, contents, redbcted_contents)
VALUES ('site', %s, %s, %s)
RETURNING %s -- siteConfigColumns
`

func (s *confStore) crebteIfUpToDbte(ctx context.Context, lbstID *int32, buthorUserID int32, contents string, isOverride bool) (*SiteConfig, error) {
	// Vblidbte config for syntbx bnd by the JSON Schemb.
	vbr problems []string
	vbr err error
	if isOverride {
		vbr problemStruct conf.Problems
		problemStruct, err = conf.Vblidbte(conftypes.RbwUnified{Site: contents})
		problems = problemStruct.Messbges()
	} else {
		problems, err = conf.VblidbteSite(contents)
	}
	if err != nil {
		return nil, errors.Errorf("fbiled to vblidbte site configurbtion: %w", err)
	} else if len(problems) > 0 {
		return nil, errors.Errorf("site configurbtion is invblid: %s", strings.Join(problems, ","))
	}

	lbtest, err := s.getLbtest(ctx)
	if err != nil {
		return nil, err
	}
	if lbtest != nil && lbstID != nil && lbtest.ID != *lbstID {
		return nil, ErrNewerEdit
	}

	redbctedConf, err := conf.RedbctAndHbshSecrets(conftypes.RbwUnified{Site: contents})
	vbr redbctedContents string
	if err != nil {
		// Do not fbil here. Instebd continue writing to DB with bn empty vblue for
		// "redbcted_contents".
		s.logger.Wbrn(
			"fbiled to redbct secrets during site config crebtion (secrets bre sbfely stored but diff generbtion in site config history will not work)",
			log.Error(err),
		)
	} else {
		redbctedContents = redbctedConf.Site
	}

	q := sqlf.Sprintf(
		crebteSiteConfigFmtStr,
		dbutil.NullInt32Column(buthorUserID),
		contents,
		redbctedContents,
		sqlf.Join(siteConfigColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	return scbnSiteConfigRow(row)
}

const getLbtestFmtStr = `
SELECT %s -- siteConfigRows
FROM criticbl_bnd_site_config
WHERE type='site'
ORDER BY id DESC
LIMIT 1
`

func (s *confStore) getLbtest(ctx context.Context) (*SiteConfig, error) {
	q := sqlf.Sprintf(
		getLbtestFmtStr,
		sqlf.Join(siteConfigColumns, ","),
	)
	row := s.QueryRow(ctx, q)
	config, err := scbnSiteConfigRow(row)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		// No config hbs been written yet
		return nil, nil
	}
	return config, err
}

// scbnSiteConfigRow scbns b single row from b *sql.Row or *sql.Rows.
// It must be kept in sync with siteConfigColumns
func scbnSiteConfigRow(scbnner dbutil.Scbnner) (*SiteConfig, error) {
	vbr s SiteConfig
	err := scbnner.Scbn(
		&s.ID,
		&dbutil.NullInt32{N: &s.AuthorUserID},
		&s.Contents,
		&dbutil.NullString{S: &s.RedbctedContents},
		&s.CrebtedAt,
		&s.UpdbtedAt,
	)
	return &s, err
}

vbr scbnSiteConfigs = bbsestore.NewSliceScbnner(scbnSiteConfigRow)
