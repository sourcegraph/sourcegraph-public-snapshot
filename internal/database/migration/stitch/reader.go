package stitch

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type rawMigration struct {
	id       string
	up       string
	down     string
	metadata string
}

// ignoreMap are valid filenames that can exist within a migration directory.
var ignoreMap = map[string]struct{}{
	"bindata.go":         {},
	"gen.go":             {},
	"migrations_test.go": {},
	"README.md":          {},
	"squashed.sql":       {},
	"BUILD.bazel":        {},
}

// readRawMigrations reads migrations from a locally available git revision for the given schema.
// This function understands the common ways we historically laid out our migration definitions
// in-tree, and will return results going back to v3.29.0 (with empty metadata where missing).
func readRawMigrations(schemaName, dir, rev string) (migrations []rawMigration, _ error) {
	entries, err := readMigrationDirectoryFilenames(schemaName, dir, rev)
	if err != nil {
		return nil, err
	}

	for _, filename := range entries {
		// Attempt to parse file as a flat migration entry
		if id, suffix, direction, ok := matchFlatPattern(filename); ok {
			if direction != "up" {
				// Reduce duplicates by choosing only .up.sql files
				continue
			}

			migration, err := readFlat(schemaName, dir, rev, id, suffix)
			if err != nil {
				return nil, err
			}

			migrations = append(migrations, migration)
			continue
		}

		// Attempt to parse file as a hierarchical migration entry
		if id, suffix, ok := matchHierarchicalPattern(filename); ok {
			migration, err := readHierarchical(schemaName, dir, rev, id, suffix)
			if err != nil {
				return nil, err
			}

			migrations = append(migrations, migration)
			continue
		}

		if _, ok := ignoreMap[filename]; !ok {
			// Throw an error if there's new file types we don't know to ignore
			return nil, errors.Newf("unrecognized entry %q", filename)
		}
	}

	return migrations, nil
}

//
// Flat migration file parsing

var flatPattern = lazyregexp.New(`(\d+)_(.+)\.(up|down)\.sql`)

// matchFlatPattern returns the text captured from the given string.
func matchFlatPattern(s string) (id, suffix, direction string, ok bool) {
	if matches := flatPattern.FindStringSubmatch(s); len(matches) > 0 {
		return matches[1], matches[2], matches[3], true
	}

	return "", "", "", false
}

// readFlat creates a raw migration from a pair of up/down SQL files in-tree.
func readFlat(schemaName, dir, rev, id, suffix string) (rawMigration, error) {
	up, err := readMigrationFileContents(schemaName, dir, rev, fmt.Sprintf("%s_%s.up.sql", id, suffix))
	if err != nil {
		return rawMigration{}, err
	}
	down, err := readMigrationFileContents(schemaName, dir, rev, fmt.Sprintf("%s_%s.down.sql", id, suffix))
	if err != nil {
		return rawMigration{}, err
	}

	return rawMigration{id, up, down, fmt.Sprintf("name: %s", strings.ReplaceAll(suffix, "_", " "))}, nil
}

//
// Hierarchical migration file parsing

var hierarchicalPattern = lazyregexp.New(`(\d+)(_.+)?/`)

// matchHierarchicalPattern returns the text captured from the given string.
func matchHierarchicalPattern(s string) (id, suffix string, ok bool) {
	if matches := hierarchicalPattern.FindStringSubmatch(s); len(matches) >= 3 {
		return matches[1], matches[2], true
	}

	return "", "", false
}

// readHierarchical creates a raw migration from a pair of up/down SQL files and a metadata
// file all located within a subdirectory in-tree.
func readHierarchical(schemaName, dir, rev, id, suffix string) (rawMigration, error) {
	up, err := readMigrationFileContents(schemaName, dir, rev, filepath.Join(id+suffix, "up.sql"))
	if err != nil {
		return rawMigration{}, err
	}
	down, err := readMigrationFileContents(schemaName, dir, rev, filepath.Join(id+suffix, "down.sql"))
	if err != nil {
		return rawMigration{}, err
	}
	metadata, err := readMigrationFileContents(schemaName, dir, rev, filepath.Join(id+suffix, "metadata.yaml"))
	if err != nil {
		return rawMigration{}, err
	}

	return rawMigration{id, up, down, metadata}, nil
}
