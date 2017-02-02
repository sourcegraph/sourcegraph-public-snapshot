package gitutil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// SymbolicRefWatcher watches a git repository for changes to a
// symbolic ref (i.e., .git/HEAD).
//
// The initial value of the symbolic ref (e.g., "refs/heads/mybranch")
// is sent on the Ref channel when the watcher is created. Whenever
// the symbolic ref changes, the new value is sent on the Ref channel.
type SymbolicRefWatcher struct {
	Ref    <-chan string
	Errors <-chan error

	watcher io.Closer
}

// NewSymbolicRefWatcher creates a watcher in the git repository at
// for the named symbolic ref. Callers must call Close when done to
// free resources.
func NewSymbolicRefWatcher(gitRepo interface {
	GitDir() string
}, name string) (*SymbolicRefWatcher, error) {
	symRefFile := filepath.Join(gitRepo.GitDir(), name)

	readRef := func() (string, error) {
		data, err := ioutil.ReadFile(symRefFile)
		if err != nil {
			return "", err
		}
		if !bytes.HasPrefix(data, []byte("ref: ")) {
			return "", fmt.Errorf("invalid symbolic ref file %q: no 'ref: ' prefix", symRefFile)
		}
		ref := string(bytes.TrimSpace(bytes.TrimPrefix(data, []byte("ref: "))))
		return ref, nil
	}
	ref, err := readRef()
	if err != nil {
		return nil, err
	}

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

	refCh := make(chan string, 1) // buffer ref's initial value
	errorsCh := make(chan error)

	refCh <- ref

	// Watch for changes.
	go func() {
	loop:
		for {
			select {
			case e, ok := <-watcher.Events:
				if !ok {
					break loop
				}
				if e.Name != symRefFile {
					continue
				}
				if e.Op&(fsnotify.Create|fsnotify.Write) == 0 {
					// TODO(sqs): how to handle deletion?
					continue
				}
				ref, err := readRef()
				if err == nil {
					refCh <- ref
				} else {
					errorsCh <- err
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					break loop
				}
				errorsCh <- err
			}
		}
		close(refCh)
		close(errorsCh)
	}()

	return &SymbolicRefWatcher{
		Ref:     refCh,
		Errors:  errorsCh,
		watcher: watcher,
	}, nil
}

// Close stops watching and closes w.Ref and w.Errors.
func (w *SymbolicRefWatcher) Close() error {
	return w.watcher.Close()
}
