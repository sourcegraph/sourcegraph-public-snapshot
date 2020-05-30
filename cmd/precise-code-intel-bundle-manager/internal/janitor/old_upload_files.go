package janitor

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
)

// removeOldUploadFiles removes all upload files that are older than the configured
// max unconverted upload age. These files are left on disk when an upload fails to
// process. These files can also be cleaned up by removeOrphanedBundleFiles when the
// upload is properly in an errored state, but we keep this cleanup routine here as
// well for good measure.
func (j *Janitor) removeOldUploadFiles() error {
	fileInfos, err := ioutil.ReadDir(paths.UploadsDir(j.bundleDir))
	if err != nil {
		return err
	}

	for _, fileInfo := range fileInfos {
		age := time.Since(fileInfo.ModTime())
		if age < j.maxUploadAge {
			continue
		}

		path := filepath.Join(paths.UploadsDir(j.bundleDir), fileInfo.Name())

		if err := os.Remove(path); err != nil {
			j.metrics.Errors.Inc()
			log15.Error("Failed to remove file", "path", path, "err", err)
			continue
		}

		log15.Debug("Removed old upload file", "path", path, "age", age)
		j.metrics.UploadFilesRemoved.Add(1)
	}

	return nil
}
