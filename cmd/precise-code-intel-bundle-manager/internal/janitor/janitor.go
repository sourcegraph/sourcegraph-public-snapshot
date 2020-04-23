package janitor

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/lsifserver/client"
)

// DeadDumpBatchSize is the maximum number of dump ids to request at once from
// the precise-code-intel-api-server.
const DeadDumpBatchSize = 100

type Janitor struct {
	bundleDir               string
	desiredPercentFree      int
	janitorInterval         time.Duration
	maxUnconvertedUploadAge time.Duration
}

type JanitorOpts struct {
	BundleDir               string
	DesiredPercentFree      int
	JanitorInterval         time.Duration
	MaxUnconvertedUploadAge time.Duration
}

func NewJanitor(opts JanitorOpts) *Janitor {
	return &Janitor{
		bundleDir:               opts.BundleDir,
		desiredPercentFree:      opts.DesiredPercentFree,
		janitorInterval:         opts.JanitorInterval,
		maxUnconvertedUploadAge: opts.MaxUnconvertedUploadAge,
	}
}

func (j *Janitor) Start() {
	for {
		if err := j.step(); err != nil {
			log15.Error("Failed to run janitor process", "err", err)
		}

		time.Sleep(j.janitorInterval)
	}
}

// step performs a best-effort cleanup. See the following methods for more specifics.
//   - cleanFailedUploads
//   - removeDeadDumps
//   - freeSpace
func (j *Janitor) step() error {
	cleanupFns := []func() error{
		j.cleanFailedUploads,
		func() error { return j.removeDeadDumps(client.DefaultClient.States) },
		func() error { return j.freeSpace(client.DefaultClient.Prune) },
	}

	for _, fn := range cleanupFns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

// cleanFailedUploads removes all upload files that are older than the configured
// max unconverted upload age.
func (j *Janitor) cleanFailedUploads() error {
	fileInfos, err := ioutil.ReadDir(paths.UploadsDir(j.bundleDir))
	if err != nil {
		return err
	}

	for _, fileInfo := range fileInfos {
		if time.Since(fileInfo.ModTime()) < j.maxUnconvertedUploadAge {
			continue
		}

		if err := os.Remove(filepath.Join(paths.UploadsDir(j.bundleDir), fileInfo.Name())); err != nil {
			return err
		}
	}

	return nil
}

// removeDeadDumps calls the precise-code-intel-api-server to get the current
// state of the dumps known by this bundle manager. Any dump on disk that is
// in an errored state or is unknown by the API is removed.
func (j *Janitor) removeDeadDumps(statesFn func(ctx context.Context, ids []int) (map[int]string, error)) error {
	pathsByID, err := j.databasePathsByID()
	if err != nil {
		return err
	}

	var ids []int
	for id := range pathsByID {
		ids = append(ids, id)
	}

	allStates := map[int]string{}
	for _, batch := range batchIntSlice(ids, DeadDumpBatchSize) {
		states, err := statesFn(context.Background(), batch)
		if err != nil {
			return err
		}

		for k, v := range states {
			allStates[k] = v
		}
	}

	for id, path := range pathsByID {
		if state, exists := allStates[id]; !exists || state == "errored" {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	}

	return nil
}

// databasePathsByID returns map of dump ids to their path on disk.
func (j *Janitor) databasePathsByID() (map[int]string, error) {
	fileInfos, err := ioutil.ReadDir(paths.DBsDir(j.bundleDir))
	if err != nil {
		return nil, err
	}

	pathsByID := map[int]string{}
	for _, fileInfo := range fileInfos {
		if id, err := strconv.Atoi(strings.Split(fileInfo.Name(), ".")[0]); err == nil {
			pathsByID[int(id)] = filepath.Join(paths.DBsDir(j.bundleDir), fileInfo.Name())
		}
	}

	return pathsByID, nil
}

// freeSpace determines the space available on the device containing the bundle directory,
// then calls cleanOldDumps to free enough space to get back below the disk usage threshold.
func (j *Janitor) freeSpace(pruneFn func(ctx context.Context) (int64, bool, error)) error {
	var fs syscall.Statfs_t
	if err := syscall.Statfs(j.bundleDir, &fs); err != nil {
		return err
	}

	diskSizeBytes := fs.Blocks * uint64(fs.Bsize)
	freeBytes := fs.Bavail * uint64(fs.Bsize)
	desiredFreeBytes := uint64(float64(diskSizeBytes) * float64(j.desiredPercentFree) / 100.0)

	if freeBytes < desiredFreeBytes {
		return j.cleanOldDumps(pruneFn, uint64(desiredFreeBytes-freeBytes))
	}

	return nil
}

// cleanOldDumps removes dumps from the database (via precise-code-intel-api-server)
// and the filesystem until at least bytesToFree, or there are no more prunable dumps.
func (j *Janitor) cleanOldDumps(pruneFn func(ctx context.Context) (int64, bool, error), bytesToFree uint64) error {
	for bytesToFree > 0 {
		bytesRemoved, pruned, err := j.cleanOldDump(pruneFn)
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

// cleanOldDump calls the precise-code-intel-api-server for the identifier of
// the oldest dump to remove then deletes the associated file. This method
// returns the size of the deleted file on success, and returns a false-valued
// flag if there are no prunable dumps.
func (j *Janitor) cleanOldDump(pruneFn func(ctx context.Context) (int64, bool, error)) (uint64, bool, error) {
	id, prunable, err := pruneFn(context.Background())
	if err != nil || !prunable {
		return 0, false, err
	}

	filename := paths.DBFilename(j.bundleDir, id)

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return 0, false, err
	}

	if err := os.Remove(filename); err != nil {
		return 0, false, err
	}

	return uint64(fileInfo.Size()), true, nil
}

// batchIntSlice returns slices of s (in order) at most batchSize in length.
func batchIntSlice(s []int, batchSize int) [][]int {
	batches := [][]int{}
	for len(s) > batchSize {
		batches = append(batches, s[:batchSize])
		s = s[batchSize:]
	}

	if len(s) > 0 {
		batches = append(batches, s)
	}

	return batches
}
