package migration

import (
	"context"
	"sort"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
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
	store                    *lsifstore.Store
	driver                   migrationDriver
	options                  migratorOptions
	selectionExpressions     []*sqlf.Query // expressions used in select query
	temporaryTableFieldNames []string      // names of fields inserted into temporary table
	temporaryTableFieldSpecs []*sqlf.Query // names of fields inserted into temporary table
	updateConditions         []*sqlf.Query // expressions used for the update statement
	updateAssignments        []*sqlf.Query // expressions used to assign to the target table
}

type migratorOptions struct {
	// tableName is the name of the table undergoing migration.
	tableName string

	// targetVersion is the value of the row's schema version after an up migration.
	targetVersion int

	// batchSize limits the number of rows that will be scanned on each call to Up/Down.
	batchSize int

	// fields is an ordered set of fields used to construct temporary tables and update queries.
	fields []fieldSpec
}

type fieldSpec struct {
	// name is the name of the column.
	name string

	// postgresType is the type (with modifiers) of the column.
	postgresType string

	// primaryKey indicates that the field is part of a composite primary key.
	primaryKey bool

	// readOnly indicates that the field should be skipped on updates.
	readOnly bool

	// updateOnly indicates that the field should be skipped on selects.
	updateOnly bool
}

type migrationDriver interface {
	// MigrateRowUp determines which fields to update for the given row. The scanner will receive
	// the values of the primary keys plus any additional non-updateOnly fields supplied via the
	// migrator's fields option.
	MigrateRowUp(scanner scanner) (updateSpec, error)

	// MigrateRowDown undoes the migration for the given row.  The scanner will receive the values
	// of the primary keys plus any additional non-updateOnly fields supplied via the migrator's
	// fields option.
	MigrateRowDown(scanner scanner) (updateSpec, error)
}

// driverFunc is the type of MigrateRowUp and MigrateRowDown.
type driverFunc func(scanner scanner) (updateSpec, error)

type scanner interface {
	Scan(dest ...interface{}) error
}

type updateSpec struct {
	// dumpID is the identifier of the associated upload record.
	dumpID int

	// fieldValues indicates the values that should be written back to the table. This must have
	// the same number of values as the set of primary keys plus any additional non-selectOnly
	// fields supplied via the migrator's fields option.
	fieldValues []interface{}
}

