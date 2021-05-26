package log

import (
	"sync"

	"github.com/hashicorp/go-multierror"
)

type LogManager interface {
	AddTask(string) (TaskLogger, error)
	Close() error
	LogFiles() []string
}

var _ LogManager = &Manager{}

type Manager struct {
	dir      string
	keepLogs bool

	tasks sync.Map
}

func NewManager(dir string, keepLogs bool) *Manager {
	return &Manager{dir: dir, keepLogs: keepLogs}
}

func (lm *Manager) AddTask(slug string) (TaskLogger, error) {
	tl, err := newTaskLogger(slug, lm.keepLogs, lm.dir)
	if err != nil {
		return nil, err
	}

	lm.tasks.Store(slug, tl)
	return tl, nil
}

func (lm *Manager) Close() error {
	var errs *multierror.Error

	lm.tasks.Range(func(_, v interface{}) bool {
		logger := v.(*FileTaskLogger)

		if err := logger.Close(); err != nil {
			errs = multierror.Append(errs, err)
		}

		return true
	})

	return errs
}

func (lm *Manager) LogFiles() []string {
	var files []string

	lm.tasks.Range(func(_, v interface{}) bool {
		files = append(files, v.(*FileTaskLogger).Path())
		return true
	})

	return files
}
