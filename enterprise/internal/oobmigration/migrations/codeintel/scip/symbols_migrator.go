package scip

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/symbols"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
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
		batchSize:     10000,
		numRoutines:   1,
		fields: []fieldSpec{
			{name: "symbol_id", postgresType: "integer not null", primaryKey: true},
			{name: "document_lookup_id", postgresType: "integer not null", primaryKey: true},
			{name: "descriptor_id", postgresType: "integer", updateOnly: true},
			{name: "descriptor_no_suffix_id", postgresType: "integer", updateOnly: true},
		},
	})
}

func (m *scipSymbolsMigrator) ID() int                 { return 24 }
func (m *scipSymbolsMigrator) Interval() time.Duration { return time.Second }

func (m *scipSymbolsMigrator) MigrateUp(ctx context.Context, uploadID int, tx *basestore.Store, rows *sql.Rows) (_ [][]any, err error) {
	// Consume symbol_id/document_id pairs from the incoming rows
	symbolInDocuments, err := readSymbolInDocuments(rows)
	if err != nil {
		return nil, err
	}
	symbolIDMap := make(map[int]struct{}, len(symbolInDocuments))
	for _, pair := range symbolInDocuments {
		symbolIDMap[pair.symbolID] = struct{}{}
	}
	symbolIDs := flattenKeys(symbolIDMap)

	// Reconstruct the full symbol names for each of the symbol IDs in this batch
	symbolNamesByID, err := readSymbolNamesBySymbolIDs(ctx, tx, uploadID, symbolIDs)
	if err != nil {
		return nil, err
	}
	symbolNames := flattenValues(symbolNamesByID)

	// An upload's symbols may be processed over several batches, and each symbol
	// identifier needs to be unique per upload, so we track the highest symbol
	// identifier written by the migration per upload. We read the last written
	// value here (or default zero), and write our next highest identifier upon
	// (successful) exit of this method.
	nextSymbolLookupID, err := getNextSymbolID(ctx, tx, uploadID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = setNextSymbolID(ctx, tx, uploadID, nextSymbolLookupID)
		}
	}()

	// NOTE(scip-migration): This behavior in the following function has been copied from the upload
	// processing procedure to match the logic for newly uploaded index files. See the original code
	// in the uploads service for more detail (.../lsifstore/insert.go).

	// Create helpers to create new tree nodes with (upload-)unique identifiers
	id := func() int { id := nextSymbolLookupID; nextSymbolLookupID++; return id }
	createSchemeNode := func() SchemeNode { return SchemeNode(newNodeWithID[PackageManagerNode](id())) }
	createPackageManagerNode := func() PackageManagerNode { return PackageManagerNode(newNodeWithID[PackageNameNode](id())) }
	createPackageNameNode := func() PackageNameNode { return PackageNameNode(newNodeWithID[PackageVersionNode](id())) }
	createPackageVersionNode := func() PackageVersionNode { return PackageVersionNode(newNodeWithID[DescriptorNode](id())) }
	createDescriptor := func() DescriptorNode { return DescriptorNode(newNodeWithID[descriptor](id())) }

	type explodedIDs struct {
		descriptorID         int
		descriptorNoSuffixID int
	}
	cache := map[string]explodedIDs{}              // Tracks symbol name -> identifiers in the scheme tree
	schemeTree := map[string]SchemeNode{}          // Tracks scheme -> manager -> name -> version -> descriptor
	descriptorsNoSuffixMap := make(map[string]int) // Tracks fuzzy descriptor

	symbolNameBatchMigrator := func(ctx context.Context, symbolLookupInserter *batch.Inserter) error {
		for _, symbolName := range symbolNames {
			symbol, err := symbols.NewExplodedSymbol(symbolName)
			if err != nil {
				return err
			}

			// Assign the parts of the exploded symbol into the scheme tree. If a prefix of
			// the exploded symbol is already in the tree then existing nodes will be re-used.
			// Laying out the exploded in a tree structure will allow us to trace parentage
			// (required for fast lookups) when we insert these into the database.

			schemeNode := getOrCreate(schemeTree, symbol.Scheme, createSchemeNode)
			packageManagerNode := getOrCreate(schemeNode.children, symbol.PackageManager, createPackageManagerNode)
			packageNameNode := getOrCreate(packageManagerNode.children, symbol.PackageName, createPackageNameNode)
			packageVersionNode := getOrCreate(packageNameNode.children, symbol.PackageVersion, createPackageVersionNode)
			descriptor := getOrCreate(packageVersionNode.children, symbol.Descriptor, createDescriptor)
			descriptorsNoSuffixID := getOrCreate(descriptorsNoSuffixMap, symbol.DescriptorNoSuffix, id)

			cache[symbolName] = explodedIDs{
				descriptorID:         descriptor.id,
				descriptorNoSuffixID: descriptorsNoSuffixID,
			}
		}

		scipNameTypeByDepth := []string{
			"SCHEME",          // depth 0
			"PACKAGE_MANAGER", // depth 1
			"PACKAGE_NAME",    // depth 2
			"PACKAGE_VERSION", // depth 3
			"DESCRIPTOR",      // depth 4
			/*              */ // depth PANIC
		}

		// Bulk insert the content of the tree
		if err := traverse(schemeTree, func(name string, id, depth int, parentID *int) error {
			return symbolLookupInserter.Insert(ctx, scipNameTypeByDepth[depth], name, id, parentID)
		}); err != nil {
			return err
		}

		// Bulk insert fuzzy descriptors
		for name, id := range descriptorsNoSuffixMap {
			if err := symbolLookupInserter.Insert(ctx, "DESCRIPTOR_NO_SUFFIX", name, id, nil); err != nil {
				return err
			}
		}

		return nil
	}

	// Batch-insert new symbol-lookup rows before batch updating the symbols table (below)
	if err := withSymbolLookupInserter(ctx, tx, uploadID, symbolNameBatchMigrator); err != nil {
		return nil, err
	}

	// Construct the updated tuples for the symbols rows we have locked in this transaction.
	// Each (original) symbol identifier is translated into the new descriptor identifier
	// pairs batch inserted into the symbols lookup table (above).
	values := make([][]any, 0, len(symbolInDocuments))
	for _, symbolInDocument := range symbolInDocuments {
		ids := cache[symbolNamesByID[symbolInDocument.symbolID]]

		values = append(values, []any{
			symbolInDocument.symbolID,
			symbolInDocument.documentLookupID,
			ids.descriptorID,
			ids.descriptorNoSuffixID,
		})
	}

	return values, nil
}

