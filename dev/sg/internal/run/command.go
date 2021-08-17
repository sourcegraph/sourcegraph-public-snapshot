package run

import (
	"context"
	"io"
	"os/exec"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type Command struct {
	Name                string
	Cmd                 string            `yaml:"cmd"`
	Install             string            `yaml:"install"`
	CheckBinary         string            `yaml:"checkBinary"`
	Env                 map[string]string `yaml:"env"`
	Watch               []string          `yaml:"watch"`
	IgnoreStdout        bool              `yaml:"ignoreStdout"`
	IgnoreStderr        bool              `yaml:"ignoreStderr"`
	DefaultArgs         string            `yaml:"defaultArgs"`
	ContinueWatchOnExit bool              `yaml:"continueWatchOnExit"`

	// ATTENTION: If you add a new field here, be sure to also handle that
	// field in `Merge` (below).
}

func (c Command) Merge(other Command) Command {
	merged := c

	if other.Name != merged.Name && other.Name != "" {
		merged.Name = other.Name
	}
	if other.Cmd != merged.Cmd && other.Cmd != "" {
		merged.Cmd = other.Cmd
	}
	if other.Install != merged.Install && other.Install != "" {
		merged.Install = other.Install
	}
	if other.IgnoreStdout != merged.IgnoreStdout && !merged.IgnoreStdout {
		merged.IgnoreStdout = other.IgnoreStdout
	}
	if other.IgnoreStderr != merged.IgnoreStderr && !merged.IgnoreStderr {
		merged.IgnoreStderr = other.IgnoreStderr
	}
	if other.DefaultArgs != merged.DefaultArgs && other.DefaultArgs != "" {
		merged.DefaultArgs = other.DefaultArgs
	}
	merged.ContinueWatchOnExit = other.ContinueWatchOnExit || merged.ContinueWatchOnExit

	for k, v := range other.Env {
		merged.Env[k] = v
	}

	if !equal(merged.Watch, other.Watch) {
		merged.Watch = other.Watch
	}

	return merged
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

type startedCmd struct {
	*exec.Cmd

	cancel func()

	stdoutBuf *prefixSuffixSaver
	stderrBuf *prefixSuffixSaver
}

func (sc *startedCmd) CapturedStdout() string {
	if sc.stdoutBuf == nil {
		return ""
	}

	return string(sc.stdoutBuf.Bytes())
}

func (sc *startedCmd) CapturedStderr() string {
	if sc.stderrBuf == nil {
		return ""
	}

	return string(sc.stderrBuf.Bytes())
}

func startCmd(ctx context.Context, dir string, cmd Command, globalEnv map[string]string) (*startedCmd, error) {
	sc := &startedCmd{
		stdoutBuf: &prefixSuffixSaver{N: 32 << 10},
		stderrBuf: &prefixSuffixSaver{N: 32 << 10},
	}

	commandCtx, cancel := context.WithCancel(ctx)
	sc.cancel = cancel

	sc.Cmd = exec.CommandContext(commandCtx, "bash", "-c", cmd.Cmd)
	sc.Cmd.Dir = dir
	sc.Cmd.Env = makeEnv(globalEnv, cmd.Env)

	logger := newCmdLogger(cmd.Name, stdout.Out)
	if cmd.IgnoreStdout {
		stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, "Ignoring stdout of %s", cmd.Name))
		sc.Cmd.Stdout = sc.stdoutBuf
	} else {
		sc.Cmd.Stdout = io.MultiWriter(logger, sc.stdoutBuf)
	}
	if cmd.IgnoreStderr {
		stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, "Ignoring stderr of %s", cmd.Name))
		sc.Cmd.Stderr = sc.stderrBuf
	} else {
		sc.Cmd.Stderr = io.MultiWriter(logger, sc.stderrBuf)
	}

	if err := sc.Start(); err != nil {
		return sc, err
	}

	return sc, nil
}
