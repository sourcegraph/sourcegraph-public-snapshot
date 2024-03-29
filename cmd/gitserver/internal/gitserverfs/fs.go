package gitserverfs

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TempDirName is the name used for the temporary directory under ReposDir.
const TempDirName = ".tmp"

// P4HomeName is the name used for the directory that git p4 will use as $HOME
// and where it will store cache data.
const P4HomeName = ".p4home"

func MakeP4HomeDir(reposDir string) (string, error) {
	p4Home := filepath.Join(reposDir, P4HomeName)
	// Ensure the directory exists
	if err := os.MkdirAll(p4Home, os.ModePerm); err != nil {
		return "", errors.Wrapf(err, "ensuring p4Home exists: %q", p4Home)
	}
	return p4Home, nil
}

func RepoDirFromName(reposDir string, name api.RepoName) common.GitDir {
	p := string(protocol.NormalizeRepo(name))
	return common.GitDir(filepath.Join(reposDir, filepath.FromSlash(p), ".git"))
}

func RepoNameFromDir(reposDir string, dir common.GitDir) api.RepoName {
	// dir == ${s.ReposDir}/${name}/.git
	parent := filepath.Dir(string(dir))                   // remove suffix "/.git"
	name := strings.TrimPrefix(parent, reposDir)          // remove prefix "${s.ReposDir}"
	name = strings.Trim(name, string(filepath.Separator)) // remove /
	name = filepath.ToSlash(name)                         // filepath -> path
	return protocol.NormalizeRepo(api.RepoName(name))
}

// TempDir is a wrapper around os.MkdirTemp, but using the given reposDir
// temporary directory filepath.Join(s.ReposDir, tempDirName).
//
// This directory is cleaned up by gitserver and will be ignored by repository
// listing operations.
func TempDir(reposDir, prefix string) (name string, err error) {
	// TODO: At runtime, this directory always exists. We only need to ensure
	// the directory exists here because tests use this function without creating
	// the directory first. Ideally, we can remove this later.
	tmp := filepath.Join(reposDir, TempDirName)
	if err := os.MkdirAll(tmp, os.ModePerm); err != nil {
		return "", err
	}
	return os.MkdirTemp(tmp, prefix)
}

func IgnorePath(reposDir string, path string) bool {
	// We ignore any path which starts with .tmp or .p4home in ReposDir
	if filepath.Dir(path) != reposDir {
		return false
	}
	base := filepath.Base(path)
	return strings.HasPrefix(base, TempDirName) || strings.HasPrefix(base, P4HomeName)
}

// RemoveRepoDirectory atomically removes a directory from reposDir.
//
// It first moves the directory to a temporary location to avoid leaving
// partial state in the event of server restart or concurrent modifications to
// the directory.
//
// Additionally, it removes parent empty directories up until reposDir.
func RemoveRepoDirectory(ctx context.Context, logger log.Logger, db database.DB, shardID string, reposDir string, gitDir common.GitDir, updateCloneStatus bool) error {
	dir := string(gitDir)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// If directory doesn't exist we can avoid all the work below and treat it as if
		// it was removed.
		return nil
	}

	// Rename out of the location, so we can atomically stop using the repo.
	tmp, err := TempDir(reposDir, "delete-repo")
	if err != nil {
		return err
	}
	defer func() {
		// Delete the atomically renamed dir.
		if err := os.RemoveAll(filepath.Join(tmp)); err != nil {
			logger.Warn("failed to cleanup after removing dir", log.String("dir", dir), log.Error(err))
		}
	}()
	if err := fileutil.RenameAndSync(dir, filepath.Join(tmp, "repo")); err != nil {
		return err
	}

	// Everything after this point is just cleanup, so any error that occurs
	// should not be returned, just logged.

	if updateCloneStatus {
		// Set as not_cloned in the database.
		if err := db.GitserverRepos().SetCloneStatus(ctx, RepoNameFromDir(reposDir, gitDir), types.CloneStatusNotCloned, shardID); err != nil {
			logger.Warn("failed to update clone status", log.Error(err))
		}
	}

	// Cleanup empty parent directories. We just attempt to remove and if we
	// have a failure we assume it's due to the directory having other
	// children. If we checked first we could race with someone else adding a
	// new clone.
	rootInfo, err := os.Stat(reposDir)
	if err != nil {
		logger.Warn("Failed to stat ReposDir", log.Error(err))
		return nil
	}
	current := dir
	for {
		parent := filepath.Dir(current)
		if parent == current {
			// This shouldn't happen, but protecting against escaping
			// ReposDir.
			break
		}
		current = parent
		info, err := os.Stat(current)
		if os.IsNotExist(err) {
			// Someone else beat us to it.
			break
		}
		if err != nil {
			logger.Warn("failed to stat parent directory", log.String("dir", current), log.Error(err))
			return nil
		}
		if os.SameFile(rootInfo, info) {
			// Stop, we are at the parent.
			break
		}

		if err := os.Remove(current); err != nil {
			// Stop, we assume remove failed due to current not being empty.
			break
		}
	}

	return nil
}

// BestEffortWalk is a filepath.WalkDir which ignores errors that can be passed
// to walkFn. This is a common pattern used in gitserver for best effort work.
//
// Note: We still respect errors returned by walkFn.
//
// filepath.Walk can return errors if we run into permission errors or a file
// disappears between readdir and the stat of the file. In either case this
// error can be ignored for best effort code.
func BestEffortWalk(root string, walkFn func(path string, entry fs.DirEntry) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		return walkFn(path, d)
	})
}

// DirSize returns the total size in bytes of all the files under d.
func DirSize(d string) int64 {
	var size int64
	// We don't return an error, so we know that err is always nil and can be
	// ignored.
	_ = BestEffortWalk(d, func(path string, d fs.DirEntry) error {
		if d.IsDir() {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			// We ignore errors for individual files.
			return nil
		}
		size += fi.Size()
		return nil
	})
	return size
}
