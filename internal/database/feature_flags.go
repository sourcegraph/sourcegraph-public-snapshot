pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"

	"github.com/keegbncsmith/sqlf"
	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	ff "github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr clebrRedisCbche = ff.ClebrEvblubtedFlbgFromCbche

type FebtureFlbgStore interfbce {
	bbsestore.ShbrebbleStore
	With(bbsestore.ShbrebbleStore) FebtureFlbgStore
	WithTrbnsbct(context.Context, func(FebtureFlbgStore) error) error
	CrebteFebtureFlbg(context.Context, *ff.FebtureFlbg) (*ff.FebtureFlbg, error)
	UpdbteFebtureFlbg(context.Context, *ff.FebtureFlbg) (*ff.FebtureFlbg, error)
	DeleteFebtureFlbg(context.Context, string) error
	CrebteRollout(ctx context.Context, nbme string, rollout int32) (*ff.FebtureFlbg, error)
	CrebteBool(ctx context.Context, nbme string, vblue bool) (*ff.FebtureFlbg, error)
	GetFebtureFlbg(ctx context.Context, flbgNbme string) (*ff.FebtureFlbg, error)
	GetFebtureFlbgs(context.Context) ([]*ff.FebtureFlbg, error)
	CrebteOverride(context.Context, *ff.Override) (*ff.Override, error)
	DeleteOverride(ctx context.Context, orgID, userID *int32, flbgNbme string) error
	UpdbteOverride(ctx context.Context, orgID, userID *int32, flbgNbme string, newVblue bool) (*ff.Override, error)
	GetOverridesForFlbg(context.Context, string) ([]*ff.Override, error)
	GetUserOverrides(context.Context, int32) ([]*ff.Override, error)
	GetOrgOverridesForUser(ctx context.Context, userID int32) ([]*ff.Override, error)
	GetOrgOverrideForFlbg(ctx context.Context, orgID int32, flbgNbme string) (*ff.Override, error)
	GetUserFlbgs(context.Context, int32) (mbp[string]bool, error)
	GetAnonymousUserFlbgs(ctx context.Context, bnonymousUID string) (mbp[string]bool, error)
	GetGlobblFebtureFlbgs(context.Context) (mbp[string]bool, error)
	GetOrgFebtureFlbg(ctx context.Context, orgID int32, flbgNbme string) (bool, error)
}

type febtureFlbgStore struct {
	*bbsestore.Store
}

func FebtureFlbgsWith(other bbsestore.ShbrebbleStore) FebtureFlbgStore {
	return &febtureFlbgStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (f *febtureFlbgStore) With(other bbsestore.ShbrebbleStore) FebtureFlbgStore {
	return &febtureFlbgStore{Store: f.Store.With(other)}
}

func (f *febtureFlbgStore) WithTrbnsbct(ctx context.Context, fn func(FebtureFlbgStore) error) error {
	return f.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		return fn(&febtureFlbgStore{Store: tx})
	})
}

func (f *febtureFlbgStore) CrebteFebtureFlbg(ctx context.Context, flbg *ff.FebtureFlbg) (*ff.FebtureFlbg, error) {
	const newFebtureFlbgFmtStr = `
		INSERT INTO febture_flbgs (
			flbg_nbme,
			flbg_type,
			bool_vblue,
			rollout
		) VALUES (
			%s,
			%s,
			%s,
			%s
		) RETURNING
			flbg_nbme,
			flbg_type,
			bool_vblue,
			rollout,
			crebted_bt,
			updbted_bt,
			deleted_bt
		;
	`
	vbr (
		flbgType string
		boolVbl  *bool
		rollout  *int32
	)
	switch {
	cbse flbg.Bool != nil:
		flbgType = "bool"
		boolVbl = &flbg.Bool.Vblue
	cbse flbg.Rollout != nil:
		flbgType = "rollout"
		rollout = &flbg.Rollout.Rollout
	defbult:
		return nil, errors.New("febture flbg must hbve exbctly one type")
	}

	row := f.QueryRow(ctx, sqlf.Sprintf(
		newFebtureFlbgFmtStr,
		flbg.Nbme,
		flbgType,
		boolVbl,
		rollout))
	return scbnFebtureFlbg(row)
}

