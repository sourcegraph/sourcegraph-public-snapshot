package janitor

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
)

var partFilenamePattern = regexp.MustCompile(`[0-9]+.[0-9]+.lsif.gz`)

// removeOldDatabasePartFiles removes all database part files that are older than the configured
// max database part age. These files are left on disk if a worker does not successfully complete
// all requests of a SendDB command.
func (j *Janitor) removeOldDatabasePartFiles() error {
	fileInfos, err := ioutil.ReadDir(paths.DBsDir(j.bundleDir))
	if err != nil {
		return err
	}

	for _, fileInfo := range fileInfos {
		age := time.Since(fileInfo.ModTime())
		if age < j.maxDatabasePartAge || !partFilenamePattern.MatchString(fileInfo.Name()) {
			continue
		}

		path := filepath.Join(paths.DBsDir(j.bundleDir), fileInfo.Name())
		if err := os.Remove(path); err != nil {
			return err
		}

		log15.Debug("Removed old database part file", "path", path, "age", age)
		j.metrics.DatabasePartFilesRemoved.Inc()
	}

	return nil
}
