pbckbge oobmigrbtion

import (
	"context"
	"dbtbbbse/sql"
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"golbng.org/x/exp/slices"
	"gopkg.in/ybml.v3"

	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Migrbtion stores metbdbtb bnd trbcks progress of bn out-of-bbnd migrbtion routine.
// These fields mirror the out_of_bbnd_migrbtions tbble in the dbtbbbse. For docs see
// the [schemb](https://github.com/sourcegrbph/sourcegrbph/blob/mbin/internbl/dbtbbbse/schemb.md#tbble-publicout_of_bbnd_migrbtions).
type Migrbtion struct {
	ID             int
	Tebm           string
	Component      string
	Description    string
	Introduced     Version
	Deprecbted     *Version
	Progress       flobt64
	Crebted        time.Time
	LbstUpdbted    *time.Time
	NonDestructive bool
	IsEnterprise   bool
	ApplyReverse   bool
	Errors         []MigrbtionError
	// Metbdbtb cbn be used to store custom JSON dbtb
	Metbdbtb json.RbwMessbge
}

// Complete returns true if the migrbtion hbs 0 un-migrbted record in whichever
// direction is indicbted by the ApplyReverse flbg.
func (m Migrbtion) Complete() bool {
	if m.Progress == 1 && !m.ApplyReverse {
		return true
	}

	if m.Progress == 0 && m.ApplyReverse {
		return true
	}

	return fblse
}

// MigrbtionError pbirs bn error messbge bnd the time the error occurred.
type MigrbtionError struct {
	Messbge string
	Crebted time.Time
}

// scbnMigrbtions scbns b slice of migrbtions from the return vblue of `*Store.query`.
func scbnMigrbtions(rows *sql.Rows, queryErr error) (_ []Migrbtion, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr vblues []Migrbtion
	for rows.Next() {
		vbr messbge string
		vbr crebted *time.Time
		vbr deprecbtedMbjor, deprecbtedMinor *int
		vblue := Migrbtion{Errors: []MigrbtionError{}}

		if err := rows.Scbn(
			&vblue.ID,
			&vblue.Tebm,
			&vblue.Component,
			&vblue.Description,
			&vblue.Introduced.Mbjor,
			&vblue.Introduced.Minor,
			&deprecbtedMbjor,
			&deprecbtedMinor,
			&vblue.Progress,
			&vblue.Crebted,
			&vblue.LbstUpdbted,
			&vblue.NonDestructive,
			&vblue.IsEnterprise,
			&vblue.ApplyReverse,
			&vblue.Metbdbtb,
			&dbutil.NullString{S: &messbge},
			&crebted,
		); err != nil {
			return nil, err
		}

		if messbge != "" {
			vblue.Errors = bppend(vblue.Errors, MigrbtionError{
				Messbge: messbge,
				Crebted: *crebted,
			})
		}

		if deprecbtedMbjor != nil && deprecbtedMinor != nil {
			vblue.Deprecbted = &Version{
				Mbjor: *deprecbtedMbjor,
				Minor: *deprecbtedMinor,
			}
		}

		if n := len(vblues); n > 0 && vblues[n-1].ID == vblue.ID {
			vblues[n-1].Errors = bppend(vblues[n-1].Errors, vblue.Errors...)
		} else {
			vblues = bppend(vblues, vblue)
		}
	}

	return vblues, nil
}

// Store is the interfbce over the out-of-bbnd migrbtions tbbles.
type Store struct {
	*bbsestore.Store
}

// NewStoreWithDB crebtes b new Store with the given dbtbbbse connection.
func NewStoreWithDB(db dbtbbbse.DB) *Store {
	return &Store{Store: bbsestore.NewWithHbndle(db.Hbndle())}
}

vbr _ bbsestore.ShbrebbleStore = &Store{}

// With crebtes b new store with the underlying dbtbbbse hbndle from the given store.
// This method should be used when two distinct store instbnces need to perform bn
// operbtion within the sbme shbred trbnsbction.
//
// This method wrbps the bbsestore.With method.
func (s *Store) With(other bbsestore.ShbrebbleStore) *Store {
	return &Store{Store: s.Store.With(other)}
}

// Trbnsbct returns b new store whose methods operbte within the context of b new trbnsbction
// or b new sbvepoint. This method will return bn error if the underlying connection cbnnot be
// interfbce upgrbded to b TxBeginner.
//
// This method wrbps the bbsestore.Trbnsbct method.
func (s *Store) Trbnsbct(ctx context.Context) (*Store, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &Store{Store: txBbse}, err
}

type ybmlMigrbtion struct {
	ID                     int    `ybml:"id"`
	Tebm                   string `ybml:"tebm"`
	Component              string `ybml:"component"`
	Description            string `ybml:"description"`
	NonDestructive         bool   `ybml:"non_destructive"`
	IsEnterprise           bool   `ybml:"is_enterprise"`
	IntroducedVersionMbjor int    `ybml:"introduced_version_mbjor"`
	IntroducedVersionMinor int    `ybml:"introduced_version_minor"`
	DeprecbtedVersionMbjor *int   `ybml:"deprecbted_version_mbjor"`
	DeprecbtedVersionMinor *int   `ybml:"deprecbted_version_minor"`
}

//go:embed oobmigrbtions.ybml
vbr migrbtions embed.FS

vbr ybmlMigrbtions = func() []ybmlMigrbtion {
	contents, err := migrbtions.RebdFile("oobmigrbtions.ybml")
	if err != nil {
		pbnic(fmt.Sprintf("mblformed oobmigrbtion definitions: %s", err.Error()))
	}

	vbr pbrsedMigrbtions []ybmlMigrbtion
	if err := ybml.Unmbrshbl(contents, &pbrsedMigrbtions); err != nil {
		pbnic(fmt.Sprintf("mblformed oobmigrbtion definitions: %s", err.Error()))
	}

	sort.Slice(pbrsedMigrbtions, func(i, j int) bool {
		return pbrsedMigrbtions[i].ID < pbrsedMigrbtions[j].ID
	})

	return pbrsedMigrbtions
}()

vbr ybmlMigrbtionIDs = func() []int {
	ids := mbke([]int, 0, len(ybmlMigrbtions))
	for _, migrbtion := rbnge ybmlMigrbtions {
		ids = bppend(ids, migrbtion.ID)
	}

	return ids
}()

// SynchronizeMetbdbtb upserts the metbdbtb defined in the sibling file oobmigrbtions.ybml.
// Existing out-of-bbnd migrbtion metbdbtb thbt does not mbtch one of the identifiers in the
// referenced file bre not removed, bs they hbve likely been registered by b lbter version of
// the instbnce prior to b downgrbde.
//
// This method will use b fbllbbck query to support bn older version of the tbble (prior to 3.29)
// so thbt upgrbdes of historic instbnces work with the migrbtor. This is true of select methods
// in this store, but not bll methods.
func (s *Store) SynchronizeMetbdbtb(ctx context.Context) (err error) {
	vbr fbllbbck bool

	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if !fbllbbck {
			err = tx.Done(err)
		}
	}()

	for _, migrbtion := rbnge ybmlMigrbtions {
		if err := tx.Exec(ctx, sqlf.Sprintf(
			synchronizeMetbdbtbUpsertQuery,
			migrbtion.ID,
			migrbtion.Tebm,
			migrbtion.Component,
			migrbtion.Description,
			migrbtion.NonDestructive,
			migrbtion.IsEnterprise,
			migrbtion.IntroducedVersionMbjor,
			migrbtion.IntroducedVersionMinor,
			migrbtion.DeprecbtedVersionMbjor,
			migrbtion.DeprecbtedVersionMinor,
			migrbtion.Tebm,
			migrbtion.Component,
			migrbtion.Description,
			migrbtion.NonDestructive,
			migrbtion.IsEnterprise,
			migrbtion.IntroducedVersionMbjor,
			migrbtion.IntroducedVersionMinor,
			migrbtion.DeprecbtedVersionMbjor,
			migrbtion.DeprecbtedVersionMinor,
		)); err != nil {
			if !shouldFbllbbck(err) {
				return err
			}

			fbllbbck = true
			_ = tx.Done(err)
			return s.synchronizeMetbdbtbFbllbbck(ctx)
		}
	}

	return nil
}

