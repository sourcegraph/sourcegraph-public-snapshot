package definition

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/keegancsmith/sqlf"

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
		upQuery, err := readQueryFromFile(fs, definition.UpFilename)
		if err != nil {
			return err
		}

		downQuery, err := readQueryFromFile(fs, definition.DownFilename)
		if err != nil {
			return err
		}

		definitions[i] = Definition{
			ID:           definition.ID,
			UpFilename:   definition.UpFilename,
			UpQuery:      upQuery,
			DownFilename: definition.DownFilename,
			DownQuery:    downQuery,
		}
	}

	return nil
}

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

	normalizedQuery := string(contents)
	normalizedQuery = strings.ReplaceAll(normalizedQuery, "%", "%%")

	return sqlf.Sprintf(normalizedQuery), nil
}
