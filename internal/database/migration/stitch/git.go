package stitch

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var gitTreePattern = lazyregexp.New("^tree .+:.+\n")

// readMigrationDirectoryFilenames reads the names of the direct children of the given migration directory
// at the given git revision.
func readMigrationDirectoryFilenames(schemaName, dir, rev string) ([]string, error) {
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", doMagicHacking(rev), migrationPath(schemaName)))
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	if ok := gitTreePattern.Match(out); !ok {
		return nil, errors.New("not a directory")
	}

	var lines []string
	for _, line := range bytes.Split(out, []byte("\n"))[1:] {
		if len(line) == 0 {
			continue
		}

		lines = append(lines, string(line))
	}

	return lines, nil
}

// readMigrationFileContents reads the contents of the migration at given path at the given git revision.
func readMigrationFileContents(schemaName, dir, rev, path string) (string, error) {
	m, err := cachedArchiveContents(dir, rev)
	if err != nil {
		return "", err
	}

	if v, ok := m[filepath.Join(migrationPath(schemaName), path)]; ok {
		return v, nil
	}

	return "", os.ErrNotExist
}

var (
	revToPathTocontentsCacheMutex sync.RWMutex
	revToPathTocontentsCache      = map[string]map[string]string{}
)

// cachedArchiveContents memoizes archiveContents by git revision and schema name.
func cachedArchiveContents(dir, rev string) (map[string]string, error) {
	revToPathTocontentsCacheMutex.Lock()
	defer revToPathTocontentsCacheMutex.Unlock()

	m, ok := revToPathTocontentsCache[rev]
	if ok {
		return m, nil
	}

	m, err := archiveContents(dir, rev)
	if err != nil {
		return nil, err
	}

	revToPathTocontentsCache[rev] = m
	return m, nil
}

// archiveContents calls git archive with the given git revision and returns a map from
// file paths to file contents.
func archiveContents(dir, rev string) (map[string]string, error) {
	cmd := exec.Command("git", "archive", "--format=tar", doMagicHacking(rev), "migrations")
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to run git archive: `%s`", out)
	}

	revContents := map[string]string{}

	r := tar.NewReader(bytes.NewReader(out))
	for {
		header, err := r.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		fileContents, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		revContents[header.Name] = string(fileContents)
	}

	return revContents, nil
}

func migrationPath(schemaName string) string {
	return filepath.Join("migrations", schemaName)
}

func doMagicHacking(rev string) string {
	if rev == "v4.0.0" {
		// NOTE: prior to the 4.0 branch cut, this can be updated to be the head of
		// the main branch to assist in testing upgrades end-to-end. Prior to the 4.0
		// branch cut, we must remove this condition.
		rev = "695b52c5151010f275eb7232eb2d6cea27a07026"
	}

	return rev
}