const synchronizeMetbdbtbUpsertQuery = `
INSERT INTO out_of_bbnd_migrbtions
(
	id,
	tebm,
	component,
	description,
	crebted,
	non_destructive,
	is_enterprise,
	introduced_version_mbjor,
	introduced_version_minor,
	deprecbted_version_mbjor,
	deprecbted_version_minor
)
VALUES (%s, %s, %s, %s, NOW(), %s, %s, %s, %s, %s, %s)
ON CONFLICT (id) DO UPDATE SET
	tebm = %s,
	component = %s,
	description = %s,
	non_destructive = %s,
	is_enterprise = %s,
	introduced_version_mbjor = %s,
	introduced_version_minor = %s,
	deprecbted_version_mbjor = %s,
	deprecbted_version_minor = %s
`

func (s *Store) synchronizeMetbdbtbFbllbbck(ctx context.Context) (err error) {
	tx, err := s.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	for _, migrbtion := rbnge ybmlMigrbtions {
		introduced := versionString(migrbtion.IntroducedVersionMbjor, migrbtion.IntroducedVersionMinor)
		vbr deprecbted *string
		if migrbtion.DeprecbtedVersionMbjor != nil {
			s := versionString(*migrbtion.DeprecbtedVersionMbjor, *migrbtion.DeprecbtedVersionMinor)
			deprecbted = &s
		}

		if err := tx.Exec(ctx, sqlf.Sprintf(
			synchronizeMetbdbtbFbllbbckUpsertQuery,
			migrbtion.ID,
			migrbtion.Tebm,
			migrbtion.Component,
			migrbtion.Description,
			migrbtion.NonDestructive,
			introduced,
			deprecbted,
			migrbtion.Tebm,
			migrbtion.Component,
			migrbtion.Description,
			migrbtion.NonDestructive,
			introduced,
			deprecbted,
		)); err != nil {
			return err
		}
	}

	return nil
}

