package stitch

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var gitTreePattern = lazyregexp.New("^tree .+:.+\n")

// readMigrationDirectoryFilenames reads the names of the direct children of the given migration directory
// at the given git revision.
func readMigrationDirectoryFilenames(schemaName, dir, rev string) ([]string, error) {
	pathForSchemaAtRev, err := migrationPath(schemaName, rev)
	if err != nil {
		return nil, err
	}

	// First we will try to look up using the version tag. This should succeed for
	// historical releases that are already tagged. If we don't find the tag we will
	// fallback below to a branch name matching the release branch.
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", rev, pathForSchemaAtRev))
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Here we will try the release branch fallback. This should be encountered for future versions, in other words
		// we are updating the max supported version to something that isn't yet tagged.
		if branch, ok := tagRevToBranch(rev); ok && strings.Contains(string(out), "fatal: invalid object name") {
			cmd := exec.Command("git", "show", fmt.Sprintf("origin/%s:%s", branch, pathForSchemaAtRev))
			cmd.Dir = dir
			out, err = cmd.CombinedOutput()
		}
		if err != nil {
			return nil, errors.Wrapf(err, "failed to run git show: %s", out)
		}
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

	pathForSchemaAtRev, err := migrationPath(schemaName, rev)
	if err != nil {
		return "", err
	}
	if v, ok := m[filepath.Join(pathForSchemaAtRev, path)]; ok {
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
	cmd := exec.Command("git", "archive", "--format=tar", rev, "migrations")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		if branch, ok := tagRevToBranch(rev); ok && strings.Contains(string(out), "fatal: not a valid object name") {
			cmd := exec.Command("git", "archive", "--format=tar", "origin/"+branch, "migrations")
			cmd.Dir = dir
			out, err = cmd.CombinedOutput()
		}
		if err != nil {
			return nil, errors.Wrapf(err, "failed to run git archive: %s", out)
		}
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

func migrationPath(schemaName, rev string) (string, error) {
	revVersion, ok := oobmigration.NewVersionFromString(rev)
	if !ok {
		return "", errors.Newf("illegal rev %q", rev)
	}
	if oobmigration.CompareVersions(revVersion, oobmigration.NewVersion(3, 21)) == oobmigration.VersionOrderBefore {
		if schemaName == "frontend" {
			// Return the root directory if we're looking for the frontend schema
			// at or before 3.20. This was the only schema in existence then.
			return "migrations", nil
		}
	}

	return filepath.Join("migrations", schemaName), nil
}

// tagRevToBranch attempts to determine the branch on which the given rev, assumed to be a tag of the
// form vX.Y.Z, belongs. This is used to support generation of stitched migrations after a branch cut
// but before the tagged commit is created.
func tagRevToBranch(rev string) (string, bool) {
	version, ok := oobmigration.NewVersionFromString(rev)
	if !ok {
		return "", false
	}

	return fmt.Sprintf("%d.%d", version.Major, version.Minor), true
}
