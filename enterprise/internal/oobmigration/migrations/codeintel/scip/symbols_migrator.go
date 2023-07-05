package scip

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/symbols"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type scipSymbolsMigrator struct {
	codeintelStore *basestore.Store
}

func NewSCIPSymbolsMigrator(codeintelStore *basestore.Store) *migrator {
	driver := &scipSymbolsMigrator{
		codeintelStore: codeintelStore,
	}

	return newMigrator(codeintelStore, driver, migratorOptions{
		tableName:     "codeintel_scip_symbols",
		targetVersion: 2,
		batchSize:     10000, // TODO
		numRoutines:   1,     // TODO
		fields: []fieldSpec{
			{name: "symbol_id", postgresType: "integer not null", primaryKey: true},
			{name: "document_lookup_id", postgresType: "integer not null", primaryKey: true},
			{name: "scheme_id", postgresType: "integer", updateOnly: true},
			{name: "package_manager_id", postgresType: "integer", updateOnly: true},
			{name: "package_name_id", postgresType: "integer", updateOnly: true},
			{name: "package_version_id", postgresType: "integer", updateOnly: true},
			{name: "descriptor_id", postgresType: "integer", updateOnly: true},
			{name: "descriptor_no_suffix_id", postgresType: "integer", updateOnly: true},
		},
	})
}

func (m *scipSymbolsMigrator) ID() int                 { return 24 }
func (m *scipSymbolsMigrator) Interval() time.Duration { return time.Second }

