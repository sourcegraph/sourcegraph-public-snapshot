package definition

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

func ReadDefinitions(fs fs.FS) (*Definitions, error) {
	filenames, err := readSQLFilenames(fs)
	if err != nil {
		return nil, err
	}

	migrationDefinitions, err := buildDefinitionStencils(filenames)
	if err != nil {
		return nil, err
	}

	if err := hydrateDefinitions(fs, migrationDefinitions); err != nil {
		return nil, err
	}

	if err := reorderDefinitions(migrationDefinitions); err != nil {
		return nil, err
	}

	if err := validateLinearizedGraph(migrationDefinitions); err != nil {
		return nil, err
	}

	return &Definitions{
		definitions: migrationDefinitions,
	}, nil
}

func readSQLFilenames(fs fs.FS) ([]string, error) {
	root, err := http.FS(fs).Open("/")
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()

	files, err := root.Readdir(0)
	if err != nil {
		return nil, err
	}

	filenames := make([]string, 0, len(files))
	for _, file := range files {
		filenames = append(filenames, file.Name())
	}
	sort.Strings(filenames)

	return filenames, nil
}

var pattern = lazyregexp.New(`^(\d+)_[^.]+\.(up|down)\.sql$`)

func buildDefinitionStencils(filenames []string) ([]Definition, error) {
	definitionMap := make(map[int]Definition, len(filenames))

	// Iterate through the set of filenames looking for things that have the shape
	// of a migration query file. Group these by identifier and match the up and down
	// direction query definitions together.

	for _, filename := range filenames {
		match := pattern.FindStringSubmatch(filename)
		if len(match) == 0 {
			continue
		}

		id, _ := strconv.Atoi(match[1])
		definition := definitionMap[id]

		if match[2] == "up" {
			// Check for duplicates before overwriting
			if definition.UpFilename != "" {
				return nil, fmt.Errorf("duplicate upgrade query for migration definition %d: %s and %s", id, definition.UpFilename, filename)
			}

			definitionMap[id] = Definition{
				UpFilename:   filename,
				DownFilename: definition.DownFilename,
			}
		} else {
			// Check for duplicates before overwriting
			if definition.DownFilename != "" {
				return nil, fmt.Errorf("duplicate downgrade query for migration definition %d: %s and %s", id, definition.DownFilename, filename)
			}

			definitionMap[id] = Definition{
				UpFilename:   definitionMap[id].UpFilename,
				DownFilename: filename,
			}
		}
	}

	// Check for migrations with only direction defined
	// Assign identifiers directly to migration definition values

	for id, definition := range definitionMap {
		if definition.UpFilename == "" {
			return nil, fmt.Errorf("upgrade query for migration definition %d not found", id)
		}
		if definition.DownFilename == "" {
			return nil, fmt.Errorf("downgrade query for migration definition %d not found", id)
		}

		definitionMap[id] = Definition{
			ID:           id,
			UpFilename:   definition.UpFilename,
			DownFilename: definition.DownFilename,
		}
	}

	// Flatten the definition map into ordered list
	definitions := make([]Definition, 0, len(definitionMap))
	for _, definition := range definitionMap {
		definitions = append(definitions, definition)
	}
	sort.Slice(definitions, func(i, j int) bool {
		return definitions[i].ID < definitions[j].ID
	})

	// Check for gaps in ids
	for i, definition := range definitions {
		if i > 0 && definition.ID != definitions[i-1].ID+1 {
			return nil, fmt.Errorf("migration identifiers jump from %d to %d", definitions[i-1].ID, definition.ID)
		}
	}

	return definitions, nil
}

func hydrateDefinitions(fs fs.FS, definitions []Definition) (err error) {
	for i, definition := range definitions {
		upQuery, metadata, err := readQueryFromFile(fs, definition.UpFilename)
		if err != nil {
			return err
		}

		downQuery, _, err := readQueryFromFile(fs, definition.DownFilename)
		if err != nil {
			return err
		}

		definitions[i] = Definition{
			ID:           definition.ID,
			Metadata:     metadata,
			UpFilename:   definition.UpFilename,
			UpQuery:      upQuery,
			DownFilename: definition.DownFilename,
			DownQuery:    downQuery,
		}
	}

	return nil
}

