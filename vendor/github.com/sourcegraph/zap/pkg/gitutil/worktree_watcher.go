package gitutil

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rjeczalik/notify"
)

// WorktreeWatcher watches a directory and its subdirectories for
// changes to files.
//
// Upon creation (with NewWorktreeWatcher), file changes are sent
// on the Events channel.
type WorktreeWatcher struct {
	Events <-chan notify.EventInfo
	Errors <-chan error

	watcher chan notify.EventInfo
}

// NewWorktreeWatcher creates a new watcher for the worktree at
// dir. Callers must call Close when done to free resources.
func NewWorktreeWatcher(repo interface {
	WorktreeDir() string
}, ignorePath func(string) (bool, error)) (*WorktreeWatcher, error) {
	dir, err := filepath.Abs(repo.WorktreeDir())
	if err != nil {
		return nil, err
	}

	if ignorePath == nil {
		// Default: ignore nothing.
		ignorePath = func(string) (bool, error) { return false, nil }
	}
	gitDir := filepath.Join(dir, ".git") + string(os.PathSeparator)

	// TODO bigger buffer? if we can't keep up, notify will drop events
	watcher := make(chan notify.EventInfo, 1)
	if err := notify.Watch(filepath.Join(dir, "..."), watcher, notify.All); err != nil {
		return nil, err
	}

	eventsCh := make(chan notify.EventInfo)
	errorsCh := make(chan error)

	go func() {
		for {
			e, ok := <-watcher
			if !ok {
				break
			}

			if strings.HasPrefix(e.Path()+string(os.PathSeparator), gitDir) {
				continue
			} else if ignore, err := ignorePath(e.Path()); err != nil {
				errorsCh <- err
				continue
			} else if ignore {
				continue
			}

			if e.Event()&notify.Remove != 0 {
				eventsCh <- e
				continue
			}

			fi, err := os.Stat(e.Path())
			if err != nil && !os.IsNotExist(err) {
				errorsCh <- err
				continue
			}
			if fi == nil {
				continue
			}

			if fi.Mode().IsRegular() {
				eventsCh <- e
			}
		}
	}()

	return &WorktreeWatcher{
		Events:  eventsCh,
		Errors:  errorsCh,
		watcher: watcher,
	}, nil
}

// Close stops watching and closes w.Events and w.Errors.
func (w *WorktreeWatcher) Close() error {
	notify.Stop(w.watcher)
	return nil
}