// TODO - rewrite
// TODO - redocument
// MigrateRowUp reads the payload of the given row and returns an updateSpec on how to
// modify the record to conform to the new schema.
func (m *scipSymbolsMigrator) MigrateUp(ctx context.Context, uploadID int, tx *basestore.Store, rows *sql.Rows) (_ [][]any, err error) {
	type symbolInDocument struct {
		symbolID         int
		documentLookupID int
	}
	var symbolPairs []symbolInDocument

	if err := func() (err error) {
		defer func() { err = basestore.CloseRows(rows, err) }()

		for rows.Next() {
			var symbolID, documentLookupID int
			if err := rows.Scan(&symbolID, &documentLookupID); err != nil {
				return err
			}

			symbolPairs = append(symbolPairs, symbolInDocument{symbolID, documentLookupID})
		}

		return nil
	}(); err != nil {
		return nil, err
	}

	var symbolIDs []int
	for _, symbol := range symbolPairs {
		symbolIDs = append(symbolIDs, symbol.symbolID)
	}

	scanner := basestore.NewMapScanner[int, string](func(s dbutil.Scanner) (symbolID int, symbolName string, _ error) {
		err := s.Scan(&symbolID, &symbolName)
		return symbolID, symbolName, err
	})

	symbolsNamesByID, err := scanner(tx.Query(ctx, sqlf.Sprintf(`
		WITH RECURSIVE
		symbols(id, upload_id, suffix, prefix_id) AS (
			(
				SELECT
					ssn.id,
					ssn.upload_id,
					ssn.name_segment AS suffix,
					ssn.prefix_id AS prefix_id
				FROM codeintel_scip_symbol_names ssn
				WHERE
					ssn.id = ANY(%s) AND
					ssn.upload_id = %s
			) UNION (
				SELECT
					s.id,
					s.upload_id,
					ssn.name_segment || s.suffix AS suffix,
					ssn.prefix_id AS prefix_id
				FROM symbols s
				JOIN codeintel_scip_symbol_names ssn ON
					ssn.upload_id = s.upload_id AND
					ssn.id = s.prefix_id
			)
		)
		SELECT s.id, s.suffix AS symbol_name
		FROM symbols s
		WHERE s.prefix_id IS NULL
	`,
		pq.Array(symbolIDs),
		uploadID,
	)))
	if err != nil {
		return nil, err
	}

	symbolNames := make([]string, 0, len(symbolsNamesByID))
	for _, symbolName := range symbolsNamesByID {
		symbolNames = append(symbolNames, symbolName)
	}

	// TODO - needs to be unique within an index?
	nextSymbolLookupID := 0

	schemes := make(map[string]int)
	managers := make(map[string]int)
	packageNames := make(map[string]int)
	packageVersions := make(map[string]int)
	descriptors := make(map[string]int)
	descriptorsNoSuffix := make(map[string]int)

	type explodedIDs struct {
		schemeID             int
		packageManagerID     int
		packageNameID        int
		packageVersionID     int
		descriptorID         int
		descriptorNoSuffixID int
	}
	cache := map[string]explodedIDs{}

	getOrSetID := func(m map[string]int, key string) int {
		if v, ok := m[key]; ok {
			return v
		}

		id := nextSymbolLookupID
		nextSymbolLookupID++
		m[key] = id
		return id
	}

	for _, symbolName := range symbolNames {
		symbol, err := symbols.NewExplodedSymbol(symbolName)
		if err != nil {
			return nil, err
		}

		schemeID := getOrSetID(schemes, symbol.Scheme)
		packageManagerID := getOrSetID(managers, symbol.PackageManager)
		packageNameID := getOrSetID(packageNames, symbol.PackageName)
		packageVersionID := getOrSetID(packageVersions, symbol.PackageVersion)
		descriptorID := getOrSetID(descriptors, symbol.Descriptor)
		descriptorNoSuffixID := getOrSetID(descriptorsNoSuffix, symbol.DescriptorNoSuffix)
		cache[symbolName] = explodedIDs{
			schemeID:             schemeID,
			packageManagerID:     packageManagerID,
			packageNameID:        packageNameID,
			packageVersionID:     packageVersionID,
			descriptorID:         descriptorID,
			descriptorNoSuffixID: descriptorNoSuffixID,
		}
	}

	maps := map[string]map[string]int{
		"SCHEME":               schemes,
		"PACKAGE_MANAGER":      managers,
		"PACKAGE_NAME":         packageNames,
		"PACKAGE_VERSION":      packageVersions,
		"DESCRIPTOR":           descriptors,
		"DESCRIPTOR_NO_SUFFIX": descriptorsNoSuffix,
	}

	const newSCIPWriterTemporarySymbolLookupTableQuery = `
		CREATE TEMPORARY TABLE t_codeintel_scip_symbols_lookup(
			id integer NOT NULL,
			upload_id integer NOT NULL,
			name text NOT NULL,
			scip_name_type text NOT NULL
		) ON COMMIT DROP
	`
	if err := tx.Exec(ctx, sqlf.Sprintf(newSCIPWriterTemporarySymbolLookupTableQuery)); err != nil {
		return nil, err
	}

	symbolLookupInserter := batch.NewInserter(
		ctx,
		tx.Handle(),
		"t_codeintel_scip_symbols_lookup",
		batch.MaxNumPostgresParameters,
		"id",
		"upload_id",
		"name",
		"scip_name_type",
	)

	for nameType, m := range maps {
		for symbolName, symbolID := range m {
			if err := symbolLookupInserter.Insert(ctx, symbolID, uploadID, symbolName, nameType); err != nil {
				return nil, err
			}
		}
	}

	if err := symbolLookupInserter.Flush(ctx); err != nil {
		return nil, err
	}

	values := make([][]any, 0, len(symbolPairs))
	for _, pair := range symbolPairs {
		ids := cache[symbolsNamesByID[pair.symbolID]]

		values = append(values, []any{
			pair.symbolID,
			pair.documentLookupID,
			ids.schemeID,
			ids.packageManagerID,
			ids.packageNameID,
			ids.packageVersionID,
			ids.descriptorID,
			ids.descriptorNoSuffixID,
		})
	}

	return values, nil
}

//
//
//

// TODO - rewrite
// TODO - redocument
// MigrateRowDown sets num_diagnostics back to zero to undo the migration up direction.
func (m *scipSymbolsMigrator) MigrateDown(ctx context.Context, uploadID int, tx *basestore.Store, rows *sql.Rows) (_ [][]any, err error) {
	defer func() { err = basestore.CloseRows(rows, err) }()

	return nil, errors.New("down unimplemented")
}
