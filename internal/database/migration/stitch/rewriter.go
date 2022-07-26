package stitch

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/mapfs"
)

var versionPattern = lazyregexp.New(`^v(\d+)\.(\d+)\.\d+$`)

// readMigrations reads migrations from a locally available git revision for the given schema, and
// rewrites old versions and explicit edge cases so that they can be more easily composed by the
// migration stitch utilities.
//
// The returned FS serves a hierarchical set of contents where the following files are available in
// a directory named equivalently to the migration identifier:
//   - up.sql
//   - down.sql
//   - metadata.yaml
//
// For historic revisions, squashed migrations are not necessarily split into privileged unprivileged
// cateogories. When there is a single squashed migration, this function will extract the privileged
// statements into a new migration. These migrations will have a negative-valued identifier, whose
// absolute value indicates the squashed migration it was split from. NOTE: Callers must take care to
// stitch these relations back together, as it can't be done easily pre-composition across versions.
//
// See the method `linkVirtualPrivilegedMigrations`.
func readMigrations(schemaName, root, rev string) (fs.FS, error) {
	migrations, err := readRawMigrations(schemaName, root, rev)
	if err != nil {
		return nil, err
	}

	replacer := strings.NewReplacer(
		// These lines cause issues with schema drift comparison
		"-- Increment tally counting tables.\n", "",
		"-- Decrement tally counting tables.\n", "",
	)

	contents := make(map[string]string, len(migrations)*3)
	for _, m := range migrations {
		contents[filepath.Join(m.id, "up.sql")] = replacer.Replace(m.up)
		contents[filepath.Join(m.id, "down.sql")] = replacer.Replace(m.down)
		contents[filepath.Join(m.id, "metadata.yaml")] = m.metadata
	}

	if matches := versionPattern.FindStringSubmatch(rev); len(matches) > 0 {
		majorVersion, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, err
		}
		minorVersion, err := strconv.Atoi(matches[2])
		if err != nil {
			return nil, err
		}

		migrationIDs, err := idsFromRawMigrations(migrations)
		if err != nil {
			return nil, err
		}

		for _, rewrite := range rewriters {
			rewrite(schemaName, majorVersion, minorVersion, migrationIDs, contents)
		}
	}

	return mapfs.New(contents), nil
}

// linkVirtualPrivilegedMigrations ensures that the parent relationships in the given migration graph
// remains well-formed after the set of rewriters defined below have been invoked. These writers may
// clean up some temporary state when being applied locally that we need to clean up once combined.
//
// This function should be called after all migrations have been composed across versions.
func linkVirtualPrivilegedMigrations(definitionMap map[int]definition.Definition) {
	// Gather migration identifiers with a virtual counterpart
	squashedIDs := make([]int, 0, len(definitionMap))
	for id := range definitionMap {
		if id < 0 {
			squashedIDs = append(squashedIDs, -id)
		}
	}
	sort.Ints(squashedIDs)

	for i, id := range squashedIDs {
		if i == 0 {
			// Keep first virtual migration only
			replaceParentsInDefinitionMap(definitionMap, -id, nil)
			replaceParentsInDefinitionMap(definitionMap, +id, []int{-id})
		} else {
			delete(definitionMap, -id)
		}
	}
}

// rewriters alter the raw migrations read from a previous git revision to resemble the format that
// is expected by the current version of the migration definition reader and validator components.
//
// Each rewriter can alter the contents map, which indexes file contents by its path within the base
// migration directory. Each rewriter is given the minor version of git revision to conditionally alter
// the state of the contents map only before or after a specific release. Each migration known at the
// beginning of the rewrite procedure will be represented in the provided slide of identifiers.
// Additional files/migrations may be added to the contents map will not be reflected in this slice for
// subsequent rewriters.
var rewriters = []func(schemaName string, majorVersion, minorVersion int, migrationIDs []int, contents map[string]string){
	rewriteInitialCodeinsightsMigration,
	ensureParentMetadataExists,
	extractPrivilegedQueriesFromSquashedMigrations,

	rewriteUnmarkedPrivilegedMigrations,
	rewriteUnmarkedConcurrentIndexCreationMigrations,
	rewriteConcurrentIndexCreationDownMigrations,
	reorderMigrations,
}

