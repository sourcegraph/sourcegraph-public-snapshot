package gitutil

import (
	"context"
	"strings"
	"sync"
	"time"
)

// RemoteRefWatcher periodically fetches a remote ref and watches for
// when it changes.
//
// Upon creation (with NewRemoteRefWatcher), ref changes are sent
// on the Update channel.
type RemoteRefWatcher struct {
	Update <-chan string // string of the ref's commit SHA (or empty string if the remote ref doesn't exist)
	Errors <-chan error

	cancel func()

	mu             sync.Mutex
	remote         string
	srcRef, dstRef string        // refspec
	value          string        // last known value of the ref (to detect when it changes)
	ready          chan struct{} // closed when SetRemoteRefspec is called the 1st time
}

func fetchRemoteRef(gitRepo BareRepo, remote, srcRef, dstRef string) (value string, err error) {
	remoteRefNotFound, err := gitRepo.Fetch(remote, srcRef+":"+dstRef)
	if remoteRefNotFound {
		// Not an error; it just means nobody else has WIP in this
		// workspace.
		err = nil
	}
	if err != nil {
		return "", err
	}
	if !remoteRefNotFound {
		value, err = gitRepo.ObjectNameSHA(dstRef + "^{commit}")
		if err != nil {
			return "", err
		}
	}
	return
}

// NewRemoteRefWatcher creates a new watcher for the git
// repository. Callers must call SetRemoteRefspec to begin watching
// (so it knows what to fetch), and Close when done (to free
// resources).
func NewRemoteRefWatcher(gitRepo BareRepo) *RemoteRefWatcher {
	ctx, cancel := context.WithCancel(context.Background())

	w := &RemoteRefWatcher{
		cancel: cancel,
		ready:  make(chan struct{}),
	}

	updateCh := make(chan string)
	errorsCh := make(chan error)
	w.Update = updateCh
	w.Errors = errorsCh

	getParams := func() (remote, srcRef, dstRef, value string) {
		w.mu.Lock()
		remote, srcRef, dstRef, value = w.remote, w.srcRef, w.dstRef, w.value
		w.mu.Unlock()
		return
	}

	ready := w.ready
	go func() {
		<-ready // wait until SetRemoteRefspec is called

		const (
			baseSleep = 100 * time.Millisecond
			maxSleep  = 15 * time.Second
		)
		sleep := time.Duration(0) // don't sleep the first time

	loop:
		for {
			select {
			case <-time.After(sleep):
				remote, srcRef, dstRef, lastValue := getParams()

				value, err := fetchRemoteRef(gitRepo, remote, srcRef, dstRef)

				// (Imperfectly) detect if the params changed (via SetRemoteRefspec)
				// while we were exec'ing, and throw away our results if so.
				if remote2, srcRef2, dstRef2, lastValue2 := getParams(); remote2 != remote || srcRef2 != srcRef || dstRef2 != dstRef || lastValue2 != lastValue {
					// Rerun immediately on updated remote refspec.
					sleep = 0
					continue
				}

				if err != nil {
					errorsCh <- err
				}
				if value != lastValue {
					updateCh <- value

					w.mu.Lock()
					w.value = value
					w.mu.Unlock()

					sleep = baseSleep
				} else {
					if sleep == 0 {
						sleep = baseSleep
					}
					sleep = sleep + sleep/10 // exponential backoff
					if sleep > maxSleep {
						sleep = maxSleep
					}
				}

			case <-ctx.Done():
				// Called by Close.
				close(updateCh)
				close(errorsCh)
				break loop
			}
		}
	}()

	return w
}

// SetRemoteRefspec updates the remote and refspec to fetch. It panics
// if refspec is an invalid refspec or if it contains '*'.
func (w *RemoteRefWatcher) SetRemoteRefspec(remote, refspec string) {
	if strings.Contains(refspec, "*") {
		panic("invalid refspec (no '*' allowed): " + refspec)
	}
	parts := strings.SplitN(refspec, ":", 2)
	if len(parts) != 2 {
		panic("invalid refspec (must be 'src:dst'): " + refspec)
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	w.remote = remote
	w.srcRef = parts[0]
	w.dstRef = parts[1]
	w.value = ""
	if w.ready != nil {
		close(w.ready)
		w.ready = nil
	}
}

// Close stops fetching/watching and closes w.Update and w.Errors.
func (w *RemoteRefWatcher) Close() error {
	w.cancel()
	return nil
}
