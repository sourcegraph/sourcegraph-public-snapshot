package gitutil

import (
	"io"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// WorktreeWatcher watches a directory and its subdirectories for
// changes to files.
//
// Upon creation (with NewWorktreeWatcher), file changes are sent
// on the Events channel.
type WorktreeWatcher struct {
	Events <-chan fsnotify.Event
	Errors <-chan error

	watcher io.Closer
}

// NewWorktreeWatcher creates a new watcher for the worktree at
// dir. Callers must call Close when done to free resources.
func NewWorktreeWatcher(repo interface {
	WorktreeDir() string
}, ignorePath func(string) (bool, error)) (*WorktreeWatcher, error) {
	dir := repo.WorktreeDir()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if ignorePath == nil {
		// Default: ignore nothing.
		ignorePath = func(string) (bool, error) { return false, nil }
	}

	recursivelyWatchDir := func(root string) error {
		// TODO(sqs): optimize by using `git ls-files`? instead of calling ignorePath many times
		if err := watcher.Add(root); err != nil {
			return err
		}
		return filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fi.Mode().IsDir() {
				if fi.Name() == ".git" {
					return filepath.SkipDir
				}
				if err := watcher.Add(path); err != nil {
					return err
				}
			}
			return nil
		})
	}

	// TODO(sqs): watch more than just the top-level dir (recursively
	// add watches, AND add watches for newly created dirs).
	if err := recursivelyWatchDir(dir); err != nil {
		watcher.Close()
		return nil, err
	}

	eventsCh := make(chan fsnotify.Event)
	errorsCh := make(chan error)

	go func() {
	loop:
		for {
			select {
			case e, ok := <-watcher.Events:
				if !ok {
					break loop
				}

				if ignorePath != nil {
					if ignore, err := ignorePath(e.Name); err != nil {
						errorsCh <- err
						continue
					} else if ignore {
						continue
					}
				}

				fi, err := os.Stat(e.Name)
				if os.IsNotExist(err) {
					// TODO(sqs): recursively remove watches
					if err := watcher.Remove(e.Name); err != nil {
						// TODO(sqs): suppress "can't remove non-existent inotify watch" error when we watch a dir and one of its files is deleted (we shouldn't remove the file's watcher (since we never actually watched the file, only the dir), but we dont know enough in that case to avoid calling watcher.Remove on the file).
						//
						// errorsCh <- err
					}
				} else if err != nil {
					errorsCh <- err
					continue
				}

				if fi != nil && fi.Mode().IsDir() {
					// Recursively watch newly added directories.
					if err := recursivelyWatchDir(e.Name); err != nil {
						errorsCh <- err
						continue
					}

					continue
				}

				// Only pass along events on files (not dirs), since
				// zap (and git itself) only deals with files.
				//
				// TODO(sqs): how to deal with symlinks? how to know if a deleted thing is a dir or file?
				isFile := fi != nil && fi.Mode().IsRegular()
				if isFile || e.Op&fsnotify.Remove != 0 {
					eventsCh <- e
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					break loop
				}
				errorsCh <- err
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
	return w.watcher.Close()
}
