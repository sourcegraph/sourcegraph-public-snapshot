package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"

	"gopkg.in/inconshreveable/log15.v2"
)

func init() {
	prometheus.MustRegister(reposRemoved)
	prometheus.MustRegister(reposRecloned)
}

const (
	// repoTTL is how often we should reclone a repository
	repoTTL = time.Hour * 24 * 45
	// repoTTLGC is how often we should reclone a repository once it is
	// reporting git gc issues.
	repoTTLGC = time.Hour * 24
)

var reposRemoved = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "repos_removed",
	Help:      "number of repos removed during cleanup",
})

var reposRecloned = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "gitserver",
	Name:      "repos_recloned",
	Help:      "number of repos removed and recloned due to age",
})

// cleanupRepos walks the repos directory and performs maintenance tasks:
//
// 1. Remove corrupt repos.
// 2. Remove stale lock files.
// 3. Remove inactive repos on sourcegraph.com
// 4. Reclone repos after a while. (simulate git gc)
func (s *Server) cleanupRepos() {
	bCtx, bCancel := s.serverContext()
	defer bCancel()

	maybeRemoveCorrupt := func(dir GitDir) (done bool, err error) {
		// We treat repositories missing HEAD to be corrupt. Both our cloning
		// and fetching ensure there is a HEAD file.
		_, err = os.Stat(dir.Path("HEAD"))
		if !os.IsNotExist(err) {
			return false, err
		}

		log15.Info("removing corrupt repo", "repo", dir)
		if err := s.removeRepoDirectory(dir); err != nil {
			return true, err
		}
		reposRemoved.Inc()
		return true, nil
	}

	ensureGitAttributes := func(dir GitDir) (done bool, err error) {
		return false, setGitAttributes(dir)
	}

	maybeReclone := func(dir GitDir) (done bool, err error) {
		recloneTime, err := getRecloneTime(dir)
		if err != nil {
			return false, err
		}

		// Add a jitter to spread out recloning of repos cloned at the same
		// time.
		var reason string
		if time.Since(recloneTime) > repoTTL+randDuration(repoTTL/4) {
			reason = "old"
		}
		if time.Since(recloneTime) > repoTTLGC+randDuration(repoTTLGC/4) {
			if gclog, err := ioutil.ReadFile(dir.Path("gc.log")); err == nil && len(gclog) > 0 {
				reason = fmt.Sprintf("git gc %s", string(bytes.TrimSpace(gclog)))
			}
		}
		if reason == "" {
			return false, nil
		}

		ctx, cancel := context.WithTimeout(bCtx, longGitCommandTimeout)
		defer cancel()

		// name is the relative path to ReposDir, but without the .git suffix.
		repo := s.name(dir)
		log15.Info("recloning expired repo", "repo", repo, "cloned", recloneTime, "reason", reason)

		remoteURL, err := repoRemoteURL(ctx, dir)
		if err != nil {
			return false, errors.Wrap(err, "failed to get remote URL")
		}

		if _, err := s.cloneRepo(ctx, repo, remoteURL, &cloneOptions{Block: true, Overwrite: true}); err != nil {
			return true, err
		}
		reposRecloned.Inc()
		return true, nil
	}

	removeStaleLocks := func(dir GitDir) (done bool, err error) {
		gitDir := string(dir)

		// if removing a lock fails, we still want to try the other locks.
		var multi error

		// config.lock should be held for a very short amount of time.
		if err := removeFileOlderThan(filepath.Join(gitDir, "config.lock"), time.Minute); err != nil {
			multi = multierror.Append(multi, err)
		}
		// packed-refs can be held for quite a while, so we are conservative
		// with the age.
		if err := removeFileOlderThan(filepath.Join(gitDir, "packed-refs.lock"), time.Hour); err != nil {
			multi = multierror.Append(multi, err)
		}
		// we use the same conservative age for locks inside of refs
		if err := filepath.Walk(filepath.Join(gitDir, "refs"), func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				// ignore
				return nil
			}

			if fi.IsDir() {
				return nil
			}

			if !strings.HasSuffix(path, ".lock") {
				return nil
			}

			return removeFileOlderThan(path, time.Hour)
		}); err != nil {
			multi = multierror.Append(multi, err)
		}

		return false, multi
	}

	type cleanupFn struct {
		Name string
		Do   func(GitDir) (bool, error)
	}
	cleanups := []cleanupFn{
		// Do some sanity checks on the repository.
		{"maybe remove corrupt", maybeRemoveCorrupt},
		// If git is interrupted it can leave lock files lying around. It does
		// not clean these up, and instead fails commands.
		{"remove stale locks", removeStaleLocks},
		// We always want to have the same git attributes file at
		// info/attributes.
		{"ensure git attributes", ensureGitAttributes},
	}
	// Old git clones accumulate loose git objects that waste space and
	// slow down git operations. Periodically do a fresh clone to avoid
	// these problems. git gc is slow and resource intensive. It is
	// cheaper and faster to just reclone the repository.
	cleanups = append(cleanups, cleanupFn{"maybe reclone", maybeReclone})

	err := filepath.Walk(s.ReposDir, func(dir string, fi os.FileInfo, fileErr error) error {
		if fileErr != nil {
			return nil
		}

		if s.ignorePath(dir) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Look for $GIT_DIR
		if !fi.IsDir() || fi.Name() != ".git" {
			return nil
		}

		// We are sure this is a GIT_DIR after the above check
		gitDir := GitDir(dir)

		for _, cfn := range cleanups {
			done, err := cfn.Do(gitDir)
			if err != nil {
				log15.Error("error running cleanup command", "name", cfn.Name, "repo", gitDir, "error", err)
			}
			if done {
				break
			}
		}
		return filepath.SkipDir
	})
	if err != nil {
		log15.Error("cleanup: error iterating over repositories", "error", err)
	}

	if s.DiskSizer == nil {
		s.DiskSizer = &StatDiskSizer{}
	}
	b, err := s.howManyBytesToFree()
	if err != nil {
		log15.Error("cleanup: ensuring free disk space", "error", err)
	}
	if err := s.freeUpSpace(b); err != nil {
		log15.Error("cleanup: error freeing up space", "error", err)
	}
}

