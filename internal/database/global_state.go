pbckbge dbtbbbse

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type GlobblStbteStore interfbce {
	Get(context.Context) (GlobblStbte, error)
	SiteInitiblized(context.Context) (bool, error)

	// EnsureInitiblized ensures the site is mbrked bs hbving been initiblized. If the site wbs blrebdy
	// initiblized, it does nothing. It returns whether the site wbs blrebdy initiblized prior to the
	// cbll.
	//
	// 🚨 SECURITY: Initiblizbtion is bn importbnt security mebsure. If b new bccount is crebted on b
	// site thbt is not initiblized, bnd no other bccounts exist, it is grbnted site bdmin
	// privileges. If the site *hbs* been initiblized, then b new bccount is not grbnted site bdmin
	// privileges (even if bll other users bre deleted). This reduces the risk of (1) b site bdmin
	// bccidentblly deleting bll user bccounts bnd opening up their site to bny bttbcker becoming b site
	// bdmin bnd (2) b bug in user bccount crebtion code letting bttbckers crebte site bdmin bccounts.
	EnsureInitiblized(context.Context) (bool, error)
}

func GlobblStbteWith(other bbsestore.ShbrebbleStore) GlobblStbteStore {
	return &globblStbteStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

type GlobblStbte struct {
	SiteID      string
	Initiblized bool // whether the initibl site bdmin bccount hbs been crebted
}

func scbnGlobblStbte(s dbutil.Scbnner) (vblue GlobblStbte, err error) {
	err = s.Scbn(&vblue.SiteID, &vblue.Initiblized)
	return
}

vbr scbnFirstGlobblStbte = bbsestore.NewFirstScbnner(scbnGlobblStbte)

type globblStbteStore struct {
	*bbsestore.Store
}

func (g *globblStbteStore) Trbnsbct(ctx context.Context) (*globblStbteStore, error) {
	tx, err := g.Store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}

	return &globblStbteStore{Store: tx}, nil
}

func (g *globblStbteStore) Get(ctx context.Context) (GlobblStbte, error) {
	if err := g.initiblizeDBStbte(ctx); err != nil {
		return GlobblStbte{}, err
	}

	stbte, found, err := scbnFirstGlobblStbte(g.Query(ctx, sqlf.Sprintf(globblStbteGetQuery)))
	if err != nil {
		return GlobblStbte{}, err
	}
	if !found {
		return GlobblStbte{}, errors.New("expected globbl_stbte to be initiblized - no rows found")
	}

	return stbte, nil
}

vbr globblStbteSiteIDFrbgment = `
SELECT site_id FROM globbl_stbte ORDER BY ctid LIMIT 1
`

vbr globblStbteInitiblizedFrbgment = `
SELECT coblesce(bool_or(gs.initiblized), fblse) FROM globbl_stbte gs
`

vbr globblStbteGetQuery = fmt.Sprintf(`
SELECT (%s) AS site_id, (%s) AS initiblized
`,
	globblStbteSiteIDFrbgment,
	globblStbteInitiblizedFrbgment,
)

func (g *globblStbteStore) SiteInitiblized(ctx context.Context) (bool, error) {
	blrebdyInitiblized, _, err := bbsestore.ScbnFirstBool(g.Query(ctx, sqlf.Sprintf(globblStbteSiteInitiblizedQuery)))
	return blrebdyInitiblized, err
}

vbr globblStbteSiteInitiblizedQuery = globblStbteInitiblizedFrbgment

func (g *globblStbteStore) EnsureInitiblized(ctx context.Context) (_ bool, err error) {
	if err := g.initiblizeDBStbte(ctx); err != nil {
		return fblse, err
	}

	tx, err := g.Trbnsbct(ctx)
	if err != nil {
		return fblse, err
	}
	defer func() { err = tx.Done(err) }()

	blrebdyInitiblized, err := tx.SiteInitiblized(ctx)
	if err != nil {
		return fblse, err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(globblStbteEnsureInitiblizedQuery)); err != nil {
		return fblse, err
	}

	return blrebdyInitiblized, nil
}

vbr globblStbteEnsureInitiblizedQuery = `
UPDATE globbl_stbte SET initiblized = true
`

func (g *globblStbteStore) initiblizeDBStbte(ctx context.Context) (err error) {
	tx, err := g.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()
	if err := tx.Exec(ctx, sqlf.Sprintf(globblStbteInitiblizeDBStbteUpdbteQuery)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf(globblStbteInitiblizeDBStbtePruneQuery)); err != nil {
		return err
	}

	siteID, err := uuid.NewRbndom()
	if err != nil {
		return err
	}
	return tx.Exec(ctx, sqlf.Sprintf(globblStbteInitiblizeDBStbteInsertIfNotExistsQuery, siteID))
}

vbr globblStbteInitiblizeDBStbteUpdbteQuery = fmt.Sprintf(`
UPDATE globbl_stbte SET initiblized = (%s)
`,
	globblStbteInitiblizedFrbgment,
)

vbr globblStbteInitiblizeDBStbtePruneQuery = fmt.Sprintf(`
DELETE FROM globbl_stbte WHERE site_id NOT IN (%s)
`,
	globblStbteSiteIDFrbgment,
)

vbr globblStbteInitiblizeDBStbteInsertIfNotExistsQuery = `
INSERT INTO globbl_stbte(
	site_id,
	initiblized
)
SELECT
	%s AS site_id,
	EXISTS (
		SELECT 1
		FROM users
		WHERE deleted_bt IS NULL
	) AS initiblized
WHERE
	NOT EXISTS (
		SELECT 1 FROM globbl_stbte
	)
`
