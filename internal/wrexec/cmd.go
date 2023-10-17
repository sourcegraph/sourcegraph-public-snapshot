package wrexec

import (
	"context"
	"io"
	"os/exec"

	"github.com/sourcegraph/log"
)

const maxRecordedOutput = 5000

// Cmd wraps an os/exec.Cmd into a thin layer than enables one to set hooks for before and after the commands.
type Cmd struct {
	*exec.Cmd
	ctx         context.Context
	logger      log.Logger
	beforeHooks []BeforeHook
	afterHooks  []AfterHook
	output      []byte
}

// BeforeHook are called before the execution of a command. Returning an error within a before
// hook prevents subsequent hooks and the command to be executed; all "running" commands such as Start, Run, Wait
// and others will return that error.
//
// The passed context is the one that was used to create the Cmd.
type BeforeHook func(context.Context, log.Logger, *exec.Cmd) error

// AfterHook are called once the execution of the command is completed, from a Go perspective. It means if
// a command were to be started with Start but Wait was never called, the after hook would never be called.
//
// The passed context is the one that was used to create the Cmd.
type AfterHook func(context.Context, log.Logger, *exec.Cmd)

// Cmder provides an interface modeled after os/exec.Cmd that enables one to operate a level higher and to
// pass around various implementations, such as RecordingCommand, without having the receiver know about it.
//
// The only new method is Unwrap() which allows one to grab the underlying os/exec.Cmd if needed.
type Cmder interface {
	CombinedOutput() ([]byte, error)
	Environ() []string
	Output() ([]byte, error)
	Run() error
	Start() error
	StderrPipe() (io.ReadCloser, error)
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.ReadCloser, error)
	String() string
	Wait() error

	Unwrap() *exec.Cmd
}

var _ Cmder = &Cmd{}

// CommandContext constructs a new Cmd wrapper with the provided context.
// If logger is nil, a no-op logger will be set.
func CommandContext(ctx context.Context, logger log.Logger, name string, args ...string) *Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	return Wrap(ctx, logger, cmd)
}

// Wrap constructs a new Cmd based of an existing os/Exec.cmd command.
// If logger is nil, a no-op logger will be set.
func Wrap(ctx context.Context, logger log.Logger, cmd *exec.Cmd) *Cmd {
	if logger == nil {
		logger = log.NoOp()
	}
	return &Cmd{
		Cmd:    cmd,
		ctx:    ctx,
		logger: logger,
	}
}

// SetBeforeHooks installs hooks that will be called just before the underlying command
// is executed.
//
// If a hook returns an error, all error returning functions from the Cmder interface
// will return that error and no subsequent hooks will be called.
func (c *Cmd) SetBeforeHooks(hooks ...BeforeHook) {
	c.beforeHooks = hooks
}

// SetAfterHooks installs hooks that will be called once the underlying command completes,
// from a Go point of view. In practice, it means that even if the underlying command exits,
// the after hooks won't be called until Wait or any other methods that waits upon completion
// are called.
func (c *Cmd) SetAfterHooks(hooks ...AfterHook) {
	c.afterHooks = hooks
}

// Unwrap returns the underlying os/exec.Cmd, that can be safely modified unless
// the Cmd has been started.
func (c *Cmd) Unwrap() *exec.Cmd {
	return c.Cmd
}

func (c *Cmd) runBeforeHooks() error {
	for _, h := range c.beforeHooks {
		if err := h(c.ctx, c.logger, c.Cmd); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cmd) runAfterHooks() {
	for _, h := range c.afterHooks {
		h(c.ctx, c.logger, c.Cmd)
	}
}

// CombinedOutput calls os/exec.Cmd.CombinedOutput after running the before hooks,
// and run the after hooks once it returns.
func (c *Cmd) CombinedOutput() ([]byte, error) {
	if err := c.runBeforeHooks(); err != nil {
		return nil, err
	}
	defer c.runAfterHooks()
	output, err := c.Cmd.CombinedOutput()
	c.setExecutionOutput(output)
	return output, err
}

// This is used to save the output of the command execution for later retrieval.
// We truncate the output at 5000 characters so we don't store too much data.
func (c *Cmd) setExecutionOutput(output []byte) {
	c.output = make([]byte, 0, maxRecordedOutput)
	if len(c.output) > maxRecordedOutput {
		c.output = output[:maxRecordedOutput]
	} else {
		c.output = output
	}
}

// This is a workaround created to retrieve the output of an executed command without
// calling `.Output()` or `.CombinedOutput()` again.
func (c *Cmd) GetExecutionOutput() string {
	return string(c.output)
}

// Environ returns the underlying command environ. It never call the hooks.
func (c *Cmd) Environ() []string {
	return c.Cmd.Environ()
}

// Output calls os/exec.Cmd.Output after running the before hooks,
// and run the after hooks once it returns.
func (c *Cmd) Output() ([]byte, error) {
	if err := c.runBeforeHooks(); err != nil {
		return nil, err
	}
	defer c.runAfterHooks()
	output, err := c.Cmd.Output()
	c.setExecutionOutput(output)
	return output, err
}

// Run calls os/exec.Cmd.Run after running the before hooks,
// and run the after hooks once it returns.
func (c *Cmd) Run() error {
	if err := c.runBeforeHooks(); err != nil {
		return err
	}
	defer c.runAfterHooks()
	return c.Cmd.Run()
}

// Start calls os/exec.Cmd.Start after running the before hooks,
// but do not run the after hooks, because the command may not
// have exited yet. Wait must be used to make sure the after hooks
// are executed.
func (c *Cmd) Start() error {
	if err := c.runBeforeHooks(); err != nil {
		return err
	}
	return c.Cmd.Start()
}

// StderrPipe calls os/exec.Cmd.StderrPipe, without running any hooks.
func (c *Cmd) StderrPipe() (io.ReadCloser, error) {
	return c.Cmd.StderrPipe()
}

// StdinPipe calls os/exec.Cmd.StderrPipe, without running any hooks.
func (c *Cmd) StdinPipe() (io.WriteCloser, error) {
	return c.Cmd.StdinPipe()
}

// StdoutPipe calls os/exec.Cmd.StderrPipe, without running any hooks.
func (c *Cmd) StdoutPipe() (io.ReadCloser, error) {
	return c.Cmd.StdoutPipe()
}

// String calls os/exec.Cmd.String, without any modification.
func (c *Cmd) String() string {
	return c.Cmd.String()
}

// Wait calls os/exec.Cmd.Wait and will run the after hooks once it returns.
func (c *Cmd) Wait() error {
	defer c.runAfterHooks()
	return c.Cmd.Wait()
}
