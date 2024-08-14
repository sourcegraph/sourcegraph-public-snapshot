package definition

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/keegancsmith/sqlf"
	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func ReadDefinitions(fs fs.FS, schemaBasePath string) (*Definitions, error) {
	migrationDefinitions, err := readDefinitions(fs, schemaBasePath)
	if err != nil {
		return nil, errors.Wrap(err, "readDefinitions")
	}

	if err := reorderDefinitions(migrationDefinitions); err != nil {
		return nil, errors.Wrap(err, "reorderDefinitions")
	}

	return newDefinitions(migrationDefinitions), nil
}

type instructionalError struct {
	class        string
	description  string
	instructions string
}

func (e instructionalError) Error() string {
	return fmt.Sprintf("%s: %s\n\n%s\n", e.class, e.description, e.instructions)
}

func readDefinitions(fs fs.FS, schemaBasePath string) ([]Definition, error) {
	root, err := http.FS(fs).Open("/")
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()

	migrations, err := root.Readdir(0)
	if err != nil {
		return nil, err
	}

	definitions := make([]Definition, 0, len(migrations))
	for _, file := range migrations {
		version, err := ParseRawVersion(file.Name())
		if err != nil {
			continue // not a versioned migration file, ignore
		}

		definition, err := readDefinition(fs, schemaBasePath, version, file.Name())
		if err != nil {
			return nil, errors.Wrapf(err, "malformed migration definition at '%s'",
				filepath.Join(schemaBasePath, file.Name()))
		}
		definitions = append(definitions, definition)
	}

	sort.Slice(definitions, func(i, j int) bool { return definitions[i].ID < definitions[j].ID })

	return definitions, nil
}

func readDefinition(fs fs.FS, schemaBasePath string, version int, filename string) (Definition, error) {
	upFilename := fmt.Sprintf("%s/up.sql", filename)
	downFilename := fmt.Sprintf("%s/down.sql", filename)
	metadataFilename := fmt.Sprintf("%s/metadata.yaml", filename)

	upQuery, err := readQueryFromFile(fs, upFilename)
	if err != nil {
		return Definition{}, err
	}

	downQuery, err := readQueryFromFile(fs, downFilename)
	if err != nil {
		return Definition{}, err
	}

	return hydrateMetadataFromFile(fs, schemaBasePath, upFilename, metadataFilename, Definition{
		ID:        version,
		UpQuery:   upQuery,
		DownQuery: downQuery,
	})
}

