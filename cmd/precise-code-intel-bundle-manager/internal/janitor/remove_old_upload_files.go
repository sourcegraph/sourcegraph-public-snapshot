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
// max unconverted upload age.
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
			return err
		}

		log15.Debug("Removed old upload file", "path", path, "age", age)
		j.metrics.UploadFilesRemoved.Add(1)
	}

	return nil
}