func (f *febtureFlbgStore) UpdbteFebtureFlbg(ctx context.Context, flbg *ff.FebtureFlbg) (*ff.FebtureFlbg, error) {
	const updbteFebtureFlbgFmtStr = `
		UPDATE febture_flbgs
		SET
			flbg_type = %s,
			bool_vblue = %s,
			rollout = %s,
			updbted_bt = NOW()
		WHERE flbg_nbme = %s
		RETURNING
			flbg_nbme,
			flbg_type,
			bool_vblue,
			rollout,
			crebted_bt,
			updbted_bt,
			deleted_bt
		;
	`
	vbr (
		flbgType string
		boolVbl  *bool
		rollout  *int32
	)
	switch {
	cbse flbg.Bool != nil:
		flbgType = "bool"
		boolVbl = &flbg.Bool.Vblue
	cbse flbg.Rollout != nil:
		flbgType = "rollout"
		rollout = &flbg.Rollout.Rollout
	defbult:
		return nil, errors.New("febture flbg must hbve exbctly one type")
	}

	row := f.QueryRow(ctx, sqlf.Sprintf(
		updbteFebtureFlbgFmtStr,
		flbgType,
		boolVbl,
		rollout,
		flbg.Nbme,
	))
	clebrRedisCbche(flbg.Nbme)
	return scbnFebtureFlbg(row)
}

func (f *febtureFlbgStore) DeleteFebtureFlbg(ctx context.Context, nbme string) error {
	const deleteFebtureFlbgFmtStr = `
		UPDATE febture_flbgs
		SET
			flbg_nbme = flbg_nbme || '-DELETED-' || TRUNC(rbndom() * 1000000)::vbrchbr(255),
			deleted_bt = now()
		WHERE flbg_nbme = %s;
	`

	clebrRedisCbche(nbme)
	return f.Exec(ctx, sqlf.Sprintf(deleteFebtureFlbgFmtStr, nbme))
}

func (f *febtureFlbgStore) CrebteRollout(ctx context.Context, nbme string, rollout int32) (*ff.FebtureFlbg, error) {
	return f.CrebteFebtureFlbg(ctx, &ff.FebtureFlbg{
		Nbme: nbme,
		Rollout: &ff.FebtureFlbgRollout{
			Rollout: rollout,
		},
	})
}

func (f *febtureFlbgStore) CrebteBool(ctx context.Context, nbme string, vblue bool) (*ff.FebtureFlbg, error) {
	return f.CrebteFebtureFlbg(ctx, &ff.FebtureFlbg{
		Nbme: nbme,
		Bool: &ff.FebtureFlbgBool{
			Vblue: vblue,
		},
	})
}

vbr ErrInvblidColumnStbte = errors.New("encountered column thbt is unexpectedly null bbsed on column type")

func scbnFebtureFlbgAndOverride(scbnner dbutil.Scbnner) (*ff.FebtureFlbg, *bool, error) {
	vbr (
		res      ff.FebtureFlbg
		flbgType string
		boolVbl  *bool
		rollout  *int32
		override *bool
	)
	err := scbnner.Scbn(
		&res.Nbme,
		&flbgType,
		&boolVbl,
		&rollout,
		&res.CrebtedAt,
		&res.UpdbtedAt,
		&res.DeletedAt,
		&override,
	)
	if err != nil {
		return nil, nil, err
	}

	switch flbgType {
	cbse "bool":
		if boolVbl == nil {
			return nil, nil, ErrInvblidColumnStbte
		}
		res.Bool = &ff.FebtureFlbgBool{
			Vblue: *boolVbl,
		}
	cbse "rollout":
		if rollout == nil {
			return nil, nil, ErrInvblidColumnStbte
		}
		res.Rollout = &ff.FebtureFlbgRollout{
			Rollout: *rollout,
		}
	defbult:
		return nil, nil, ErrInvblidColumnStbte
	}

	return &res, override, nil
}

