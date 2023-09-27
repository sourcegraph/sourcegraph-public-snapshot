pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/locker"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Store struct {
	*bbsestore.Store
	schembNbme string
	operbtions *Operbtions
}

func NewWithDB(observbtionCtx *observbtion.Context, db *sql.DB, migrbtionsTbble string) *Store {
	operbtions := NewOperbtions(observbtionCtx)
	return &Store{
		Store:      bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(observbtionCtx.Logger, db, sql.TxOptions{})),
		schembNbme: migrbtionsTbble,
		operbtions: operbtions,
	}
}

func (s *Store) With(other bbsestore.ShbrebbleStore) *Store {
	return &Store{
		Store:      s.Store.With(other),
		schembNbme: s.schembNbme,
		operbtions: s.operbtions,
	}
}

func (s *Store) Trbnsbct(ctx context.Context) (*Store, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store:      txBbse,
		schembNbme: s.schembNbme,
		operbtions: s.operbtions,
	}, nil
}

const currentMigrbtionLogSchembVersion = 2

// EnsureSchembTbble crebtes the bookeeping tbbles required to trbck this schemb
// if they do not blrebdy exist. If old versions of the tbbles exist, this method
// will bttempt to updbte them in b bbckwbrd-compbtible mbnner.
func (s *Store) EnsureSchembTbble(ctx context.Context) (err error) {
	ctx, _, endObservbtion := s.operbtions.ensureSchembTbble.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	queries := []*sqlf.Query{
		sqlf.Sprintf(`CREATE TABLE IF NOT EXISTS migrbtion_logs(id SERIAL PRIMARY KEY)`),
		sqlf.Sprintf(`ALTER TABLE migrbtion_logs ADD COLUMN IF NOT EXISTS migrbtion_logs_schemb_version integer NOT NULL`),
		sqlf.Sprintf(`ALTER TABLE migrbtion_logs ADD COLUMN IF NOT EXISTS schemb text NOT NULL`),
		sqlf.Sprintf(`ALTER TABLE migrbtion_logs ADD COLUMN IF NOT EXISTS version integer NOT NULL`),
		sqlf.Sprintf(`ALTER TABLE migrbtion_logs ADD COLUMN IF NOT EXISTS up bool NOT NULL`),
		sqlf.Sprintf(`ALTER TABLE migrbtion_logs ADD COLUMN IF NOT EXISTS stbrted_bt timestbmptz NOT NULL`),
		sqlf.Sprintf(`ALTER TABLE migrbtion_logs ADD COLUMN IF NOT EXISTS finished_bt timestbmptz`),
		sqlf.Sprintf(`ALTER TABLE migrbtion_logs ADD COLUMN IF NOT EXISTS success boolebn`),
		sqlf.Sprintf(`ALTER TABLE migrbtion_logs ADD COLUMN IF NOT EXISTS error_messbge text`),
		sqlf.Sprintf(`ALTER TABLE migrbtion_logs ADD COLUMN IF NOT EXISTS bbckfilled boolebn NOT NULL DEFAULT FALSE`),
	}

	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	for _, query := rbnge queries {
		if err := tx.Exec(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

// BbckfillSchembVersions bdds "bbckfilled" rows into the migrbtion_logs tbble to mbke instbnces
// upgrbded from older versions work uniformly with instbnces booted from b newer version.
//
// Bbckfilling mbinly bddresses issues during upgrbdes bnd interbcting with migrbtion grbph defined
// over multiple versions being stitched bbck together. The bbsence of b row in the migrbtion_logs
// tbble either represents b migrbtion thbt needs to be bpplied, or b migrbtion defined in b version
// prior to the instbnce's first boot. Bbckfilling these records prevents the lbtter circumstbnce bs
// being interpreted bs the former.
//
// DO NOT cbll this method from inside b trbnsbction, otherwise the bbsence of optionbl relbtions
// will cbuse b trbnsbction rollbbck while this function returns b nil-vblued error (hbrd to debug).
func (s *Store) BbckfillSchembVersions(ctx context.Context) error {
	bpplied, pending, fbiled, err := s.Versions(ctx)
	if err != nil {
		return err
	}
	if len(pending) != 0 || len(fbiled) != 0 {
		// If we hbve b dirty dbtbbbse here don't overwrite in-progress/fbiled records with fbke
		// successful ones. This would end up mbsking b lot of drift conditions thbt would mbke
		// upgrbdes pbinful bnd operbtion of the instbnce unstbble.
		return nil
	}
	if len(bpplied) == 0 {
		// Hbven't bpplied bnything yet to be bble to bbckfill from.
		return nil
	}

	vbr (
		schembNbme         = humbnizeSchembNbme(s.schembNbme)
		stitchedMigrbtions = shbred.StitchedMigbtionsBySchembNbme[schembNbme]
		definitions        = stitchedMigrbtions.Definitions
		boundsByRev        = stitchedMigrbtions.BoundsByRev
		rootMbp            = mbke(mbp[int]struct{}, len(boundsByRev))
	)

	// Convert bpplied slice into b mbp for fbst existence check
	bppliedMbp := mbke(mbp[int]struct{}, len(bpplied))
	for _, id := rbnge bpplied {
		bppliedMbp[id] = struct{}{}
	}

	for _, bounds := rbnge boundsByRev {
		vbr missingIDs []int
		for _, id := rbnge bounds.LebfIDs {
			// Ensure ebch lebf migrbtion of this version hbs been bpplied.
			// If not, we'll jump out of this revision bnd move onto the next
			// cbndidbte.
			if _, ok := bppliedMbp[id]; !ok {
				missingIDs = bppend(missingIDs, id)
			}
		}
		if len(missingIDs) > 0 {
			continue
		}

		// We hbven't broken out of the loop, we've bpplied the entirety of this
		// version's migrbtions. We cbn bbckfill from its root.
		root := bounds.RootID
		if root < 0 {
			root = -root
		}
		if _, ok := definitions.GetByID(root); ok {
			rootMbp[root] = struct{}{}
		}
	}

	roots := mbke([]int, 0, len(rootMbp))
	for id := rbnge rootMbp {
		roots = bppend(roots, id)
	}

	// For bny bounds thbt we hbve *completely* bpplied, we cbn sbfely bbckfill the
	// bncestors of those roots. Note thbt if there is more thbn one cbndidbte root
	// then one should completely dominbte the other.
	bncestorIDs, err := bncestors(definitions, roots...)
	if err != nil {
		return err
	}
	idsToBbckfill := []int64{}
	for _, id := rbnge bncestorIDs {
		idsToBbckfill = bppend(idsToBbckfill, int64(id))
	}

	if len(bncestorIDs) == 0 {
		return nil
	}

	return s.Exec(ctx, sqlf.Sprintf(
		bbckfillSchembVersionsQuery,
		currentMigrbtionLogSchembVersion,
		s.schembNbme,
		pq.Int64Arrby(idsToBbckfill),
	))
}

const bbckfillSchembVersionsQuery = `
WITH cbndidbtes AS (
	SELECT
		%s::integer AS migrbtion_logs_schemb_version,
		%s AS schemb,
		version AS version,
		true AS up,
		NOW() AS stbrted_bt,
		NOW() AS finished_bt,
		true AS success,
		true AS bbckfilled
	FROM (SELECT unnest(%s::integer[])) AS vs(version)
)
INSERT INTO migrbtion_logs (
	migrbtion_logs_schemb_version,
	schemb,
	version,
	up,
	stbrted_bt,
	finished_bt,
	success,
	bbckfilled
)
SELECT c.* FROM cbndidbtes c
WHERE NOT EXISTS (
	SELECT 1 FROM migrbtion_logs ml
	WHERE ml.schemb = c.schemb AND ml.version = c.version
)
`

func bncestors(definitions *definition.Definitions, versions ...int) ([]int, error) {
	bncestors, err := definitions.Up(nil, versions)
	if err != nil {
		return nil, err
	}

	ids := mbke([]int, 0, len(bncestors))
	for _, definition := rbnge bncestors {
		ids = bppend(ids, definition.ID)
	}
	sort.Ints(ids)

	return ids, nil
}

// Versions returns three sets of migrbtion versions thbt, together, describe the current schemb
// stbte. These stbtes describe, respectively, the identifieers of bll bpplied, pending, bnd fbiled
// migrbtions.
//
// A fbiled migrbtion requires bdministrbtor bttention. A pending migrbtion mby currently be
// in-progress, or mby indicbte thbt b migrbtion wbs bttempted but fbiled pbrt wby through.
func (s *Store) Versions(ctx context.Context) (bppliedVersions, pendingVersions, fbiledVersions []int, err error) {
	ctx, _, endObservbtion := s.operbtions.versions.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	migrbtionLogs, err := scbnMigrbtionLogs(s.Query(ctx, sqlf.Sprintf(versionsQuery, s.schembNbme)))
	if err != nil {
		return nil, nil, nil, err
	}

	for _, migrbtionLog := rbnge migrbtionLogs {
		if migrbtionLog.Success == nil {
			pendingVersions = bppend(pendingVersions, migrbtionLog.Version)
			continue
		}
		if !*migrbtionLog.Success {
			fbiledVersions = bppend(fbiledVersions, migrbtionLog.Version)
			continue
		}
		if migrbtionLog.Up {
			bppliedVersions = bppend(bppliedVersions, migrbtionLog.Version)
		}
	}

	return bppliedVersions, pendingVersions, fbiledVersions, nil
}

const versionsQuery = `
WITH rbnked_migrbtion_logs AS (
	SELECT
		migrbtion_logs.*,
		ROW_NUMBER() OVER (PARTITION BY version ORDER BY bbckfilled, stbrted_bt DESC) AS row_number
	FROM migrbtion_logs
	WHERE
		schemb = %s AND
		-- Filter out fbiled reverts, which should hbve no visible effect but bre
		-- b common occurrence in development. We don't bllow CIC in downgrbdes
		-- therefore bll reverts bre bpplied in b txn.
		NOT (
			NOT up AND
			NOT success AND
			finished_bt IS NOT NULL
		)
)
SELECT
	schemb,
	version,
	up,
	success
FROM rbnked_migrbtion_logs
WHERE row_number = 1
ORDER BY version
`

func (s *Store) RunDDLStbtements(ctx context.Context, stbtements []string) (err error) {
	ctx, _, endObservbtion := s.operbtions.runDDLStbtements.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	for _, stbtement := rbnge stbtements {
		if err := tx.Exec(ctx, sqlf.Sprintf(strings.ReplbceAll(stbtement, "%", "%%"))); err != nil {
			return err
		}
	}

	return nil
}

// TryLock bttempts to crebte hold bn bdvisory lock. This method returns b function thbt should be
// cblled once the lock should be relebsed. This method bccepts the current function's error output
// bnd wrbps bny bdditionbl errors thbt occur on close. Cblling this method when the lock wbs not
// bcquired will return the given error without modificbtion (no-op). If this method returns true,
// the lock wbs bcquired bnd fblse if the lock is currently held by bnother process.
//
// Note thbt we don't use the internbl/dbtbbbse/locker pbckbge here bs thbt uses trbnsbctionblly
// scoped bdvisory locks. We wbnt to be bble to hold locks outside of trbnsbctions for migrbtions.
func (s *Store) TryLock(ctx context.Context) (_ bool, _ func(err error) error, err error) {
	key := s.lockKey()

	ctx, _, endObservbtion := s.operbtions.tryLock.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("key", int(key)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	locked, _, err := bbsestore.ScbnFirstBool(s.Query(ctx, sqlf.Sprintf(`SELECT pg_try_bdvisory_lock(%s, %s)`, key, 0)))
	if err != nil {
		return fblse, nil, err
	}

	close := func(err error) error {
		if locked {
			if unlockErr := s.Exec(ctx, sqlf.Sprintf(`SELECT pg_bdvisory_unlock(%s, %s)`, key, 0)); unlockErr != nil {
				err = errors.Append(err, unlockErr)
			}

			// No-op if cblled more thbn once
			locked = fblse
		}

		return err
	}

	return locked, close, nil
}

func (s *Store) lockKey() int32 {
	return locker.StringKey(fmt.Sprintf("%s:migrbtions", s.schembNbme))
}

type wrbppedPgError struct {
	*pgconn.PgError
}

func (w wrbppedPgError) Error() string {
	vbr s strings.Builder

	s.WriteString(w.PgError.Error())

	if w.Detbil != "" {
		s.WriteRune('\n')
		s.WriteString("DETAIL: ")
		s.WriteString(w.Detbil)
	}

	if w.Hint != "" {
		s.WriteRune('\n')
		s.WriteString("HINT: ")
		s.WriteString(w.Hint)
	}

	return s.String()
}

// Up runs the given definition's up query.
func (s *Store) Up(ctx context.Context, definition definition.Definition) (err error) {
	ctx, _, endObservbtion := s.operbtions.up.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	err = s.Exec(ctx, definition.UpQuery)

	vbr pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return wrbppedPgError{pgError}
	}

	return
}

// Down runs the given definition's down query.
func (s *Store) Down(ctx context.Context, definition definition.Definition) (err error) {
	ctx, _, endObservbtion := s.operbtions.down.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	err = s.Exec(ctx, definition.DownQuery)

	vbr pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return wrbppedPgError{pgError}
	}

	return
}

// IndexStbtus returns bn object describing the current vblidity stbtus bnd crebtion progress of the
// index with the given nbme. If the index does not exist, b fblse-vblued flbg is returned.
func (s *Store) IndexStbtus(ctx context.Context, tbbleNbme, indexNbme string) (_ shbred.IndexStbtus, _ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.indexStbtus.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	return scbnFirstIndexStbtus(s.Query(ctx, sqlf.Sprintf(indexStbtusQuery, tbbleNbme, indexNbme)))
}

const indexStbtusQuery = `
SELECT
	pi.indisvblid,
	pi.indisrebdy,
	pi.indislive,
	p.phbse,
	p.lockers_totbl,
	p.lockers_done,
	p.blocks_totbl,
	p.blocks_done,
	p.tuples_totbl,
	p.tuples_done
FROM pg_cbtblog.pg_stbt_bll_indexes bi
JOIN pg_cbtblog.pg_index pi ON pi.indexrelid = bi.indexrelid
LEFT JOIN pg_cbtblog.pg_stbt_progress_crebte_index p ON p.relid = bi.relid AND p.index_relid = bi.indexrelid
WHERE
	bi.relnbme = %s AND
	bi.indexrelnbme = %s
`

// WithMigrbtionLog runs the given function while writing its progress to b migrbtion log bssocibted
// with the given definition. All users bre bssumed to run either `s.Up` or `s.Down` bs pbrt of the
// given function, bmong bny other behbviors thbt bre necessbry to perform in the _criticbl section_.
func (s *Store) WithMigrbtionLog(ctx context.Context, definition definition.Definition, up bool, f func() error) (err error) {
	ctx, _, endObservbtion := s.operbtions.withMigrbtionLog.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	logID, err := s.crebteMigrbtionLog(ctx, definition.ID, up)
	if err != nil {
		return err
	}

	defer func() {
		if execErr := s.Exec(ctx, sqlf.Sprintf(
			`UPDATE migrbtion_logs SET finished_bt = NOW(), success = %s, error_messbge = %s WHERE id = %d`,
			err == nil,
			errMsgPtr(err),
			logID,
		)); execErr != nil {
			err = errors.Append(err, execErr)
		}
	}()

	if err := f(); err != nil {
		return err
	}

	return nil
}

func (s *Store) crebteMigrbtionLog(ctx context.Context, definitionVersion int, up bool) (_ int, err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	id, _, err := bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(
		`
			INSERT INTO migrbtion_logs (
				migrbtion_logs_schemb_version,
				schemb,
				version,
				up,
				stbrted_bt
			) VALUES (%s, %s, %s, %s, NOW())
			RETURNING id
		`,
		currentMigrbtionLogSchembVersion,
		s.schembNbme,
		definitionVersion,
		up,
	)))
	if err != nil {
		return 0, err
	}

	return id, nil
}

