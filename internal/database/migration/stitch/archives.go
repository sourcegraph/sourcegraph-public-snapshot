package stitch

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/migrations"
)

// migrationEntries are a map whose keys are filepaths and value are
// the content of the file at that path.
type migrationEntries map[string]string

// migrationArchives holds all migrations for each major or minor releases made so far.
// Content is populated through a folder containing tarballs for each release migrations and
// stored on GCP CloudStorare.
type migrationArchives struct {
	// currentVersion points at the version we should map to the migrations files at the root of the
	// repo, for the current revision.
	currentVersion string
	m              map[string]migrationEntries
}

func NewMigrationArchives(path string, currentVersion string) (*migrationArchives, error) {
	a := migrationArchives{
		currentVersion: currentVersion,
	}
	if err := a.load(path); err != nil {
		return nil, err
	}
	return &a, nil
}

// filenameRegexp matches the version of a given tarball of a given release migrations.
var filenameRegexp = regexp.MustCompile(`migrations-(v\d+\.\d+\.\d+)\.tar\.gz`)

// load populates the migrationsArchiveStore from the files on disk.
func (s *migrationArchives) load(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// Load all versions from the archives.
	for _, e := range entries {
		f, err := os.Open(filepath.Join(path, e.Name()))
		if err != nil {
			return err
		}

		gzipReader, err := gzip.NewReader(f)
		if err != nil {
			return err
		}

		contents := map[string]string{}
		r := tar.NewReader(gzipReader)
		for {
			header, err := r.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}

			fileContents, err := io.ReadAll(r)
			if err != nil {
				return err
			}
			// We don't want to deal with directories in the map.
			if header.Typeflag == tar.TypeDir {
				continue
			}
			contents[header.Name] = string(fileContents)
		}

		matches := filenameRegexp.FindStringSubmatch(e.Name())
		if len(matches) < 2 {
			return fmt.Errorf("invalid filename format: %s, can't extract version number", e.Name())
		}

		s.m[matches[1]] = contents
	}

	// Load the current version from the files on disk at that revision.
	contents := map[string]string{}
	err = fs.WalkDir(migrations.QueryDefinitions, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			b, err := fs.ReadFile(migrations.QueryDefinitions, path)
			if err != nil {
				return err
			}
			// When we read them from the embedded FS, they don't have the "migrations" prefix that we
			// have everywhere else.
			contents[filepath.Join("migrations", path)] = string(b)
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.m[s.currentVersion] = contents
	return nil
}

// Get returns a map of filepaths to their contents or an error if the version is not found in the archives.
func (s *migrationArchives) Get(version string) (map[string]string, error) {
	migrations, ok := s.m[version]
	if !ok {
		migrations, ok = s.m["v"+version]
		if !ok {
			return nil, fmt.Errorf("version %s not found", version)
		}
	}
	return migrations, nil
}
