package janitor

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
)

// cleanOldUploads removes all upload files that are older than the configured
// max unconverted upload age.
func (j *Janitor) cleanOldUploads() error {
	fileInfos, err := ioutil.ReadDir(paths.UploadsDir(j.BundleDir))
	if err != nil {
		return err
	}

	count := 0
	for _, fileInfo := range fileInfos {
		age := time.Since(fileInfo.ModTime())
		if age < j.MaxUploadAge {
			continue
		}

		path := filepath.Join(paths.UploadsDir(j.BundleDir), fileInfo.Name())
		if err := os.Remove(path); err != nil {
			return err
		}

		count++
		log15.Debug("Removed old upload", "path", fileInfo.Name(), "age", age)
	}

	j.Metrics.OldUploads.Add(float64(count))
	return nil
}
