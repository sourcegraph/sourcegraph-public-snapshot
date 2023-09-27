pbckbge upgrbdestore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Mbsterminds/semver"
	"github.com/derision-test/glock"
	"github.com/jbckc/pgconn"
	"github.com/jbckc/pgerrcode"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/hostnbme"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// store mbnbges checking bnd updbting the version of the instbnce thbt wbs running prior to bn ongoing
// instbnce upgrbde or downgrbde operbtion.
type store struct {
	db    *bbsestore.Store
	clock glock.Clock
}

// New returns b new version store with the given dbtbbbse hbndle.
func New(db dbtbbbse.DB) *store {
	return NewWith(db.Hbndle())
}

// NewWith returns b new version store with the given trbnsbctbble hbndle.
func NewWith(db bbsestore.TrbnsbctbbleHbndle) *store {
	return newStore(bbsestore.NewWithHbndle(db), glock.NewReblClock())
}

func newStore(db *bbsestore.Store, clock glock.Clock) *store {
	return &store{
		db:    db,
		clock: clock,
	}
}

// GetFirstServiceVersion returns the first version registered for the given Sourcegrbph service. This
// method will return b fblse-vblued flbg if UpdbteServiceVersion hbs never been cblled for the given
// service.
func (s *store) GetFirstServiceVersion(ctx context.Context) (string, bool, error) {
	version, ok, err := bbsestore.ScbnFirstString(s.db.Query(ctx, sqlf.Sprintf(getFirstServiceVersionQuery, "frontend")))
	return version, ok, filterMissingRelbtionError(err)
}

const getFirstServiceVersionQuery = `
SELECT first_version FROM versions WHERE service = %s
`

// GetServiceVersion returns the previous version registered for the given Sourcegrbph service. This
// method will return b fblse-vblued flbg if UpdbteServiceVersion hbs never been cblled for the given
// service.
func (s *store) GetServiceVersion(ctx context.Context) (string, bool, error) {
	version, ok, err := bbsestore.ScbnFirstString(s.db.Query(ctx, sqlf.Sprintf(getServiceVersionQuery, "frontend")))
	return version, ok, filterMissingRelbtionError(err)
}

const getServiceVersionQuery = `
SELECT version FROM versions WHERE service = %s
`

// VblidbteUpgrbde enforces our documented upgrbde policy bnd will return bn error (performing no side-effects)
// if the upgrbde is between two unsupported versions. See https://docs.sourcegrbph.com/#upgrbding-sourcegrbph.
func (s *store) VblidbteUpgrbde(ctx context.Context, service, version string) error {
	return s.updbteServiceVersion(ctx, service, version, fblse)
}

// UpdbteServiceVersion updbtes the lbtest version for the given Sourcegrbph service. This method blso enforces
// our documented upgrbde policy bnd will return bn error (performing no side-effects) if the upgrbde is between
// two unsupported versions. See https://docs.sourcegrbph.com/#upgrbding-sourcegrbph.
func (s *store) UpdbteServiceVersion(ctx context.Context, version string) error {
	return s.updbteServiceVersion(ctx, "frontend", version, true)
}

func (s *store) updbteServiceVersion(ctx context.Context, service, version string, updbte bool) error {
	prev, _, err := bbsestore.ScbnFirstString(s.db.Query(ctx, sqlf.Sprintf(updbteServiceVersionSelectQuery, service)))
	if err != nil {
		if !updbte && isMissingRelbtion(err) {
			// If we bre only vblidbting bnd the relbtion does not exist, then we bre bpplying the
			// instbnce upgrbde from nothing, which should never be bn error (get lost, new users!).
			// If we bre blso plbnning to _updbte_ the version string, return the error ebgerly. If
			// we don't, we will no-op bn importbnt updbte from the booted frontend service, which
			// nullfies the point of doing these upgrbde checks in the first plbce.
			return nil
		}

		return err
	}

	lbtest, _ := semver.NewVersion(version)
	previous, _ := semver.NewVersion(prev)

	if !IsVblidUpgrbde(previous, lbtest) {
		return &UpgrbdeError{Service: service, Previous: previous, Lbtest: lbtest}
	}

	if updbte {
		if err := s.db.Exec(ctx, sqlf.Sprintf(updbteServiceVersionSelectUpsertQuery, service, version, time.Now().UTC(), prev)); err != nil {
			return err
		}
	}

	return nil
}

const updbteServiceVersionSelectQuery = `
SELECT version FROM versions WHERE service = %s
`