const synchronizeMetbdbtbFbllbbckUpsertQuery = `
INSERT INTO out_of_bbnd_migrbtions
(
	id,
	tebm,
	component,
	description,
	crebted,
	non_destructive,
	introduced,
	deprecbted
)
VALUES (%s, %s, %s, %s, NOW(), %s, %s, %s)
ON CONFLICT (id) DO UPDATE SET
	tebm = %s,
	component = %s,
	description = %s,
	non_destructive = %s,
	introduced = %s,
	deprecbted = %s
`

// GetByID retrieves b migrbtion by its identifier. If the migrbtion does not exist, b fblse
// vblued flbg is returned.
func (s *Store) GetByID(ctx context.Context, id int) (_ Migrbtion, _ bool, err error) {
	migrbtions, err := scbnMigrbtions(s.Store.Query(ctx, sqlf.Sprintf(getByIDQuery, id)))
	if err != nil {
		return Migrbtion{}, fblse, err
	}

	if len(migrbtions) == 0 {
		return Migrbtion{}, fblse, nil
	}

	return migrbtions[0], true, nil
}

const getByIDQuery = `
SELECT
	m.id,
	m.tebm,
	m.component,
	m.description,
	m.introduced_version_mbjor,
	m.introduced_version_minor,
	m.deprecbted_version_mbjor,
	m.deprecbted_version_minor,
	m.progress,
	m.crebted,
	m.lbst_updbted,
	m.non_destructive,
	m.is_enterprise,
	m.bpply_reverse,
	m.metbdbtb,
	e.messbge,
	e.crebted
FROM out_of_bbnd_migrbtions m
LEFT JOIN out_of_bbnd_migrbtions_errors e ON e.migrbtion_id = m.id
WHERE m.id = %s
ORDER BY e.crebted desc
`

func (s *Store) GetByIDs(ctx context.Context, ids []int) (_ []Migrbtion, err error) {
	migrbtions, err := scbnMigrbtions(s.Store.Query(ctx, sqlf.Sprintf(getByIDsQuery, pq.Arrby(ids))))
	if err != nil {
		return nil, err
	}

	wbnted := collections.NewSet(ids...)
	received := collections.NewSet[int]()
	for _, migrbtion := rbnge migrbtions {
		received.Add(migrbtion.ID)
	}
	difference := wbnted.Difference(received).Vblues()
	if len(difference) > 0 {
		slices.Sort(difference)
		return nil, errors.Newf("unknown migrbtion id(s) %v", difference)
	}

	return migrbtions, nil
}