// DiskSizer gets information about disk size and free space.
type DiskSizer interface {
	BytesFreeOnDisk(mountPoint string) (uint64, error)
	DiskSizeBytes(mountPoint string) (uint64, error)
}

// howManyBytesToFree returns the number of bytes that should be freed to make sure
// there is sufficient disk space free to satisfy s.DesiredPercentFree.
func (s *Server) howManyBytesToFree() (int64, error) {
	// Check how much disk space is available.
	mountPoint, err := findMountPoint(s.ReposDir)
	if err != nil {
		return 0, errors.Wrap(err, "finding mount point for dir containing repos")
	}
	actualFreeBytes, err := s.DiskSizer.BytesFreeOnDisk(mountPoint)
	if err != nil {
		return 0, errors.Wrap(err, "finding the amount of space free on disk")
	}

	// Free up space if necessary.
	diskSizeBytes, err := s.DiskSizer.DiskSizeBytes(mountPoint)
	if err != nil {
		return 0, errors.Wrap(err, "getting disk size")
	}
	desiredFreeBytes := uint64(float64(s.DesiredPercentFree) / 100.0 * float64(diskSizeBytes))
	howManyBytesToFree := int64(desiredFreeBytes - actualFreeBytes)
	if howManyBytesToFree < 0 {
		howManyBytesToFree = 0
	}
	const G = float64(1024 * 1024 * 1024)
	log15.Debug("cleanup",
		"desired percent free", s.DesiredPercentFree,
		"actual percent free", float64(actualFreeBytes)/float64(diskSizeBytes)*100.0,
		"amount to free in GiB", float64(howManyBytesToFree)/G,
		"mount point", mountPoint)
	return howManyBytesToFree, nil
}

type StatDiskSizer struct{}

func (s *StatDiskSizer) BytesFreeOnDisk(mountPoint string) (uint64, error) {
	var fs syscall.Statfs_t
	if err := syscall.Statfs(mountPoint, &fs); err != nil {
		return 0, errors.Wrap(err, "statting")
	}
	free := fs.Bavail * uint64(fs.Bsize)
	return free, nil
}

