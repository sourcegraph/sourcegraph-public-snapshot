package stitch

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/migrations"
)

// migrationFiles are a map whose keys are filepaths and value are
// the content of the file at that path. Just for convenience, as
// writing map[string]map[string]string is horrendous.
type migrationFiles map[string]string

// MigrationsReader provides an interface to reading migrations at a given version.
type MigrationsReader interface {
	Get(version string) (map[string]string, error)
}

// LazyMigrationsReader downloads the migrations archive for a given version, ideal for
// sg migration commands, as it doesn't require any GCP authentication, as the migrations are
// stored in a public GCS bucket.
type LazyMigrationsReader struct {
	baseUrl string
}

func NewLazyMigrationsReader() *LazyMigrationsReader {
	return &LazyMigrationsReader{
		baseUrl: "https://storage.googleapis.com/schemas-migrations/migrations",
	}
}

func (l *LazyMigrationsReader) urlForVersion(version string) string {
	return fmt.Sprintf("%s/migrations-%s.tar.gz", l.baseUrl, version)
}

func (l *LazyMigrationsReader) Get(version string) (map[string]string, error) {
	url := l.urlForVersion(version)
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"LazyMigrationReader, failed to load migration archive for version %q from %q",
			version,
			url,
		)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("unexpected status code (%d) for version %q from %q", resp.StatusCode, version, url)
	}
	contents, err := readFromTarball(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "faild to read migrations archive for version %q from tarball", version)
	}
	return contents, nil
}

// LocalMigrationsReader holds all migrations for each major or minor releases made so far.
// Content is populated through a folder containing tarballs for each release migrations and
// stored on GCP CloudStorare.
type LocalMigrationsReader struct {
	// currentVersion points at the version we should map to the migrations files at the root of the
	// repo, for the current revision.
	currentVersion string
	m              map[string]migrationFiles
}

// NewLocalMigrationsReader initialize the archives with the migration and the current version.
// It assumes that a given path, we'll find all the tarballs for each versions we require.
func NewLocalMigrationsReader(path string, currentVersion string) (*LocalMigrationsReader, error) {
	a := LocalMigrationsReader{
		m:              make(map[string]migrationFiles),
		currentVersion: "v" + currentVersion,
	}
	if err := a.load(path); err != nil {
		return nil, errors.Wrap(err, "failed to read local migrations")
	}
	return &a, nil
}

// readFromTarball returns a map of filenames to content for the given tarball.
func readFromTarball(r io.Reader) (migrationFiles, error) {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read from tarball")
	}
	contents := map[string]string{}
	tr := tar.NewReader(gzipReader)
	for {
		header, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, errors.Wrap(err, "failed to read from tarball")
		}

		fileContents, err := io.ReadAll(tr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read from tarball")
		}
		// We don't want to deal with directories in the map.
		if header.Typeflag == tar.TypeDir {
			continue
		}
		contents[header.Name] = string(fileContents)
	}
	return contents, nil
}

// filenameRegexp matches the version of a given tarball of a given release migrations.
var filenameRegexp = regexp.MustCompile(`migrations-(v\d+\.\d+\.\d+)\.tar\.gz`)

// load populates the migrationsArchiveStore from the files on disk.
func (s *LocalMigrationsReader) load(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// Load all versions from the archives.
	for _, e := range entries {
		f, err := os.Open(filepath.Join(path, e.Name()))
		if err != nil {
			return errors.Wrapf(err, "failed to open file %q", e.Name())
		}
		defer f.Close()

		contents, err := readFromTarball(f)
		if err != nil {
			return errors.Wrapf(err, "failed to read tarball %q", e.Name())
		}

		matches := filenameRegexp.FindStringSubmatch(e.Name())
		if len(matches) < 2 {
			return errors.Newf("invalid filename format: %s, can't extract version number", e.Name())
		}

		s.m[matches[1]] = contents
	}

	// If we already have a tarball for the current version, it's most likely because someone didn't bump
	// the constants yet, so let's use this one instead.
	if _, ok := s.m[s.currentVersion]; !ok {
		// Load the current version from the files on disk at that revision.
		contents := map[string]string{}
		err = fs.WalkDir(migrations.QueryDefinitions, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return errors.Wrapf(err, "failed to walk directory for current migrations (%q)", path)
			}
			if !d.IsDir() {
				b, err := fs.ReadFile(migrations.QueryDefinitions, path)
				if err != nil {
					return errors.Wrapf(err, "failed to read file for current migrations (%q)", path)
				}
				// When we read them from the embedded FS, they don't have the "migrations" prefix that we
				// have everywhere else.
				contents[filepath.Join("migrations", path)] = string(b)
			}
			return nil
		})
		if err != nil {
			return errors.Wrap(err, "failed to load current migrations from current migrations")
		}
		s.m[s.currentVersion] = contents
	} else {
		fmt.Printf("WARNING: a tarball for %s already exists, constant is out of date\n", s.currentVersion)
	}

	return nil
}

// Get returns a map of filepaths to their contents or an error if the version is not found in the archives.
func (s *LocalMigrationsReader) Get(version string) (map[string]string, error) {
	migrations, ok := s.m[version]
	if !ok {
		migrations, ok = s.m["v"+version]
		if !ok {
			return nil, errors.Newf("version key %s not found in migrations archive", version)
		}
	}
	return migrations, nil
}