// readQueryFromFile returns the parsed query and extracted metadata read from
// the given file.
func readQueryFromFile(fs fs.FS, filepath string) (_ *sqlf.Query, metadata Metadata, _ error) {
	file, err := fs.Open(filepath)
	if err != nil {
		return nil, metadata, err
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, metadata, err
	}

	query, metadata, err := extractMetadata(string(contents))
	if err != nil {
		return nil, metadata, errors.Wrap(err, "failed to extract metadata")
	}

	// Stringify -> SQL-ify the contents of the file. We first replace any
	// SQL placeholder values with an escaped version so that the sqlf.Sprintf
	// call does not try to interpolate the text with variables we don't have.
	return sqlf.Sprintf(strings.ReplaceAll(query, "%", "%%")), metadata, nil
}

var (
	metadataFence   = `-- +++`
	metadataPattern = lazyregexp.New(`[\s\S]*` + regexp.QuoteMeta(metadataFence) + `([.\s\S]+)` + regexp.QuoteMeta(metadataFence))
)

// extractMetadata splits the given migration file contents into query and optional metadata.
// Metadata can be supplied alongside a query by attaching a SQL comment at the top of the file
// with the following shape:
//
// -- +++
// -- key: value
// -- allowed: here
// -- because: this
// -- is: interpreted
// -- as: yaml
// -- +++
// -- anything remaining is returned as part of the SQL query.
//
// Note that thee SQL query itself can have additional comments; metadata need only be at the top
// of the file for extraction.
func extractMetadata(contents string) (_ string, metadata Metadata, _ error) {
	match := metadataPattern.FindStringSubmatch(contents)

	if len(match) > 0 {
		if err := yaml.Unmarshal([]byte(extractCommentPrefix(match[1])), &metadata); err != nil {
			return "", metadata, err
		}

		return strings.TrimSpace(contents[len(match[0]):]), metadata, nil
	}

	return contents, metadata, nil
}

// extractCommentPrefix removes the `-- ` on each line of the given multiline string.
func extractCommentPrefix(match string) string {
	lines := strings.Split(match, "\n")
	for i := range lines {
		lines[i] = strings.TrimPrefix(lines[i], "-- ")
	}

	return strings.Join(lines, "\n")
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

var (
	ErrNoRoots       = fmt.Errorf("no roots")
	ErrMultipleRoots = fmt.Errorf("multiple roots")
	ErrCycle         = fmt.Errorf("cycle")
)

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
		children = children(migrationDefinitions)

		dfs func(id int) error
	)

	dfs = func(id int) error {
		if marks[id] == MarkTypeVisiting {
			// currently processing
			return ErrCycle
		}
		if marks[id] == MarkTypeVisited {
			// already visited
			return nil
		}

		marks[id] = MarkTypeVisiting
		defer func() { marks[id] = MarkTypeVisited }()

		for _, child := range children[id] {
			if err := dfs(child); err != nil {
				return err
			}
		}

		// Add self _after_ adding all children recursively
		order = append(order, id)
		return nil
	}

	if err := dfs(root); err != nil {
		return nil, err
	}
	if len(order) != len(migrationDefinitions) {
		return nil, ErrCycle
	}

	return order, nil
}

// root returns the unique migration definition with no parent. An error is returned
// if there is not exactly one root.
func root(migrationDefinitions []Definition) (int, error) {
	roots := make([]int, 0, 1)
	for _, migrationDefinition := range migrationDefinitions {
		if migrationDefinition.Metadata.Parent == 0 {
			roots = append(roots, migrationDefinition.ID)
		}
	}
	if len(roots) == 0 {
		return 0, ErrNoRoots
	}
	if len(roots) > 1 {
		return 0, ErrMultipleRoots
	}

	return roots[0], nil
}

// children constructs map from migration identifiers to the set of identifiers of all
// dependent migrations.
func children(migrationDefinitions []Definition) map[int][]int {
	children := make(map[int][]int, len(migrationDefinitions))
	for _, migrationDefinition := range migrationDefinitions {
		if parent := migrationDefinition.Metadata.Parent; parent != 0 {
			children[parent] = append(children[parent], migrationDefinition.ID)
		}
	}

	return children
}

// validateLinearizedGraph returns an error if the given sequence of migrations are
// not in linear order. This requires that each migration definition's parent is marked
// as the one that proceeds it in file order.
//
// This check is here to maintain backwards compatibility with the sequential migration
// numbers required by golang migrate. This will be lifted once we build support for non
// sequential migrations in the background.
func validateLinearizedGraph(migrationDefinitions []Definition) error {
	if len(migrationDefinitions) == 0 {
		return nil
	}

	if migrationDefinitions[0].Metadata.Parent != 0 {
		return fmt.Errorf("unexpected parent for root definition")
	}

	for _, definition := range migrationDefinitions[1:] {
		if definition.Metadata.Parent != definition.ID-1 {
			return fmt.Errorf("unexpected parent declared in definition %d", definition.ID)
		}
	}

	return nil
}
