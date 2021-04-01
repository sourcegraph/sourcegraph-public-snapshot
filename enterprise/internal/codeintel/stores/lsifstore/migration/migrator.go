package migration

import (
	"context"
	"sort"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

// Migrator is a code-intelligence-specific out-of-band migration runner. This migrator can
// be configured by supplying a different driver instance, which controls the update behavior
// over every matching row in the migration set.
//
// Code intelligence tables are very large and using a full table scan count is too expensvie
// to use in an out-of-band migration. For each table we need to perform a migration over, we
// introduce a second aggregate table that tracks the minimum and maximum schema version of
// each data record  associated with a particular upload record.
//
// We have the following assumptions about the schema (for a configured table T):
//
// 1. There is an index on T.dump_id
// 2. For each distinct dump_id in table T, there is a corresponding row in table
//    T_schema_version. This invariant is kept up to date via triggers on insert.
// 3. Table T_schema_version has the following schema:
//
//    CREATE TABLE T_schema_versions (
//        dump_id            integer PRIMARY KEY NOT NULL,
//        min_schema_version integer,
//        max_schema_version integer
//    );
//
// When selecting a set of candidate records to migrate, we first use the each upload record's
// schema version bounds to determine if there are still records associated with that upload
// that may still need migrating. This set allows us to use the dump_id index on the target
// table. These counts can be maintained efficiently within the same transaction as a batch
// of migration updates. This requires counting within a small indexed subset of the original
// table. When checking progress, we can efficiently do a full-table on the much smaller
// aggregate table.
type Migrator struct {
	store   *lsifstore.Store
	driver  migrationDriver
	options migratorOptions
}

type migratorOptions struct {
	// tableName is the name of the table undergoing migration.
	tableName string

	// selectionFields is a list of fields that should be scanned and made available to
	// the migration driver.
	selectionFields []string

	// targetVersion is the value of the row's schema version after an up migration.
	targetVersion int

	// batchSize limits the number of rows that will be scanned on each call to Up/Down.
	batchSize int
}

type migrationDriver interface {
	// MigrateRowUp determines which fields to update for the given row.
	MigrateRowUp(scanner scanner) (updateSpec, error)

	// MigrateRowDown undoes the migration for the given row.
	MigrateRowDown(scanner scanner) (updateSpec, error)
}

// driverFunc is the type of MigrateRowUp and MigrateRowDown.
type driverFunc func(scanner scanner) (updateSpec, error)

type scanner interface {
	Scan(dest ...interface{}) error
}

type updateSpec struct {
	// DumpID is the identifier of the associated upload record.
	DumpID int

	// Conditions is a map from field names to values or SQL expressions. This map should
	// include the expected value of each primary key field of the current row, as well as
	// any additional field/value mappings that are expected to exist for this row. If not
	// supplied, the dump identifier will be implicitly added to this map prior to update.
	Conditions map[string]interface{}

	// Assignments is a map from field names to values or SQL expressions. This map should
	// include a value for each field that should be updated. If not supplied, the target
	// schema version will be implicitly added to this map prior to update.
	Assignments map[string]interface{}
}

func newMigrator(store *lsifstore.Store, driver migrationDriver, options migratorOptions) oobmigration.Migrator {
	return &Migrator{
		store:   store,
		driver:  driver,
		options: options,
	}
}

// Progress returns the ratio between the number of upload records that have been completely
// migrated over the total number of upload records. A record is migrated if its schema version
// is no less than the target migration version.
func (m *Migrator) Progress(ctx context.Context) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(
		migratorProgressQuery,
		sqlf.Sprintf(m.options.tableName),
		m.options.targetVersion,
		sqlf.Sprintf(m.options.tableName),
	)))
	if err != nil {
		return 0, err
	}

	return progress, nil
}

const migratorProgressQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/migration/migrator.go:Progress
SELECT CASE c2.count WHEN 0 THEN 1 ELSE cast(c1.count as float) / cast(c2.count as float) END FROM
	(SELECT COUNT(*) as count FROM %s_schema_versions WHERE min_schema_version >= %s) c1,
	(SELECT COUNT(*) as count FROM %s_schema_versions) c2
`

// Up runs a batch of the migration.
func (m *Migrator) Up(ctx context.Context) (err error) {
	return m.run(ctx, m.options.targetVersion-1, m.options.targetVersion, m.driver.MigrateRowUp)
}

// Down runs a batch of the migration in reverse.
func (m *Migrator) Down(ctx context.Context) error {
	return m.run(ctx, m.options.targetVersion, m.options.targetVersion-1, m.driver.MigrateRowDown)
}

// run performs a batch of updates with the given driver function. Records with the given source version
// will be selected for candidacy, and their version will match the given target version after an update.
func (m *Migrator) run(ctx context.Context, sourceVersion, targetVersion int, driverFunc driverFunc) (err error) {
	tx, err := m.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Perform the actual batch of migrations
	ids, err := m.selectAndUpdate(ctx, tx, sourceVersion, targetVersion, driverFunc)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	idQueries := make([]*sqlf.Query, 0, len(ids))
	for _, id := range ids {
		idQueries = append(idQueries, sqlf.Sprintf("%s", id))
	}

	return tx.Exec(ctx, sqlf.Sprintf(runQuery, sqlf.Sprintf(m.options.tableName), sqlf.Sprintf(m.options.tableName), sqlf.Join(idQueries, ", ")))
}

const runQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/migration/migrator.go:run
INSERT INTO %s_schema_versions
SELECT
	dump_id,
	MIN(schema_version) as min_schema_version,
	MAX(schema_version) as max_schema_version
FROM
	%s
WHERE
	dump_id IN (%s)
GROUP BY
	dump_id
ON CONFLICT (dump_id) DO UPDATE SET
	min_schema_version = EXCLUDED.min_schema_version,
	max_schema_version = EXCLUDED.max_schema_version
`