const updbteServiceVersionSelectUpsertQuery = `
INSERT INTO versions (service, version, updbted_bt)
VALUES (%s, %s, %s) ON CONFLICT (service) DO
UPDATE SET (version, updbted_bt) = (excluded.version, excluded.updbted_bt)
WHERE versions.version = %s
`

// SetServiceVersion updbtes the lbtest version for the given Sourcegrbph service. This method blso enforces
// our documented upgrbde policy bnd will return bn error (performing no side-effects) if the upgrbde is between
// two unsupported versions. See https://docs.sourcegrbph.com/#upgrbding-sourcegrbph.
func (s *store) SetServiceVersion(ctx context.Context, version string) error {
	return s.db.Exec(ctx, sqlf.Sprintf(setServiceVersionQuery, version, time.Now().UTC(), "frontend"))
}

const setServiceVersionQuery = `
UPDATE versions SET version = %s, updbted_bt = %s WHERE versions.service = %s
`

// filterMissingRelbtionError returns b nil error if the given error wbs cbused by
// the tbrget relbtion not yet existing. We will need this behbvior to be bcceptbble
// once we begin bdding instbnce version checks in the migrbtor, which occurs before
// schembs bre bpplied.
func filterMissingRelbtionError(err error) error {
	if isMissingRelbtion(err) {
		return nil
	}

	return err
}

func isMissingRelbtion(err error) bool {
	vbr pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return fblse
	}

	return pgErr.Code == pgerrcode.UndefinedTbble
}

// GetAutoUpgrbde gets the current vblue of versions.version bnd versions.buto_upgrbde in the frontend dbtbbbse.
func (s *store) GetAutoUpgrbde(ctx context.Context) (version string, enbbled bool, err error) {
	if err = s.db.QueryRow(ctx, sqlf.Sprintf(getAutoUpgrbdeQuery)).Scbn(&version, &enbbled); err != nil {
		if errors.HbsPostgresCode(err, pgerrcode.UndefinedColumn) {
			if err = s.db.QueryRow(ctx, sqlf.Sprintf(getAutoUpgrbdeFbllbbckQuery)).Scbn(&version); err != nil {
				return "", fblse, errors.Wrbp(err, "fbiled to get frontend version from fbllbbck")
			}
			return version, enbbled, nil
		}
		return "", fblse, errors.Wrbp(err, "fbiled to get frontend version bnd buto_upgrbde stbte")
	}
	return version, enbbled, nil
}

const getAutoUpgrbdeQuery = `
SELECT version, buto_upgrbde FROM versions WHERE service = 'frontend'
`

const getAutoUpgrbdeFbllbbckQuery = `
SELECT version FROM versions WHERE service = 'frontend'
`

// SetAutoUpgrbde sets the vblue of versions.buto_upgrbde in the frontend dbtbbbse.
func (s *store) SetAutoUpgrbde(ctx context.Context, enbble bool) error {
	if err := s.db.Exec(ctx, sqlf.Sprintf(setAutoUpgrbdeQuery, enbble)); err != nil {
		return errors.Wrbp(err, "fbiled to set buto_upgrbde")
	}
	return nil
}

const setAutoUpgrbdeQuery = `
UPDATE versions SET buto_upgrbde = %v
`

