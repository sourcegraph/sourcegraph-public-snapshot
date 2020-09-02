package janitor

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-bundle-manager/internal/paths"
)

// removeOldUploadFiles removes all upload files that are older than the configured
// max unconverted upload age. These files are left on disk when an upload fails to
// process. These files can also be cleaned up by removeOrphanedBundleFiles when the
// upload is properly in an errored state, but we keep this cleanup routine here as
// well for good measure.
func (j *Janitor) removeOldUploadFiles(ctx context.Context) error {
	return j.removeOldFiles(paths.UploadsDir(j.bundleDir), j.maxUploadAge, func(path string, age time.Duration) {
		log15.Debug("Removed old upload file", "path", path, "age", age)
		j.metrics.UploadFilesRemoved.Inc()
	})
}

// removeOldUploadPartFiles removes all upload part files that are older than the configured max
// upload part age. These files are left on disk if an upload does not complete within a CI run.
func (j *Janitor) removeOldUploadPartFiles(ctx context.Context) error {
	return j.removeOldFiles(paths.UploadPartsDir(j.bundleDir), j.maxUploadPartAge, func(path string, age time.Duration) {
		log15.Debug("Removed old upload part file", "path", path, "age", age)
		j.metrics.PartFilesRemoved.Inc()
	})
}

// removeOldDatabasePartFiles removes all database part files that are older than the configured
// max database part age. These files are left on disk if a worker does not successfully complete
// all requests of a SendDB command.
func (j *Janitor) removeOldDatabasePartFiles(ctx context.Context) error {
	return j.removeOldFiles(paths.DBPartsDir(j.bundleDir), j.maxDatabasePartAge, func(path string, age time.Duration) {
		log15.Debug("Removed old database part file", "path", path, "age", age)
		j.metrics.PartFilesRemoved.Inc()
	})
}

// removeOldFiles removes all part files within the given directrory that are older than the given
// age. The onRemove function is called when a file or directory is successfully unlinked.
func (j *Janitor) removeOldFiles(root string, maxAge time.Duration, onRemove func(path string, age time.Duration)) error {
	fileInfos, err := ioutil.ReadDir(root)
	if err != nil {
		return err
	}

	for _, fileInfo := range fileInfos {
		age := time.Since(fileInfo.ModTime())
		if age <= maxAge {
			continue
		}

		if path := filepath.Join(root, fileInfo.Name()); j.remove(path) {
			onRemove(path, age)
		}
	}

	return nil
}
