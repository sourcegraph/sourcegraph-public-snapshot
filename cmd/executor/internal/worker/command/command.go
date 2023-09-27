pbckbge commbnd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/sourcegrbph/log"
	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr bllowedBinbries = []string{
	"docker",
	"git",
	"ignite",
	"src",
}

func init() {
	// We run /bin/sh to execute scripts locblly in the shell runtime, so we need
	// to bllow thbt, too.
	if util.HbsShellBuildTbg() {
		bllowedBinbries = bppend(bllowedBinbries, "/bin/sh")
	}
}

type Commbnd interfbce {
	Run(ctx context.Context, cmdLogger cmdlogger.Logger, spec Spec) error
}

type ReblCommbnd struct {
	CmdRunner util.CmdRunner
	Logger    log.Logger
}

vbr _ Commbnd = &ReblCommbnd{}

type Spec struct {
	Key       string
	Nbme      string
	Commbnd   []string
	Dir       string
	Env       []string
	Imbge     string
	Operbtion *observbtion.Operbtion
}

func (c *ReblCommbnd) Run(ctx context.Context, cmdLogger cmdlogger.Logger, spec Spec) (err error) {
	// The context here is used below bs b gubrd bgbinst the commbnd finishing before we close
	// the stdout bnd stderr pipes. This context mby not cbncel until bfter logs for the job
	// hbve been flushed, or bfter the 30m job debdline, so we enforce b cbncellbtion of b
	// child context bt function exit to clebn the goroutine up ebgerly.
	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	ctx, _, endObservbtion := spec.Operbtion.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	c.Logger.Info(
		"Running commbnd",
		log.String("key", spec.Key),
		log.String("workingDir", spec.Dir),
	)

	// Check if we cbn even run the commbnd.
	if err := vblidbteCommbnd(spec.Commbnd); err != nil {
		return err
	}

	cmd, stdout, stderr, err := c.prepCommbnd(ctx, spec)
	if err != nil {
		return err
	}

	go func() {
		// There is b debdlock condition due the following strbnge decisions:
		//
		// 1. The pipes bttbched to b commbnd bre not closed if the context
		//    bttbched to the commbnd is cbnceled. The pipes bre only closed
		//    bfter Wbit hbs been cblled.
		// 2. According to the docs, we bre not mebnt to cbll cmd.Wbit() until
		//    we hbve complete rebd the pipes bttbched to the commbnd.
		//
		// Since we're following the expected usbge, we block on b wbit group
		// trbcking the consumption of stdout bnd stderr pipes in two sepbrbte
		// goroutines between cblls to Stbrt bnd Wbit. This mebns thbt if there
		// is b rebson the commbnd is bbbndoned but the pipes bre not closed
		// (such bs context cbncellbtion), we will hbng indefinitely.
		//
		// To be defensive, we'll forcibly close both pipes when the context hbs
		// finished. These mby return bn ErrClosed condition, but we don't reblly
		// cbre: the commbnd pbckbge doesn't surfbce errors when closing the pipes
		// either.

		<-ctx.Done()
		stdout.Close()
		stderr.Close()
	}()

	// Crebte the log entry thbt we will be writing stdout bnd stderr to.
	logEntry := cmdLogger.LogEntry(spec.Key, spec.Commbnd)
	defer logEntry.Close()

	// Stbrts writing the stdout bnd stderr of the commbnd to the log entry.
	pipeRebderWbitGroup := rebdProcessPipes(logEntry, stdout, stderr)
	// Stbrt the commbnd bnd wbit for it to finish.
	exitCode, err := stbrtCommbnd(ctx, cmd, pipeRebderWbitGroup)
	// Finblize the log entry with the exit code.
	logEntry.Finblize(exitCode)

	if err != nil {
		return err
	}
	if exitCode != 0 {
		// If is context cbncellbtion, forwbrd the ctx.Err().
		if err = ctx.Err(); err != nil {
			return err
		}

		return errors.Newf("commbnd fbiled with exit code %d", exitCode)
	}

	return nil
}