func (s *StatDiskSizer) DiskSizeBytes(mountPoint string) (uint64, error) {
	var fs syscall.Statfs_t
	if err := syscall.Statfs(mountPoint, &fs); err != nil {
		return 0, errors.Wrap(err, "statting")
	}
	free := fs.Blocks * uint64(fs.Bsize)
	return free, nil
}

// findMountPoint searches upwards starting from the directory d to find the mount point.
func findMountPoint(d string) (string, error) {
	d, err := filepath.Abs(d)
	if err != nil {
		return "", errors.Wrapf(err, "getting absolute version of %s", d)
	}
	for {
		m, err := isMount(d)
		if err != nil {
			return "", errors.Wrapf(err, "finding out if %s is a mount point", d)
		}
		if m {
			return d, nil
		}
		d2 := filepath.Dir(d)
		if d2 == d {
			return d2, nil
		}
		d = d2
	}
}

// isMount tells whether the directory d is a mount point.
func isMount(d string) (bool, error) {
	ddev, err := device(d)
	if err != nil {
		return false, errors.Wrapf(err, "gettting device id for %s", d)
	}
	parent := filepath.Dir(d)
	if parent == d {
		return true, nil
	}
	pdev, err := device(parent)
	if err != nil {
		return false, errors.Wrapf(err, "getting device id for %s", parent)
	}
	return pdev != ddev, nil
}

// device gets the device id of a file f.
func device(f string) (int64, error) {
	fi, err := os.Stat(f)
	if err != nil {
		return 0, errors.Wrapf(err, "running stat on %s", f)
	}
	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("failed to get stat details for %s", f)
	}
	return int64(stat.Dev), nil
}

// freeUpSpace removes git directories under ReposDir, in order from least
// recently to most recently used, until it has freed howManyBytesToFree.
func (s *Server) freeUpSpace(howManyBytesToFree int64) error {
	// Get the git directories and their mod times.
	gitDirs, err := s.findGitDirs()
	if err != nil {
		return errors.Wrap(err, "finding git dirs")
	}
	dirModTimes := make(map[GitDir]time.Time, len(gitDirs))
	for _, d := range gitDirs {
		mt, err := gitDirModTime(d)
		if err != nil {
			return errors.Wrap(err, "computing mod time of git dir")
		}
		dirModTimes[d] = mt
	}

	// Sort the repos from least to most recently used.
	sort.Slice(gitDirs, func(i, j int) bool {
		return dirModTimes[gitDirs[i]].Before(dirModTimes[gitDirs[j]])
	})

	// Remove repos until howManyBytesToFree is met or exceeded.
	var spaceFreed int64
	mountPoint, err := findMountPoint(s.ReposDir)
	if err != nil {
		return errors.Wrap(err, "finding mount point")
	}
	diskSizeBytes, err := s.DiskSizer.DiskSizeBytes(mountPoint)
	if err != nil {
		return errors.Wrap(err, "getting disk size")
	}
	for _, d := range gitDirs {
		if spaceFreed >= howManyBytesToFree {
			return nil
		}
		delta, err := dirSize(string(d))
		if err != nil {
			return errors.Wrapf(err, "computing size of directory %s", d)
		}
		if err := s.removeRepoDirectory(d); err != nil {
			return errors.Wrap(err, "removing repo directory")
		}
		spaceFreed += delta

		// Report the new disk usage situation after removing this repo.
		actualFreeBytes, err := s.DiskSizer.BytesFreeOnDisk(mountPoint)
		if err != nil {
			return errors.Wrap(err, "finding the amount of space free on disk")
		}
		G := float64(1024 * 1024 * 1024)
		log15.Warn("cleanup: removed least recently used repo",
			"repo", d,
			"how old", time.Since(dirModTimes[d]),
			"free space in GiB", float64(actualFreeBytes)/G,
			"actual percent of disk space free", float64(actualFreeBytes)/float64(diskSizeBytes)*100.0,
			"desired percent of disk space free", float64(s.DesiredPercentFree),
			"space freed in GiB", float64(spaceFreed)/G,
			"how much space to free in GiB", float64(howManyBytesToFree)/G)
	}

	// Check.
	if spaceFreed < howManyBytesToFree {
		return fmt.Errorf("only freed %d bytes, wanted to free %d", spaceFreed, howManyBytesToFree)
	}
	return nil
}

