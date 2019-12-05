package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// GitDir is an absolute path to a GIT_DIR.
// They will all follow the form:
//
//    ${s.ReposDir}/${name}/.git
type GitDir string

// Path is a helper which returns filepath.Join(dir, elem...)
func (dir GitDir) Path(elem ...string) string {
	return filepath.Join(append([]string{string(dir)}, elem...)...)
}

func (s *Server) dir(name api.RepoName) GitDir {
	path := string(protocol.NormalizeRepo(name))
	return GitDir(filepath.Join(s.ReposDir, filepath.FromSlash(path), ".git"))
}

func (s *Server) name(dir GitDir) api.RepoName {
	// dir == ${s.ReposDir}/${name}/.git
	parent := filepath.Dir(string(dir))                   // remove suffix "/.git"
	name := strings.TrimPrefix(parent, s.ReposDir)        // remove prefix "${s.ReposDir}"
	name = strings.Trim(name, string(filepath.Separator)) // remove /
	name = filepath.ToSlash(name)                         // filepath -> path
	return protocol.NormalizeRepo(api.RepoName(name))
}

func isAlwaysCloningTest(name api.RepoName) bool {
	return protocol.NormalizeRepo(name) == "github.com/sourcegraphtest/alwayscloningtest"
}

// checkSpecArgSafety returns a non-nil err if spec begins with a "-", which could
// cause it to be interpreted as a git command line argument.
func checkSpecArgSafety(spec string) error {
	if strings.HasPrefix(spec, "-") {
		return errors.Errorf("invalid git revision spec %q (begins with '-')", spec)
	}
	return nil
}

// runWithRemoteOpts runs the command after applying the remote options.
// If progress is not nil, all output is written to it in a separate goroutine.
func runWithRemoteOpts(ctx context.Context, cmd *exec.Cmd, progress io.Writer) ([]byte, error) {
	configureGitCommand(cmd)

	var b interface {
		Bytes() []byte
	}

	if progress != nil {
		var pw progressWriter
		r, w := io.Pipe()
		defer w.Close()
		mr := io.MultiWriter(&pw, w)
		cmd.Stdout = mr
		cmd.Stderr = mr
		go func() {
			if _, err := io.Copy(progress, r); err != nil {
				log15.Error("error while copying progress", "error", err)
			}
		}()
		b = &pw
	} else {
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		b = &buf
	}

	_, err := runCommand(ctx, cmd)
	return b.Bytes(), err
}

func configureGitCommand(cmd *exec.Cmd) {
	if cmd.Args[0] != "git" {
		panic("Only git commands are supported")
	}

	cmd.Env = append(cmd.Env, "GIT_ASKPASS=true") // disable password prompt

	// Suppress asking to add SSH host key to known_hosts (which will hang because
	// the command is non-interactive).
	//
	// And set a timeout to avoid indefinite hangs if the server is unreachable.
	cmd.Env = append(cmd.Env, "GIT_SSH_COMMAND=ssh -o BatchMode=yes -o ConnectTimeout=30")

	extraArgs := []string{
		// Unset credential helper because the command is non-interactive.
		"-c", "credential.helper=",
	}

	if len(cmd.Args) > 1 && cmd.Args[1] != "ls-remote" {
		// Use Git protocol version 2 for all commands except for ls-remote because it actually decreases the performance of ls-remote.
		// https://opensource.googleblog.com/2018/05/introducing-git-protocol-version-2.html
		extraArgs = append(extraArgs, "-c", "protocol.version=2")
	}

	cmd.Args = append(cmd.Args[:1], append(extraArgs, cmd.Args[1:]...)...)
}

// repoCloned checks if dir or `${dir}/.git` is a valid GIT_DIR.
var repoCloned = func(dir GitDir) bool {
	_, err := os.Stat(dir.Path("HEAD"))
	return !os.IsNotExist(err)
}

