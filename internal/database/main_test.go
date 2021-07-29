package database

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

func init() {
	dbtesting.DBNameSuffix = "internal-database"
}

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}

	// An ephemeral DB. Targetting this test package first as an experiment,
	// will make more general so all tests benefit from it.
	//
	// 2021-07-29(keegan) - testing on my machine this pkg goes from 78s to
	// 4s.
	dsn, cleanup, err := startEphemeralDB()
	if err != nil && testing.Verbose() {
		log.Println("failed to setup ephemeral db:", err)
	}
	if dsn != "" {
		// Hack, this is the first envvar we read for dbtest.NewDB.
		os.Setenv("PGDATASOURCE", dsn)
	}

	code := m.Run()

	if cleanup != nil {
		cleanup()
	}

	os.Exit(code)
}

func startEphemeralDB() (dsn string, _ func(), _ error) {
	// Like defer, but runs in the returned cleanup function.
	var cleanup deferList

	// If we fail, run any cleanups we may have.
	defer func() {
		if dsn == "" {
			cleanup.Run()
		}
	}()

	// allows us to adjust the command before running. Use case is for logging
	// and to sudo on CI.
	cmdPreRun := func(c *exec.Cmd) {
		if !testing.Verbose() {
			return
		}
		if c.Stdout == nil {
			c.Stdout = os.Stdout
		}
		if c.Stderr == nil {
			c.Stderr = os.Stderr
		}
	}

	// HACK: CI doesn't have postgres on the PATH. Hardcode
	// it in until we update our agents.
	if os.Getenv("CI") != "" {
		os.Setenv("PATH", os.Getenv("PATH")+":/usr/lib/postgresql/12/bin")
	}

	// This only works if postgres is on PATH.
	if _, err := exec.LookPath("postgres"); err != nil {
		return "", nil, err
	}

	// Create an environment without outside PG*
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "PG") {
			env = append(env, e)
		}
	}

	// Store postgres data in a temp dir
	pgHost, err := os.MkdirTemp("", filepath.Base(os.Args[0])+"-pghost-*")
	if err != nil {
		return "", nil, err
	}
	cleanup.Add(func() { os.RemoveAll(pgHost) })
	pgData := filepath.Join(pgHost, "data")
	env = append(env,
		"PGHOST="+pgHost,
		"PGDATA="+pgData,
		"PGDATABASE=postgres",
	)

	// CI runs as root but postgres can't :O
	if os.Getenv("CI") != "" && syscall.Getuid() == 0 {
		sudo, err := exec.LookPath("sudo")
		if err != nil {
			return "", nil, err
		}

		// need the tmp dir to be owned by postgres user. Easier to use
		// exec than looking up uid.
		if err := exec.Command("chown", "postgres", pgHost).Run(); err != nil {
			return "", nil, err
		}

		// need to wrap commands so they run as postgres user
		old := cmdPreRun
		cmdPreRun = func(c *exec.Cmd) {
			c.Args = append([]string{"sudo", "-u", "postgres", "-E", c.Path}, c.Args[1:]...)
			c.Path = sudo
			old(c)
		}
	}

	// Create the database without auth and without fsync.
	cmd := exec.Command("initdb", pgData, "--nosync", "-E", "UNICODE", "--auth=trust")
	cmd.Env = env
	cmdPreRun(cmd)
	if err := cmd.Run(); err != nil {
		return "", nil, err
	}

	// Tune configuration for speed with ephemeral test data.
	if err := appendFile(filepath.Join(pgData, "postgresql.conf"), []byte(`
unix_socket_directories = '`+pgHost+`'
listen_addresses = ''
shared_buffers = 12MB
fsync = off
synchronous_commit = off
full_page_writes = off
log_min_duration_statement = 0
`), 0644); err != nil {
		return "", nil, err
	}

	// start postgres as a subprocess. If the test process dies hopefully the
	// OS will come in and cleanup. We parse stderr looking for the log
	// message saying it is ready.
	cmd = exec.Command("postgres", "-D", pgData)
	cmd.Env = env
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", nil, err
	}
	cleanup.Add(func() { stderr.Close() })

	// Postgres logs with test output \o/
	var r io.Reader = stderr
	if testing.Verbose() {
		r = io.TeeReader(r, os.Stderr)
	}

	cmdPreRun(cmd)
	if err := cmd.Start(); err != nil {
		return "", nil, err
	}

	// We use ready to signal that postgres has started.
	ready := make(chan bool, 1)
	cleanup.Add(func() {
		_ = cmd.Process.Signal(os.Interrupt)
		_ = cmd.Wait()
		// make sure scanner goroutine has stopped
		<-ready
	})

	go func() {
		defer close(ready)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if bytes.Contains(scanner.Bytes(), []byte("listening on Unix socket")) {
				ready <- true
				break
			}
		}
		io.Copy(io.Discard, r) // drain r
	}()

	ok := <-ready
	if !ok {
		return "", nil, errors.New("failed to find ready message in log output")
	}

	// This DSN points to the directory postgres will put unix sockets.
	return "postgresql:///postgres?host=" + pgHost, cleanup.Run, nil
}

type deferList []func()

func (l *deferList) Add(f func()) {
	*l = append(*l, f)
}

func (l *deferList) Run() {
	for i := len(*l) - 1; i >= 0; i-- {
		(*l)[i]()
	}
}

// WriteFile but with os.O_APPEND
func appendFile(name string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}