// rewriteInitialCodeinsightsMigration renames the initial codeinsights migration file to include the expected
// title of "squashed migration".
func rewriteInitialCodeinsightsMigration(schemaName string, _, _ int, _ []int, contents map[string]string) {
	if schemaName != "codeinsights" {
		return
	}

	mapContents(contents, migrationFilename(1000000000, "metadata.yaml"), func(oldMetadata string) string {
		return fmt.Sprintf("name: %s", squashedMigrationPrefix)
	})
}

// ensureParentMetadataExists adds parent information to the metadata file of each migration, prior to 3.37,
// in which metadata files did not exist and parentage was implied by linear migration identifiers.
func ensureParentMetadataExists(_ string, majorVersion, minorVersion int, migrationIDs []int, contents map[string]string) {
	// 3.37 and above enforces this structure
	if !(majorVersion == 3 && minorVersion < 37) || len(migrationIDs) == 0 {
		return
	}

	for _, id := range migrationIDs[1:] {
		mapContents(contents, migrationFilename(id, "metadata.yaml"), func(oldMetadata string) string {
			return replaceParents(oldMetadata, id-1)
		})
	}
}

// extractPrivilegedQueriesFromSquashedMigrations splits the squashed migration into a distinct set of
// privileged and unprivileged queries. Prior to 3.38, privileged migrations were not distinct. The current
// code that reads migration definitions require that privileged migrations are expilcitly marked.
func extractPrivilegedQueriesFromSquashedMigrations(_ string, majorVersion, minorVersion int, migrationIDs []int, contents map[string]string) {
	if !(majorVersion == 3 && minorVersion < 38) || len(migrationIDs) == 0 {
		// 3.38 and above enforces this structure
		return
	}

	squashID := migrationIDs[0]
	oldMetadata := contents[migrationFilename(squashID, "metadata.yaml")]
	oldUpQuery := contents[migrationFilename(squashID, "up.sql")]
	newMetadata := "name: 'squashed migrations (privileged)'\nprivileged: true"
	privilegedUpQuery, unprivilegedUpQuery := partitionPrivilegedQueries(oldUpQuery)

	// Add new privileged squashed migration
	contents[migrationFilename(-squashID, "up.sql")] = privilegedUpQuery
	contents[migrationFilename(-squashID, "down.sql")] = ""
	contents[migrationFilename(-squashID, "metadata.yaml")] = newMetadata

	// Remove privileged statements from unprivileged squashed migration
	contents[migrationFilename(squashID, "up.sql")] = unprivilegedUpQuery

	// Make unprivileged squashed migration a direct child of the new privileged squashed migration
	contents[migrationFilename(squashID, "metadata.yaml")] = replaceParents(oldMetadata, -squashID)
}

var unmarkedPrivilegedMigrationsMap = map[string][]int{
	"frontend":     {1528395953},
	"codeintel":    {1000000020},
	"codeinsights": {1000000001, 1000000027},
}

// rewriteUnmarkedPrivilegedMigrations adds an explicit privileged marker to the metadata of migration
// definitions that modify extensions (prior to the privileged/unprivileged split).
func rewriteUnmarkedPrivilegedMigrations(schemaName string, _, _ int, _ []int, contents map[string]string) {
	for _, id := range unmarkedPrivilegedMigrationsMap[schemaName] {
		mapContents(contents, migrationFilename(id, "metadata.yaml"), func(oldMetadata string) string {
			return fmt.Sprintf("%s\nprivileged: true", oldMetadata)
		})
	}
}

var unmarkedConcurrentIndexCreationMigrationsMap = map[string][]int{
	"frontend":     {1528395797, 1528395877, 1528395878, 1528395886, 1528395887, 1528395888, 1528395893, 1528395894, 1528395896, 1528395897, 1528395899, 1528395900, 1528395935, 1528395936, 1528395954},
	"codeintel":    {1000000009, 1000000010, 1000000011},
	"codeinsights": {},
}

// rewriteUnmarkedConcurrentIndexCreationMigrations adds an explicit marker to the metadata of migrations that
// define a concurrent index (prior to the introduction of the migrator).
func rewriteUnmarkedConcurrentIndexCreationMigrations(schemaName string, _, _ int, _ []int, contents map[string]string) {
	for _, id := range unmarkedConcurrentIndexCreationMigrationsMap[schemaName] {
		mapContents(contents, migrationFilename(id, "metadata.yaml"), func(oldMetadata string) string {
			return fmt.Sprintf("%s\ncreateIndexConcurrently: true", oldMetadata)
		})
	}
}

