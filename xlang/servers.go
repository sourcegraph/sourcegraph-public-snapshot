package xlang

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

// ServersByMode registers build/lang servers. It should only be
// accessed by other packages at init time.
//
// This is populated by the addServersFromEnv func.
var ServersByMode = map[string]func() (io.ReadWriteCloser, error){}

// connectToServer opens a connection to the server that is registered
// for the given mode (e.g., "go" or "typescript").
func connectToServer(ctx context.Context, mode string) (io.ReadWriteCloser, error) {
	if connect, ok := ServersByMode[mode]; ok {
		return connect()
	}
	return nil, fmt.Errorf("xlang server proxy: no server registered for mode %q", mode)
}

// RegisterServersFromEnv registers a lang/build server for each
// environment variable of the form `LANGSERVER_XYZ=addr-or-program`
// (where XYZ is the case-insensitive "mode", such as "go" or
// "typescript").
//
// addr-or-program can be any of:
//
//   tcp://addr:port (connect to TCP listener)
//   path/to/executable (exec subprocess and connect to its stdio)
func RegisterServersFromEnv() error {
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 || parts[1] == "" {
			continue
		}
		name, val := parts[0], parts[1]
		if prefix := "LANGSERVER_"; strings.HasPrefix(name, prefix) {
			mode := strings.ToLower(strings.TrimPrefix(name, prefix))
			if _, present := ServersByMode[mode]; present {
				return fmt.Errorf("invalid language server registration from env var %s: a server is already registered for the mode %q", name, mode)
			}
			switch {
			case strings.HasPrefix(val, "tcp://"):
				log15.Info("Registering language server listener.", "mode", mode, "listener", val)
				ServersByMode[mode] = func() (io.ReadWriteCloser, error) {
					return net.Dial("tcp", strings.TrimPrefix(val, "tcp://"))
				}
			case strings.Contains(val, ":"):
				return fmt.Errorf(`invalid language server URL %q (you probably mean "tcp://%s")`, val, val)
			default:
				log15.Info("Registering language server executable (will exec and manage subprocesses, communicating over stdin/stdout).", "mode", mode, "path", val)
				ServersByMode[mode] = func() (io.ReadWriteCloser, error) {
					cmd := exec.Command(val)
					cmd.Stderr = &prefixWriter{w: os.Stderr, prefix: filepath.Base(val) + ": "}
					in, err := cmd.StdinPipe()
					if err != nil {
						return nil, err
					}
					out, err := cmd.StdoutPipe()
					if err != nil {
						return nil, err
					}
					if err := cmd.Start(); err != nil {
						return nil, err
					}
					return readWriteCloser{out, in, cmd.Process.Kill}, nil
				}
			}
		}
	}
	if len(ServersByMode) == 0 {
		log15.Warn("No language servers registered.")
	}
	return nil
}

type readWriteCloser struct {
	rc             io.ReadCloser
	wc             io.WriteCloser
	otherCloseFunc func() error
}

func (rwc readWriteCloser) Read(p []byte) (int, error) {
	return rwc.rc.Read(p)
}

func (rwc readWriteCloser) Write(p []byte) (int, error) {
	return rwc.wc.Write(p)
}

func (rwc readWriteCloser) Close() error {
	if err := rwc.rc.Close(); err != nil {
		return err
	}
	if err := rwc.wc.Close(); err != nil {
		return err
	}
	if rwc.otherCloseFunc != nil {
		if err := rwc.otherCloseFunc(); err != nil {
			return err
		}
	}
	return nil
}

type prefixWriter struct {
	w      io.Writer
	prefix string
}

func (w *prefixWriter) Write(p []byte) (int, error) {
	lines := bytes.Split(p, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if _, err := fmt.Fprintf(w.w, "%s%s\n", w.prefix, line); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

// InMemoryPeerConns is a convenience helper that returns a pair of
// io.ReadWriteClosers that are each other's peer.
//
// It can be used, for example, to run an in-memory JSON-RPC handler
// that speaks to an in-memory client, without needin to open a Unix
// or TCP connection.
func InMemoryPeerConns() (io.ReadWriteCloser, io.ReadWriteCloser) {
	sr, cw := io.Pipe()
	cr, sw := io.Pipe()
	return readWriteNoopCloser{sr, sw}, readWriteNoopCloser{cr, cw}
}

type readWriteNoopCloser struct {
	io.Reader
	io.Writer
}

func (readWriteNoopCloser) Close() error { return nil }
