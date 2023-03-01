package run

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/rjeczalik/notify"
	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/lib/process"
)

type BazelCommand struct {
	Name        string
	Description string            `yaml:"description"`
	Target      string            `yaml:"target"`
	Args        string            `yaml:"args"`
	Env         map[string]string `yaml:"env"`
	Static      bool              `yaml:"static"`
}

type IBazel struct {
	pwd     string
	targets []string
	cancel  func()
}

// binLocation returns the path on disk where Bazel is putting the binary
// associated with a given target.
func binLocation(target string) (string, error) {
	// Get the output directory from Bazel, which varies depending on which OS
	// we're running against.
	cmd := exec.Command("bazel", "info", "output_path")
	b, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// Trim "bazel-out" because the next bazel query will include it.
	outputPath := strings.TrimSuffix(strings.TrimSpace(string(b)), "bazel-out")

	// Get the binary from the specific target.
	cmd = exec.Command("bazel", "cquery", target, "--output=files")
	b, err = cmd.Output()
	if err != nil {
		return "", err
	}
	binPath := strings.TrimSpace(string(b))

	return fmt.Sprintf("%s%s", outputPath, binPath), nil
}

// newIBazel returns a runner to interact with ibazel.
func newIBazel(pwd string, targets ...string) *IBazel {
	return &IBazel{
		pwd:     pwd,
		targets: targets,
	}
}

func (ib *IBazel) Start(ctx context.Context, dir string) error {
	args := append([]string{"build"}, ib.targets...)
	ctx, ib.cancel = context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, "ibazel", args...)

	sc := &startedCmd{
		stdoutBuf: &prefixSuffixSaver{N: 32 << 10},
		stderrBuf: &prefixSuffixSaver{N: 32 << 10},
	}

	sc.cancel = ib.cancel
	sc.Cmd = cmd
	sc.Cmd.Dir = dir

	var stdoutWriter, stderrWriter io.Writer
	logger := newCmdLogger(ctx, "iBazel", std.Out.Output)
	stdoutWriter = io.MultiWriter(logger, sc.stdoutBuf)
	stderrWriter = io.MultiWriter(logger, sc.stderrBuf)
	eg, err := process.PipeOutputUnbuffered(ctx, sc.Cmd, stdoutWriter, stderrWriter)
	if err != nil {
		return err
	}
	sc.outEg = eg

	return sc.Start()
}

func (ib *IBazel) Stop() error {
	ib.cancel()
	return nil
}

func BazelCommands(ctx context.Context, parentEnv map[string]string, verbose bool, cmds ...BazelCommand) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	var targets []string
	for _, cmd := range cmds {
		targets = append(targets, cmd.Target)
	}
	ibazel := newIBazel(repoRoot, targets...)

	p := pool.New().WithContext(ctx).WithCancelOnError()
	p.Go(func(ctx context.Context) error {
		return ibazel.Start(ctx, repoRoot)
	})

	for _, bc := range cmds {
		bc := bc
		p.Go(func(ctx context.Context) error {
			return bc.Start(ctx, repoRoot, parentEnv)
		})
	}

	return p.Wait()
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
	if bc.Static {
		return nil
	}

	std.Out.WriteLine(output.Styledf(output.StylePending, "Running %s...", bc.Name))

	// Run the binary for the first time.
	cancel, err := bc.start(ctx, dir, parentEnv)
	if err != nil {
		return err
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
	sc.Cmd = exec.CommandContext(commandCtx, "bash", "-c", binLocation)
	sc.Cmd.Dir = dir

	// secretsEnv, err := getSecrets(ctx, cmd)
	// if err != nil {
	// 	std.Out.WriteLine(output.Styledf(output.StyleWarning, "[%s] %s %s",
	// 		cmd.Name, output.EmojiFailure, err.Error()))
	// }
	//
	sc.Cmd.Env = makeEnv(parentEnv, bc.Env) // TODO secrets env

	var stdoutWriter, stderrWriter io.Writer
	logger := newCmdLogger(commandCtx, bc.Name, std.Out.Output)
	stdoutWriter = io.MultiWriter(logger, sc.stdoutBuf)
	stderrWriter = io.MultiWriter(logger, sc.stderrBuf)

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
