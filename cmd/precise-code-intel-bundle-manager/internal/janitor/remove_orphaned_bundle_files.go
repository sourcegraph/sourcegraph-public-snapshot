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
	"github.com/sourcegraph/sourcegraph/internal/codeintel/lsifserver/client"
)

// OrphanedBundleBatchSize is the maximum number of bundle ids to request at
// once from the precise-code-intel-api-server.
const OrphanedBundleBatchSize = 100

type StatesFn func(ctx context.Context, ids []int) (map[int]string, error)

func defaultStatesFn(ctx context.Context, ids []int) (map[int]string, error) {
	states, err := client.DefaultClient.States(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, "lsifserver.States")
	}

	return states, nil
}

// removeOrphanedBundleFiles calls the precise-code-intel-api-server to get the
// current state of the bundle known by this bundle manager. Any bundle on disk
// that is in an errored state or is unknown by the API is removed.
func (j *Janitor) removeOrphanedBundleFiles(statesFn StatesFn) error {
	pathsByID, err := j.databasePathsByID()
	if err != nil {
		return err
	}

	var ids []int
	for id := range pathsByID {
		ids = append(ids, id)
	}

	allStates := map[int]string{}
	for _, batch := range batchIntSlice(ids, OrphanedBundleBatchSize) {
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

			log15.Debug("Removed orphaned bundle file", "id", id, "path", path)
			j.metrics.OprphanedBundleFilesRemoved.Add(1)
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
			pathsByID[int(id)] = filepath.Join(paths.DBsDir(j.bundleDir), fileInfo.Name())
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
