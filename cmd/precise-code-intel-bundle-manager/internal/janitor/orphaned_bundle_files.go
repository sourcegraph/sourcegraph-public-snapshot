package janitor

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
)

// OrphanedBundleBatchSize is the maximum number of bundle ids to request the state
// of from the database at once.
const OrphanedBundleBatchSize = 100

// removeOrphanedBundleFiles calls the removes any bundle on disk that is in an errored
// state or has no associated record in the database.
func (j *Janitor) removeOrphanedBundleFiles() error {
	pathsByID, err := j.databasePathsByID()
	if err != nil {
		return err
	}

	var ids []int
	for id := range pathsByID {
		ids = append(ids, id)
	}

	states := map[int]string{}
	for _, batch := range batchIntSlice(ids, OrphanedBundleBatchSize) {
		batchStates, err := j.db.GetStates(context.Background(), batch)
		if err != nil {
			return errors.Wrap(err, "db.GetStates")
		}

		for k, v := range batchStates {
			states[k] = v
		}
	}

	for id, path := range pathsByID {
		if state, exists := states[id]; !exists || state == "errored" {
			if err := os.Remove(path); err != nil {
				j.metrics.Errors.Inc()
				log15.Error("Failed to remove file", "path", path, "err", err)
				continue
			}

			log15.Debug("Removed orphaned bundle file", "id", id, "path", path)
			j.metrics.OrphanedBundleFilesRemoved.Add(1)
		}
	}

	return nil
}

// databasePathsByID returns map of bundle ids to their path on disk.
func (j *Janitor) databasePathsByID() (map[int]string, error) {
	fileInfos, err := ioutil.ReadDir(paths.DBsDir(j.bundleDir))
	if err != nil {
		return nil, err
	}

	pathsByID := map[int]string{}
	for _, fileInfo := range fileInfos {
		if id, err := strconv.Atoi(strings.Split(fileInfo.Name(), ".")[0]); err == nil {
			pathsByID[id] = filepath.Join(paths.DBsDir(j.bundleDir), fileInfo.Name())
		}
	}

	return pathsByID, nil
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