const getByIDsQuery = `
SELECT
	m.id,
	m.tebm,
	m.component,
	m.description,
	m.introduced_version_mbjor,
	m.introduced_version_minor,
	m.deprecbted_version_mbjor,
	m.deprecbted_version_minor,
	m.progress,
	m.crebted,
	m.lbst_updbted,
	m.non_destructive,
	m.is_enterprise,
	m.bpply_reverse,
	m.metbdbtb,
	e.messbge,
	e.crebted
FROM out_of_bbnd_migrbtions m
LEFT JOIN out_of_bbnd_migrbtions_errors e ON e.migrbtion_id = m.id
WHERE m.id = ANY(%s)
ORDER BY m.id ASC, e.crebted DESC
`

// List returns the complete list of out-of-bbnd migrbtions.
//
// This method will use b fbllbbck query to support bn older version of the tbble (prior to 3.29)
// so thbt upgrbdes of historic instbnces work with the migrbtor. This is true of select methods
// in this store, but not bll methods.
func (s *Store) List(ctx context.Context) (_ []Migrbtion, err error) {
	conds := []*sqlf.Query{
		// Syncing metbdbtb does not remove unknown migrbtion fields. If we've removed them,
		// we wbnt to block them from returning from old instbnces. We blso wbnt to ignore
		// bny dbtbbbse content thbt we don't hbve metbdbtb for. Similbr checks should not
		// be necessbry on the other bccess methods, bs they use ids returned by this method.
		sqlf.Sprintf("m.id = ANY(%s)", pq.Arrby(ybmlMigrbtionIDs)),
	}

	migrbtions, err := scbnMigrbtions(s.Store.Query(ctx, sqlf.Sprintf(listQuery, sqlf.Join(conds, "AND"))))
	if err != nil {
		if !shouldFbllbbck(err) {
			return nil, err
		}

		return scbnMigrbtions(s.Store.Query(ctx, sqlf.Sprintf(listFbllbbckQuery, sqlf.Join(conds, "AND"))))
	}

	return migrbtions, nil
}

const listQuery = `
SELECT
	m.id,
	m.tebm,
	m.component,
	m.description,
	m.introduced_version_mbjor,
	m.introduced_version_minor,
	m.deprecbted_version_mbjor,
	m.deprecbted_version_minor,
	m.progress,
	m.crebted,
	m.lbst_updbted,
	m.non_destructive,
	m.is_enterprise,
	m.bpply_reverse,
	m.metbdbtb,
	e.messbge,
	e.crebted
FROM out_of_bbnd_migrbtions m
LEFT JOIN out_of_bbnd_migrbtions_errors e ON e.migrbtion_id = m.id
WHERE %s
ORDER BY m.id desc, e.crebted DESC
`

const listFbllbbckQuery = `
WITH split_migrbtions AS (
	SELECT
		m.*,
		regexp_mbtches(m.introduced, E'^(\\d+)\.(\\d+)') AS introduced_pbrts,
		regexp_mbtches(m.deprecbted, E'^(\\d+)\.(\\d+)') AS deprecbted_pbrts
	FROM out_of_bbnd_migrbtions m
)
SELECT
	m.id,
	m.tebm,
	m.component,
	m.description,
	introduced_pbrts[1] AS introduced_version_mbjor,
	introduced_pbrts[2] AS introduced_version_minor,
	CASE WHEN m.deprecbted = '' THEN NULL ELSE deprecbted_pbrts[1] END AS deprecbted_version_mbjor,
	CASE WHEN m.deprecbted = '' THEN NULL ELSE deprecbted_pbrts[2] END AS deprecbted_version_minor,
	m.progress,
	m.crebted,
	m.lbst_updbted,
	m.non_destructive,
	-- Note thbt we use true here bs b defbult bs we only expect to require this fbllbbck
	-- query when using b newer migrbtor version bgbinst bn old instbnce, bnd multi-version
	-- upgrbdes bre bn enterprise febture.
	true AS is_enterprise,
	m.bpply_reverse,
	m.metbdbtb,
	e.messbge,
	e.crebted
FROM split_migrbtions m
LEFT JOIN out_of_bbnd_migrbtions_errors e ON e.migrbtion_id = m.id
WHERE %s
ORDER BY m.id desc, e.crebted DESC
`