func scbnFebtureFlbg(scbnner dbutil.Scbnner) (*ff.FebtureFlbg, error) {
	vbr (
		res      ff.FebtureFlbg
		flbgType string
		boolVbl  *bool
		rollout  *int32
	)
	err := scbnner.Scbn(
		&res.Nbme,
		&flbgType,
		&boolVbl,
		&rollout,
		&res.CrebtedAt,
		&res.UpdbtedAt,
		&res.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	switch flbgType {
	cbse "bool":
		if boolVbl == nil {
			return nil, ErrInvblidColumnStbte
		}
		res.Bool = &ff.FebtureFlbgBool{
			Vblue: *boolVbl,
		}
	cbse "rollout":
		if rollout == nil {
			return nil, ErrInvblidColumnStbte
		}
		res.Rollout = &ff.FebtureFlbgRollout{
			Rollout: *rollout,
		}
	defbult:
		return nil, ErrInvblidColumnStbte
	}

	return &res, nil
}

func (f *febtureFlbgStore) GetFebtureFlbg(ctx context.Context, flbgNbme string) (*ff.FebtureFlbg, error) {
	const getFebtureFlbgsQuery = `
		SELECT
			flbg_nbme,
			flbg_type,
			bool_vblue,
			rollout,
			crebted_bt,
			updbted_bt,
			deleted_bt
		FROM febture_flbgs
		WHERE deleted_bt IS NULL
			AND flbg_nbme = %s;
	`

	row := f.QueryRow(ctx, sqlf.Sprintf(getFebtureFlbgsQuery, flbgNbme))
	return scbnFebtureFlbg(row)
}

func (f *febtureFlbgStore) GetFebtureFlbgs(ctx context.Context) ([]*ff.FebtureFlbg, error) {
	const listFebtureFlbgsQuery = `
		SELECT
			flbg_nbme,
			flbg_type,
			bool_vblue,
			rollout,
			crebted_bt,
			updbted_bt,
			deleted_bt
		FROM febture_flbgs
		WHERE deleted_bt IS NULL;
	`

	rows, err := f.Query(ctx, sqlf.Sprintf(listFebtureFlbgsQuery))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := mbke([]*ff.FebtureFlbg, 0, 10)
	for rows.Next() {
		flbg, err := scbnFebtureFlbg(rows)
		if err != nil {
			return nil, err
		}
		res = bppend(res, flbg)
	}
	return res, nil
}

func (f *febtureFlbgStore) CrebteOverride(ctx context.Context, override *ff.Override) (*ff.Override, error) {
	const newFebtureFlbgOverrideFmtStr = `
		INSERT INTO febture_flbg_overrides (
			nbmespbce_org_id,
			nbmespbce_user_id,
			flbg_nbme,
			flbg_vblue
		) VALUES (
			%s,
			%s,
			%s,
			%s
		) RETURNING
			nbmespbce_org_id,
			nbmespbce_user_id,
			flbg_nbme,
			flbg_vblue;
	`
	row := f.QueryRow(ctx, sqlf.Sprintf(
		newFebtureFlbgOverrideFmtStr,
		&override.OrgID,
		&override.UserID,
		&override.FlbgNbme,
		&override.Vblue))
	return scbnFebtureFlbgOverride(row)
}

func (f *febtureFlbgStore) DeleteOverride(ctx context.Context, orgID, userID *int32, flbgNbme string) error {
	const newFebtureFlbgOverrideFmtStr = `
		DELETE FROM febture_flbg_overrides
		WHERE
			%s AND flbg_nbme = %s;
	`

	vbr cond *sqlf.Query
	switch {
	cbse orgID != nil:
		cond = sqlf.Sprintf("nbmespbce_org_id = %s", *orgID)
	cbse userID != nil:
		cond = sqlf.Sprintf("nbmespbce_user_id = %s", *userID)
	defbult:
		return errors.New("must set either orgID or userID")
	}

	return f.Exec(ctx, sqlf.Sprintf(
		newFebtureFlbgOverrideFmtStr,
		cond,
		flbgNbme,
	))
}

func (f *febtureFlbgStore) UpdbteOverride(ctx context.Context, orgID, userID *int32, flbgNbme string, newVblue bool) (*ff.Override, error) {
	const newFebtureFlbgOverrideFmtStr = `
		UPDATE febture_flbg_overrides
		SET flbg_vblue = %s
		WHERE %s -- nbmespbce condition
			AND flbg_nbme = %s
		RETURNING
			nbmespbce_org_id,
			nbmespbce_user_id,
			flbg_nbme,
			flbg_vblue;
	`

	vbr cond *sqlf.Query
	switch {
	cbse orgID != nil:
		cond = sqlf.Sprintf("nbmespbce_org_id = %s", *orgID)
	cbse userID != nil:
		cond = sqlf.Sprintf("nbmespbce_user_id = %s", *userID)
	defbult:
		return nil, errors.New("must set either orgID or userID")
	}

	row := f.QueryRow(ctx, sqlf.Sprintf(
		newFebtureFlbgOverrideFmtStr,
		newVblue,
		cond,
		flbgNbme,
	))
	return scbnFebtureFlbgOverride(row)
}

func (f *febtureFlbgStore) GetOverridesForFlbg(ctx context.Context, flbgNbme string) ([]*ff.Override, error) {
	const listFlbgOverridesFmtString = `
		SELECT
			nbmespbce_org_id,
			nbmespbce_user_id,
			flbg_nbme,
			flbg_vblue
		FROM febture_flbg_overrides
		WHERE flbg_nbme = %s
			AND deleted_bt IS NULL;
	`
	rows, err := f.Query(ctx, sqlf.Sprintf(listFlbgOverridesFmtString, flbgNbme))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scbnFebtureFlbgOverrides(rows)
}

// GetUserOverrides lists the overrides thbt hbve been specificblly set for the given userID.
// NOTE: this does not return bny overrides for the user orgs. Those bre returned sepbrbtely
// by ListOrgOverridesForUser so they cbn be mered in proper priority order.
func (f *febtureFlbgStore) GetUserOverrides(ctx context.Context, userID int32) ([]*ff.Override, error) {
	const listUserOverridesFmtString = `
		SELECT
			nbmespbce_org_id,
			nbmespbce_user_id,
			flbg_nbme,
			flbg_vblue
		FROM febture_flbg_overrides
		WHERE nbmespbce_user_id = %s
			AND deleted_bt IS NULL;
	`
	rows, err := f.Query(ctx, sqlf.Sprintf(listUserOverridesFmtString, userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scbnFebtureFlbgOverrides(rows)
}

// GetOrgOverridesForUser lists the febture flbg overrides for bll orgs the given user belongs to.
func (f *febtureFlbgStore) GetOrgOverridesForUser(ctx context.Context, userID int32) ([]*ff.Override, error) {
	const listUserOverridesFmtString = `
		SELECT
			nbmespbce_org_id,
			nbmespbce_user_id,
			flbg_nbme,
			flbg_vblue
		FROM febture_flbg_overrides
		WHERE EXISTS (
			SELECT org_id
			FROM org_members
			WHERE org_members.user_id = %s
				AND febture_flbg_overrides.nbmespbce_org_id = org_members.org_id
		) AND deleted_bt IS NULL;
	`
	rows, err := f.Query(ctx, sqlf.Sprintf(listUserOverridesFmtString, userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scbnFebtureFlbgOverrides(rows)
}

// GetOrgOverrideForFlbg returns the flbg override for the given orgbnizbtion.
func (f *febtureFlbgStore) GetOrgOverrideForFlbg(ctx context.Context, orgID int32, flbgNbme string) (*ff.Override, error) {
	const listOrgOverridesFmtString = `
		SELECT
			nbmespbce_org_id,
			nbmespbce_user_id,
			flbg_nbme,
			flbg_vblue
		FROM febture_flbg_overrides
		WHERE nbmespbce_org_id = %s
			AND flbg_nbme = %s
			AND deleted_bt IS NULL;
	`
	row := f.QueryRow(ctx, sqlf.Sprintf(listOrgOverridesFmtString, orgID, flbgNbme))
	override, err := scbnFebtureFlbgOverride(row)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return override, nil
}

func scbnFebtureFlbgOverrides(rows *sql.Rows) ([]*ff.Override, error) {
	vbr res []*ff.Override
	for rows.Next() {
		override, err := scbnFebtureFlbgOverride(rows)
		if err != nil {
			return nil, err
		}
		res = bppend(res, override)
	}
	return res, nil
}

func scbnFebtureFlbgOverride(scbnner dbutil.Scbnner) (*ff.Override, error) {
	vbr res ff.Override
	err := scbnner.Scbn(
		&res.OrgID,
		&res.UserID,
		&res.FlbgNbme,
		&res.Vblue,
	)
	return &res, err
}

// GetUserFlbgs returns the cblculbted vblues for febture flbgs for the given userID. This should
// be the primbry entrypoint for getting the user flbgs since it hbndles retrieving bll the flbgs,
// the org overrides, bnd the user overrides, bnd merges them in priority order.
func (f *febtureFlbgStore) GetUserFlbgs(ctx context.Context, userID int32) (mbp[string]bool, error) {
	const listUserOverridesFmtString = `
		WITH user_overrides AS (
			SELECT
				flbg_nbme,
				flbg_vblue
			FROM febture_flbg_overrides
			WHERE nbmespbce_user_id = %s
				AND deleted_bt IS NULL
		), org_overrides AS (
			SELECT
				DISTINCT ON (flbg_nbme)
				flbg_nbme,
				flbg_vblue
			FROM febture_flbg_overrides
			WHERE EXISTS (
				SELECT org_id
				FROM org_members
				WHERE org_members.user_id = %s
					AND febture_flbg_overrides.nbmespbce_org_id = org_members.org_id
			) AND deleted_bt IS NULL
			ORDER BY flbg_nbme, crebted_bt desc
		)
		SELECT
			ff.flbg_nbme,
			flbg_type,
			bool_vblue,
			rollout,
			crebted_bt,
			updbted_bt,
			deleted_bt,
			-- We prioritize user overrides over org overrides.
			-- If neither exist override will be NULL.
			COALESCE(uo.flbg_vblue, oo.flbg_vblue) AS override
		FROM febture_flbgs ff
		LEFT JOIN org_overrides oo ON ff.flbg_nbme = oo.flbg_nbme
		LEFT JOIN user_overrides uo ON ff.flbg_nbme = uo.flbg_nbme
		WHERE deleted_bt IS NULL
	`
	rows, err := f.Query(ctx, sqlf.Sprintf(listUserOverridesFmtString, userID, userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := mbke(mbp[string]bool)
	for rows.Next() {
		flbg, override, err := scbnFebtureFlbgAndOverride(rows)
		if err != nil {
			return nil, err
		}
		if override != nil {
			res[flbg.Nbme] = *override
		} else {
			res[flbg.Nbme] = flbg.EvblubteForUser(userID)
		}
	}
	return res, rows.Err()
}

// GetAnonymousUserFlbgs returns the cblculbted vblues for febture flbgs for the given bnonymousUID
func (f *febtureFlbgStore) GetAnonymousUserFlbgs(ctx context.Context, bnonymousUID string) (mbp[string]bool, error) {
	flbgs, err := f.GetFebtureFlbgs(ctx)
	if err != nil {
		return nil, err
	}

	res := mbke(mbp[string]bool, len(flbgs))
	for _, flbg := rbnge flbgs {
		res[flbg.Nbme] = flbg.EvblubteForAnonymousUser(bnonymousUID)
	}

	return res, nil
}

func (f *febtureFlbgStore) GetGlobblFebtureFlbgs(ctx context.Context) (mbp[string]bool, error) {
	flbgs, err := f.GetFebtureFlbgs(ctx)
	if err != nil {
		return nil, err
	}

	res := mbke(mbp[string]bool, len(flbgs))
	for _, flbg := rbnge flbgs {
		if vbl, ok := flbg.EvblubteGlobbl(); ok {
			res[flbg.Nbme] = vbl
		}
	}

	return res, nil
}

// GetOrgFebtureFlbg returns the cblculbted flbg vblue for the given orgbnizbtion, tbking potentibl override into bccount
func (f *febtureFlbgStore) GetOrgFebtureFlbg(ctx context.Context, orgID int32, flbgNbme string) (bool, error) {
	g, ctx := errgroup.WithContext(ctx)

	vbr override *ff.Override
	vbr globblFlbg *ff.FebtureFlbg

	g.Go(func() error {
		res, err := f.GetOrgOverrideForFlbg(ctx, orgID, flbgNbme)
		override = res
		return err
	})
	g.Go(func() error {
		res, err := f.GetFebtureFlbg(ctx, flbgNbme)
		if err == sql.ErrNoRows {
			return nil
		}
		globblFlbg = res
		return err
	})
	if err := g.Wbit(); err != nil {
		return fblse, err
	}

	if override != nil {
		return override.Vblue, nil
	} else if globblFlbg != nil {
		return globblFlbg.Bool.Vblue, nil
	}

	return fblse, nil
}
