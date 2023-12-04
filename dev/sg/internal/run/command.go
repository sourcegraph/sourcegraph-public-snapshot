package run

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"

	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/lib/process"
)

type Command struct {
	Name                string
	Cmd                 string            `yaml:"cmd"`
	Install             string            `yaml:"install"`
	InstallFunc         string            `yaml:"install_func"`
	CheckBinary         string            `yaml:"checkBinary"`
	Env                 map[string]string `yaml:"env"`
	Watch               []string          `yaml:"watch"`
	IgnoreStdout        bool              `yaml:"ignoreStdout"`
	IgnoreStderr        bool              `yaml:"ignoreStderr"`
	DefaultArgs         string            `yaml:"defaultArgs"`
	ContinueWatchOnExit bool              `yaml:"continueWatchOnExit"`
	// Preamble is a short and visible message, displayed when the command is launched.
	Preamble string `yaml:"preamble"`

	ExternalSecrets map[string]secrets.ExternalSecret `yaml:"external_secrets"`
	Description     string                            `yaml:"description"`

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
	if other.InstallFunc != merged.InstallFunc && other.InstallFunc != "" {
		merged.InstallFunc = other.InstallFunc
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
	if other.Preamble != merged.Preamble && other.Preamble != "" {
		merged.Preamble = other.Preamble
	}
	if other.Description != merged.Description && other.Description != "" {
		merged.Description = other.Description
	}
	merged.ContinueWatchOnExit = other.ContinueWatchOnExit || merged.ContinueWatchOnExit

	for k, v := range other.Env {
		if merged.Env == nil {
			merged.Env = make(map[string]string)
		}
		merged.Env[k] = v
	}

	for k, v := range other.ExternalSecrets {
		if merged.ExternalSecrets == nil {
			merged.ExternalSecrets = make(map[string]secrets.ExternalSecret)
		}
		merged.ExternalSecrets[k] = v
	}

	if !equal(merged.Watch, other.Watch) && len(other.Watch) != 0 {
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

	outEg *pool.ErrorPool
}

func (sc *startedCmd) Wait() error {
	if err := sc.outEg.Wait(); err != nil {
		return err
	}
	return sc.Cmd.Wait()
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

func getSecrets(ctx context.Context, name string, extSecrets map[string]secrets.ExternalSecret) (map[string]string, error) {
	secretsEnv := map[string]string{}

	if len(extSecrets) == 0 {
		return secretsEnv, nil
	}

	secretsStore, err := secrets.FromContext(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to get secrets store: %v", err)
	}

	var errs error
	for envName, secret := range extSecrets {
		secretsEnv[envName], err = secretsStore.GetExternal(ctx, secret)
		if err != nil {
			errs = errors.Append(errs,
				errors.Wrapf(err, "failed to access secret %q for command %q", envName, name))
		}
	}
	return secretsEnv, errs
}

var sgConn net.Conn

func OpenUnixSocket() error {
	var err error
	sgConn, err = net.Dial("unix", "/tmp/sg.sock")
	return err
}

func startCmd(ctx context.Context, dir string, cmd Command, parentEnv map[string]string) (*startedCmd, error) {
	sc := &startedCmd{
		stdoutBuf: &prefixSuffixSaver{N: 32 << 10},
		stderrBuf: &prefixSuffixSaver{N: 32 << 10},
	}

	commandCtx, cancel := context.WithCancel(ctx)
	sc.cancel = cancel

	sc.Cmd = exec.CommandContext(commandCtx, "bash", "-c", cmd.Cmd)
	sc.Cmd.Dir = dir

	secretsEnv, err := getSecrets(ctx, cmd.Name, cmd.ExternalSecrets)
	if err != nil {
		std.Out.WriteLine(output.Styledf(output.StyleWarning, "[%s] %s %s",
			cmd.Name, output.EmojiFailure, err.Error()))
	}

	sc.Cmd.Env = makeEnv(parentEnv, secretsEnv, cmd.Env)

	var stdoutWriter, stderrWriter io.Writer
	logger := newCmdLogger(commandCtx, cmd.Name, std.Out.Output)

	// TODO(JH) sgtail experiment going on, this is a bit ugly, that will do it
	// for the demo day.
	if sgConn != nil {
		sink := func(data string) {
			sgConn.Write([]byte(fmt.Sprintf("%s: %s\n", cmd.Name, data)))
		}
		sgConnLog := process.NewLogger(ctx, sink)

		if cmd.IgnoreStdout {
			std.Out.WriteLine(output.Styledf(output.StyleSuggestion, "Ignoring stdout of %s", cmd.Name))
			stdoutWriter = sc.stdoutBuf
		} else {
			stdoutWriter = io.MultiWriter(logger, sc.stdoutBuf, sgConnLog)
		}
		if cmd.IgnoreStderr {
			std.Out.WriteLine(output.Styledf(output.StyleSuggestion, "Ignoring stderr of %s", cmd.Name))
			stderrWriter = sc.stderrBuf
		} else {
			stderrWriter = io.MultiWriter(logger, sc.stderrBuf, sgConnLog)
		}
	} else {
		if cmd.IgnoreStdout {
			std.Out.WriteLine(output.Styledf(output.StyleSuggestion, "Ignoring stdout of %s", cmd.Name))
			stdoutWriter = sc.stdoutBuf
		} else {
			stdoutWriter = io.MultiWriter(logger, sc.stdoutBuf)
		}
		if cmd.IgnoreStderr {
			std.Out.WriteLine(output.Styledf(output.StyleSuggestion, "Ignoring stderr of %s", cmd.Name))
			stderrWriter = sc.stderrBuf
		} else {
			stderrWriter = io.MultiWriter(logger, sc.stderrBuf)
		}
	}

	if cmd.Preamble != "" {
		std.Out.WriteLine(output.Styledf(output.StyleOrange, "[%s] %s %s", cmd.Name, output.EmojiInfo, cmd.Preamble))
	}
	eg, err := process.PipeOutputUnbuffered(ctx, sc.Cmd, stdoutWriter, stderrWriter)
	if err != nil {
		return nil, err
	}
	sc.outEg = eg

	if err := sc.Start(); err != nil {
		return sc, err
	}

	return sc, nil
}
