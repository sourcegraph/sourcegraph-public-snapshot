package janitor

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-bundle-manager/internal/paths"
)

// GetStateBatchSize is the maximum number of bundle ids to request the state of from the
// database at once.
const GetStateBatchSize = 100

// MinimumUploadAge is the minimum age an upload has to be before it can be considered orphaned.
// We allow a grace period here because the transaction that writes the initial record may not
// have committed by the time this cleanup function runs.
const MinimumUploadAge = time.Minute

// removeOrphanedUploadFiles removes any upload file on disk that is associated with an
// errored (or missing) entry in the database.
func (j *Janitor) removeOrphanedUploadFiles() error {
	pathsByID, err := j.uploadPathsByID()
	if err != nil {
		return err
	}

	return j.removeOrphans(pathsByID, func(id int, path string) {
		log15.Debug("Removed orphaned upload file", "id", id, "path", path)
		j.metrics.OrphanedFilesRemoved.Inc()
	})
}

// removeOrphanedUploadFiles removes any bundle file on disk that is associated with an
// errored (or missing) entry in the database.
func (j *Janitor) removeOrphanedBundleFiles() error {
	pathsByID, err := j.databasePathsByID()
	if err != nil {
		return err
	}

	return j.removeOrphans(pathsByID, func(id int, path string) {
		log15.Debug("Removed orphaned bundle file", "id", id, "path", path)
		j.metrics.OrphanedFilesRemoved.Inc()
	})
}

// removeOrphans removes files from the given mapping if the upload identifier matches an
// errored (or missing) entry in the database. The onRemove function is called when a file
// or directory is successfully unlinked.
func (j *Janitor) removeOrphans(pathsByID map[int]string, onRemove func(id int, path string)) error {
	var ids []int
	for id := range pathsByID {
		ids = append(ids, id)
	}

	states := map[int]string{}
	for _, batch := range batchIntSlice(ids, GetStateBatchSize) {
		batchStates, err := j.store.GetStates(context.Background(), batch)
		if err != nil {
			return errors.Wrap(err, "store.GetStates")
		}

		for k, v := range batchStates {
			states[k] = v
		}
	}

	for id, path := range pathsByID {
		if state, exists := states[id]; !exists || state == "errored" {
			if j.remove(path) {
				onRemove(id, path)
			}
		}
	}

	return nil
}

// uploadPathsByID returns map of bundle ids to their upload file on disk.
func (j *Janitor) uploadPathsByID() (map[int]string, error) {
	fileInfos, err := ioutil.ReadDir(paths.UploadsDir(j.bundleDir))
	if err != nil {
		return nil, err
	}

	pathsByID := map[int]string{}
	for _, fileInfo := range fileInfos {
		if age := time.Since(fileInfo.ModTime()); age <= MinimumUploadAge {
			continue
		}

		if id, err := strconv.Atoi(strings.Split(fileInfo.Name(), ".")[0]); err == nil {
			pathsByID[id] = filepath.Join(paths.UploadsDir(j.bundleDir), fileInfo.Name())
		}
	}

	return pathsByID, nil
}

// databasePathsByID returns map of bundle ids to their path on disk.
func (j *Janitor) databasePathsByID() (map[int]string, error) {
	fileInfos, err := ioutil.ReadDir(paths.DBsDir(j.bundleDir))
	if err != nil {
		return nil, err
	}

	pathsByID := map[int]string{}
	for _, fileInfo := range fileInfos {
		if id, err := strconv.Atoi(fileInfo.Name()); err == nil {
			pathsByID[id] = filepath.Join(paths.DBsDir(j.bundleDir), fileInfo.Name())
		}
	}

	return pathsByID, nil
}
