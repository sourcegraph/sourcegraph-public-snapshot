package archive

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	VERSION_FILE = ".sourcegraph-version"
)

// Archiver types implement the methods to archive and restore storage backends
type Archiver interface {
	archive(root string) error
	restore(root string) error
	root() string
}

// archiveSources - storage backends which should be included in the archive
// TODO(jac): allow user to select which sources to archive?
func archiveSources() []Archiver {
	sources := make([]Archiver, 0, 3)
	sources = append(sources, dbSources()...)
	return sources
}

// archiveIdentify - identify storage backends included in the archive
// TODO(jac): allow user to choose which identified sources to restore?
func archiveIdentify(path string) []Archiver {
	sources := make([]Archiver, 0, 3)
	sources = append(sources, dbIdentify(path)...)
	return sources
}

func CreateArchive(c context.Context, logger log.Logger) (io.Reader, error) {
	dir, err := os.MkdirTemp(os.TempDir(), "archive-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	if err = writeSiteVersion(dir); err != nil {
		return nil, err
	}

	sources := archiveSources()

	for _, source := range sources {
		root, err := createRoot(dir, source.root())
		if err != nil {
			return nil, err
		}
		err = source.archive(root)
		if err != nil {
			return nil, err
		}
	}

	files, err := archiveFiles(dir)
	if err != nil {
		return nil, err
	}

	zip, err := createZip(dir, files)
	if err != nil {
		return nil, err
	}

	return zip, err
}

func writeSiteVersion(path string) error {
	f, err := os.Create(filepath.Join(path, VERSION_FILE))
	if err != nil {
		return err
	}
	defer f.Close()
	f.Write([]byte(version.Version()))
	return nil
}

func compareSiteVersion(path string) error {
	f, err := os.Open(filepath.Join(path, VERSION_FILE))
	if err != nil {
		return err
	}
	defer f.Close()

	c, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	if string(c) != version.Version() {
		return errors.New("archive version does not match running instance")
	}

	return nil
}

// create storage backend root folder inside archive root dir
func createRoot(archiveRoot string, root string) (string, error) {
	path := filepath.Join(archiveRoot, root)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return path, err
		}
	}

	return path, nil
}

func RestoreFromArchive(c context.Context, logger log.Logger, archive string) error {
	path, err := unZip(archive)
	if err != nil {
		return err
	}

	if err = compareSiteVersion(path); err != nil {
		return err
	}

	sources := archiveIdentify(path)
	for _, source := range sources {
		if err := source.restore(filepath.Join(path, source.root())); err != nil {
			return err
		}
	}

	return nil
}
