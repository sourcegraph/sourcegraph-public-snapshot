package log

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ LogManager = &DiskManager{}

type DiskManager struct {
	dir      string
	keepLogs bool

	tasks sync.Map
}

func NewDiskManager(dir string, keepLogs bool) *DiskManager {
	return &DiskManager{dir: dir, keepLogs: keepLogs}
}

func (lm *DiskManager) AddTask(slug string) (TaskLogger, error) {
	tl, err := newTaskLogger(slug, lm.keepLogs, lm.dir)
	if err != nil {
		return nil, err
	}

	lm.tasks.Store(slug, tl)
	return tl, nil
}

func (lm *DiskManager) Close() error {
	var errs errors.MultiError

	lm.tasks.Range(func(_, v interface{}) bool {
		logger := v.(*FileTaskLogger)

		if err := logger.Close(); err != nil {
			errs = errors.Append(errs, err)
		}

		return true
	})

	return errs
}

func (lm *DiskManager) LogFiles() []string {
	var files []string

	lm.tasks.Range(func(_, v interface{}) bool {
		files = append(files, v.(*FileTaskLogger).Path())
		return true
	})

	return files
}
