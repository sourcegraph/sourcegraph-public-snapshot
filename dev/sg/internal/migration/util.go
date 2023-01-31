package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
)

// readDefinitions returns definitions from the given database object.
func readDefinitions(database db.Database) (*definition.Definitions, error) {
	fs, err := database.FS()
	if err != nil {
		return nil, err
	}

	return definition.ReadDefinitions(fs, filepath.Join("migrations", database.Name))
}

type MigrationFiles struct {
	UpFile       string
	DownFile     string
	MetadataFile string
}

// makeMigrationFilenames makes a pair of (absolute) paths to migration files with the given migration index.
func makeMigrationFilenames(database db.Database, migrationIndex int, name string) (MigrationFiles, error) {
	baseDir, err := migrationDirectoryForDatabase(database)
	if err != nil {
		return MigrationFiles{}, err
	}

	return makeMigrationFilenamesFromDir(baseDir, migrationIndex, name)
}

var nonAlphaNumericOrUnderscore = regexp.MustCompile("[^a-z0-9_]+")

func makeMigrationFilenamesFromDir(baseDir string, migrationIndex int, name string) (MigrationFiles, error) {
	sanitizedName := nonAlphaNumericOrUnderscore.ReplaceAllString(
		strings.ReplaceAll(strings.ToLower(name), " ", "_"), "",
	)
	var dirName string
	if sanitizedName == "" {
		// No name associated with this migration, we just use the index
		dirName = fmt.Sprintf("%d", migrationIndex)
	} else {
		// Include both index and simplified name
		dirName = fmt.Sprintf("%d_%s", migrationIndex, sanitizedName)
	}
	return MigrationFiles{
		UpFile:       filepath.Join(baseDir, fmt.Sprintf("%s/up.sql", dirName)),
		DownFile:     filepath.Join(baseDir, fmt.Sprintf("%s/down.sql", dirName)),
		MetadataFile: filepath.Join(baseDir, fmt.Sprintf("%s/metadata.yaml", dirName)),
	}, nil
}

// migrationDirectoryForDatabase returns the directory where migration files are stored for the
// given database.
func migrationDirectoryForDatabase(database db.Database) (string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(repoRoot, "migrations", database.Name), nil
}

// writeMigrationFiles writes the contents of migrationFileTemplate to the given filepaths.
func writeMigrationFiles(contents map[string]string) (err error) {
	defer func() {
		if err != nil {
			for path := range contents {
				// undo any changes to the fs on error
				_ = os.Remove(path)
			}
		}
	}()

	for path, contents := range contents {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(path, []byte(contents), os.FileMode(0644)); err != nil {
			return err
		}
	}

	return nil
}

// parseVersions takes a list of filepaths (the output of some git command) and a base
// migrations directory and returns the versions of migrations present in the list.
func parseVersions(lines []string, migrationsDir string) []int {
	var (
		pathSeparator       = string(os.PathSeparator)
		prefixesToTrim      = []string{migrationsDir, pathSeparator}
		separatorsToSplitBy = []string{pathSeparator, "_"}
	)

	versionMap := make(map[int]struct{}, len(lines))
	for _, rawVersion := range lines {
		// Remove leading migration directory if it exists
		for _, prefix := range prefixesToTrim {
			rawVersion = strings.TrimPrefix(rawVersion, prefix)
		}

		// Remove trailing filepath (if dir) or name prefix (if old migration)
		for _, separator := range separatorsToSplitBy {
			rawVersion = strings.Split(rawVersion, separator)[0]
		}

		// Should be left with only a version number
		if version, err := definition.ParseRawVersion(rawVersion); err == nil {
			versionMap[version] = struct{}{}
		}
	}

	versions := make([]int, 0, len(versionMap))
	for version := range versionMap {
		versions = append(versions, version)
	}
	sort.Ints(versions)

	return versions
}

// rootRelative removes the repo root prefix from the given path.
func rootRelative(path string) string {
	if repoRoot, _ := root.RepositoryRoot(); repoRoot != "" {
		sep := string(os.PathSeparator)
		rootWithTrailingSep := strings.TrimRight(repoRoot, sep) + sep
		return strings.TrimPrefix(path, rootWithTrailingSep)
	}

	return path
}
