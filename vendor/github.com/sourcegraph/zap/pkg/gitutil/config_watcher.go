package gitutil

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConfigWatcher watches a git repository for changes to its
// configuration (in the "config" file for bare repos or ".git/config"
// for non-bare repos).
//
// Whenever the config changes, the new value is sent on the Config
// channel.
type ConfigWatcher struct {
	Config <-chan []byte
	Errors <-chan error

	watcher io.Closer
}

// NewConfigWatcher creates a watcher in the git repository at for the
// repository's configuration. Callers must call Close when done to
// free resources.
func NewConfigWatcher(gitRepo interface {
	GitDir() string
}) (*ConfigWatcher, error) {
	configFile := filepath.Join(gitRepo.GitDir(), "config")

	// Watching the whole dir, not just the single file, is more
	// portable (otherwise we sometimes only see remove/chmod events
	// for the file on Linux).
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := watcher.Add(gitRepo.GitDir()); err != nil {
		_ = watcher.Close()
		return nil, err
	}

	configCh := make(chan []byte)
	errorsCh := make(chan error)

	readFile := func() ([]byte, error) {
		// When the config file is in the middle of being
		// written to disk, sometimes reading it results in an
		// empty byte array. Give it some time to finish
		// writing to disk.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		for {
			data, err := ioutil.ReadFile(configFile)
			if len(data) > 0 || err != nil {
				return data, err
			}

			if ctx.Err() != nil {
				// Timed out; the file is probably intentionally empty.
				return data, err
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Watch for changes.
	go func() {
		var lastValue []byte
	loop:
		for {
			select {
			case e, ok := <-watcher.Events:
				if !ok {
					break loop
				}
				if e.Name != configFile {
					continue
				}
				if e.Op&(fsnotify.Create|fsnotify.Write) == 0 {
					// TODO(sqs): how to handle deletion?
					continue
				}
				data, err := readFile()
				if err == nil {
					if !bytes.Equal(data, lastValue) {
						configCh <- data
						lastValue = data
					}
				} else {
					errorsCh <- err
					lastValue = nil
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					break loop
				}
				errorsCh <- err
			}
		}
		close(configCh)
		close(errorsCh)
	}()

	return &ConfigWatcher{
		Config:  configCh,
		Errors:  errorsCh,
		watcher: watcher,
	}, nil
}

// Close stops watching and closes w.Ref and w.Errors.
func (w *ConfigWatcher) Close() error {
	return w.watcher.Close()
}