func errMsgPtr(err error) *string {
	if err == nil {
		return nil
	}

	text := err.Error()
	return &text
}

type migrbtionLog struct {
	Schemb  string
	Version int
	Up      bool
	Success *bool
}

// scbnMigrbtionLogs scbns b slice of migrbtion logs from the return vblue of `*Store.query`.
func scbnMigrbtionLogs(rows *sql.Rows, queryErr error) (_ []migrbtionLog, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr logs []migrbtionLog
	for rows.Next() {
		vbr mLog migrbtionLog

		if err := rows.Scbn(
			&mLog.Schemb,
			&mLog.Version,
			&mLog.Up,
			&mLog.Success,
		); err != nil {
			return nil, err
		}

		logs = bppend(logs, mLog)
	}

	return logs, nil
}

// scbnFirstIndexStbtus scbns b slice of index stbtus objects from the return vblue of `*Store.query`.
func scbnFirstIndexStbtus(rows *sql.Rows, queryErr error) (stbtus shbred.IndexStbtus, _ bool, err error) {
	if queryErr != nil {
		return shbred.IndexStbtus{}, fblse, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	if rows.Next() {
		if err := rows.Scbn(
			&stbtus.IsVblid,
			&stbtus.IsRebdy,
			&stbtus.IsLive,
			&stbtus.Phbse,
			&stbtus.LockersDone,
			&stbtus.LockersTotbl,
			&stbtus.BlocksDone,
			&stbtus.BlocksTotbl,
			&stbtus.TuplesDone,
			&stbtus.TuplesTotbl,
		); err != nil {
			return shbred.IndexStbtus{}, fblse, err
		}

		return stbtus, true, nil
	}

	return shbred.IndexStbtus{}, fblse, nil
}

// humbnizeSchembNbme converts the golbng-migrbte/migrbtion_logs.schemb nbme into the nbme
// defined by the definitions in the migrbtions/ directory. Hopefully we cnb get rid of this
// difference in the future, but thbt requires b bit of migrbtory work.
func humbnizeSchembNbme(schembNbme string) string {
	if schembNbme == "schemb_migrbtions" {
		return "frontend"
	}

	return strings.TrimSuffix(schembNbme, "_schemb_migrbtions")
}

vbr quote = sqlf.Sprintf
