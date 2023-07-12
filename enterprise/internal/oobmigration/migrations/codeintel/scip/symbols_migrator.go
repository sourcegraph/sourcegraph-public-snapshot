package scip

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/symbols"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func getEnv(name string, defaultValue int) int {
	if value, _ := strconv.Atoi(os.Getenv(name)); value != 0 {
		return value
	}

	return defaultValue
}

var (
	// NOTE: modified in tests
	symbolsMigratorConcurrencyLevel      = getEnv("SYMBOLS_MIGRATOR_CONCURRENCY_LEVEL", 1)
	symbolsMigratorSymbolRecordBatchSize = getEnv("SYMBOLS_MIGRATOR_UPLOAD_BATCH_SIZE", 10000)
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
		batchSize:     symbolsMigratorSymbolRecordBatchSize,
		numRoutines:   symbolsMigratorConcurrencyLevel,
		fields: []fieldSpec{
			{name: "symbol_id", postgresType: "integer not null", primaryKey: true},
			{name: "document_lookup_id", postgresType: "integer not null", primaryKey: true},
		},
	})
}

func (m *scipSymbolsMigrator) ID() int                 { return 24 }
func (m *scipSymbolsMigrator) Interval() time.Duration { return time.Second }

func (m *scipSymbolsMigrator) MigrateUp(ctx context.Context, uploadID int, tx *basestore.Store, rows *sql.Rows) (_ [][]any, err error) {
	start := time.Now()

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

	// Convert symbol names into a tree structure we'll insert into the database
	// All identifiers here are created ahead of the insertion so we do not need
	// to do multiple round-trips to get new insertion identifiers for pending
	// data - everything is known up-front.

	id := func() int { id := nextSymbolLookupID; nextSymbolLookupID++; return id }
	cache, traverser, err := constructSymbolLookupTable(symbolNames, id)
	if err != nil {
		return nil, err
	}

	// Bulk insert the content of the tree / descriptor-no-suffix map
	symbolNamePartInserter := func(ctx context.Context, symbolLookupInserter *batch.Inserter) error {
		visit := func(segmentType, name string, id int, parentID *int) error {
			return symbolLookupInserter.Insert(ctx, segmentType, name, id, parentID)
		}
		return traverser(visit)
	}

	// In the same transaction but ouf-of-band from the row updates, batch-insert new
	// symbol-lookup rows. These identifiers need to exist before the batch is complete.
	if err := withSymbolLookupInserter(ctx, tx, uploadID, symbolNamePartInserter); err != nil {
		return nil, err
	}

	// Bulk insert descriptor/descriptor-no-suffix pairs with relations to their symbol
	symbolRelationshipInserter := func(ctx context.Context, symbolLookupLeavesInserter *batch.Inserter) error {
		for _, symbolInDocument := range symbolInDocuments {
			symbolID := symbolInDocument.symbolID
			ids := cache[symbolNamesByID[symbolID]]

			if err := symbolLookupLeavesInserter.Insert(ctx, symbolID, ids.descriptorSuffixID, ids.fuzzyDescriptorSuffixID); err != nil {
				return err
			}
		}

		return nil
	}

	// In the same transaction but ouf-of-band from the row updates, batch-insert new
	// symbol-lookup-leaves rows. These identifiers need to exist before the batch is complete.
	if err := withSymbolLookupLeavesInserter(ctx, tx, uploadID, symbolRelationshipInserter); err != nil {
		return nil, err
	}

	// Construct the updated tuples for the symbols rows we have locked in this transaction.
	// Each (original) symbol identifier is translated into the new descriptor identifier
	// pairs batch inserted into the symbols lookup table (above). We're not supplying any
	// additional information here, so we'll end up just writing foreign keys _to_ these rows
	// and bumping the schema version.
	values := make([][]any, 0, len(symbolInDocuments))
	for _, symbolInDocument := range symbolInDocuments {
		values = append(values, []any{
			symbolInDocument.symbolID,
			symbolInDocument.documentLookupID,
		})
	}

	fmt.Printf("> updated %d rows and inserted %d symbols in %s\n", len(values), len(symbolNames), time.Since(start))
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
					ssn.prefix_id
				FROM codeintel_scip_symbol_names ssn
				WHERE
					ssn.id = ANY(%s) AND
					ssn.upload_id = %s
			) UNION (
				SELECT
					s.id,
					s.upload_id,
					ssn.name_segment || s.suffix AS suffix,
					ssn.prefix_id
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
			segment_type SymbolNameSegmentType NOT NULL,
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
		"segment_type",
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
		INSERT INTO codeintel_scip_symbols_lookup (id, upload_id, name, segment_type, parent_id)
		SELECT id, %s, name, segment_type, parent_id
		FROM t_codeintel_scip_symbols_lookup
	`,
		uploadID,
	))
}

func withSymbolLookupLeavesInserter(ctx context.Context, tx *basestore.Store, uploadID int, f inserterFunc) error {
	if err := tx.Exec(ctx, sqlf.Sprintf(`
		CREATE TEMPORARY TABLE t_codeintel_scip_symbols_lookup_leaves(
			symbol_id integer NOT NULL,
			descriptor_suffix_id integer NOT NULL,
			fuzzy_descriptor_suffix_id integer NOT NULL
		) ON COMMIT DROP
	`)); err != nil {
		return err
	}

	symbolLookupLeavesInserter := batch.NewInserter(
		ctx,
		tx.Handle(),
		"t_codeintel_scip_symbols_lookup_leaves",
		batch.MaxNumPostgresParameters,
		"symbol_id",
		"descriptor_suffix_id",
		"fuzzy_descriptor_suffix_id",
	)

	if err := f(ctx, symbolLookupLeavesInserter); err != nil {
		return err
	}
	if err := symbolLookupLeavesInserter.Flush(ctx); err != nil {
		return err
	}

	return tx.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO codeintel_scip_symbols_lookup_leaves (upload_id, symbol_id, descriptor_suffix_id, fuzzy_descriptor_suffix_id)
		SELECT %s, symbol_id, descriptor_suffix_id, fuzzy_descriptor_suffix_id
		FROM t_codeintel_scip_symbols_lookup_leaves
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

// NOTE(scip-migration): This behavior in the following function has been copied from the upload
// processing procedure to match the logic for newly uploaded index files. See the original code
// in the uploads service for more detail (.../lsifstore/insert.go).

type explodedIDs struct {
	descriptorSuffixID      int
	fuzzyDescriptorSuffixID int
}

type visitFunc func(segmentType, name string, id int, parentID *int) error

func constructSymbolLookupTable(symbolNames []string, id func() int) (map[string]explodedIDs, func(visit visitFunc) error, error) {
	// Create helpers to create new tree nodes with (upload-)unique identifiers
	createSchemeNode := func() SchemeNode { return SchemeNode(newNodeWithID[PackageManagerNode](id())) }
	createPackageManagerNode := func() PackageManagerNode { return PackageManagerNode(newNodeWithID[PackageNameNode](id())) }
	createPackageNameNode := func() PackageNameNode { return PackageNameNode(newNodeWithID[PackageVersionNode](id())) }
	createPackageVersionNode := func() PackageVersionNode { return PackageVersionNode(newNodeWithID[NamespaceNode](id())) }
	createNamespaceNode := func() NamespaceNode { return NamespaceNode(newNodeWithID[DescriptorNode](id())) }
	createDescriptor := func() DescriptorNode { return DescriptorNode(newNodeWithID[descriptor](id())) }

	cache := map[string]explodedIDs{}                // Tracks symbol name -> identifiers in the scheme tree
	schemeTree := map[string]SchemeNode{}            // Tracks scheme -> manager -> name -> version -> descriptor (namespace, suffix)
	fuzzyDescriptorSuffixMap := make(map[string]int) // Tracks fuzzy descriptor

	for _, symbolName := range symbolNames {
		symbol, err := symbols.NewExplodedSymbol(symbolName)
		if err != nil {
			return nil, nil, err
		}

		// Assign the parts of the exploded symbol into the scheme tree. If a prefix of
		// the exploded symbol is already in the tree then existing nodes will be re-used.
		// Laying out the exploded in a tree structure will allow us to trace parentage
		// (required for fast lookups) when we insert these into the database.

		schemeNode := getOrCreate(schemeTree, symbol.Scheme, createSchemeNode)                                       // depth 0
		packageManagerNode := getOrCreate(schemeNode.children, symbol.PackageManager, createPackageManagerNode)      // depth 1
		packageNameNode := getOrCreate(packageManagerNode.children, symbol.PackageName, createPackageNameNode)       // depth 2
		packageVersionNode := getOrCreate(packageNameNode.children, symbol.PackageVersion, createPackageVersionNode) // depth 3
		namespace := getOrCreate(packageVersionNode.children, symbol.DescriptorNamespace, createNamespaceNode)       // depth 4
		descriptor := getOrCreate(namespace.children, symbol.DescriptorSuffix, createDescriptor)                     // depth 5
		fuzzyDescriptorsSuffixID := getOrCreate(fuzzyDescriptorSuffixMap, symbol.FuzzyDescriptorSuffix, id)          // map insertion

		cache[symbolName] = explodedIDs{
			descriptorSuffixID:      descriptor.id,
			fuzzyDescriptorSuffixID: fuzzyDescriptorsSuffixID,
		}
	}

	segmentTypeByDepth := []string{
		"SCHEME",               // depth 0
		"PACKAGE_MANAGER",      // depth 1
		"PACKAGE_NAME",         // depth 2
		"PACKAGE_VERSION",      // depth 3
		"DESCRIPTOR_NAMESPACE", // depth 4
		"DESCRIPTOR_SUFFIX",    // depth 5
		/*                   */ // depth PANIC
	}

	traverser := func(visit visitFunc) error {
		// Call visit on each node of the tree
		if err := traverse(schemeTree, func(name string, id, depth int, parentID *int) error {
			return visit(segmentTypeByDepth[depth], name, id, parentID)
		}); err != nil {
			return err
		}

		// Call visit on each element in the descriptor-no-suffix map
		for name, id := range fuzzyDescriptorSuffixMap {
			if err := visit("DESCRIPTOR_SUFFIX_FUZZY", name, id, nil); err != nil {
				return err
			}
		}

		return nil
	}

	return cache, traverser, nil
}