// repoLastFetched returns the mtime of the repo's FETCH_HEAD, which is the date of the last successful `git remote
// update` or `git fetch` (even if nothing new was fetched). As a special case when the repo has been cloned but
// none of those other two operations have been run (and so FETCH_HEAD does not exist), it will return the mtime of HEAD.
//
// This breaks on file systems that do not record mtime and if Git ever changes this undocumented behavior.
var repoLastFetched = func(dir GitDir) (time.Time, error) {
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
// setLastChanged via doRepoUpdate2.
//
// As a special case, tries both the directory given, and the .git subdirectory,
// because we're a bit inconsistent about which name to use.
var repoLastChanged = func(dir GitDir) (time.Time, error) {
	fi, err := os.Stat(dir.Path("sg_refhash"))
	if os.IsNotExist(err) {
		return repoLastFetched(dir)
	}
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

// repoRemoteURL returns the "origin" remote fetch URL for the Git repository in dir. If the repository
// doesn't exist or the remote doesn't exist and have a fetch URL, an error is returned. If there are
// multiple fetch URLs, only the first is returned.
var repoRemoteURL = func(ctx context.Context, dir GitDir) (string, error) {
	// We do not pass in context since this command is quick. We do a lot of
	// logging around what this function does, so having to handle context
	// failures in each case is verbose. Rather just prevent that.
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = string(dir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	_, err := runCommand(ctx, cmd)
	if err != nil {
		stderr := stderr.Bytes()
		if len(stderr) > 200 {
			stderr = stderr[:200]
		}
		return "", fmt.Errorf("git %s failed: %s (%q)", cmd.Args, err, stderr)
	}
	remoteURLs := strings.SplitN(strings.TrimSpace(stdout.String()), "\n", 2)
	if len(remoteURLs) == 0 || remoteURLs[0] == "" {
		return "", fmt.Errorf("no remote URL for repo %s", dir)
	}
	return remoteURLs[0], nil
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
	flusher http.Flusher
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
	flusher := hackilyGetHTTPFlusher(w)
	if flusher == nil {
		logUnflushableResponseWriterOnce.Do(func() {
			log15.Warn("Unable to flush HTTP response bodies. Diff search performance and completeness will be affected.", "type", reflect.TypeOf(w).String())
		})
		return nil
	}

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

// progressWriter is an io.Writer that writes to a buffer.
// '\r' resets the write offset to the index after last '\n' in the buffer,
// or the beginning of the buffer if a '\n' has not been written yet.
type progressWriter struct {
	// writeOffset is the offset in buf where the next write should begin.
	writeOffset int

	// afterLastNewline is the index after the last '\n' in buf
	// or 0 if there is no '\n' in buf.
	afterLastNewline int

	buf []byte
}

func (w *progressWriter) Write(p []byte) (n int, err error) {
	l := len(p)
	for {
		if len(p) == 0 {
			// If p ends in a '\r' we still want to include that in the buffer until it is overwritten.
			break
		}
		idx := bytes.IndexAny(p, "\r\n")
		if idx == -1 {
			w.buf = append(w.buf[:w.writeOffset], p...)
			w.writeOffset = len(w.buf)
			break
		}
		switch p[idx] {
		case '\n':
			w.buf = append(w.buf[:w.writeOffset], p[:idx+1]...)
			w.writeOffset = len(w.buf)
			w.afterLastNewline = len(w.buf)
			p = p[idx+1:]
		case '\r':
			w.buf = append(w.buf[:w.writeOffset], p[:idx+1]...)
			// Record that our next write should overwrite the data after the most recent newline.
			// Don't slice it off immediately here, because we want to be able to return that output
			// until it is overwritten.
			w.writeOffset = w.afterLastNewline
			p = p[idx+1:]
		default:
			panic(fmt.Sprintf("unexpected char %q", p[idx]))
		}
	}
	return l, nil
}

// String returns the contents of the buffer as a string.
func (w *progressWriter) String() string {
	return string(w.buf)
}

// Bytes returns the contents of the buffer.
func (w *progressWriter) Bytes() []byte {
	return w.buf
}

// mapToLog15Ctx translates a map to log15 context fields.
func mapToLog15Ctx(m map[string]interface{}) []interface{} {
	// sort so its stable
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	ctx := make([]interface{}, len(m)*2)
	for i, k := range keys {
		j := i * 2
		ctx[j] = k
		ctx[j+1] = m[k]
	}
	return ctx
}

// updateFileIfDifferent will atomically update the file if the contents are
// different. If it does an update ok is true.
func updateFileIfDifferent(path string, content []byte) (bool, error) {
	current, err := ioutil.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		// If the file doesn't exist we write a new file.
		return false, err
	}

	if bytes.Equal(current, content) {
		return false, nil
	}

	// We write to a tempfile first to do the atomic update (via rename)
	f, err := ioutil.TempFile(filepath.Dir(path), filepath.Base(path))
	if err != nil {
		return false, err
	}
	// We always remove the tempfile. In the happy case it won't exist.
	defer os.Remove(f.Name())

	if n, err := f.Write(content); err != nil {
		f.Close()
		return false, err
	} else if n != len(content) {
		f.Close()
		return false, io.ErrShortWrite
	}

	// fsync to ensure the disk contents are written. This is important, since
	// we are not gaurenteed that os.Rename is recorded to disk after f's
	// contents.
	if err := f.Sync(); err != nil {
		f.Close()
		return false, err
	}
	if err := f.Close(); err != nil {
		return false, err
	}
	return true, renameAndSync(f.Name(), path)
}

// renameAndSync will do an os.Rename followed by fsync to ensure the rename
// is recorded
func renameAndSync(oldpath, newpath string) error {
	err := os.Rename(oldpath, newpath)
	if err != nil {
		return err
	}

	oldparent, newparent := filepath.Dir(oldpath), filepath.Dir(newpath)
	err = fsync(newparent)
	if oldparent != newparent {
		if err1 := fsync(oldparent); err == nil {
			err = err1
		}
	}
	return err
}

func fsync(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

// bestEffortWalk is a filepath.Walk which ignores errors that can be passed
// to walkFn. This is a common pattern used in gitserver for best effort work.
//
// Note: We still respect errors returned by walkFn.
//
// filepath.Walk can return errors if we run into permission errors or a file
// disappears between readdir and the stat of the file. In either case this
// error can be ignored for best effort code.
func bestEffortWalk(root string, walkFn func(path string, info os.FileInfo) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		return walkFn(path, info)
	})
}
