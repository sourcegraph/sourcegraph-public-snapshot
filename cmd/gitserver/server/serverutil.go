package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/util"
)

// runWithRemoteOpts runs the command after applying the remote options.
func (s *Server) runWithRemoteOpts(cmd *exec.Cmd, repoURI string) ([]byte, error) {
	cmd.Env = append(cmd.Env, "GIT_ASKPASS=true") // disable password prompt

	// Add github creds if we have them configured. This should never run for
	// Sourcegraph.com, but does run on our dogfood server.
	if s.GithubAccessToken != "" && strings.HasPrefix(repoURI, "github.com/") {
		gitPassHelperDir, err := makeGitPassHelper("x-oauth-token", s.GithubAccessToken)
		if err != nil {
			return nil, err
		}
		if gitPassHelperDir != "" {
			defer os.RemoveAll(gitPassHelperDir)
		}
		cmd.Args = append(cmd.Args[:1], append([]string{"-c", "credential.helper=gitserver-helper"}, cmd.Args[1:]...)...)
		env := environ(os.Environ())
		env.Set("PATH", gitPassHelperDir+string(filepath.ListSeparator)+os.Getenv("PATH"))
		cmd.Env = env
	} else {
		// Suppress asking to add SSH host key to known_hosts (which will hang because
		// the command is non-interactive).
		cmd.Env = append(cmd.Env, "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=yes")

		cmd.Args = append(cmd.Args[:1], append([]string{"-c", "credential.helper="}, cmd.Args[1:]...)...)
	}

	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	err, _ := runCommand(cmd)
	return b.Bytes(), err
}

// makeGitSSHWrapper writes a GIT_SSH wrapper that runs ssh with the
// private key. You should remove the sshWrapper, sshWrapperDir and
// the keyFile after using them.
func (s *Server) makeGitSSHWrapper(privKey []byte) (sshWrapper, sshWrapperDir, keyFile string, err error) {
	var otherOpt string
	if s.InsecureSkipCheckVerifySSH {
		otherOpt = "-o StrictHostKeyChecking=no"
	}

	kf, err := ioutil.TempFile("", "go-vcs-gitcmd-key")
	if err != nil {
		return "", "", "", err
	}
	keyFile = kf.Name()
	err = util.WriteFileWithPermissions(keyFile, privKey, 0600)
	if err != nil {
		return "", "", keyFile, err
	}

	tmpFile, tmpFileDir, err := gitSSHWrapper(keyFile, otherOpt)
	return tmpFile, tmpFileDir, keyFile, err
}

// gitSSHWrapper makes system-dependent SSH wrapper.
func gitSSHWrapper(keyFile string, otherOpt string) (sshWrapperFile string, tempDir string, err error) {
	// TODO(sqs): encrypt and store the key in the env so that
	// attackers can't decrypt if they have disk access after our
	// process dies

	var script string

	if runtime.GOOS == "windows" {
		script = `
	@echo off
	ssh -o ControlMaster=no -o ControlPath=none ` + otherOpt + ` -i ` + filepath.ToSlash(keyFile) + ` "%@"
`
	} else {
		script = `
	#!/bin/sh
	exec /usr/bin/ssh -o ControlMaster=no -o ControlPath=none ` + otherOpt + ` -i ` + filepath.ToSlash(keyFile) + ` "$@"
`
	}

	sshWrapperName, tempDir, err := util.ScriptFile("go-vcs-gitcmd")
	if err != nil {
		return sshWrapperName, tempDir, err
	}

	err = util.WriteFileWithPermissions(sshWrapperName, []byte(script), 0500)
	return sshWrapperName, tempDir, err
}

// makeGitPassHelper writes a git credential helper that supplies username and password over stdout.
// Its name is "git-credential-gitserver-helper" and it's located inside gitPassHelperDir.
// If err is nil, the caller is responsible for removing gitPassHelperDir after it's done using it.
func makeGitPassHelper(user, pass string) (gitPassHelperDir string, err error) {
	tempDir, err := ioutil.TempDir("", "gitserver_")
	if err != nil {
		return "", err
	}

	// Write the credentials content to credentialsFile file.
	// This is done to avoid code injection attacks.
	// Usernames and passwords are untrusted arbitrary user data. It's hard to escape
	// strings in shell scripts, so we opt to `cat` this non-executable credentials file instead.
	credentialsFile := filepath.Join(tempDir, "credentials-content")
	{
		// Always provide username and password via git credential helper.
		// Do this even if some of the values are blank strings.
		// Otherwise, git will fallback to asking via other means.
		content := fmt.Sprintf("username=%s\npassword=%s\n", user, pass)

		err := util.WriteFileWithPermissions(credentialsFile, []byte(content), 0600)
		if err != nil {
			os.RemoveAll(tempDir)
			return "", err
		}
	}

	// Write the credential helper executable that uses credentialsFile.
	{
		// We assume credentialsFile can be escaped with a simple wrapping of single
		// quotes. The path is not user controlled so this assumption should
		// not be violated.
		content := fmt.Sprintf("#!/bin/sh\ncat '%s'\n", credentialsFile)

		path := filepath.Join(tempDir, "git-credential-gitserver-helper")
		err := util.WriteFileWithPermissions(path, []byte(content), 0500)
		if err != nil {
			os.RemoveAll(tempDir)
			return "", err
		}
	}

	return tempDir, nil
}

// repoCloned checks if dir is a valid GIT_DIR.
var repoCloned = func(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "HEAD"))
	return !os.IsNotExist(err)
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

// newFlushingResponseWriter creates a new flushing response writer. Callers
// must call Close to free the resources created by the writer.
func newFlushingResponseWriter(w http.ResponseWriter) *flushingResponseWriter {
	// We panic if we don't implement the needed interfaces.
	flusher := w.(http.Flusher)

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
