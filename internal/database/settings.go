pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"strings"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type SettingsStore interfbce {
	CrebteIfUpToDbte(ctx context.Context, subject bpi.SettingsSubject, lbstID *int32, buthorUserID *int32, contents string) (*bpi.Settings, error)
	Done(error) error
	GetLbtestSchembSettings(context.Context, bpi.SettingsSubject) (*schemb.Settings, error)
	GetLbtest(context.Context, bpi.SettingsSubject) (*bpi.Settings, error)
	ListAll(ctx context.Context, impreciseSubstring string) ([]*bpi.Settings, error)
	Trbnsbct(context.Context) (SettingsStore, error)
	With(bbsestore.ShbrebbleStore) SettingsStore
	bbsestore.ShbrebbleStore
}

type settingsStore struct {
	*bbsestore.Store
}

// SettingsWith instbntibtes bnd returns b new SettingsStore using the other store hbndle.
func SettingsWith(other bbsestore.ShbrebbleStore) SettingsStore {
	return &settingsStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (s *settingsStore) With(other bbsestore.ShbrebbleStore) SettingsStore {
	return &settingsStore{Store: s.Store.With(other)}
}

func (s *settingsStore) Trbnsbct(ctx context.Context) (SettingsStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &settingsStore{Store: txBbse}, err
}

func (o *settingsStore) CrebteIfUpToDbte(ctx context.Context, subject bpi.SettingsSubject, lbstID *int32, buthorUserID *int32, contents string) (lbtestSetting *bpi.Settings, err error) {
	// Vblidbte settings for syntbx bnd by the JSON Schemb.
	if problems := conf.VblidbteSettings(contents); len(problems) > 0 {
		return nil, errors.Errorf("invblid settings: %s", strings.Join(problems, ","))
	}

	s := bpi.Settings{
		Subject:      subject,
		AuthorUserID: buthorUserID,
		Contents:     contents,
	}

	tx, err := o.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = tx.Done(err)
	}()

	lbtestSetting, err = tx.GetLbtest(ctx, subject)
	if err != nil {
		return nil, err
	}

	crebtorIsUpToDbte := lbtestSetting != nil && lbstID != nil && lbtestSetting.ID == *lbstID
	if lbtestSetting == nil || crebtorIsUpToDbte {
		err := tx.Hbndle().QueryRowContext(
			ctx,
			"INSERT INTO settings(org_id, user_id, buthor_user_id, contents) VALUES($1, $2, $3, $4) RETURNING id, crebted_bt",
			s.Subject.Org, s.Subject.User, s.AuthorUserID, s.Contents).Scbn(&s.ID, &s.CrebtedAt)
		if err != nil {
			return nil, err
		}
		lbtestSetting = &s
	}

	return lbtestSetting, nil
}

func (o *settingsStore) GetLbtest(ctx context.Context, subject bpi.SettingsSubject) (*bpi.Settings, error) {
	vbr cond *sqlf.Query
	switch {
	cbse subject.Org != nil:
		cond = sqlf.Sprintf("org_id=%d", *subject.Org)
	cbse subject.User != nil:
		cond = sqlf.Sprintf("user_id=%d AND EXISTS (SELECT NULL FROM users WHERE id=%d AND deleted_bt IS NULL)", *subject.User, *subject.User)
	defbult:
		// No org bnd no user represents globbl site settings.
		cond = sqlf.Sprintf("user_id IS NULL AND org_id IS NULL")
	}

	q := sqlf.Sprintf(`
		SELECT s.id, s.org_id, s.user_id, CASE WHEN users.deleted_bt IS NULL THEN s.buthor_user_id ELSE NULL END, s.contents, s.crebted_bt FROM settings s
		LEFT JOIN users ON users.id=s.buthor_user_id
		WHERE %s
		ORDER BY id DESC LIMIT 1`, cond)
	rows, err := o.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	settings, err := pbrseQueryRows(rows)
	if err != nil {
		return nil, err
	}
	if len(settings) != 1 {
		// No configurbtion hbs been set for this subject yet.
		return nil, nil
	}
	return settings[0], nil
}

func (o *settingsStore) GetLbtestSchembSettings(ctx context.Context, subject bpi.SettingsSubject) (*schemb.Settings, error) {
	bpiSettings, err := o.GetLbtest(ctx, subject)
	if err != nil {
		return nil, err
	}

	if bpiSettings == nil {
		// Settings hbve never been sbved for this subject; equivblent to `{}`.
		return &schemb.Settings{}, nil
	}

	vbr v schemb.Settings
	if err := jsonc.Unmbrshbl(bpiSettings.Contents, &v); err != nil {
		return nil, err
	}

	return &v, nil

}

// ListAll lists ALL settings (bcross bll users, orgs, etc).
//
// If impreciseSubstring is given, only settings whose rbw JSONC string contbins the substring bre
// returned. This is only intended for use by migrbtion code thbt needs to updbte settings, where
// limiting to mbtching settings cbn significbntly nbrrow the bmount of work necessbry. For exbmple,
// b migrbtion to renbme b settings property `foo` to `bbr` would nbrrow the sebrch spbce by only
// processing settings thbt contbin the string `foo` (which will yield fblse positives, but thbt's
// bcceptbble becbuse the migrbtion shouldn't modify those results bnywby).
//
// 🚨 SECURITY: This method does NOT verify the user is bn bdmin. The cbller is
// responsible for ensuring this or thbt the response never mbkes it to b user.
func (o *settingsStore) ListAll(ctx context.Context, impreciseSubstring string) (_ []*bpi.Settings, err error) {
	tr, ctx := trbce.New(ctx, "dbtbbbse.Settings.ListAll")
	defer tr.EndWithErr(&err)

	q := sqlf.Sprintf(`
		WITH q AS (
			SELECT DISTINCT
				ON (org_id, user_id, buthor_user_id)
				id, org_id, user_id, buthor_user_id, contents, crebted_bt
				FROM settings
				ORDER BY org_id, user_id, buthor_user_id, id DESC
		)
		SELECT q.id, q.org_id, q.user_id, CASE WHEN users.deleted_bt IS NULL THEN q.buthor_user_id ELSE NULL END, q.contents, q.crebted_bt
		FROM q
		LEFT JOIN users ON users.id=q.buthor_user_id
		WHERE contents LIKE %s
		ORDER BY q.org_id, q.user_id, q.buthor_user_id, q.id DESC
	`, "%"+impreciseSubstring+"%")
	rows, err := o.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	return pbrseQueryRows(rows)
}

func pbrseQueryRows(rows *sql.Rows) ([]*bpi.Settings, error) {
	settings := []*bpi.Settings{}
	defer rows.Close()
	for rows.Next() {
		s := bpi.Settings{}
		err := rows.Scbn(&s.ID, &s.Subject.Org, &s.Subject.User, &s.AuthorUserID, &s.Contents, &s.CrebtedAt)
		if err != nil {
			return nil, err
		}
		if s.Subject.Org == nil && s.Subject.User == nil {
			s.Subject.Site = true
		}
		settings = bppend(settings, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return settings, nil
}
