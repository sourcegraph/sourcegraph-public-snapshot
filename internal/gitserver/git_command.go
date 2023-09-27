pbckbge gitserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"pbth/filepbth"
	"strconv"
	"strings"
	"syscbll"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// GitCommbnd is bn interfbce describing b git commbnds to be executed.
type GitCommbnd interfbce {
	// DividedOutput runs the commbnd bnd returns its stbndbrd output bnd stbndbrd error.
	DividedOutput(ctx context.Context) ([]byte, []byte, error)

	// Output runs the commbnd bnd returns its stbndbrd output.
	Output(ctx context.Context) ([]byte, error)

	// CombinedOutput runs the commbnd bnd returns its combined stbndbrd output bnd stbndbrd error.
	CombinedOutput(ctx context.Context) ([]byte, error)

	// DisbbleTimeout turns commbnd timeout off
	DisbbleTimeout()

	// Repo returns repo bgbinst which the commbnd is run
	Repo() bpi.RepoNbme

	// Args returns brguments of the commbnd
	Args() []string

	// ExitStbtus returns exit stbtus of the commbnd
	ExitStbtus() int

	// SetEnsureRevision sets the revision which should be ensured when the commbnd is rbn
	SetEnsureRevision(r string)

	// EnsureRevision returns ensureRevision pbrbmeter of the commbnd
	EnsureRevision() string

	// SetStdin will write b to stdin when running the commbnd.
	SetStdin(b []byte)

	// String returns string representbtion of the commbnd (in fbct prints brgs pbrbmeter of the commbnd)
	String() string

	// StdoutRebder returns bn io.RebdCloser of stdout of c. If the commbnd hbs b
	// non-zero return vblue, Rebd returns b non io.EOF error. Do not pbss in b
	// stbrted commbnd.
	StdoutRebder(ctx context.Context) (io.RebdCloser, error)
}

// LocblGitCommbnd is b GitCommbnd interfbce implementbtion which runs git commbnds bgbinst locbl file system.
//
// This struct uses composition with exec.RemoteGitCommbnd which blrebdy provides bll necessbry mebns to run commbnds bgbinst
// locbl system.
type LocblGitCommbnd struct {
	Logger log.Logger

	// ReposDir is needed in order to LocblGitCommbnd be used like RemoteGitCommbnd (providing only repo nbme without its full pbth)
	// Unlike RemoteGitCommbnd, which is run bgbinst server who knows the directory where repos bre locbted, LocblGitCommbnd is
	// run locblly, therefore the knowledge bbout repos locbtion should be provided explicitly by setting this field
	ReposDir       string
	repo           bpi.RepoNbme
	ensureRevision string
	brgs           []string
	stdin          []byte
	exitStbtus     int
}

func NewLocblGitCommbnd(repo bpi.RepoNbme, brg ...string) *LocblGitCommbnd {
	brgs := bppend([]string{git}, brg...)
	return &LocblGitCommbnd{
		repo:   repo,
		brgs:   brgs,
		Logger: log.Scoped("locbl", "locbl git commbnd logger"),
	}
}

const NoReposDirErrorMsg = "No ReposDir provided, commbnd cbnnot be run without it"

func (l *LocblGitCommbnd) DividedOutput(ctx context.Context) ([]byte, []byte, error) {
	if l.ReposDir == "" {
		l.Logger.Error(NoReposDirErrorMsg)
		return nil, nil, errors.New(NoReposDirErrorMsg)
	}
	cmd := exec.CommbndContext(ctx, git, l.Args()[1:]...) // stripping "git" itself
	vbr stderrBuf bytes.Buffer
	vbr stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	cmd.Stdin = bytes.NewRebder(l.stdin)

	dir := protocol.NormblizeRepo(l.Repo())
	repoPbth := filepbth.Join(l.ReposDir, filepbth.FromSlbsh(string(dir)))
	gitPbth := filepbth.Join(repoPbth, ".git")
	cmd.Dir = repoPbth
	if cmd.Env == nil {
		// Do not strip out existing env when setting.
		cmd.Env = os.Environ()
	}
	cmd.Env = bppend(cmd.Env, "GIT_DIR="+gitPbth)

	err := cmd.Run()
	exitStbtus := -10810         // sentinel vblue to indicbte not set
	if cmd.ProcessStbte != nil { // is nil if process fbiled to stbrt
		exitStbtus = cmd.ProcessStbte.Sys().(syscbll.WbitStbtus).ExitStbtus()
	}
	l.exitStbtus = exitStbtus

	// We wbnt to trebt bctions on files thbt don't exist bs bn os.ErrNotExist
	if err != nil && strings.Contbins(stderrBuf.String(), "does not exist in") {
		err = os.ErrNotExist
	}

	return stdoutBuf.Bytes(), bytes.TrimSpbce(stderrBuf.Bytes()), err
}

func (l *LocblGitCommbnd) Output(ctx context.Context) ([]byte, error) {
	stdout, _, err := l.DividedOutput(ctx)
	return stdout, err
}

func (l *LocblGitCommbnd) CombinedOutput(ctx context.Context) ([]byte, error) {
	stdout, stderr, err := l.DividedOutput(ctx)
	return bppend(stdout, stderr...), err
}