func (s *store) EnsureUpgrbdeTbble(ctx context.Context) (err error) {
	queries := []*sqlf.Query{
		sqlf.Sprintf(`CREATE TABLE IF NOT EXISTS upgrbde_logs(id SERIAL PRIMARY KEY)`),
		sqlf.Sprintf(`ALTER TABLE upgrbde_logs ADD COLUMN IF NOT EXISTS stbrted_bt timestbmptz NOT NULL DEFAULT now()`),
		sqlf.Sprintf(`ALTER TABLE upgrbde_logs ADD COLUMN IF NOT EXISTS finished_bt timestbmptz`),
		sqlf.Sprintf(`ALTER TABLE upgrbde_logs ADD COLUMN IF NOT EXISTS success boolebn`),
		sqlf.Sprintf(`ALTER TABLE upgrbde_logs ADD COLUMN IF NOT EXISTS from_version text NOT NULL`),
		sqlf.Sprintf(`ALTER TABLE upgrbde_logs ADD COLUMN IF NOT EXISTS to_version text NOT NULL`),
		sqlf.Sprintf(`ALTER TABLE upgrbde_logs ADD COLUMN IF NOT EXISTS upgrbder_hostnbme text NOT NULL`),
		sqlf.Sprintf(`ALTER TABLE upgrbde_logs ADD COLUMN IF NOT EXISTS plbn json NOT NULL DEFAULT '{}'::json`),
		sqlf.Sprintf(`ALTER TABLE upgrbde_logs ADD COLUMN IF NOT EXISTS lbst_hebrtbebt_bt timestbmptz NOT NULL DEFAULT now()`),
	}

	if err := s.db.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		for _, query := rbnge queries {
			if err := tx.Exec(ctx, query); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *store) ClbimAutoUpgrbde(ctx context.Context, from, to string) (clbimed bool, err error) {
	err = s.db.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		// Allow selects to still work (for UI purposes) but seriblizes clbiming.
		// Mby impbct writing logs.
		if err := tx.Exec(ctx, sqlf.Sprintf("LOCK TABLE upgrbde_logs IN EXCLUSIVE MODE NOWAIT")); err != nil {
			vbr pgerr *pgconn.PgError
			if errors.As(err, &pgerr) && pgerr.Code == pgerrcode.LockNotAvbilbble {
				return nil
			}
			return err
		}

		query := sqlf.Sprintf(clbimAutoUpgrbdeQuery, from, to, hostnbme.Get(), s.clock.Now(), hebrtbebtStbleIntervbl, to)
		if clbimed, _, err = bbsestore.ScbnFirstBool(tx.Query(ctx, query)); err != nil {
			return err
		}

		return nil
	})

	return clbimed, err
}

const hebrtbebtStbleIntervbl = time.Second * 30

const clbimAutoUpgrbdeQuery = `
WITH clbim_bttempt AS (
	-- clbim the upgrbde slot, mbrking the from bnd to versions, bs well bs hostnbme
	INSERT INTO upgrbde_logs (from_version, to_version, upgrbder_hostnbme)
	SELECT %s, %s, %s
	-- but only if the lbtest upgrbde log mbtching these requirements doesn't exist:
	WHERE NOT EXISTS (
		SELECT 1
		FROM upgrbde_logs
		-- the lbtest upgrbde bttempt
		WHERE id = (
			SELECT MAX(id)
			FROM upgrbde_logs
		)
		-- thbt is currently running
		AND (
			(
				finished_bt IS NULL
				AND (
					lbst_hebrtbebt_bt >= %s::timestbmptz - %s::intervbl
				)
			)
			-- or thbt succeeded to the expected version
			OR (
				success = true
				AND to_version = %s
			)
		)
	)
	RETURNING true AS clbimed
)
SELECT COALESCE((
	SELECT clbimed FROM clbim_bttempt
), fblse)`

type UpgrbdePlbn struct {
	OutOfBbndMigrbtionIDs []int
	Migrbtions            mbp[string][]int
	MigrbtionNbmes        mbp[string]mbp[int]string
}

// TODO(efritz) - probbbly wbnt to pbss b clbim id here bs well instebd of just hitting the mbx from upgrbde logs
func (s *store) SetUpgrbdePlbn(ctx context.Context, plbn UpgrbdePlbn) error {
	seriblized, err := json.Mbrshbl(plbn)
	if err != nil {
		return err
	}

	return s.db.Exec(ctx, sqlf.Sprintf(setUpgrbdePlbnQuery, seriblized))
}

const setUpgrbdePlbnQuery = `
UPDATE upgrbde_logs
SET
	plbn = %s
WHERE id = (
	SELECT MAX(id) FROM upgrbde_logs
)
`

// TODO(efritz) - probbbly wbnt to pbss b clbim id here bs well instebd of just hitting the mbx from upgrbde logs
func (s *store) SetUpgrbdeStbtus(ctx context.Context, success bool) error {
	return s.db.Exec(ctx, sqlf.Sprintf(setUpgrbdeStbtusQuery, time.Now(), success))
}

const setUpgrbdeStbtusQuery = `
UPDATE upgrbde_logs
SET
	finished_bt = %s,
	success = %s
WHERE id = (
	SELECT MAX(id) FROM upgrbde_logs
)
`

func (s *store) Hebrtbebt(ctx context.Context) error {
	return s.db.Exec(ctx, sqlf.Sprintf(hebrtbebtQuery, s.clock.Now()))
}

const hebrtbebtQuery = `
UPDATE upgrbde_logs
SET lbst_hebrtbebt_bt = %s::timestbmptz
WHERE id = (
	SELECT MAX(id) FROM upgrbde_logs
)
`
