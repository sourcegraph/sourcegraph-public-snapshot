package server

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Server) dir(name api.RepoName) common.GitDir {
	p := string(protocol.NormalizeRepo(name))
	return common.GitDir(filepath.Join(s.ReposDir, filepath.FromSlash(p), ".git"))
}

func repoNameFromDir(reposDir string, dir common.GitDir) api.RepoName {
	// dir == ${s.ReposDir}/${name}/.git
	parent := filepath.Dir(string(dir))                   // remove suffix "/.git"
	name := strings.TrimPrefix(parent, reposDir)          // remove prefix "${s.ReposDir}"
	name = strings.Trim(name, string(filepath.Separator)) // remove /
	name = filepath.ToSlash(name)                         // filepath -> path
	return protocol.NormalizeRepo(api.RepoName(name))
}

func cloneStatus(cloned, cloning bool) types.CloneStatus {
	switch {
	case cloned:
		return types.CloneStatusCloned
	case cloning:
		return types.CloneStatusCloning
	}
	return types.CloneStatusNotCloned
}

func isAlwaysCloningTest(name api.RepoName) bool {
	return protocol.NormalizeRepo(name).Equal("github.com/sourcegraphtest/alwayscloningtest")
}

func isAlwaysCloningTestRemoteURL(remoteURL *vcs.URL) bool {
	return strings.EqualFold(remoteURL.Host, "github.com") &&
		strings.EqualFold(remoteURL.Path, "sourcegraphtest/alwayscloningtest")
}

// checkSpecArgSafety returns a non-nil err if spec begins with a "-", which could
// cause it to be interpreted as a git command line argument.
func checkSpecArgSafety(spec string) error {
	if strings.HasPrefix(spec, "-") {
		return errors.Errorf("invalid git revision spec %q (begins with '-')", spec)
	}
	return nil
}

// repoLastFetched returns the mtime of the repo's FETCH_HEAD, which is the date of the last successful `git remote
// update` or `git fetch` (even if nothing new was fetched). As a special case when the repo has been cloned but
// none of those other two operations have been run (and so FETCH_HEAD does not exist), it will return the mtime of HEAD.
//
// This breaks on file systems that do not record mtime and if Git ever changes this undocumented behavior.
var repoLastFetched = func(dir common.GitDir) (time.Time, error) {
	fi, err := os.Stat(dir.Path("FETCH_HEAD"))
	if os.IsNotExist(err) {
		fi, err = os.Stat(dir.Path("HEAD"))
	}
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

// repoLastChanged returns the mtime of the repo's sg_refhash, which is the
// cached timestamp of the most recent commit we could find in the tree. As a
// special case when sg_refhash is missing we return repoLastFetched(dir).
//
// This breaks on file systems that do not record mtime. This is a Sourcegraph
// extension to track last time a repo changed. The file is updated by
// setLastChanged via doBackgroundRepoUpdate.
//
// As a special case, tries both the directory given, and the .git subdirectory,
// because we're a bit inconsistent about which name to use.
var repoLastChanged = func(dir common.GitDir) (time.Time, error) {
	fi, err := os.Stat(dir.Path("sg_refhash"))
	if os.IsNotExist(err) {
		return repoLastFetched(dir)
	}
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

// writeCounter wraps an io.Writer and keeps track of bytes written.
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

// limitWriter is a io.Writer that writes to an W but discards after N bytes.
type limitWriter struct {
	W io.Writer // underling writer
	N int       // max bytes remaining
}

func (l *limitWriter) Write(p []byte) (int, error) {
	if l.N <= 0 {
		return len(p), nil
	}
	origLen := len(p)
	if len(p) > l.N {
		p = p[:l.N]
	}
	n, err := l.W.Write(p)
	l.N -= n
	if l.N <= 0 {
		// If we have written limit bytes, then we can include the discarded
		// part of p in the count.
		n = origLen
	}
	return n, err
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
	flusher http.Flusher
	closed  bool
	doFlush bool
}

var logUnflushableResponseWriterOnce sync.Once

// newFlushingResponseWriter creates a new flushing response writer. Callers
// must call Close to free the resources created by the writer.
//
// If w does not support flushing, it returns nil.
func newFlushingResponseWriter(logger log.Logger, w http.ResponseWriter) *flushingResponseWriter {
	// We panic if we don't implement the needed interfaces.
	flusher := hackilyGetHTTPFlusher(w)
	if flusher == nil {
		logUnflushableResponseWriterOnce.Do(func() {
			logger.Warn("unable to flush HTTP response bodies - Diff search performance and completeness will be affected",
				log.String("type", reflect.TypeOf(w).String()))
		})
		return nil
	}

	w.Header().Set("Transfer-Encoding", "chunked")

	f := &flushingResponseWriter{w: w, flusher: flusher}
	go f.periodicFlush()
	return f
}

// hackilyGetHTTPFlusher attempts to get an http.Flusher from w. It (hackily) handles the case where w is a
// nethttp.statusCodeTracker (which wraps http.ResponseWriter and does not implement http.Flusher). See
// https://github.com/opentracing-contrib/go-stdlib/pull/11#discussion_r164295773 and
// https://github.com/sourcegraph/sourcegraph/issues/9045.
//
// I (@sqs) wrote this hack instead of fixing it upstream immediately because seems to be some reluctance to merge
// a fix (because it'd make the http.ResponseWriter falsely appear to implement many interfaces that it doesn't
// actually implement, so it would break the correctness of Go type-assertion impl checks).
func hackilyGetHTTPFlusher(w http.ResponseWriter) http.Flusher {
	if f, ok := w.(http.Flusher); ok {
		return f
	}
	if reflect.TypeOf(w).String() == "*nethttp.statusCodeTracker" {
		v := reflect.ValueOf(w).Elem()
		if v.Kind() == reflect.Struct {
			if rwv := v.FieldByName("ResponseWriter"); rwv.IsValid() {
				f, ok := rwv.Interface().(http.Flusher)
				if ok {
					return f
				}
			}
		}
	}
	return nil
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

func (f *flushingResponseWriter) periodicFlush() {
	for {
		time.Sleep(100 * time.Millisecond)
		f.mu.Lock()
		if f.closed {
			f.mu.Unlock()
			break
		}
		if f.doFlush {
			f.flusher.Flush()
		}
		f.mu.Unlock()
	}
}

// Close signals to the flush goroutine to stop.
func (f *flushingResponseWriter) Close() {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
}

// mapToLoggerField translates a map to log context fields.
func mapToLoggerField(m map[string]any) []log.Field {
	LogFields := []log.Field{}

	for i, v := range m {

		LogFields = append(LogFields, log.String(i, fmt.Sprint(v)))
	}

	return LogFields
}

// bestEffortWalk is a filepath.WalkDir which ignores errors that can be passed
// to walkFn. This is a common pattern used in gitserver for best effort work.
//
// Note: We still respect errors returned by walkFn.
//
// filepath.Walk can return errors if we run into permission errors or a file
// disappears between readdir and the stat of the file. In either case this
// error can be ignored for best effort code.
func bestEffortWalk(root string, walkFn func(path string, entry fs.DirEntry) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		return walkFn(path, d)
	})
}

// hostnameMatch checks whether the hostname matches the given address.
// If we don't find an exact match, we look at the initial prefix.
func hostnameMatch(shardID, addr string) bool {
	if !strings.HasPrefix(addr, shardID) {
		return false
	}
	if addr == shardID {
		return true
	}
	// We know that shardID is shorter than addr so we can safely check the next
	// char
	next := addr[len(shardID)]
	return next == '.' || next == ':'
}