var concurrentIndexCreationDownMigrationsMap = map[string][]int{
	"frontend":     {1528395895, 1528395901, 1528395902, 1528395903, 1528395904, 1528395905, 1528395906},
	"codeintel":    {},
	"codeinsights": {},
}

// rewriteConcurrentIndexCreationDownMigrations removes CONCURRENTLY from down migrations, which is now unsupported.
func rewriteConcurrentIndexCreationDownMigrations(schemaName string, _, _ int, _ []int, contents map[string]string) {
	for _, id := range concurrentIndexCreationDownMigrationsMap[schemaName] {
		mapContents(contents, migrationFilename(id, "down.sql"), func(oldQuery string) string {
			return strings.ReplaceAll(oldQuery, " CONCURRENTLY", "")
		})
	}
}

// reorderMigrations reproduces an explicit (historic) reodering of several migration files. For versions where
// these files exist and haven't yet been renamed, we do the renaming at this time to make it match later versions.
//
// See https://github.com/sourcegraph/sourcegraph/pull/29395.
func reorderMigrations(schemaName string, majorVersion, minorVersion int, _ []int, contents map[string]string) {
	if schemaName != "frontend" || !(majorVersion == 3 && minorVersion < 36) {
		// Rename occurred at v3.36
		return
	}

	for oldID, newID := range map[int]int{
		1528395945: 1528395961,
		1528395946: 1528395962,
		1528395947: 1528395963,
		1528395948: 1528395964,
	} {
		if _, ok := contents[migrationFilename(oldID, "metadata.yaml")]; !ok {
			// File doesn't exist at this verson (nothing to rewrite)
			continue
		}

		// Move new contents and replace previous contents
		noopContents := "-- NO-OP to fix out of sequence migrations"
		contents[migrationFilename(newID, "up.sql")] = contents[migrationFilename(oldID, "up.sql")]
		contents[migrationFilename(newID, "down.sql")] = contents[migrationFilename(oldID, "down.sql")]
		contents[migrationFilename(oldID, "up.sql")] = noopContents
		contents[migrationFilename(oldID, "down.sql")] = noopContents

		// Write new metadata
		oldMetadata := contents[migrationFilename(oldID, "metadata.yaml")]
		contents[migrationFilename(newID, "metadata.yaml")] = replaceParents(oldMetadata, newID-1)
	}
}

func idsFromRawMigrations(rawMigrations []rawMigration) ([]int, error) {
	ids := make([]int, 0, len(rawMigrations))
	for _, rawMigration := range rawMigrations {
		id, err := strconv.Atoi(rawMigration.id)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	sort.Ints(ids)
	return ids, nil
}

func migrationFilename(id int, filename string) string {
	return filepath.Join(strconv.Itoa(id), filename)
}

// mapContents transforms and replaces the contents of the given filename, if it is already present in the map.
// An absent entry results in a no-op.
func mapContents(contents map[string]string, filename string, f func(v string) string) {
	if v, ok := contents[filename]; ok {
		contents[filename] = f(v)
	}
}

var yamlParentsPattern = lazyregexp.New(`parents: \[[\d,]+\]`)

// removeParents removes the `parents: ` line from the given YAML file contents.
func removeParents(contents string) string {
	return yamlParentsPattern.ReplaceAllString(contents, "")
}

// replacesParents removes the `parents: ` line from the given YAML file contents and inserts a new line with the
// given parent identifiers.
func replaceParents(contents string, parents ...int) string {
	strParents := make([]string, 0, len(parents))
	for _, id := range parents {
		strParents = append(strParents, strconv.Itoa(id))
	}

	return removeParents(contents) + fmt.Sprintf("\nparents: [%s]", strings.Join(strParents, ", "))
}

// replaceParentsInDefinitionMap updates the `parents` field of the definition with the given identifier.
func replaceParentsInDefinitionMap(definitionMap map[int]definition.Definition, id int, parents []int) {
	definition := definitionMap[id]
	definition.Parents = parents
	definitionMap[id] = definition
}

var alterExtensionPattern = lazyregexp.New(`(?:CREATE|COMMENT ON|DROP)\s+EXTENSION.*;`)

// partitionPrivilegedQueries partitions the lines of the given query into privileged and unprivileged queries.
func partitionPrivilegedQueries(query string) (privileged string, unprivileged string) {
	var matches []string
	for _, match := range alterExtensionPattern.FindAllStringSubmatch(query, -1) {
		matches = append(matches, match[0])
	}

	return strings.Join(matches, "\n\n"), alterExtensionPattern.ReplaceAllString(query, "")
}