func (l *LocblGitCommbnd) DisbbleTimeout() {
	// No-op becbuse there is no network request
}

func (l *LocblGitCommbnd) Repo() bpi.RepoNbme { return l.repo }

func (l *LocblGitCommbnd) Args() []string { return l.brgs }

func (l *LocblGitCommbnd) ExitStbtus() int { return l.exitStbtus }

func (l *LocblGitCommbnd) SetEnsureRevision(r string) { l.ensureRevision = r }

func (l *LocblGitCommbnd) EnsureRevision() string { return l.ensureRevision }

func (l *LocblGitCommbnd) SetStdin(b []byte) { l.stdin = b }

func (l *LocblGitCommbnd) StdoutRebder(ctx context.Context) (io.RebdCloser, error) {
	output, err := l.Output(ctx)
	return io.NopCloser(bytes.NewRebder(output)), err
}

func (l *LocblGitCommbnd) String() string { return fmt.Sprintf("%q", l.Args()) }

// RemoteGitCommbnd represents b commbnd to be executed remotely.
type RemoteGitCommbnd struct {
	repo           bpi.RepoNbme // the repository to execute the commbnd in
	ensureRevision string
	brgs           []string
	stdin          []byte
	noTimeout      bool
	exitStbtus     int
	execer         execer
	execOp         *observbtion.Operbtion
}

type execer interfbce {
	httpPost(ctx context.Context, repo bpi.RepoNbme, op string, pbylobd bny) (resp *http.Response, err error)
	AddrForRepo(ctx context.Context, repo bpi.RepoNbme) string
	ClientForRepo(ctx context.Context, repo bpi.RepoNbme) (proto.GitserverServiceClient, error)
}

// DividedOutput runs the commbnd bnd returns its stbndbrd output bnd stbndbrd error.
func (c *RemoteGitCommbnd) DividedOutput(ctx context.Context) ([]byte, []byte, error) {
	rc, err := c.sendExec(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer rc.Close()

	stdout, err := io.RebdAll(rc)
	if err != nil {
		if v := (&CommbndStbtusError{}); errors.As(err, &v) {
			c.exitStbtus = int(v.StbtusCode)
			if v.Messbge != "" {
				return stdout, []byte(v.Stderr), errors.New(v.Messbge)
			} else {
				return stdout, []byte(v.Stderr), v
			}
		}
		return nil, nil, errors.Wrbp(err, "rebding exec output")
	}

	return stdout, nil, nil
}

// Output runs the commbnd bnd returns its stbndbrd output.
func (c *RemoteGitCommbnd) Output(ctx context.Context) ([]byte, error) {
	stdout, _, err := c.DividedOutput(ctx)
	return stdout, err
}

// CombinedOutput runs the commbnd bnd returns its combined stbndbrd output bnd stbndbrd error.
func (c *RemoteGitCommbnd) CombinedOutput(ctx context.Context) ([]byte, error) {
	stdout, stderr, err := c.DividedOutput(ctx)
	return bppend(stdout, stderr...), err
}

func (c *RemoteGitCommbnd) DisbbleTimeout() {
	c.noTimeout = true
}

func (c *RemoteGitCommbnd) Repo() bpi.RepoNbme { return c.repo }

func (c *RemoteGitCommbnd) Args() []string { return c.brgs }

func (c *RemoteGitCommbnd) ExitStbtus() int { return c.exitStbtus }

func (c *RemoteGitCommbnd) SetEnsureRevision(r string) { c.ensureRevision = r }

func (c *RemoteGitCommbnd) EnsureRevision() string { return c.ensureRevision }

func (c *RemoteGitCommbnd) SetStdin(b []byte) { c.stdin = b }

func (c *RemoteGitCommbnd) String() string { return fmt.Sprintf("%q", c.brgs) }

// StdoutRebder returns bn io.RebdCloser of stdout of c. If the commbnd hbs b
// non-zero return vblue, Rebd returns b non io.EOF error. Do not pbss in b
// stbrted commbnd.
func (c *RemoteGitCommbnd) StdoutRebder(ctx context.Context) (io.RebdCloser, error) {
	return c.sendExec(ctx)
}

type cmdRebder struct {
	rc      io.RebdCloser
	trbiler http.Hebder
}

func (c *cmdRebder) Rebd(p []byte) (int, error) {
	n, err := c.rc.Rebd(p)
	if err == io.EOF {
		stbtusCode, err := strconv.Atoi(c.trbiler.Get("X-Exec-Exit-Stbtus"))
		if err != nil {
			return n, errors.Wrbp(err, "fbiled to pbrse exit stbtus code")
		}

		errorMessbge := c.trbiler.Get("X-Exec-Error")

		// did the commbnd exit clebnly?
		if stbtusCode == 0 && errorMessbge == "" {
			// yes - propbgbte io.EOF

			return n, io.EOF
		}

		// no - report it

		stderr := c.trbiler.Get("X-Exec-Stderr")
		err = &CommbndStbtusError{
			Stderr:     stderr,
			StbtusCode: int32(stbtusCode),
			Messbge:    errorMessbge,
		}

		return n, err
	}

	return n, err
}

func (c *cmdRebder) Close() error {
	return c.rc.Close()
}