//
//
//

func (m *scipSymbolsMigrator) MigrateDown(ctx context.Context, uploadID int, tx *basestore.Store, rows *sql.Rows) (_ [][]any, err error) {
	// Consume symbol_id/document_id pairs from the incoming rows
	symbolInDocuments, err := readSymbolInDocuments(rows)
	if err != nil {
		return nil, err
	}

	// Remove the keys we added in the up direction
	values := make([][]any, 0, len(symbolInDocuments))
	for _, symbolInDocument := range symbolInDocuments {
		values = append(values, []any{
			symbolInDocument.symbolID,
			symbolInDocument.documentLookupID,
			nil,
			nil,
		})
	}

	return values, nil
}

//
//
//

type symbolInDocument struct {
	symbolID         int
	documentLookupID int
}

func readSymbolInDocuments(rows *sql.Rows) ([]symbolInDocument, error) {
	return scanSymbolInDocuments(rows, nil)
}

var scanSymbolInDocuments = basestore.NewSliceScanner(func(s dbutil.Scanner) (sd symbolInDocument, _ error) {
	err := s.Scan(&sd.symbolID, &sd.documentLookupID)
	return sd, err
})

//
//

func readSymbolNamesBySymbolIDs(ctx context.Context, tx *basestore.Store, uploadID int, symbolIDs []int) (map[int]string, error) {
	return scanSymbolNamesByID(tx.Query(ctx, sqlf.Sprintf(`
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
}

var scanSymbolNamesByID = basestore.NewMapScanner(func(s dbutil.Scanner) (symbolID int, symbolName string, _ error) {
	err := s.Scan(&symbolID, &symbolName)
	return symbolID, symbolName, err
})

//
//

func getNextSymbolID(ctx context.Context, tx *basestore.Store, uploadID int) (int, error) {
	nextSymbolLookupID, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(`
		SELECT symbol_id
		FROM codeintel_scip_symbols_migration_progress
		WHERE upload_id = %s
	`,
		uploadID,
	)))
	return nextSymbolLookupID, err
}

func setNextSymbolID(ctx context.Context, tx *basestore.Store, uploadID, id int) error {
	return tx.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO codeintel_scip_symbols_migration_progress (upload_id, symbol_id)
		VALUES (%s, %s)
		ON CONFLICT (upload_id) DO UPDATE SET symbol_id = EXCLUDED.symbol_id
	`,
		uploadID,
		id,
	))
}

//
//

type inserterFunc func(ctx context.Context, symbolLookupInserter *batch.Inserter) error

func withSymbolLookupInserter(ctx context.Context, tx *basestore.Store, uploadID int, f inserterFunc) error {
	if err := tx.Exec(ctx, sqlf.Sprintf(`
		CREATE TEMPORARY TABLE t_codeintel_scip_symbols_lookup(
			name text NOT NULL,
			id integer NOT NULL,
			scip_name_type text NOT NULL,
			parent_id integer
		) ON COMMIT DROP
	`)); err != nil {
		return err
	}

	symbolLookupInserter := batch.NewInserter(
		ctx,
		tx.Handle(),
		"t_codeintel_scip_symbols_lookup",
		batch.MaxNumPostgresParameters,
		"scip_name_type",
		"name",
		"id",
		"parent_id",
	)

	if err := f(ctx, symbolLookupInserter); err != nil {
		return err
	}
	if err := symbolLookupInserter.Flush(ctx); err != nil {
		return err
	}

	return tx.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO codeintel_scip_symbols_lookup (id, upload_id, name, scip_name_type, parent_id)
		SELECT id, %s, name, scip_name_type, parent_id
		FROM t_codeintel_scip_symbols_lookup
	`,
		uploadID,
	))
}

//
//

func flattenKeys[K int | string, V any](m map[K]V) []K {
	ss := make([]K, 0, len(m))
	for v := range m {
		ss = append(ss, v)
	}
	sort.Slice(ss, func(i, j int) bool { return ss[i] < ss[j] })

	return ss
}

func flattenValues[K comparable, V int | string](m map[K]V) []V {
	ss := make([]V, 0, len(m))
	for _, v := range m {
		ss = append(ss, v)
	}
	sort.Slice(ss, func(i, j int) bool { return ss[i] < ss[j] })

	return ss
}
