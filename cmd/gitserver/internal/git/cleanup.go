package git

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

// CleanTmpPackFiles tries to remove tmp_pack_* files from .git/objects/pack.
// These files can be created by an interrupted fetch operation,
// and would be purged by `git gc --prune=now`, but `git gc` is
// very slow. Removing these files while they're in use will cause
// an operation to fail, but not damage the repository.
func CleanTmpPackFiles(logger log.Logger, dir common.GitDir) {
	logger = logger.Scoped("cleanup.cleanTmpFiles")

	now := time.Now()
	packdir := dir.Path("objects", "pack")
	err := gitserverfs.BestEffortWalk(packdir, func(path string, d fs.DirEntry) error {
		if path != packdir && d.IsDir() {
			return filepath.SkipDir
		}
		file := filepath.Base(path)
		if strings.HasPrefix(file, "tmp_pack_") {
			info, err := d.Info()
			if err != nil {
				return err
			}
			if now.Sub(info.ModTime()) > conf.GitLongCommandTimeout() {
				err := os.Remove(path)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		logger.Error("error removing tmp_pack_* files", log.Error(err))
	}
}
