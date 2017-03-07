package gitutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// RefWatcher watches a git ref for changes to the object it refers
// to.
//
// The initial value of the ref is sent on the OID channel when the
// watcher is created. Whenever the ref changes, the new value is sent
// on the OID channel. If the ref does not exist, an empty string is
// sent.
type RefWatcher struct {
	OID    <-chan string // git object ID of what the ref points to
	Errors <-chan error

	closed chan struct{}

	mu      sync.Mutex
	ref     string
	lastOID string // last known value of the ref (to detect when it changes)
}

// NewRefWatcher creates a new watcher for the git repository at
// detects changes to what the named ref points to.
func NewRefWatcher(gitRepo interface {
	GitDir() string
}, ref string) (*RefWatcher, error) {
	refFile := filepath.Join(gitRepo.GitDir(), ref)

	readRef := func() (string, error) {
		data, err := ioutil.ReadFile(refFile)
		if os.IsNotExist(err) {
			// If the ref doesn't exist, return "". (But if the whole
			// repo was deleted, return an error.)
			if _, err := os.Stat(gitRepo.GitDir()); os.IsNotExist(err) {
				return "", err
			}
			return "", nil
		} else if err != nil {
			return "", err
		}
		data = bytes.TrimRight(data, "\r\n")
		if len(data) != 40 {
			return "", fmt.Errorf("invalid ref file %q: no 'ref: ' prefix", refFile)
		}
		return string(data), nil
	}
	oid, err := readRef()
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
	if err := watcher.Add(filepath.Dir(refFile)); err != nil {
		_ = watcher.Close()
		return nil, err
	}

	oidCh := make(chan string, 1) // buffer ref's initial value
	errorsCh := make(chan error)
	w := &RefWatcher{
		OID:     oidCh,
		Errors:  errorsCh,
		lastOID: oid,
		closed:  make(chan struct{}),
	}

	// Send the initial value on the channel.
	oidCh <- oid

	// Watch for changes.
	go func() {
	loop:
		for {
			select {
			case e, ok := <-watcher.Events:
				if !ok {
					break loop
				}

				if e.Name != refFile {
					continue
				}
				if e.Op&(fsnotify.Create|fsnotify.Write) == 0 {
					// TODO(sqs): how to handle deletion?
					continue
				}

				oid, err := readRef()
				if err != nil {
					errorsCh <- err
					continue
				}

				w.mu.Lock()
				lastOID := w.lastOID
				w.mu.Unlock()
				if oid == lastOID {
					continue
				}

				oidCh <- string(oid)

			case err, ok := <-watcher.Errors:
				if !ok {
					break loop
				}
				errorsCh <- err

			case <-w.closed:
				if err := watcher.Close(); err != nil {
					errorsCh <- err
				}
				close(oidCh)
				close(errorsCh)
				break loop
			}
		}
	}()

	return w, nil
}

// Close stops watching and closes w.OID and w.Errors.
func (w *RefWatcher) Close() error {
	if w.closed != nil {
		close(w.closed)
	}
	return nil
}