// hydrateMetadataFromFile populates the given definition with metdata parsed
// from the given file. The mutated definition is returned.
func hydrateMetadataFromFile(fs fs.FS, schemaBasePath, upFilename, metadataFilename string, definition Definition) (_ Definition, _ error) {
	file, err := fs.Open(metadataFilename)
	if err != nil {
		return Definition{}, err
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		return Definition{}, err
	}

	var payload struct {
		Name                    string `yaml:"name"`
		Parent                  int    `yaml:"parent"`
		Parents                 []int  `yaml:"parents"`
		CreateIndexConcurrently bool   `yaml:"createIndexConcurrently"`
		Privileged              bool   `yaml:"privileged"`
		NonIdempotent           bool   `yaml:"nonIdempotent"`
	}
	if err := yaml.Unmarshal(contents, &payload); err != nil {
		return Definition{}, err
	}

	definition.Name = payload.Name
	definition.Privileged = payload.Privileged
	definition.NonIdempotent = payload.NonIdempotent

	parents := payload.Parents
	if payload.Parent != 0 {
		parents = append(parents, payload.Parent)
	}
	sort.Ints(parents)
	definition.Parents = parents

	schemaPath := filepath.Join(schemaBasePath, strconv.Itoa(definition.ID))
	upPath := filepath.Join(schemaBasePath, upFilename)
	metadataPath := filepath.Join(schemaBasePath, metadataFilename)

	if _, ok := parseIndexMetadata(definition.DownQuery.Query(sqlf.PostgresBindVar)); ok {
		return Definition{}, instructionalError{
			class:       "malformed concurrent index creation",
			description: fmt.Sprintf("did not expect down query of migration at '%s' to contain concurrent creation of an index", schemaPath),
			instructions: strings.Join([]string{
				"Remove `CONCURRENTLY` when re-creating an old index in down migrations (if you're seeing this in a local dev environment, try running `sg update` to see if it fixes the issue first).",
				"Downgrades indicate an instance stability error which generally requires a maintenance window.",
			}, " "),
		}
	}

	upQueryText := definition.UpQuery.Query(sqlf.PostgresBindVar)
	if indexMetadata, ok := parseIndexMetadata(upQueryText); ok {
		if !payload.CreateIndexConcurrently {
			return Definition{}, instructionalError{
				class:       "malformed concurrent index creation",
				description: fmt.Sprintf("did not expect up query of migration at '%s' to contain concurrent creation of an index", schemaPath),
				instructions: strings.Join([]string{
					fmt.Sprintf("Add `createIndexConcurrently: true` to the metadata file '%s'.", metadataPath),
				}, " "),
			}
		} else if removeConcurrentIndexCreation(upQueryText) != "" {
			return Definition{}, instructionalError{
				class:       "malformed concurrent index creation",
				description: fmt.Sprintf("did not expect up query of migration at '%s' to contain additional statements", schemaPath),
				instructions: strings.Join([]string{
					fmt.Sprintf("Split the index creation from '%s' into a new migration file.", upPath),
				}, " "),
			}
		}

		definition.IsCreateIndexConcurrently = true
		definition.IndexMetadata = indexMetadata
	} else if payload.CreateIndexConcurrently {
		return Definition{}, instructionalError{
			class:       "malformed concurrent index creation",
			description: fmt.Sprintf("expected up query of migration at '%s' to contain concurrent creation of an index", schemaPath),
			instructions: strings.Join([]string{
				fmt.Sprintf("Remove `createIndexConcurrently: true` from the metadata file '%s'.", metadataPath),
			}, " "),
		}
	}

	if isPrivileged(definition.UpQuery.Query(sqlf.PostgresBindVar)) || isPrivileged(definition.DownQuery.Query(sqlf.PostgresBindVar)) {
		if !payload.Privileged {
			return Definition{}, instructionalError{
				class:       "malformed Postgres extension modification",
				description: fmt.Sprintf("did not expect queries of migration at '%s' to require elevated permissions", schemaPath),
				instructions: strings.Join([]string{
					fmt.Sprintf("Add `privileged: true` to the metadata file '%s'.", metadataPath),
				}, " "),
			}
		}
	}

	return definition, nil
}

// readQueryFromFile returns the query parsed from the given file.
func readQueryFromFile(fs fs.FS, filepath string) (*sqlf.Query, error) {
	file, err := fs.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return queryFromString(string(contents)), nil
}

// queryFromString creates a sqlf Query object from the conetents of a file or serialized
// string literal. The resulting query is canonicalized. SQL placeholder values are also
// escaped, so when sqlf.Query renders it the placeholders will be valid and not replaced
// by a "missing" parameterized value.
func queryFromString(query string) *sqlf.Query {
	return sqlf.Sprintf(strings.ReplaceAll(CanonicalizeQuery(query), "%", "%%"))
}

// CanonicalizeQuery removes old cruft from historic definitions to make them conform to
// the new standards. This includes YAML metadata frontmatter as well as explicit tranaction
// blocks around golang-migrate-era migration definitions.
func CanonicalizeQuery(query string) string {
	// Strip out embedded yaml frontmatter (existed temporarily)
	parts := strings.SplitN(query, "-- +++\n", 3)
	if len(parts) == 3 {
		query = parts[2]
	}

	// Strip outermost transactions
	return strings.TrimSpace(
		strings.TrimSuffix(
			strings.TrimPrefix(
				strings.TrimSpace(query),
				"BEGIN;",
			),
			"COMMIT;",
		),
	)
}

var createIndexConcurrentlyPattern = lazyregexp.New(`CREATE\s+(?:UNIQUE\s+)?INDEX\s+CONCURRENTLY\s+(?:IF\s+NOT\s+EXISTS\s+)?([A-Za-z0-9_]+)\s+ON\s+([A-Za-z0-9_]+)`)