// selectAndUpdate selects a batch of records from the configured table with the given source
// version, then performs an update on each matching row with the new values given from an
// invocation of the given driver function. This method returns a deduplicated and ordered set
// of upload identifiers denoting the complete set of uploads whose records were modified
// by this batch of updates.
func (m *Migrator) selectAndUpdate(ctx context.Context, tx *lsifstore.Store, sourceVersion, targetVersion int, driverFunc driverFunc) ([]int, error) {
	// Note: we can't pipeline this as you can't have an open rows object and
	// execute another unrelated query using the same database handle.
	updateSpecs, err := m.selectAndProcess(ctx, tx, sourceVersion, driverFunc)
	if err != nil {
		return nil, err
	}

	for _, spec := range updateSpecs {
		defaultConditions := map[string]interface{}{"dump_id": spec.DumpID}
		defaultAssignments := map[string]interface{}{"schema_version": targetVersion}

		if err := tx.Exec(ctx, sqlf.Sprintf(
			selectAndUpdateQuery,
			sqlf.Sprintf(m.options.tableName),
			sqlf.Join(formatFieldValuePairs(fillInDefaults(spec.Assignments, defaultAssignments)), ", "),  // SET k1 = v1, k2 = v2
			sqlf.Join(formatFieldValuePairs(fillInDefaults(spec.Conditions, defaultConditions)), " AND "), // WHERE k1 = v1 AND k2 = v2
		)); err != nil {
			return nil, err
		}
	}

	return dumpIDsFromUpdateSpecs(updateSpecs), nil
}

const selectAndUpdateQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/migration/migrator.go:selectAndUpdate
UPDATE %s SET %s WHERE %s
`

// selectAndProcess selects a batch of records from the configured table with the given version and
// returns the update specifications after running the given driver function on each matching row.
// The records selected by this method are locked (via select for update) in the given transaction.
func (m *Migrator) selectAndProcess(ctx context.Context, tx *lsifstore.Store, version int, driverFunc driverFunc) ([]updateSpec, error) {
	fieldQueries := make([]*sqlf.Query, 0, len(m.options.selectionFields))
	for _, field := range m.options.selectionFields {
		fieldQueries = append(fieldQueries, sqlf.Sprintf(field))
	}

	rows, err := tx.Query(ctx, sqlf.Sprintf(
		selectAndProcessQuery,
		sqlf.Sprintf(m.options.tableName),
		version,
		version,
		sqlf.Join(fieldQueries, ", "),
		sqlf.Sprintf(m.options.tableName),
		version,
		m.options.batchSize,
	))
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	updateSpecs := make([]updateSpec, 0, m.options.batchSize)
	for rows.Next() {
		spec, err := driverFunc(rows)
		if err != nil {
			return nil, err
		}

		updateSpecs = append(updateSpecs, spec)
	}

	return updateSpecs, nil
}

const selectAndProcessQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/migration/migrator.go:selectAndProcess
WITH candidates AS (
	SELECT dump_id
	FROM %s_schema_versions
	WHERE
		min_schema_version <= %s AND
		max_schema_version >= %s
)
SELECT dump_id, %s
FROM %s
WHERE
	dump_id IN (SELECT dump_id FROM candidates) AND
	schema_version = %s
ORDER BY dump_id
LIMIT %s
FOR UPDATE SKIP LOCKED
`

// fillInDefaults assigns each (k, v) pair in the given defaults map into the target map m
// if the key does not already exist. This method returns the target map m.
func fillInDefaults(m, defaults map[string]interface{}) map[string]interface{} {
	for k, v := range defaults {
		if _, ok := m[k]; !ok {
			m[k] = v
		}
	}

	return m
}

// formatFieldValuePairs returns a slice of sqlf.Query values whose values are `{k} = {v}`
// for each (k, v) in the given map. The output slice is ordered by field name.
func formatFieldValuePairs(union map[string]interface{}) []*sqlf.Query {
	keys := make([]string, 0, len(union))
	for k := range union {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	queries := make([]*sqlf.Query, 0, len(union))
	for _, k := range keys {
		queries = append(queries, sqlf.Sprintf("%s = %s", sqlf.Sprintf(k), union[k]))
	}

	return queries
}

// dumpIDsFromUpdateSpecs returns the identifiers of the uploads referenced in the given slice
// of update spec values.
func dumpIDsFromUpdateSpecs(updateSpecs []updateSpec) []int {
	idMap := map[int]struct{}{}
	for _, spec := range updateSpecs {
		idMap[spec.DumpID] = struct{}{}
	}

	ids := make([]int, 0, len(idMap))
	for id := range idMap {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	return ids
}
