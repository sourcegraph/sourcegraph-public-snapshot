package janitor

import (
	"context"
	"os"
	"syscall"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
)

// freeSpace determines the space available on the device containing the bundle directory,
// then calls cleanOldBundles to free enough space to get back below the disk usage threshold.
func (j *Janitor) freeSpace() error {
	var fs syscall.Statfs_t
	if err := syscall.Statfs(j.bundleDir, &fs); err != nil {
		return err
	}

	diskSizeBytes := fs.Blocks * uint64(fs.Bsize)
	freeBytes := fs.Bavail * uint64(fs.Bsize)
	desiredFreeBytes := uint64(float64(diskSizeBytes) * float64(j.desiredPercentFree) / 100.0)
	if freeBytes >= desiredFreeBytes {
		return nil
	}

	return j.evictBundles(desiredFreeBytes - freeBytes)
}

// evictBundles removes completed upload recors from the database and then deletes the
// associated bundle file from the filesystem until at least bytesToFree, or there are
// no more prunable bundles.
func (j *Janitor) evictBundles(bytesToFree uint64) error {
	for bytesToFree > 0 {
		bytesRemoved, pruned, err := j.evictBundle()
		if err != nil {
			return err
		}
		if !pruned {
			break
		}

		if bytesRemoved >= bytesToFree {
			break
		}

		bytesToFree -= bytesRemoved
	}

	return nil
}

// evictBundle removes the oldest bundle from the database, then deletes the associated file.
// This method returns the size of the deleted file on success, and returns a false-valued
// flag if there are no prunable bundles.
func (j *Janitor) evictBundle() (uint64, bool, error) {
	id, prunable, err := j.db.DeleteOldestDump(context.Background())
	if err != nil {
		return 0, false, errors.Wrap(err, "db.DeleteOldestDump")
	}
	if !prunable {
		return 0, false, nil
	}

	path := paths.DBFilename(j.bundleDir, int64(id))

	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Bundle file already gone, continue
			return 0, true, nil
		}

		return 0, false, err
	}

	if err := os.Remove(path); err != nil {
		j.metrics.Errors.Inc()
		log15.Error("Failed to remove file", "path", path, "err", err)
		return 0, true, nil
	}

	log15.Debug("Removed evicted bundle file", "id", id, "path", path)
	j.metrics.EvictedBundleFilesRemoved.Add(1)
	return uint64(fileInfo.Size()), true, nil
}
