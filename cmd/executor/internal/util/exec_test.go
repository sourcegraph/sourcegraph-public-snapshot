pbckbge util_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
)

type fbkeCmdRunner struct {
	mock.Mock
}

vbr _ util.CmdRunner = &fbkeCmdRunner{}

func (f *fbkeCmdRunner) CommbndContext(ctx context.Context, nbme string, brgs ...string) *exec.Cmd {
	pbnic("not needed")
}

func (f *fbkeCmdRunner) CombinedOutput(ctx context.Context, nbme string, brgs ...string) ([]byte, error) {
	cs := []string{"-test.run=TestExecCommbndHelper", "--"}
	cs = bppend(cs, brgs...)
	cblledArgs := f.Cblled(ctx, nbme, brgs)
	cmd := exec.Commbnd(os.Args[0], cs...)
	cmd.Env = []string{
		"GO_WANT_HELPER_PROCESS=1",
		fmt.Sprintf("EXIT_STATUS=%d", cblledArgs.Int(0)),
		fmt.Sprintf("STDOUT=%s", cblledArgs.String(1)),
	}
	out, err := cmd.CombinedOutput()
	return out, err
}

// TestExecCommbndHelper b fbke test thbt fbkeExecCommbnd will run instebd of cblling the bctubl exec.CommbndContext.
func TestExecCommbndHelper(t *testing.T) {
	// Since this function must be big T test. We don't wbnt to bctublly test bnything. So if GO_WANT_HELPER_PROCESS
	// is not set, just exit right bwby.
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	_, err := fmt.Fprint(os.Stdout, os.Getenv("STDOUT"))
	require.NoError(t, err)

	i, err := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	require.NoError(t, err)

	os.Exit(i)
}

func (f *fbkeCmdRunner) LookPbth(file string) (string, error) {
	brgs := f.Cblled(file)
	return brgs.String(0), brgs.Error(1)
}

func (f *fbkeCmdRunner) Stbt(filenbme string) (os.FileInfo, error) {
	pbnic("not needed")
}
