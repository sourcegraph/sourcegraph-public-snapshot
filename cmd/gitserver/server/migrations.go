package server

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"
)

// migrate runs any missing migrations for gitserver rootDir. This should be
// done before the server starts accepting requests or does background jobs.
//
// Migrations:
// 1. Old directory structure to .git directory structure.
func migrate(rootDir string) error {
	mf, err := loadManifest(rootDir)
	if err != nil {
		return err
	}

	// No migration needed
	if mf.Version == 1 {
		return nil
	}

	if err := migrateGitDir(rootDir); err != nil {
		return errors.Wrap(err, "failed to migrate to .git based directory structure")
	}
	mf.Version = 1

	return storeManifest(mf, rootDir)
}

func migrateGitDir(rootDir string) error {
	tmp, err := ioutil.TempDir(rootDir, "migrate-git-dir-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// We only care about directories
		if !info.IsDir() {
			return nil
		}

		// New style git directory layout
		if filepath.Base(path) == ".git" {
			return filepath.SkipDir
		}

		// For old-style directory layouts we need to rename them to the new
		// style .git layout. The presence of HEAD outside of a .git dir
		// indicates our old layout.
		if _, err := os.Stat(filepath.Join(path, "HEAD")); os.IsNotExist(err) {
			// HEAD doesn't exist, so keep recursing
			return nil
		} else if err != nil {
			return err
		}

		// path is an old style git repo since it contains HEAD. We need to do
		// two renames to end up in the new directory layout since it is a
		// child of the current path:
		//
		//   want /repos/example.com/foo/bar -> /repos/example.com/foo/bar/.git
		//
		//   do  mkdir /repos/$tmp/bar
		//       /repos/example.com/foo/bar -> /repos/$tmp/bar/.git
		//       /repos/$tmp/bar -> /repos/example.com/foo/bar
		//
		middle := filepath.Join(tmp, filepath.Base(path))
		log15.Info("migrating git clone location", "src", path, "dst", filepath.Join(path, ".git"))
		if err := os.Mkdir(middle, os.ModePerm); err != nil {
			return err
		}
		if err := os.Rename(path, filepath.Join(middle, ".git")); err != nil {
			return err
		}
		if err := os.Rename(middle, path); err != nil {
			// Failing here means we have renamed out the clone but not put it
			// in place. Returning an error means we have lost the clone. This
			// is fine since is should be rare and the clone is just a cache
			// of the clone from the code host.
			return err
		}

		return filepath.SkipDir
	})
}

// manifestName is the name of the manifest file present in the root of a gitserver DataDir
const manifestName = "gitserver.json"

// manifest holds manifest file data for gitserver.
type manifest struct {
	// Version is a number used to track the manifest and gitserver layout.
	//
	// 0. No manifest
	// 1. Added manifest and migrated directories
	Version int
}

// loadManifest loads the manifest for rootDir. If missing it returns an empty
// Manifest (version 0).
func loadManifest(rootDir string) (*manifest, error) {
	b, err := ioutil.ReadFile(filepath.Join(rootDir, manifestName))
	if err != nil {
		if os.IsNotExist(err) {
			return &manifest{}, nil
		}
		return nil, errors.Wrap(err, "failed to open gitserver manifest")
	}

	var m manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Wrap(err, "failed to parse gitserver manifest")
	}
	return &m, nil
}

// storeManifest marshalls to disk the manifest for rootDir.
func storeManifest(m *manifest, rootDir string) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal gitserver manifest")
	}
	if err := ioutil.WriteFile(filepath.Join(rootDir, manifestName), b, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to write gitserver manifest")
	}
	return nil
}
