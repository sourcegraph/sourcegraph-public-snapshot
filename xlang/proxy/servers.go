package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/pkg/conf"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

var (
	// serversByModeMu protects ServersByMode after init time.
	serversByModeMu sync.RWMutex

	// ServersByMode registers build/lang servers. It should only be
	// accessed by other packages at init time.
	//
	// This is populated by the addServersFromEnv func.
	ServersByMode map[string]func() (jsonrpc2.ObjectStream, error)
)

// connectToServer opens a connection to the server that is registered
// for the given mode (e.g., "go" or "typescript").
func connectToServer(ctx context.Context, mode string) (jsonrpc2.ObjectStream, error) {
	serversByModeMu.RLock()
	connect, ok := ServersByMode[mode]
	serversByModeMu.RUnlock()

	if ok {
		return connect()
	}
	return nil, &jsonrpc2.Error{
		Code:    CodeModeNotFound,
		Message: fmt.Sprintf("xlang server proxy: no server registered for mode %q", mode),
	}
}

func RegisterServers() {
	conf.Watch(func() {
		serversByModeMu.Lock()
		defer serversByModeMu.Unlock()

		ServersByMode = make(map[string]func() (jsonrpc2.ObjectStream, error))

		err := registerServersFromEnv()
		if err != nil {
			log15.Error("error registering language servers from env", "error", err)
		}
		err = registerServersFromConfig()
		if err != nil {
			log15.Error("error registering language servers from config", "error", err)
		}
		if len(ServersByMode) == 0 {
			log15.Info("No language servers registered")
		}
	})
}

func registerServersFromConfig() error {
	for _, l := range conf.EnabledLangservers() {
		if l.Address != "" {
			err := registerTCPServer(l.Language, l.Address, "config")
			if err != nil {
				return err
			}
		} else {
			log15.Debug("missing address in langserver config (it must be set by env LANGSERVER_XYZ)", "lang", l.Language)
		}
	}
	return nil
}

// registerServersFromEnv registers a lang/build server for each
// environment variable of the form `LANGSERVER_XYZ=addr-or-program`
// (where XYZ is the case-insensitive "mode", such as "go" or
// "typescript").
//
// addr-or-program can be any of:
//
//   tcp://addr:port (connect to TCP listener)
//   path/to/executable (exec subprocess and connect to its stdio)
func registerServersFromEnv() error {
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 || parts[1] == "" {
			continue
		}
		name, val := parts[0], parts[1]
		if prefix := "LANGSERVER_"; strings.HasPrefix(name, prefix) && !strings.HasSuffix(name, "_ARGS_JSON") {
			mode := strings.ToLower(strings.TrimPrefix(name, prefix))
			if strings.HasPrefix(val, "tcp://") {
				err := registerTCPServer(mode, val, "env")
				if err != nil {
					return err
				}
			} else if strings.HasPrefix(val, ":") {
				return fmt.Errorf(`invalid language server URL %q (you probably mean "tcp://%s")`, val, val)
			} else {
				// Allow specifying extra command-line args to
				// language server executables in
				// LANGSERVER_name_ARGS_JSON env vars.
				var args []string
				if v := os.Getenv(name + "_ARGS_JSON"); v != "" {
					if err := json.Unmarshal([]byte(v), &args); err != nil {
						return fmt.Errorf("%s_ARGS_JSON: %s", name, err)
					}
				}

				log15.Info("Registering language server executable", "mode", mode, "path", val)
				ServersByMode[mode] = execServer(val, args)
			}
		}
	}
	return nil
}

func registerTCPServer(mode, addr, scope string) error {
	if _, present := ServersByMode[mode]; present {
		// TODO(john): at the moment it's possible to have services networked in a way that diverges
		// from the site configuration (by setting `LANGSERVER_XYZ=foo' and also `"languages": [{
		// "language": "xyz", "address": "bar" }]` in the site config).  We ignore subsequent
		// registrations and currently prefer env variable addresses.
		log15.Debug("A language server is already registered, skipping...", "mode", mode, "scope", scope)
		return nil
	}
	log15.Info("Registering language server listener", "mode", mode, "listener", addr)
	ServersByMode[mode] = tcpServer(strings.TrimPrefix(addr, "tcp://"))
	return nil
}

const connectTimeout = 10 * time.Second

func tcpServer(addr string) func() (jsonrpc2.ObjectStream, error) {
	return func() (jsonrpc2.ObjectStream, error) {
		conn, err := net.DialTimeout("tcp", addr, connectTimeout)
		if err != nil {
			return nil, err
		}
		return jsonrpc2.NewBufferedStream(conn, jsonrpc2.VSCodeObjectCodec{}), nil
	}
}

func execServer(name string, args []string) func() (jsonrpc2.ObjectStream, error) {
	return func() (jsonrpc2.ObjectStream, error) {
		cmd := exec.Command(name, args...)
		cmd.Stderr = &prefixWriter{w: os.Stderr, prefix: filepath.Base(name) + ": "}
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
		return jsonrpc2.NewBufferedStream(readWriteCloser{out, in, cmd.Process.Kill}, jsonrpc2.VSCodeObjectCodec{}), nil
	}
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
func InMemoryPeerConns() (jsonrpc2.ObjectStream, jsonrpc2.ObjectStream) {
	sr, cw := io.Pipe()
	cr, sw := io.Pipe()
	return jsonrpc2.NewBufferedStream(&pipeReadWriteCloser{sr, sw}, jsonrpc2.VSCodeObjectCodec{}),
		jsonrpc2.NewBufferedStream(&pipeReadWriteCloser{cr, cw}, jsonrpc2.VSCodeObjectCodec{})
}

type pipeReadWriteCloser struct {
	*io.PipeReader
	*io.PipeWriter
}

func (c *pipeReadWriteCloser) Close() error {
	err1 := c.PipeReader.Close()
	err2 := c.PipeWriter.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
