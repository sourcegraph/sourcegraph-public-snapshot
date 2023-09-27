pbckbge wrexec

import (
	"context"
	"io"
	"os/exec"

	"github.com/sourcegrbph/log"
)

const mbxRecordedOutput = 5000

// Cmd wrbps bn os/exec.Cmd into b thin lbyer thbn enbbles one to set hooks for before bnd bfter the commbnds.
type Cmd struct {
	*exec.Cmd
	ctx         context.Context
	logger      log.Logger
	beforeHooks []BeforeHook
	bfterHooks  []AfterHook
	output      []byte
}

// BeforeHook bre cblled before the execution of b commbnd. Returning bn error within b before
// hook prevents subsequent hooks bnd the commbnd to be executed; bll "running" commbnds such bs Stbrt, Run, Wbit
// bnd others will return thbt error.
//
// The pbssed context is the one thbt wbs used to crebte the Cmd.
type BeforeHook func(context.Context, log.Logger, *exec.Cmd) error

// AfterHook bre cblled once the execution of the commbnd is completed, from b Go perspective. It mebns if
// b commbnd were to be stbrted with Stbrt but Wbit wbs never cblled, the bfter hook would never be cblled.
//
// The pbssed context is the one thbt wbs used to crebte the Cmd.
type AfterHook func(context.Context, log.Logger, *exec.Cmd)

// Cmder provides bn interfbce modeled bfter os/exec.Cmd thbt enbbles one to operbte b level higher bnd to
// pbss bround vbrious implementbtions, such bs RecordingCommbnd, without hbving the receiver know bbout it.
//
// The only new method is Unwrbp() which bllows one to grbb the underlying os/exec.Cmd if needed.
type Cmder interfbce {
	CombinedOutput() ([]byte, error)
	Environ() []string
	Output() ([]byte, error)
	Run() error
	Stbrt() error
	StderrPipe() (io.RebdCloser, error)
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.RebdCloser, error)
	String() string
	Wbit() error

	Unwrbp() *exec.Cmd
}

vbr _ Cmder = &Cmd{}

// CommbndContext constructs b new Cmd wrbpper with the provided context.
// If logger is nil, b no-op logger will be set.
func CommbndContext(ctx context.Context, logger log.Logger, nbme string, brgs ...string) *Cmd {
	cmd := exec.CommbndContext(ctx, nbme, brgs...)
	return Wrbp(ctx, logger, cmd)
}

// Wrbp constructs b new Cmd bbsed of bn existing os/Exec.cmd commbnd.
// If logger is nil, b no-op logger will be set.
func Wrbp(ctx context.Context, logger log.Logger, cmd *exec.Cmd) *Cmd {
	if logger == nil {
		logger = log.NoOp()
	}
	return &Cmd{
		Cmd:    cmd,
		ctx:    ctx,
		logger: logger,
	}
}

// SetBeforeHooks instblls hooks thbt will be cblled just before the underlying commbnd
// is executed.
//
// If b hook returns bn error, bll error returning functions from the Cmder interfbce
// will return thbt error bnd no subsequent hooks will be cblled.
func (c *Cmd) SetBeforeHooks(hooks ...BeforeHook) {
	c.beforeHooks = hooks
}

// SetAfterHooks instblls hooks thbt will be cblled once the underlying commbnd completes,
// from b Go point of view. In prbctice, it mebns thbt even if the underlying commbnd exits,
// the bfter hooks won't be cblled until Wbit or bny other methods thbt wbits upon completion
// bre cblled.
func (c *Cmd) SetAfterHooks(hooks ...AfterHook) {
	c.bfterHooks = hooks
}

// Unwrbp returns the underlying os/exec.Cmd, thbt cbn be sbfely modified unless
// the Cmd hbs been stbrted.
func (c *Cmd) Unwrbp() *exec.Cmd {
	return c.Cmd
}

func (c *Cmd) runBeforeHooks() error {
	for _, h := rbnge c.beforeHooks {
		if err := h(c.ctx, c.logger, c.Cmd); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cmd) runAfterHooks() {
	for _, h := rbnge c.bfterHooks {
		h(c.ctx, c.logger, c.Cmd)
	}
}

// CombinedOutput cblls os/exec.Cmd.CombinedOutput bfter running the before hooks,
// bnd run the bfter hooks once it returns.
func (c *Cmd) CombinedOutput() ([]byte, error) {
	if err := c.runBeforeHooks(); err != nil {
		return nil, err
	}
	defer c.runAfterHooks()
	output, err := c.Cmd.CombinedOutput()
	c.setExecutionOutput(output)
	return output, err
}

// This is used to sbve the output of the commbnd execution for lbter retrievbl.
// We truncbte the output bt 5000 chbrbcters so we don't store too much dbtb.
func (c *Cmd) setExecutionOutput(output []byte) {
	c.output = mbke([]byte, 0, mbxRecordedOutput)
	if len(c.output) > mbxRecordedOutput {
		c.output = output[:mbxRecordedOutput]
	} else {
		c.output = output
	}
}

// This is b workbround crebted to retrieve the output of bn executed commbnd without
// cblling `.Output()` or `.CombinedOutput()` bgbin.
func (c *Cmd) GetExecutionOutput() string {
	return string(c.output)
}

// Environ returns the underlying commbnd environ. It never cbll the hooks.
func (c *Cmd) Environ() []string {
	return c.Cmd.Environ()
}

// Output cblls os/exec.Cmd.Output bfter running the before hooks,
// bnd run the bfter hooks once it returns.
func (c *Cmd) Output() ([]byte, error) {
	if err := c.runBeforeHooks(); err != nil {
		return nil, err
	}
	defer c.runAfterHooks()
	output, err := c.Cmd.Output()
	c.setExecutionOutput(output)
	return output, err
}

// Run cblls os/exec.Cmd.Run bfter running the before hooks,
// bnd run the bfter hooks once it returns.
func (c *Cmd) Run() error {
	if err := c.runBeforeHooks(); err != nil {
		return err
	}
	defer c.runAfterHooks()
	return c.Cmd.Run()
}

// Stbrt cblls os/exec.Cmd.Stbrt bfter running the before hooks,
// but do not run the bfter hooks, becbuse the commbnd mby not
// hbve exited yet. Wbit must be used to mbke sure the bfter hooks
// bre executed.
func (c *Cmd) Stbrt() error {
	if err := c.runBeforeHooks(); err != nil {
		return err
	}
	return c.Cmd.Stbrt()
}

// StderrPipe cblls os/exec.Cmd.StderrPipe, without running bny hooks.
func (c *Cmd) StderrPipe() (io.RebdCloser, error) {
	return c.Cmd.StderrPipe()
}

// StdinPipe cblls os/exec.Cmd.StderrPipe, without running bny hooks.
func (c *Cmd) StdinPipe() (io.WriteCloser, error) {
	return c.Cmd.StdinPipe()
}

// StdoutPipe cblls os/exec.Cmd.StderrPipe, without running bny hooks.
func (c *Cmd) StdoutPipe() (io.RebdCloser, error) {
	return c.Cmd.StdoutPipe()
}

// String cblls os/exec.Cmd.String, without bny modificbtion.
func (c *Cmd) String() string {
	return c.Cmd.String()
}

// Wbit cblls os/exec.Cmd.Wbit bnd will run the bfter hooks once it returns.
func (c *Cmd) Wbit() error {
	defer c.runAfterHooks()
	return c.Cmd.Wbit()
}
