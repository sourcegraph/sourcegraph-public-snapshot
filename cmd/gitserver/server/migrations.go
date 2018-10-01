package server

import (
	"os"
	"path/filepath"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

func (s *Server) migrateGitDir() {
	tmp, err := s.tempDir("migrate-git-dir-")
	if err != nil {
		log15.Warn("git clone location migration failed", "error", err)
		return
	}
	defer os.RemoveAll(tmp)

	err = filepath.Walk(s.ReposDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// The directory may not exist if it is a path we want to ignore
			// (as e.g. tmp dirs may be created in a race where filepath.Walk
			// sees it initially but it has been deleted). This is safe to
			// ignore because we would've ignored this directory a few lines
			// below here anyway.
			if !os.IsNotExist(err) && s.ignorePath(path) {
				log15.Warn("(1) ignoring path in git clone location migration", "path", path, "error", err)
			}
			return nil
		}

		if s.ignorePath(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// We only care about directories
		if !info.IsDir() {
			return nil
		}

		// Take this opportunity to best-effort ensure our permissions aren't
		// whack. We want to be able to rwx
		// directories. https://github.com/sourcegraph/sourcegraph/issues/12234
		if info.Mode()&0700 != 0700 {
			os.Chmod(path, (info.Mode()&os.ModePerm)|0700)
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
			log15.Warn("(2) ignoring path in git clone location migration", "path", path, "error", err)
			return filepath.SkipDir
		}

		lock, ok := s.locker.TryAcquire(path, "migrating repository clone")
		if !ok {
			log15.Warn("failed to acquire directory lock in git clone location migration", "path", path)
			return filepath.SkipDir
		}
		defer lock.Release()

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
			log15.Warn("(3) ignoring path in git clone location migration", "path", path, "error", err)
			return filepath.SkipDir
		}
		// If something goes wrong, ensure we clean up the temporary location.
		defer os.RemoveAll(middle)
		if err := os.Rename(path, filepath.Join(middle, ".git")); err != nil {
			log15.Warn("(4) ignoring path in git clone location migration", "path", path, "error", err)
			return filepath.SkipDir
		}
		if err := os.Rename(middle, path); err != nil {
			// Failing here means we have renamed out the clone but not put it
			// in place. So we have lost the clone. This is fine since is
			// should be rare and the clone is just a cache of the clone from
			// the code host.
			log15.Warn("(5) ignoring path in git clone location migration", "path", path, "error", err)
			return filepath.SkipDir
		}

		return filepath.SkipDir
	})

	if err != nil {
		log15.Warn("git clone location migration failed", "error", err)
	}
}
