// Package goreman implements a process supervisor for a Procfile.
package goreman

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// -- process information structure.
type procInfo struct {
	proc    string
	cmdline string
	stopped bool // true if we stopped it
	cmd     *exec.Cmd
	mu      sync.Mutex
	cond    *sync.Cond
	waitErr error
}

// process informations named with proc.
var procs map[string]*procInfo
var procM sync.Mutex

var maxProcNameLength int

// read Procfile and parse it.
func readProcfile(content []byte) (newProcs []string) {
	procM.Lock()
	defer procM.Unlock()

	if len(procs) == 0 {
		procs = map[string]*procInfo{}
	}

	re := lazyregexp.New(`\$([a-zA-Z]+[a-zA-Z0-9_]+)`)
	for _, line := range strings.Split(string(content), "\n") {
		tokens := strings.SplitN(line, ":", 2)
		if len(tokens) != 2 || tokens[0][0] == '#' {
			continue
		}
		k, v := strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1])
		if runtime.GOOS == "windows" {
			v = re.ReplaceAllStringFunc(v, func(s string) string {
				return "%" + s[1:] + "%"
			})
		}
		p := &procInfo{proc: k, cmdline: v}
		p.cond = sync.NewCond(&p.mu)
		procs[k] = p
		newProcs = append(newProcs, k)
		if len(k) > maxProcNameLength {
			maxProcNameLength = len(k)
		}
	}
	return newProcs
}

// ProcDiedAction specifies the behaviour Goreman takes if a process exits
// with a non-zero exit code.
type ProcDiedAction uint

const (
	// Shutdown will shutdown Goreman if any process shuts down with a
	// non-zero exit code.
	Shutdown ProcDiedAction = iota

	// Ignore will continue running Goreman and will leave not restart the
	// dead process.
	Ignore
)

// procDiedAction is the ProcDiedAction to take. Goreman still is globals
// everywhere \o/
var procDiedAction ProcDiedAction

type Options struct {
	// RPCAddr is the address to listen for Goreman RPCs.
	RPCAddr string

	// ProcDiedAction specifies the behaviour to take when a process dies.
	ProcDiedAction ProcDiedAction
}

var startOnce sync.Once

// Start starts up the Procfile.
func Start(contents []byte, opts Options) error {
	var err error
	startOnce.Do(func() {
		if opts.ProcDiedAction > Ignore {
			err = errors.Errorf("invalid ProcDiedAction %v", opts.ProcDiedAction)
			return
		}
		procDiedAction = opts.ProcDiedAction
		if opts.RPCAddr != "" {
			if err = os.Setenv("GOREMAN_RPC_ADDR", opts.RPCAddr); err != nil {
				return
			}

			if err = startServer(opts.RPCAddr); err != nil {
				return
			}
		}
	})
	if err != nil {
		return err
	}

	newProcs := readProcfile(contents)
	if len(newProcs) == 0 {
		return errors.New("No valid entry")
	}

	for _, proc := range newProcs {
		_ = startProc(proc)
	}

	return waitProcs()
}
