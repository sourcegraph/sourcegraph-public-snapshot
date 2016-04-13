package jsserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/sysreq"

	"golang.org/x/net/context"
)

func init() {
	sysreq.AddCheck("Node.js", func(ctx context.Context) (problem, fix string, err error) {
		if _, err = exec.LookPath("node"); err != nil {
			problem = "Could not find Node.js 'node' program in your PATH (which is required to render React components on the server, for improved performance)."
			fix = "Install Node.js and ensure the 'node' interpreter is in your PATH."
			return
		}
		return "", "", nil
	})
}

// Server represents an interface to a JavaScript function.
type Server interface {
	Call(ctx context.Context, arg json.RawMessage) (json.RawMessage, error)
	Close() error
}

// New creates a new background node process that runs the given
// JavaScript code.
func New(js []byte) (Server, error) {
	tmpFile, err := ioutil.TempFile("", "jsserver")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(js); err != nil {
		tmpFile.Close()
		return nil, err
	}
	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	var errBuf bytes.Buffer

	cmd := exec.Command("node", tmpFile.Name())
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "JSSERVER=1")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	s := &server{
		c:   cmd,
		in:  in,
		out: out,
		err: &errBuf,
	}

	resp, err := s.recv()
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(resp, []byte(`"ready"`)) {
		return nil, fmt.Errorf(`jsserver: expected "ready" response, got %s`, resp)
	}

	return s, nil
}

type server struct {
	mu  sync.Mutex
	c   *exec.Cmd
	in  io.WriteCloser
	out io.ReadCloser
	err *bytes.Buffer

	closed bool
}

func (s *server) recv() (json.RawMessage, error) {
	var resp json.RawMessage
	dec := json.NewDecoder(s.out)
	if err := dec.Decode(&resp); err != nil {
		if other, _ := ioutil.ReadAll(dec.Buffered()); len(other) > 0 {
			err = fmt.Errorf("%s (remaining in buffer: %q)", err, other)
		}
		return nil, fmt.Errorf("jsserver: recv: %s", err)
	}
	return resp, nil
}

// Call calls the node process with the given argument.
func (s *server) Call(ctx context.Context, arg json.RawMessage) (json.RawMessage, error) {
	s.mu.Lock()
	err := json.NewEncoder(s.in).Encode(&arg)
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}

	var resp json.RawMessage
	done := make(chan struct{})
	go func() {
		s.mu.Lock()
		resp, err = s.recv()
		s.mu.Unlock()
		close(done)
	}()
	select {
	case <-done:
		return resp, err

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close kills the node process. After calling Close, the behavior of
// calling Call is undefined.
func (s *server) Close() error {
	s.closed = true
	if err := s.c.Process.Kill(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.c.Wait()
	if _, ok := err.(*exec.ExitError); ok {
		return nil
	}
	return err
}
