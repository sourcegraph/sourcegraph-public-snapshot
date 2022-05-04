package run

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"golang.org/x/sync/errgroup"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/lib/process"
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
	Secrets             map[string]Secret `yaml:"secrets"`
	// Preamble is a short and visible message, displayed when the command is launched.
	Preamble string `yaml:"preamble"`

	// ATTENTION: If you add a new field here, be sure to also handle that
	// field in `Merge` (below).
}

type Secret struct {
	Provider string `yaml:"provider"`
	Project  string `yaml:"project"`
	Name     string `yaml:"name"`
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
	if other.Preamble != merged.Preamble && other.Preamble != "" {
		merged.Preamble = other.Preamble
	}
	merged.ContinueWatchOnExit = other.ContinueWatchOnExit || merged.ContinueWatchOnExit

	for k, v := range other.Env {
		merged.Env[k] = v
	}

	for k, v := range other.Secrets {
		merged.Secrets[k] = v
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

	outEg *errgroup.Group
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

func getSecrets(ctx context.Context, cmd Command) (map[string]string, error) {
	secretsEnv := map[string]string{}

	if len(cmd.Secrets) == 0 {
		return secretsEnv, nil
	}

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to create secretmanager client: %v", err)
	}
	for envName, secret := range cmd.Secrets {
		if secret.Provider != "gcloud" {
			errors.Newf("Unknown secrets provider %s", secret.Provider)
		}
		path := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", secret.Project, secret.Name)
		req := &secretmanagerpb.AccessSecretVersionRequest{
			Name: path,
		}
		result, err := client.AccessSecretVersion(ctx, req)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to access secret %s from %s", secret.Name, secret.Project)
		}
		secretsEnv[envName] = string(result.Payload.Data)
	}
	return secretsEnv, nil
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

	secretsEnv, err := getSecrets(ctx, cmd)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot fetch secrets")
	}

	sc.Cmd.Env = makeEnv(parentEnv, secretsEnv, cmd.Env)

	var stdoutWriter, stderrWriter io.Writer
	logger := newCmdLogger(commandCtx, cmd.Name, stdout.Out)
	if cmd.IgnoreStdout {
		stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, "Ignoring stdout of %s", cmd.Name))
		stdoutWriter = sc.stdoutBuf
	} else {
		stdoutWriter = io.MultiWriter(logger, sc.stdoutBuf)
	}
	if cmd.IgnoreStderr {
		stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, "Ignoring stderr of %s", cmd.Name))
		stderrWriter = sc.stderrBuf
	} else {
		stderrWriter = io.MultiWriter(logger, sc.stderrBuf)
	}

	if cmd.Preamble != "" {
		stdout.Out.WriteLine(output.Linef("", output.StyleOrange, "[%s] %s %s", cmd.Name, output.EmojiInfo, cmd.Preamble))
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