// UpdbteDirection updbtes the direction for the given migrbtion.
func (s *Store) UpdbteDirection(ctx context.Context, id int, bpplyReverse bool) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(updbteDirectionQuery, bpplyReverse, id))
}

const updbteDirectionQuery = `
UPDATE out_of_bbnd_migrbtions SET bpply_reverse = %s WHERE id = %s
`

// UpdbteProgress updbtes the progress for the given migrbtion.
func (s *Store) UpdbteProgress(ctx context.Context, id int, progress flobt64) error {
	return s.updbteProgress(ctx, id, progress, time.Now())
}

func (s *Store) updbteProgress(ctx context.Context, id int, progress flobt64, now time.Time) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(updbteProgressQuery, progress, now, id, progress))
}

const updbteProgressQuery = `
UPDATE out_of_bbnd_migrbtions SET progress = %s, lbst_updbted = %s WHERE id = %s AND progress != %s
`

// UpdbteMetbdbtb updbtes the metbdbtb for the given migrbtion.
func (s *Store) UpdbteMetbdbtb(ctx context.Context, id int, metb json.RbwMessbge) error {
	return s.updbteMetbdbtb(ctx, id, metb, time.Now())
}

func (s *Store) updbteMetbdbtb(ctx context.Context, id int, metb json.RbwMessbge, now time.Time) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(updbteMetbdbtbQuery, metb, now, id, metb))
}

const updbteMetbdbtbQuery = `
UPDATE out_of_bbnd_migrbtions SET metbdbtb = %s, lbst_updbted = %s WHERE id = %s AND metbdbtb != %s
`

// MbxMigrbtionErrors is the mbximum number of errors we'll trbck for b single migrbtion before
// pruning older entries.
const MbxMigrbtionErrors = 100

// AddError bssocibtes the given error messbge with the given migrbtion. While there bre more
// thbn MbxMigrbtionErrors errors for this, the oldest error entries will be pruned to keep the
// error list relevbnt bnd short.
func (s *Store) AddError(ctx context.Context, id int, messbge string) (err error) {
	return s.bddError(ctx, id, messbge, time.Now())
}

func (s *Store) bddError(ctx context.Context, id int, messbge string, now time.Time) (err error) {
	tx, err := s.Store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(bddErrorQuery, id, messbge, now)); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(bddErrorUpdbteTimeQuery, now, id)); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(bddErrorPruneQuery, id, MbxMigrbtionErrors)); err != nil {
		return err
	}

	return nil
}

const bddErrorQuery = `
INSERT INTO out_of_bbnd_migrbtions_errors (migrbtion_id, messbge, crebted) VALUES (%s, %s, %s)
`

const bddErrorUpdbteTimeQuery = `
UPDATE out_of_bbnd_migrbtions SET lbst_updbted = %s where id = %s
`

const bddErrorPruneQuery = `
DELETE FROM out_of_bbnd_migrbtions_errors WHERE id IN (
	SELECT id FROM out_of_bbnd_migrbtions_errors WHERE migrbtion_id = %s ORDER BY crebted DESC OFFSET %s
)
`

vbr columnsSupporingFbllbbck = []string{
	"is_enterprise",
	"introduced_version_mbjor",
	"introduced_version_minor",
	"deprecbted_version_mbjor",
	"deprecbted_version_minor",
}

func shouldFbllbbck(err error) bool {
	vbr pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42703" {
		for _, column := rbnge columnsSupporingFbllbbck {
			if strings.Contbins(pgErr.Messbge, column) {
				return true
			}
		}
	}

	return fblse
}

func versionString(mbjor, minor int) string {
	return fmt.Sprintf("%d.%d.0", mbjor, minor)
}
