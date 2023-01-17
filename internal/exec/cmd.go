package exec

import (
	"context"
	"io"
	"os/exec"

	"github.com/sourcegraph/log"
)

type Cmd struct {
	cmd         *exec.Cmd
	ctx         context.Context
	logger      log.Logger
	beforeHooks []BeforeHook
	afterHooks  []AfterHook
}

type BeforeHook func(context.Context, log.Logger, *exec.Cmd) error
type AfterHook func(context.Context, log.Logger, *exec.Cmd)

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

func Command(ctx context.Context, logger log.Logger, name string, args ...string) *Cmd { // TODO
	cmd := exec.CommandContext(ctx, name, args...)
	return Wrap(ctx, logger, cmd)
}

func Wrap(ctx context.Context, logger log.Logger, cmd *exec.Cmd) *Cmd {
	// TODO?
	if logger == nil {
		logger = log.NoOp()
	}
	return &Cmd{
		cmd:    cmd,
		ctx:    ctx,
		logger: logger,
	}
}

func (c *Cmd) SetBeforeHooks(hooks ...BeforeHook) {
	c.beforeHooks = hooks
}

func (c *Cmd) SetAfterHooks(hooks ...AfterHook) {
	c.afterHooks = hooks
}

func (c *Cmd) Unwrap() *exec.Cmd {
	return c.cmd
}

func (c *Cmd) runBeforeHooks() error {
	for _, h := range c.beforeHooks {
		if err := h(c.ctx, c.logger, c.cmd); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cmd) runAfterHooks() {
	for _, h := range c.afterHooks {
		h(c.ctx, c.logger, c.cmd)
	}
}

func (c *Cmd) CombinedOutput() ([]byte, error) {
	if err := c.runBeforeHooks(); err != nil {
		return nil, err
	}
	defer c.runAfterHooks()
	return c.cmd.CombinedOutput()
}

func (c *Cmd) Environ() []string {
	return c.cmd.Environ()
}

func (c *Cmd) Output() ([]byte, error) {
	if err := c.runBeforeHooks(); err != nil {
		return nil, err
	}
	defer c.runAfterHooks()
	return c.cmd.Output()
}

func (c *Cmd) Run() error {
	if err := c.runBeforeHooks(); err != nil {
		return err
	}
	defer c.runAfterHooks()
	return c.cmd.Run()
}

func (c *Cmd) Start() error {
	if err := c.runBeforeHooks(); err != nil {
		return err
	}
	return c.cmd.Start()
}

func (c *Cmd) StderrPipe() (io.ReadCloser, error) {
	return c.cmd.StderrPipe()
}

func (c *Cmd) StdinPipe() (io.WriteCloser, error) {
	return c.cmd.StdinPipe()
}

func (c *Cmd) StdoutPipe() (io.ReadCloser, error) {
	return c.cmd.StdoutPipe()
}

func (c *Cmd) String() string {
	return c.cmd.String()
}

func (c *Cmd) Wait() error {
	defer c.runAfterHooks()
	return c.cmd.Wait()
}
