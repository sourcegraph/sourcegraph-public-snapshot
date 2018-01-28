package server

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

// runWithRemoteOpts runs the command after applying the remote options.
func (s *Server) runWithRemoteOpts(ctx context.Context, cmd *exec.Cmd) ([]byte, error) {
	cmd.Env = append(cmd.Env, "GIT_ASKPASS=true") // disable password prompt

	// Suppress asking to add SSH host key to known_hosts (which will hang because
	// the command is non-interactive).
	//
	// And set a timeout to avoid indefinite hangs if the server is unreachable.
	cmd.Env = append(cmd.Env, "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=yes -o ConnectTimeout=7")

	// Unset credential helper because the command is non-interactive.
	cmd.Args = append(cmd.Args[:1], append([]string{"-c", "credential.helper="}, cmd.Args[1:]...)...)

	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	err, _ := runCommand(ctx, cmd)
	return b.Bytes(), err
}

// repoCloned checks if dir or `${dir}/.git` is a valid GIT_DIR.
var repoCloned = func(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "HEAD")); !os.IsNotExist(err) {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, ".git", "HEAD")); !os.IsNotExist(err) {
		return true
	}
	return false
}

// repoLastFetched returns the mtime of the repo's FETCH_HEAD, which is the date of the last successful `git remote
// update` or `git fetch` (even if nothing new was fetched). As a special case when the repo has been cloned but
// none of those other two operations have been run (and so FETCH_HEAD does not exist), it will return the mtime of HEAD.
//
// This breaks on file systems that do not record mtime and if Git ever changes this undocumented behavior.
var repoLastFetched = func(dir string) (time.Time, error) {
	fi, err := os.Stat(filepath.Join(dir, "FETCH_HEAD"))
	if os.IsNotExist(err) {
		fi, err = os.Stat(filepath.Join(dir, ".git", "FETCH_HEAD"))
	}
	if os.IsNotExist(err) {
		fi, err = os.Stat(filepath.Join(dir, "HEAD"))
	}
	if os.IsNotExist(err) {
		fi, err = os.Stat(filepath.Join(dir, ".git", "HEAD"))
	}
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

// environ is a slice of strings representing the environment, in the form "key=value".
type environ []string

// Set environment variable key to value.
func (e *environ) Set(key, value string) {
	for i := range *e {
		if strings.HasPrefix((*e)[i], key+"=") {
			(*e)[i] = key + "=" + value
			return
		}
	}
	// If we get here, it's because the key isn't already present, so add a new one.
	*e = append(*e, key+"="+value)
}

// Unset environment variable key.
func (e *environ) Unset(key string) {
	for i := range *e {
		if strings.HasPrefix((*e)[i], key+"=") {
			(*e)[i] = (*e)[len(*e)-1]
			*e = (*e)[:len(*e)-1]
			return
		}
	}
}

// writeCounter wraps an io.WriterCloser and keeps track of bytes written.
type writeCounter struct {
	w io.Writer
	// n is the number of bytes written to w
	n int64
}

func (c *writeCounter) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	c.n += int64(n)
	return
}

// flushingResponseWriter is a http.ResponseWriter that flushes all writes
// to the underlying connection within a certain time period after Write is
// called (instead of buffering them indefinitely).
//
// This lets, e.g., clients with a context deadline see as much partial response
// body as possible.
type flushingResponseWriter struct {
	// mu ensures we don't concurrently call Flush and Write. It also protects
	// state.
	mu      sync.Mutex
	w       http.ResponseWriter
	closed  bool
	doFlush bool
}

var logUnflushableResponseWriterOnce sync.Once

// newFlushingResponseWriter creates a new flushing response writer. Callers
// must call Close to free the resources created by the writer.
//
// If w does not support flushing, it returns nil.
func newFlushingResponseWriter(w http.ResponseWriter) *flushingResponseWriter {
	// We panic if we don't implement the needed interfaces.
	flusher, ok := w.(http.Flusher)
	if !ok {
		logUnflushableResponseWriterOnce.Do(func() {
			log15.Warn("Unable to flush HTTP response bodies. Diff search performance and completeness will be affected.", "type", reflect.TypeOf(w).String())
		})
		return nil
	}

	f := &flushingResponseWriter{w: w}
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			f.mu.Lock()
			if f.closed {
				f.mu.Unlock()
				break
			}
			if f.doFlush {
				flusher.Flush()
			}
			f.mu.Unlock()
		}
	}()
	return f
}

// Header implements http.ResponseWriter.
func (f *flushingResponseWriter) Header() http.Header { return f.w.Header() }

// WriteHeader implements http.ResponseWriter.
func (f *flushingResponseWriter) WriteHeader(code int) { f.w.WriteHeader(code) }

// Write implements http.ResponseWriter.
func (f *flushingResponseWriter) Write(p []byte) (int, error) {
	f.mu.Lock()
	n, err := f.w.Write(p)
	if n > 0 {
		f.doFlush = true
	}
	f.mu.Unlock()
	return n, err
}

// Close signals to the flush goroutine to stop.
func (f *flushingResponseWriter) Close() {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
}