func parseIndexMetadata(queryText string) (*IndexMetadata, bool) {
	matches := createIndexConcurrentlyPattern.FindStringSubmatch(queryText)
	if len(matches) == 0 {
		return nil, false
	}

	return &IndexMetadata{
		TableName: matches[2],
		IndexName: matches[1],
	}, true
}

var createIndexConcurrentlyFullPattern = lazyregexp.New(createIndexConcurrentlyPattern.Re().String() + `[^;]+;`)

func removeConcurrentIndexCreation(query string) string {
	if matches := createIndexConcurrentlyFullPattern.FindStringSubmatch(query); len(matches) > 0 {
		query = strings.Replace(query, matches[0], "", 1)
	}

	return removeComments(query)
}

func removeComments(query string) string {
	filtered := []string{}
	for _, line := range strings.Split(query, "\n") {
		l := strings.TrimSpace(strings.Split(line, "--")[0])
		if l != "" {
			filtered = append(filtered, l)
		}
	}

	return strings.TrimSpace(strings.Join(filtered, "\n"))
}

var alterExtensionPattern = lazyregexp.New(`(CREATE|COMMENT ON|DROP)\s+EXTENSION`)

func isPrivileged(queryText string) bool {
	matches := alterExtensionPattern.FindStringSubmatch(queryText)
	return len(matches) != 0
}

// reorderDefinitions will re-order the given migration definitions in-place so that
// migrations occur before their dependents in the slice. An error is returned if the
// given migration definitions do not form a single-root directed acyclic graph.
func reorderDefinitions(migrationDefinitions []Definition) error {
	if len(migrationDefinitions) == 0 {
		return nil
	}

	// Stash migration definitions by identifier
	migrationDefinitionMap := make(map[int]Definition, len(migrationDefinitions))
	for _, migrationDefinition := range migrationDefinitions {
		migrationDefinitionMap[migrationDefinition.ID] = migrationDefinition
	}

	for _, migrationDefinition := range migrationDefinitions {
		for _, parent := range migrationDefinition.Parents {
			if _, ok := migrationDefinitionMap[parent]; !ok {
				return unknownMigrationError(parent, &migrationDefinition.ID)
			}
		}
	}

	// Find topological order of migrations
	order, err := findDefinitionOrder(migrationDefinitions)
	if err != nil {
		return err
	}

	for i, id := range order {
		// Re-order migration definitions slice to be in topological order. The order
		// returned by findDefinitionOrder is reversed; we want parents _before_ their
		// dependencies, so we fill this slice in backwards.
		migrationDefinitions[len(migrationDefinitions)-1-i] = migrationDefinitionMap[id]
	}

	return nil
}