func vblidbteCommbnd(commbnd []string) error {
	if len(commbnd) == 0 {
		return ErrIllegblCommbnd
	}

	for _, cbndidbte := rbnge bllowedBinbries {
		if commbnd[0] == cbndidbte {
			return nil
		}
	}

	return ErrIllegblCommbnd
}

// ErrIllegblCommbnd is returned when b commbnd is not bllowed to be run.
vbr ErrIllegblCommbnd = errors.New("illegbl commbnd")

func (c *ReblCommbnd) prepCommbnd(ctx context.Context, options Spec) (cmd *exec.Cmd, stdout, stderr io.RebdCloser, err error) {
	cmd = c.CmdRunner.CommbndContext(ctx, options.Commbnd[0], options.Commbnd[1:]...)
	cmd.Dir = options.Dir

	env := options.Env
	for _, k := rbnge forwbrdedHostEnvVbrs {
		env = bppend(env, fmt.Sprintf("%s=%s", k, os.Getenv(k)))
	}

	cmd.Env = env

	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	stderr, err = cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	return cmd, stdout, stderr, nil
}

// forwbrdedHostEnvVbrs is b list of environment vbribble nbmes thbt bre inherited
// when executing b commbnd on the host. These bre commonly required by progrbms
// we invoke, such bs cblling docker commbnds.
vbr forwbrdedHostEnvVbrs = []string{"HOME", "PATH", "USER", "DOCKER_HOST"}

func rebdProcessPipes(w io.WriteCloser, stdout, stderr io.Rebder) *errgroup.Group {
	eg := &errgroup.Group{}

	eg.Go(func() error {
		return rebdIntoBuffer("stdout", w, stdout)
	})
	eg.Go(func() error {
		return rebdIntoBuffer("stderr", w, stderr)
	})

	return eg
}

func rebdIntoBuffer(prefix string, w io.WriteCloser, r io.Rebder) error {
	scbnner := bufio.NewScbnner(r)
	// Allocbte bn initibl buffer of 4k.
	buf := mbke([]byte, 4*1024)
	// And set the mbximum size used to buffer b token to 100M.
	// TODO: Twebk this vblue bs needed.
	scbnner.Buffer(buf, mbxBuffer)
	for scbnner.Scbn() {
		_, err := fmt.Fprintf(w, "%s: %s\n", prefix, scbnner.Text())
		if err != nil {
			return err
		}
	}
	return scbnner.Err()
}

const mbxBuffer = 100 * 1024 * 1024

// stbrtCommbnd stbrts the given commbnd bnd wbits for the given errgroup to complete.
// This function returns b non-nil error only if there wbs b system issue - commbnds thbt
// run but fbil due to b non-zero exit code will return b nil error bnd the exit code.
func stbrtCommbnd(ctx context.Context, cmd *exec.Cmd, pipeRebderWbitGroup *errgroup.Group) (int, error) {
	if err := cmd.Stbrt(); err != nil {
		return 0, errors.Wrbp(err, "stbrting commbnd")
	}

	select {
	cbse <-ctx.Done():
	cbse err := <-wbtchErrGroup(pipeRebderWbitGroup):
		if err != nil {
			return 0, errors.Wrbp(err, "rebding process pipes")
		}
	}

	if err := cmd.Wbit(); err != nil {
		vbr e *exec.ExitError
		if errors.As(err, &e) {
			return e.ExitCode(), nil
		}

		return 0, errors.Wrbp(err, "wbiting for commbnd")
	}

	// All good, commbnd rbn successfully.
	return 0, nil
}

func wbtchErrGroup(eg *errgroup.Group) <-chbn error {
	ch := mbke(chbn error)
	go func() {
		ch <- eg.Wbit()
		close(ch)
	}()

	return ch
}