func newMigrator(store *lsifstore.Store, driver migrationDriver, options migratorOptions) oobmigration.Migrator {
	selectionExpressions := make([]*sqlf.Query, 0, len(options.fields))
	temporaryTableFieldNames := make([]string, 0, len(options.fields))
	temporaryTableFieldSpecs := make([]*sqlf.Query, 0, len(options.fields))
	updateConditions := make([]*sqlf.Query, 0, len(options.fields))
	updateAssignments := make([]*sqlf.Query, 0, len(options.fields))

	for _, field := range options.fields {
		if field.primaryKey {
			updateConditions = append(updateConditions, sqlf.Sprintf("dest."+field.name+" = src."+field.name))
		}
		if !field.updateOnly {
			selectionExpressions = append(selectionExpressions, sqlf.Sprintf(field.name))
		}
		if !field.readOnly {
			temporaryTableFieldNames = append(temporaryTableFieldNames, field.name)
			temporaryTableFieldSpecs = append(temporaryTableFieldSpecs, sqlf.Sprintf(field.name+" "+field.postgresType))

			if !field.primaryKey {
				updateAssignments = append(updateAssignments, sqlf.Sprintf(field.name+" = src."+field.name))
			}
		}
	}

	return &Migrator{
		store:                    store,
		driver:                   driver,
		options:                  options,
		selectionExpressions:     selectionExpressions,
		temporaryTableFieldNames: temporaryTableFieldNames,
		temporaryTableFieldSpecs: temporaryTableFieldSpecs,
		updateConditions:         updateConditions,
		updateAssignments:        updateAssignments,
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
func (m *Migrator) selectAndUpdate(ctx context.Context, tx *lsifstore.Store, sourceVersion, targetVersion int, driverFunc driverFunc) (_ []int, err error) {
	// Note: we can't pipeline this as you can't have an open rows object and
	// execute another unrelated query using the same database handle.
	updateSpecs, err := m.selectAndProcess(ctx, tx, sourceVersion, driverFunc)
	if err != nil {
		return nil, err
	}

	idMap := map[int]struct{}{}
	rowValues := make(chan []interface{}, len(updateSpecs))
	for _, spec := range updateSpecs {
		rowValues <- spec.fieldValues
		idMap[spec.dumpID] = struct{}{}
	}
	close(rowValues)

	temporaryTableName := "t_migration_payload"
	temporaryTableExpression := sqlf.Sprintf(temporaryTableName)

	// Create temporary table symmetric to the target table without read-only fields
	if err := tx.Exec(ctx, sqlf.Sprintf(
		selectAndUpdateTemporaryTableQuery,
		temporaryTableExpression,
		sqlf.Join(m.temporaryTableFieldSpecs, ", "),
	)); err != nil {
		return nil, err
	}

	// Bulk insert all the unique column values into the temporary table
	if err := batch.InsertValues(
		ctx,
		tx.Handle().DB(),
		temporaryTableName,
		m.temporaryTableFieldNames,
		rowValues,
	); err != nil {
		return nil, err
	}

	// Update the values from the temporary table in the target table. We assign a
	// parameterized schema version here since it is the same for all rows in this
	// operation.
	if err := tx.Exec(ctx, sqlf.Sprintf(
		selectAndUpdateUpdateQuery,
		sqlf.Sprintf(m.options.tableName),
		sqlf.Join(m.updateAssignments, ", "),
		targetVersion,
		temporaryTableExpression,
		sqlf.Join(m.updateConditions, " AND "),
	)); err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(idMap))
	for id := range idMap {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	return ids, nil
}

const selectAndUpdateTemporaryTableQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/migration/migrator.go:selectAndUpdate
CREATE TEMPORARY TABLE %s (%s) ON COMMIT DROP
`

const selectAndUpdateUpdateQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/migration/migrator.go:selectAndUpdate
UPDATE %s dest SET %s, schema_version = %s FROM %s src WHERE %s
`

// selectAndProcess selects a batch of records from the configured table with the given version and
// returns the update specifications after running the given driver function on each matching row.
// The records selected by this method are locked (via select for update) in the given transaction.
//
// This method will lock as many records to be processed as can fit in a batch, but all records will
// belong to the same index (due to the query construction). Therefore the dump id for each batch is
// going to be the same for each record.
func (m *Migrator) selectAndProcess(ctx context.Context, tx *lsifstore.Store, version int, driverFunc driverFunc) ([]updateSpec, error) {
	rows, err := tx.Query(ctx, sqlf.Sprintf(
		selectAndProcessQuery,
		sqlf.Join(m.selectionExpressions, ", "),
		sqlf.Sprintf(m.options.tableName),
		sqlf.Sprintf(m.options.tableName),
		version,
		version,
		sqlf.Sprintf(m.options.tableName),
		version,
		version,
		m.options.batchSize,
	))
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var specs []updateSpec
	for rows.Next() {
		spec, err := driverFunc(rows)
		if err != nil {
			return nil, err
		}

		specs = append(specs, spec)
	}

	return specs, nil
}

const selectAndProcessQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/migration/migrator.go:selectAndProcess
SELECT %s
FROM %s t1
WHERE
	dump_id = (
		-- First, we select an index that has at least one record with the target schema
		-- version. We do this so that we can efficiently select from the target table,
		-- which are all keyed on dump_id.
		--
		-- This is more efficient than scanning the entire target table looking for a matching
		-- schema version especially when that schema version is a small subset of the table.

		SELECT sv.dump_id
		FROM %s_schema_versions sv
		WHERE
			-- Check if we have a schema version in range
			sv.min_schema_version <= %s AND
			sv.max_schema_version >= %s AND

			-- Ensure we actually have a row with the target schema version. This condition may
			-- be true numerically but may not have any rows with this particular schema version;
			--
			-- For example: an index with a min schema version of 3 and max schema version of 5
			-- may have no rows with a schema version of 4. We want to skip over these indexes
			-- before moving on to the query so we don't always pull back an empty batch unable
			-- to migrate a legitimate index stuck behind the head of the queue.
			EXISTS (SELECT 1 FROM %s t2 WHERE t2.dump_id = sv.dump_id AND t2.schema_version = %s)

		-- Encourage index scan of pk
		ORDER BY dump_id

		-- All records in a migration batch will belong to a single dump
		LIMIT 1
	) AND
	schema_version = %s
LIMIT %s
FOR UPDATE SKIP LOCKED
`