// findDefinitionOrder returns an order of migration definition identifiers such that
// migrations occur only after their dependencies (parents). This assumes that the set
// of definitions provided form a single-root directed acyclic graph and fails with an
// error if this is not the case.
func findDefinitionOrder(migrationDefinitions []Definition) ([]int, error) {
	root, err := root(migrationDefinitions)
	if err != nil {
		return nil, err
	}

	// Use depth-first-search to topologically sort the migration definition sets as a
	// graph. At this point we know we have a single root; this means that the given set
	// of definitions either (a) form a connected acyclic graph, or (b) form a disconnected
	// set of graphs containing at least one cycle (by construction). In either case, we'll
	// return an error indicating that a cycle exists and that the set of definitions are
	// not well-formed.
	//
	// See the following Wikipedia article for additional intuition and description of the
	// `marks` array to detect cycles.
	// https://en.wikipedia.org/wiki/Topological_sorting#Depth-first_search

	type MarkType uint
	const (
		MarkTypeUnvisited MarkType = iota
		MarkTypeVisiting
		MarkTypeVisited
	)

	var (
		order    = make([]int, 0, len(migrationDefinitions))
		marks    = make(map[int]MarkType, len(migrationDefinitions))
		childMap = children(migrationDefinitions)

		dfs func(id int, parents []int) error
	)

	for _, children := range childMap {
		// Reverse-order each child slice. This will end up giving the output slice the
		// property that migrations not related via ancestry will be ordered by their
		// version number. This gives a nice, determinstic, and intuitive order in which
		// migrations will be applied.
		sort.Sort(sort.Reverse(sort.IntSlice(children)))
	}

	dfs = func(id int, parents []int) error {
		if marks[id] == MarkTypeVisiting {
			// We're currently processing the descendants of this node, so we have a paths in
			// both directions between these two nodes.

			// Peel off the head of the parent list until we reach the target node. This leaves
			// us with a slice starting with the target node, followed by the path back to itself.
			// We'll use this instance of a cycle in the error description.
			for len(parents) > 0 && parents[0] != id {
				parents = parents[1:]
			}
			if len(parents) == 0 || parents[0] != id {
				panic("unreachable")
			}
			cycle := append(parents, id)

			return instructionalError{
				class:       "migration dependency cycle",
				description: fmt.Sprintf("migrations %d and %d declare each other as dependencies", parents[len(parents)-1], id),
				instructions: strings.Join([]string{
					fmt.Sprintf("Break one of the links in the following cycle:\n%s", strings.Join(intsToStrings(cycle), " -> ")),
				}, " "),
			}
		}
		if marks[id] == MarkTypeVisited {
			// already visited
			return nil
		}

		marks[id] = MarkTypeVisiting
		defer func() { marks[id] = MarkTypeVisited }()

		for _, child := range childMap[id] {
			if err := dfs(child, append(append([]int(nil), parents...), id)); err != nil {
				return err
			}
		}

		// Add self _after_ adding all children recursively
		order = append(order, id)
		return nil
	}

	// Perform a depth-first traversal from the single root we found above
	if err := dfs(root, nil); err != nil {
		return nil, err
	}
	if len(order) < len(migrationDefinitions) {
		// We didn't visit every node, but we also do not have more than one root. There necessarily
		// exists a cycle that we didn't enter in the traversal from our root. Continue the traversal
		// starting from each unvisited node until we return a cycle.
		for _, migrationDefinition := range migrationDefinitions {
			if _, ok := marks[migrationDefinition.ID]; !ok {
				if err := dfs(migrationDefinition.ID, nil); err != nil {
					return nil, err
				}
			}
		}

		panic("unreachable")
	}

	return order, nil
}

// root returns the unique migration definition with no parent or an error of no such migration exists.
func root(migrationDefinitions []Definition) (int, error) {
	roots := make([]int, 0, 1)
	for _, migrationDefinition := range migrationDefinitions {
		if len(migrationDefinition.Parents) == 0 {
			roots = append(roots, migrationDefinition.ID)
		}
	}
	if len(roots) == 0 {
		return 0, instructionalError{
			class:       "no roots",
			description: "every migration declares a parent",
			instructions: strings.Join([]string{
				`There is no migration defined in this schema that does not declare a parent.`,
				`This indicates either a migration dependency cycle or a reference to a parent migration that no longer exists.`,
			}, " "),
		}
	}

	if len(roots) > 1 {
		strRoots := intsToStrings(roots)
		sort.Strings(strRoots)

		return 0, instructionalError{
			class:       "multiple roots",
			description: fmt.Sprintf("expected exactly one migration to have no parent but found %d (%v)", len(roots), roots),
			instructions: strings.Join([]string{
				`There are multiple migrations defined in this schema that do not declare a parent.`,
				`This indicates a new migration that did not correctly attach itself to an existing migration.`,
				`This may also indicate the presence of a duplicate squashed migration.`,
			}, " "),
		}
	}

	return roots[0], nil
}

func children(migrationDefinitions []Definition) map[int][]int {
	childMap := make(map[int][]int, len(migrationDefinitions))
	for _, migrationDefinition := range migrationDefinitions {
		for _, parent := range migrationDefinition.Parents {
			childMap[parent] = append(childMap[parent], migrationDefinition.ID)
		}
	}

	return childMap
}

func intsToStrings(ints []int) []string {
	strs := make([]string, 0, len(ints))
	for _, value := range ints {
		strs = append(strs, strconv.Itoa(value))
	}

	return strs
}

// ParseRawVersion returns the migration version for a given 'raw version', i.e. the
// filename of a mgiration.
//
// For example, for migration '1648115472_do_the_thing', we discard everything after
// the first '_' as a name, and return the verison 1648115472.
func ParseRawVersion(rawVersion string) (int, error) {
	nameParts := strings.SplitN(rawVersion, "_", 2)
	return strconv.Atoi(nameParts[0])
}