func gitDirModTime(d GitDir) (time.Time, error) {
	head, err := os.Stat(d.Path("HEAD"))
	if err != nil {
		return time.Time{}, errors.Wrap(err, "getting repository modification time")
	}
	return head.ModTime(), nil
}

func (s *Server) findGitDirs() ([]GitDir, error) {
	var dirs []GitDir
	err := filepath.Walk(s.ReposDir, func(path string, fi os.FileInfo, fileErr error) error {
		if fileErr != nil {
			return nil
		}
		if s.ignorePath(path) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !fi.IsDir() || fi.Name() != ".git" {
			return nil
		}
		dirs = append(dirs, GitDir(path))
		return filepath.SkipDir
	})
	if err != nil {
		return nil, errors.Wrap(err, "findGitDirs")
	}
	return dirs, nil
}

// dirSize returns the total size in bytes of all the files under d.
func dirSize(d string) (int64, error) {
	var size int64
	err := filepath.Walk(d, func(path string, fi os.FileInfo, fileErr error) error {
		if fileErr != nil {
			return nil
		}
		if fi.IsDir() {
			return nil
		}
		size += fi.Size()
		return nil
	})
	if err != nil {
		return 0, errors.Wrapf(err, "walking dir tree from %s to find size", d)
	}
	return size, nil
}

