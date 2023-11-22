package run

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/rjeczalik/notify"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/lib/process"
)

// A BazelCommand is a command definition for sg run/start that uses
// bazel under the hood. It will handle restarting itself autonomously,
// as long as iBazel is running and watch that specific target.
type BazelCommand struct {
	Name            string
	Description     string                            `yaml:"description"`
	Target          string                            `yaml:"target"`
	Args            string                            `yaml:"args"`
	PreCmd          string                            `yaml:"precmd"`
	Env             map[string]string                 `yaml:"env"`
	IgnoreStdout    bool                              `yaml:"ignoreStdout"`
	IgnoreStderr    bool                              `yaml:"ignoreStderr"`
	ExternalSecrets map[string]secrets.ExternalSecret `yaml:"external_secrets"`
}

func (bc *BazelCommand) BinLocation() (string, error) {
	return binLocation(bc.Target)
}

func (bc *BazelCommand) watch(ctx context.Context) (<-chan struct{}, error) {
	// Grab the location of the binary in bazel-out.
	binLocation, err := bc.BinLocation()
	if err != nil {
		return nil, err
	}

	// Set up the watcher.
	restart := make(chan struct{})
	events := make(chan notify.EventInfo, 1)
	if err := notify.Watch(binLocation, events, notify.All); err != nil {
		return nil, err
	}

	// Start watching for a freshly compiled version of the binary.
	go func() {
		defer close(events)
		defer notify.Stop(events)

		for {
			select {
			case <-ctx.Done():
				return
			case e := <-events:
				if e.Event() != notify.Remove {
					restart <- struct{}{}
				}
			}

		}
	}()

	return restart, nil
}

func (bc *BazelCommand) Start(ctx context.Context, dir string, parentEnv map[string]string) error {
	std.Out.WriteLine(output.Styledf(output.StylePending, "Running %s...", bc.Name))

	// Run the binary for the first time.
	cancel, err := bc.start(ctx, dir, parentEnv)
	if err != nil {
		return errors.Wrapf(err, "failed to start Bazel command %q", bc.Name)
	}

	// Restart when the binary change.
	wantRestart, err := bc.watch(ctx)
	if err != nil {
		return err
	}

	// Wait forever until we're asked to stop or that restarting returns an error.
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-wantRestart:
			std.Out.WriteLine(output.Styledf(output.StylePending, "Restarting %s...", bc.Name))
			cancel()
			cancel, err = bc.start(ctx, dir, parentEnv)
			if err != nil {
				return err
			}
		}
	}
}

func (bc *BazelCommand) start(ctx context.Context, dir string, parentEnv map[string]string) (func(), error) {
	binLocation, err := bc.BinLocation()
	if err != nil {
		return nil, err
	}

	sc := &startedCmd{
		stdoutBuf: &prefixSuffixSaver{N: 32 << 10},
		stderrBuf: &prefixSuffixSaver{N: 32 << 10},
	}

	commandCtx, cancel := context.WithCancel(ctx)
	sc.cancel = cancel
	sc.Cmd = exec.CommandContext(commandCtx, "bash", "-c", fmt.Sprintf("%s\n%s", bc.PreCmd, binLocation))
	sc.Cmd.Dir = dir

	secretsEnv, err := getSecrets(ctx, bc.Name, bc.ExternalSecrets)
	if err != nil {
		std.Out.WriteLine(output.Styledf(output.StyleWarning, "[%s] %s %s",
			bc.Name, output.EmojiFailure, err.Error()))
	}

	sc.Cmd.Env = makeEnv(parentEnv, secretsEnv, bc.Env)

	var stdoutWriter, stderrWriter io.Writer
	logger := newCmdLogger(commandCtx, bc.Name, std.Out.Output)
	if bc.IgnoreStdout {
		std.Out.WriteLine(output.Styledf(output.StyleSuggestion, "Ignoring stdout of %s", bc.Name))
		stdoutWriter = sc.stdoutBuf
	} else {
		stdoutWriter = io.MultiWriter(logger, sc.stdoutBuf)
	}
	if bc.IgnoreStderr {
		std.Out.WriteLine(output.Styledf(output.StyleSuggestion, "Ignoring stderr of %s", bc.Name))
		stderrWriter = sc.stderrBuf
	} else {
		stderrWriter = io.MultiWriter(logger, sc.stderrBuf)
	}

	eg, err := process.PipeOutputUnbuffered(ctx, sc.Cmd, stdoutWriter, stderrWriter)
	if err != nil {
		return nil, err
	}
	sc.outEg = eg

	if err := sc.Start(); err != nil {
		return nil, err
	}

	return cancel, nil
}
