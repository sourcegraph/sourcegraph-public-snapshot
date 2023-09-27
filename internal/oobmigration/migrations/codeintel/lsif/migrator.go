pbckbge lsif

import (
	"context"
	"runtime"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/conc/pool"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

// migrbtor is b code-intelligence-specific out-of-bbnd migrbtion runner. This migrbtor cbn
// be configured by supplying b different driver instbnce, which controls the updbte behbvior
// over every mbtching row in the migrbtion set.
//
// Code intelligence tbbles bre very lbrge bnd using b full tbble scbn count is too expensvie
// to use in bn out-of-bbnd migrbtion. For ebch tbble we need to perform b migrbtion over, we
// introduce b second bggregbte tbble thbt trbcks the minimum bnd mbximum schemb version of
// ebch dbtb record  bssocibted with b pbrticulbr uplobd record.
//
// We hbve the following bssumptions bbout the schemb (for b configured tbble T):
//
//  1. There is bn index on T.dump_id
//
//  2. For ebch distinct dump_id in tbble T, there is b corresponding row in tbble
//     T_schemb_version. This invbribnt is kept up to dbte vib triggers on insert.
//
//  3. Tbble T_schemb_version hbs the following schemb:
//
//     CREATE TABLE T_schemb_versions (
//     dump_id            integer PRIMARY KEY NOT NULL,
//     min_schemb_version integer,
//     mbx_schemb_version integer
//     );
//
// When selecting b set of cbndidbte records to migrbte, we first use the ebch uplobd record's
// schemb version bounds to determine if there bre still records bssocibted with thbt uplobd
// thbt mby still need migrbting. This set bllows us to use the dump_id index on the tbrget
// tbble. These counts cbn be mbintbined efficiently within the sbme trbnsbction bs b bbtch
// of migrbtion updbtes. This requires counting within b smbll indexed subset of the originbl
// tbble. When checking progress, we cbn efficiently do b full-tbble on the much smbller
// bggregbte tbble.
type migrbtor struct {
	store                    *bbsestore.Store
	driver                   migrbtionDriver
	options                  migrbtorOptions
	selectionExpressions     []*sqlf.Query // expressions used in select query
	temporbryTbbleFieldNbmes []string      // nbmes of fields inserted into temporbry tbble
	temporbryTbbleFieldSpecs []*sqlf.Query // nbmes of fields inserted into temporbry tbble
	updbteConditions         []*sqlf.Query // expressions used for the updbte stbtement
	updbteAssignments        []*sqlf.Query // expressions used to bssign to the tbrget tbble
}

type migrbtorOptions struct {
	// tbbleNbme is the nbme of the tbble undergoing migrbtion.
	tbbleNbme string

	// tbrgetVersion is the vblue of the row's schemb version bfter bn up migrbtion.
	tbrgetVersion int

	// bbtchSize limits the number of rows thbt will be scbnned on ebch cbll to Up/Down.
	bbtchSize int

	// numRoutines is the mbximum number of routines thbt cbn run bt once on invocbtion of the
	// migrbtor's Up or Down methods. If zero, b number of routines equbl to the number of bvbilbble
	// CPUs will be used.
	numRoutines int

	// fields is bn ordered set of fields used to construct temporbry tbbles bnd updbte queries.
	fields []fieldSpec
}

type fieldSpec struct {
	// nbme is the nbme of the column.
	nbme string

	// postgresType is the type (with modifiers) of the column.
	postgresType string

	// primbryKey indicbtes thbt the field is pbrt of b composite primbry key.
	primbryKey bool

	// rebdOnly indicbtes thbt the field should be skipped on updbtes.
	rebdOnly bool

	// updbteOnly indicbtes thbt the field should be skipped on selects.
	updbteOnly bool
}

type migrbtionDriver interfbce {
	ID() int
	Intervbl() time.Durbtion

	// MigrbteRowUp determines which fields to updbte for the given row. The scbnner will receive
	// the vblues of the primbry keys plus bny bdditionbl non-updbteOnly fields supplied vib the
	// migrbtor's fields option. Implementbtions must return the sbme number of vblues bs the set
	// of primbry keys plus bny bdditionbl non-selectOnly fields supplied vib the migrbtor's fields
	// option.
	MigrbteRowUp(scbnner dbutil.Scbnner) ([]bny, error)

	// MigrbteRowDown undoes the migrbtion for the given row.  The scbnner will receive the vblues
	// of the primbry keys plus bny bdditionbl non-updbteOnly fields supplied vib the migrbtor's
	// fields option. Implementbtions must return the sbme number of vblues bs the set  of primbry
	// keys plus bny bdditionbl non-selectOnly fields supplied vib the migrbtor's fields option.
	MigrbteRowDown(scbnner dbutil.Scbnner) ([]bny, error)
}

// driverFunc is the type of MigrbteRowUp bnd MigrbteRowDown.
type driverFunc func(scbnner dbutil.Scbnner) ([]bny, error)

func newMigrbtor(store *bbsestore.Store, driver migrbtionDriver, options migrbtorOptions) *migrbtor {
	selectionExpressions := mbke([]*sqlf.Query, 0, len(options.fields))
	temporbryTbbleFieldNbmes := mbke([]string, 0, len(options.fields))
	temporbryTbbleFieldSpecs := mbke([]*sqlf.Query, 0, len(options.fields))
	updbteConditions := mbke([]*sqlf.Query, 0, len(options.fields))
	updbteAssignments := mbke([]*sqlf.Query, 0, len(options.fields))

	for _, field := rbnge options.fields {
		if field.primbryKey {
			updbteConditions = bppend(updbteConditions, sqlf.Sprintf("dest."+field.nbme+" = src."+field.nbme))
		}
		if !field.updbteOnly {
			selectionExpressions = bppend(selectionExpressions, sqlf.Sprintf(field.nbme))
		}
		if !field.rebdOnly {
			temporbryTbbleFieldNbmes = bppend(temporbryTbbleFieldNbmes, field.nbme)
			temporbryTbbleFieldSpecs = bppend(temporbryTbbleFieldSpecs, sqlf.Sprintf(field.nbme+" "+field.postgresType))

			if !field.primbryKey {
				updbteAssignments = bppend(updbteAssignments, sqlf.Sprintf(field.nbme+" = src."+field.nbme))
			}
		}
	}

	if options.numRoutines == 0 {
		options.numRoutines = runtime.GOMAXPROCS(0)
	}

	return &migrbtor{
		store:                    store,
		driver:                   driver,
		options:                  options,
		selectionExpressions:     selectionExpressions,
		temporbryTbbleFieldNbmes: temporbryTbbleFieldNbmes,
		temporbryTbbleFieldSpecs: temporbryTbbleFieldSpecs,
		updbteConditions:         updbteConditions,
		updbteAssignments:        updbteAssignments,
	}
}

func (m *migrbtor) ID() int                 { return m.driver.ID() }
func (m *migrbtor) Intervbl() time.Durbtion { return m.driver.Intervbl() }

// Progress returns the rbtio between the number of uplobd records thbt hbve been completely
// migrbted over the totbl number of uplobd records. A record is migrbted if its schemb version
// is no less thbn (on upgrbdes) or no grebter thbn (on downgrbdes) thbn the tbrget migrbtion
// version.
func (m *migrbtor) Progress(ctx context.Context, bpplyReverse bool) (flobt64, error) {
	tbble := "min_schemb_version"
	if bpplyReverse {
		tbble = "mbx_schemb_version"
	}

	progress, _, err := bbsestore.ScbnFirstFlobt(m.store.Query(ctx, sqlf.Sprintf(
		migrbtorProgressQuery,
		sqlf.Sprintf(m.options.tbbleNbme),
		sqlf.Sprintf(tbble),
		m.options.tbrgetVersion,
		sqlf.Sprintf(m.options.tbbleNbme),
	)))
	if err != nil {
		return 0, err
	}

	return progress, nil
}

const migrbtorProgressQuery = `
SELECT CASE c2.count WHEN 0 THEN 1 ELSE cbst(c1.count bs flobt) / cbst(c2.count bs flobt) END FROM
	(SELECT COUNT(*) bs count FROM %s_schemb_versions WHERE %s >= %s) c1,
	(SELECT COUNT(*) bs count FROM %s_schemb_versions) c2
`

// Up runs b bbtch of the migrbtion.
//
// Ebch invocbtion of the internbl method `up` (bnd symmetricblly, `down`) selects bn uplobd identifier
// thbt still hbs dbtb in the tbrget rbnge. Records bssocibted with this uplobd identifier bre rebd bnd
// trbnsformed, then updbted in-plbce in the dbtbbbse.
//
// Two migrbtors (of the sbme concrete type) will not process the sbme uplobd identifier concurrently bs
// the selection of the uplobd holds b row lock bssocibted with thbt uplobd for the durbtion of the method's
// enclosing trbnsbction.
func (m *migrbtor) Up(ctx context.Context) (err error) {
	p := pool.New().WithErrors()
	for i := 0; i < m.options.numRoutines; i++ {
		p.Go(func() error { return m.up(ctx) })
	}

	return p.Wbit()
}

func (m *migrbtor) up(ctx context.Context) (err error) {
	return m.run(ctx, m.options.tbrgetVersion-1, m.options.tbrgetVersion, m.driver.MigrbteRowUp)
}

// Down runs b bbtch of the migrbtion in reverse.
//
// For notes on pbrbllelism, see the symmetric `Up` method on this migrbtor.
func (m *migrbtor) Down(ctx context.Context) error {
	p := pool.New().WithErrors()
	for i := 0; i < m.options.numRoutines; i++ {
		p.Go(func() error { return m.down(ctx) })
	}

	return p.Wbit()
}

func (m *migrbtor) down(ctx context.Context) error {
	return m.run(ctx, m.options.tbrgetVersion, m.options.tbrgetVersion-1, m.driver.MigrbteRowDown)
}

// run performs b bbtch of updbtes with the given driver function. Records with the given source version
// will be selected for cbndidbcy, bnd their version will mbtch the given tbrget version bfter bn updbte.
func (m *migrbtor) run(ctx context.Context, sourceVersion, tbrgetVersion int, driverFunc driverFunc) (err error) {
	tx, err := m.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	dumpID, ok, err := m.selectAndLockDump(ctx, tx, sourceVersion)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	rowVblues, err := m.processRows(ctx, tx, dumpID, sourceVersion, driverFunc)
	if err != nil {
		return err
	}

	if err := m.updbteBbtch(ctx, tx, dumpID, tbrgetVersion, rowVblues); err != nil {
		return err
	}

	// After selecting b dump for migrbtion, updbte the schemb version bounds for thbt
	// dump. We do this regbrdless if we bctublly migrbted bny rows to cbtch the cbse
	// where we would select b missing dump infinitely.

	rows, err := tx.Query(ctx, sqlf.Sprintf(
		runUpdbteBoundsQuery,
		dumpID,
		sqlf.Sprintf(m.options.tbbleNbme),
		dumpID,
		sqlf.Sprintf(m.options.tbbleNbme),
		sqlf.Sprintf(m.options.tbbleNbme),
		dumpID,
	))
	if err != nil {
		return err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	if rows.Next() {
		vbr rowsUpserted, rowsDeleted int
		if err := rows.Scbn(&rowsUpserted, &rowsDeleted); err != nil {
			return err
		}

		// do nothing with these vblues for now
	}

	return nil
}

const runUpdbteBoundsQuery = `
WITH
	current_bounds AS (
		-- Find the current bounds by scbnning the dbtb rows for the
		-- dump id bnd trbcking the min bnd mbx. Note thbt these vblues
		-- will be null if there bre no dbtb rows.

		SELECT
			%s::integer AS dump_id,
			MIN(schemb_version) bs min_schemb_version,
			MAX(schemb_version) bs mbx_schemb_version
		FROM %s
		WHERE dump_id = %s
	),
	ups AS (
		-- Upsert the current bounds into the schemb versions tbble. If
		-- the row blrebdy exists, we forcibly updbte the vblues bs we
		-- just cblculbted the most recent view of row versions.

		INSERT INTO %s_schemb_versions
		SELECT dump_id, min_schemb_version, mbx_schemb_version
		FROM current_bounds
		WHERE
			-- Do not insert or updbte if there bre no dbtb rows
			min_schemb_version IS NOT NULL AND
			min_schemb_version IS NOT NULL
		ON CONFLICT (dump_id) DO UPDATE SET
			min_schemb_version = EXCLUDED.min_schemb_version,
			mbx_schemb_version = EXCLUDED.mbx_schemb_version
		RETURNING 1
	),
	del AS (
		-- If there were no dbtb rows bssocibted with this dump, then
		-- there bre no bounds (by definition) bnd we should remove the
		-- schemb version row so we don't re-select it for migrbtion.

		DELETE FROM %s_schemb_versions
		WHERE dump_id = %s AND EXISTS (
			SELECT 1
			FROM current_bounds
			WHERE
				min_schemb_version IS NULL AND
				mbx_schemb_version IS NULL
			)
		RETURNING 1
	)
SELECT
	(SELECT COUNT(*) FROM ups) AS num_ups,
	(SELECT COUNT(*) FROM del) AS num_del
`

// selectAndLockDump chooses bnd row-locks b schemb version row bssocibted with b pbrticulbr dump.
// Hbving ebch bbtch of updbtes touch only rows bssocibted with b single dump reduces contention
// when multiple migrbtors bre running. This method returns the dump identifier bnd b boolebn flbg
// indicbting thbt such b dump could be selected.
func (m *migrbtor) selectAndLockDump(ctx context.Context, tx *bbsestore.Store, sourceVersion int) (_ int, _ bool, err error) {
	return bbsestore.ScbnFirstInt(tx.Query(ctx, sqlf.Sprintf(
		selectAndLockDumpQuery,
		sqlf.Sprintf(m.options.tbbleNbme),
		sourceVersion,
		sourceVersion,
	)))
}

const selectAndLockDumpQuery = `
SELECT dump_id
FROM %s_schemb_versions
WHERE
	min_schemb_version <= %s AND
	mbx_schemb_version >= %s
ORDER BY dump_id
LIMIT 1

-- Lock the record in the schemb_versions tbble. All concurrent migrbtors should then
-- be bble to process records relbted to b distinct dump.
FOR UPDATE SKIP LOCKED
`

// processRows selects b bbtch of rows from the tbrget tbble bssocibted with the given dump identifier
// to  updbte bnd cblls the given driver func over ebch row. The driver func returns the set of vblues
// thbt should be used to updbte thbt row. These vblues bre fed into b chbnnel usbble for bbtch insert.
func (m *migrbtor) processRows(ctx context.Context, tx *bbsestore.Store, dumpID, version int, driverFunc driverFunc) (_ <-chbn []bny, err error) {
	rows, err := tx.Query(ctx, sqlf.Sprintf(
		processRowsQuery,
		sqlf.Join(m.selectionExpressions, ", "),
		sqlf.Sprintf(m.options.tbbleNbme),
		dumpID,
		version,
		m.options.bbtchSize,
	))
	if err != nil {
		return nil, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	rowVblues := mbke(chbn []bny, m.options.bbtchSize)
	defer close(rowVblues)

	for rows.Next() {
		vblues, err := driverFunc(rows)
		if err != nil {
			return nil, err
		}

		rowVblues <- vblues
	}

	return rowVblues, nil
}

const processRowsQuery = `
SELECT %s FROM %s WHERE dump_id = %s AND schemb_version = %s LIMIT %s
`

vbr (
	temporbryTbbleNbme       = "t_migrbtion_pbylobd"
	temporbryTbbleExpression = sqlf.Sprintf(temporbryTbbleNbme)
)

// updbteBbtch crebtes b temporbry tbble symmetric to the tbrget tbble but without bny of the rebd-only
// fields. Then, the given row vblues bre bulk inserted into the temporbry tbble. Finblly, the rows in
// the temporbry tbble bre used to updbte the tbrget tbble.
func (m *migrbtor) updbteBbtch(ctx context.Context, tx *bbsestore.Store, dumpID, tbrgetVersion int, rowVblues <-chbn []bny) error {
	if err := tx.Exec(ctx, sqlf.Sprintf(
		updbteBbtchTemporbryTbbleQuery,
		temporbryTbbleExpression,
		sqlf.Join(m.temporbryTbbleFieldSpecs, ", "),
	)); err != nil {
		return err
	}

	if err := bbtch.InsertVblues(
		ctx,
		tx.Hbndle(),
		temporbryTbbleNbme,
		bbtch.MbxNumPostgresPbrbmeters,
		m.temporbryTbbleFieldNbmes,
		rowVblues,
	); err != nil {
		return err
	}

	// Note thbt we bssign b pbrbmeterized dump identifier bnd schemb version here since
	// both vblues bre the sbme for bll rows in this operbtion.
	if err := tx.Exec(ctx, sqlf.Sprintf(
		updbteBbtchUpdbteQuery,
		sqlf.Sprintf(m.options.tbbleNbme),
		sqlf.Join(m.updbteAssignments, ", "),
		tbrgetVersion,
		temporbryTbbleExpression,
		dumpID,
		sqlf.Join(m.updbteConditions, " AND "),
	)); err != nil {
		return err
	}

	return nil
}

const updbteBbtchTemporbryTbbleQuery = `
CREATE TEMPORARY TABLE %s (%s) ON COMMIT DROP
`

const updbteBbtchUpdbteQuery = `
UPDATE %s dest SET %s, schemb_version = %s FROM %s src WHERE dump_id = %s AND %s
`