// removeRepoDirectory atomically removes a directory from s.ReposDir.
//
// It first moves the directory to a temporary location to avoid leaving
// partial state in the event of server restart or concurrent modifications to
// the directory.
//
// Additionally it removes parent empty directories up until s.ReposDir.
func (s *Server) removeRepoDirectory(gitDir GitDir) error {
	dir := string(gitDir)

	// Rename out of the location so we can atomically stop using the repo.
	tmp, err := s.tempDir("delete-repo")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if err := renameAndSync(dir, filepath.Join(tmp, "repo")); err != nil {
		return err
	}

	// Everything after this point is just cleanup, so any error that occurs
	// should not be returned, just logged.

	// Cleanup empty parent directories. We just attempt to remove and if we
	// have a failure we assume it's due to the directory having other
	// children. If we checked first we could race with someone else adding a
	// new clone.
	rootInfo, err := os.Stat(s.ReposDir)
	if err != nil {
		log15.Warn("Failed to stat ReposDir", "error", err)
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
			log15.Warn("failed to stat parent directory", "dir", current, "error", err)
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

	// Delete the atomically renamed dir. We do this last since if it fails we
	// will rely on a janitor job to clean up for us.
	if err := os.RemoveAll(filepath.Join(tmp, "repo")); err != nil {
		log15.Warn("failed to cleanup after removing dir", "dir", dir, "error", err)
	}

	return nil
}

// cleanTmpFiles tries to remove tmp_pack_* files from .git/objects/pack.
// These files can be created by an interrupted fetch operation,
// and would be purged by `git gc --prune=now`, but `git gc` is
// very slow. Removing these files while they're in use will cause
// an operation to fail, but not damage the repository.
func (s *Server) cleanTmpFiles(dir GitDir) {
	now := time.Now()
	packdir := dir.Path("objects", "pack")
	err := filepath.Walk(packdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// ignore
			return nil
		}
		if path != packdir && info.IsDir() {
			return filepath.SkipDir
		}
		file := filepath.Base(path)
		if strings.HasPrefix(file, "tmp_pack_") {
			if now.Sub(info.ModTime()) > longGitCommandTimeout {
				err := os.Remove(path)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		log15.Error("error removing tmp_pack_* files", "error", err)
	}
}

// SetupAndClearTmp sets up the the tempdir for ReposDir as well as clearing it
// out. It returns the temporary directory location.
func (s *Server) SetupAndClearTmp() (string, error) {
	// Additionally we create directories with the prefix .tmp-old which are
	// asynchronously removed. We do not remove in place since it may be a
	// slow operation to block on. Our tmp dir will be ${s.ReposDir}/.tmp
	dir := filepath.Join(s.ReposDir, tempDirName) // .tmp
	oldPrefix := tempDirName + "-old"
	if _, err := os.Stat(dir); err == nil {
		// Rename the current tmp file so we can asynchronously remove it. Use
		// a consistent pattern so if we get interrupted, we can clean it
		// another time.
		oldTmp, err := ioutil.TempDir(s.ReposDir, oldPrefix)
		if err != nil {
			return "", err
		}
		// oldTmp dir exists, so we need to use a child of oldTmp as the
		// rename target.
		if err := os.Rename(dir, filepath.Join(oldTmp, tempDirName)); err != nil {
			return "", err
		}
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}

	// Asynchronously remove old temporary directories
	files, err := ioutil.ReadDir(s.ReposDir)
	if err != nil {
		log15.Error("failed to do tmp cleanup", "error", err)
	} else {
		for _, f := range files {
			// Remove older .tmp directories as well as our older tmp-
			// directories we would place into ReposDir. In September 2018 we
			// can remove support for removing tmp- directories.
			if !strings.HasPrefix(f.Name(), oldPrefix) && !strings.HasPrefix(f.Name(), "tmp-") {
				continue
			}
			go func(path string) {
				if err := os.RemoveAll(path); err != nil {
					log15.Error("cleanup: failed to remove old temporary directory", "path", path, "error", err)
				}
			}(filepath.Join(s.ReposDir, f.Name()))
		}
	}

	return dir, nil
}

// getRecloneTime returns an approximate time a repository is cloned. If the
// value is not stored in the repository, the reclone time for the repository
// is set to now.
func getRecloneTime(dir GitDir) (time.Time, error) {
	// We store the time we recloned the repository. If the value is missing,
	// we store the current time. This decouples this timestamp from the
	// different ways a clone can appear in gitserver.
	update := func() (time.Time, error) {
		now := time.Now()
		cmd := exec.Command("git", "config", "--add", "sourcegraph.recloneTimestamp", strconv.FormatInt(time.Now().Unix(), 10))
		cmd.Dir = string(dir)
		if _, err := cmd.Output(); err != nil {
			return now, errors.Wrap(wrapCmdError(cmd, err), "failed to update recloneTimestamp")
		}
		return now, nil
	}

	cmd := exec.Command("git", "config", "--get", "sourcegraph.recloneTimestamp")
	cmd.Dir = string(dir)
	out, err := cmd.Output()
	if err != nil {
		// Exit code 1 means the key is not set.
		if ee, ok := err.(*exec.ExitError); ok && ee.Sys().(syscall.WaitStatus).ExitStatus() == 1 {
			return update()
		}
		return time.Unix(0, 0), errors.Wrap(wrapCmdError(cmd, err), "failed to determine clone timestamp")
	}

	sec, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 0)
	if err != nil {
		// If the value is bad update it to the current time
		now, err2 := update()
		if err2 != nil {
			err = err2
		}
		return now, err
	}

	return time.Unix(sec, 0), nil
}

// randDuration returns a psuedo-random duration between [0, d)
func randDuration(d time.Duration) time.Duration {
	return time.Duration(rand.Int63n(int64(d)))
}

// wrapCmdError will wrap errors for cmd to include the arguments. If the error
// is an exec.ExitError and cmd was invoked with Output(), it will also include
// the captured stderr.
func wrapCmdError(cmd *exec.Cmd, err error) error {
	if err == nil {
		return nil
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return errors.Wrapf(err, "%s %s failed with stderr: %s", cmd.Path, strings.Join(cmd.Args, " "), string(ee.Stderr))
	}
	return errors.Wrapf(err, "%s %s failed", cmd.Path, strings.Join(cmd.Args, " "))
}

// removeFileOlderThan removes path if its mtime is older than maxAge. If the
// file is missing, no error is returned.
func removeFileOlderThan(path string, maxAge time.Duration) error {
	fi, err := os.Stat(filepath.Join(path))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	age := time.Since(fi.ModTime())
	if age < maxAge {
		return nil
	}

	log15.Debug("removing stale lock file", "path", path, "age", age)
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
