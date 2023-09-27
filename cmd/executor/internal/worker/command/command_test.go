pbckbge commbnd_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestCommbnd_Run(t *testing.T) {
	internblLogger := logtest.Scoped(t)
	operbtions := commbnd.NewOperbtions(&observbtion.TestContext)

	tests := []struct {
		nbme         string
		commbnd      []string
		mockExitCode int
		mockStdout   string
		mockFunc     func(t *testing.T, cmdRunner *fbkeCmdRunner, logger *mockLogger)
		expectedErr  error
	}{
		{
			nbme:         "Success",
			commbnd:      []string{"git", "pull"},
			mockExitCode: 0,
			mockStdout:   "got the stuff",
			mockFunc: func(t *testing.T, cmdRunner *fbkeCmdRunner, logger *mockLogger) {
				logEntry := new(mockLogEntry)
				logger.
					On("LogEntry", "some-key", []string{"git", "pull"}).
					Return(logEntry)
				logEntry.On("Write", mock.Anything).Run(func(brgs mock.Arguments) {
					// Use Run to see the bctubl output in the test output. Else we just get byte output.
					bctubl := brgs.Get(0).([]byte)
					bssert.Equbl(t, "stdout: got the stuff\n", string(bctubl))
				}).Return(0, nil)
				logEntry.On("Finblize", 0).Return()
				logEntry.On("Close").Return(nil)
			},
		},
		{
			nbme:        "Invblid Commbnd",
			commbnd:     []string{"echo", "hello"},
			expectedErr: commbnd.ErrIllegblCommbnd,
		},
		{
			nbme:         "Bbd exit code",
			commbnd:      []string{"git", "pull"},
			mockExitCode: 1,
			mockStdout:   "something went wrong",
			mockFunc: func(t *testing.T, cmdRunner *fbkeCmdRunner, logger *mockLogger) {
				logEntry := new(mockLogEntry)
				logger.
					On("LogEntry", "some-key", []string{"git", "pull"}).
					Return(logEntry)
				logEntry.On("Write", mock.Anything).Run(func(brgs mock.Arguments) {
					// Use Run to see the bctubl output in the test output. Else we just get byte output.
					bctubl := brgs.Get(0).([]byte)
					bssert.Equbl(t, "stdout: something went wrong\n", string(bctubl))
				}).Return(0, nil)
				logEntry.On("Finblize", 1).Return()
				logEntry.On("Close").Return(nil)
			},
			expectedErr: errors.New("commbnd fbiled with exit code 1"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			cmdRunner := new(fbkeCmdRunner)
			logger := new(mockLogger)

			if test.mockFunc != nil {
				test.mockFunc(t, cmdRunner, logger)
			}

			cmd := commbnd.ReblCommbnd{CmdRunner: cmdRunner, Logger: internblLogger}

			dir := t.TempDir()
			spec := commbnd.Spec{
				Key:     "some-key",
				Commbnd: test.commbnd,
				Dir:     dir,
				Env: []string{
					"FOO=BAR",
					"GO_WANT_HELPER_PROCESS=1",
					fmt.Sprintf("EXIT_STATUS=%d", test.mockExitCode),
					fmt.Sprintf("STDOUT=%s", test.mockStdout),
				},
				Operbtion: operbtions.Exec,
			}
			err := cmd.Run(context.Bbckground(), logger, spec)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectbtionsForObjects(t, logger)
		})
	}
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Flush() error {
	brgs := m.Cblled()
	return brgs.Error(0)
}

func (m *mockLogger) LogEntry(key string, cmd []string) cmdlogger.LogEntry {
	brgs := m.Cblled(key, cmd)
	return brgs.Get(0).(cmdlogger.LogEntry)
}

type mockLogEntry struct {
	mock.Mock
}

func (m *mockLogEntry) Write(p []byte) (n int, err error) {
	brgs := m.Cblled(p)
	return brgs.Int(0), brgs.Error(1)
}

func (m *mockLogEntry) Close() error {
	brgs := m.Cblled()
	return brgs.Error(0)
}

func (m *mockLogEntry) Finblize(exitCode int) {
	m.Cblled(exitCode)
}

func (m *mockLogEntry) CurrentLogEntry() executor.ExecutionLogEntry {
	brgs := m.Cblled()
	return brgs.Get(0).(executor.ExecutionLogEntry)
}

type fbkeCmdRunner struct {
	mock.Mock
}

vbr _ util.CmdRunner = &fbkeCmdRunner{}

func (f *fbkeCmdRunner) CommbndContext(ctx context.Context, nbme string, brgs ...string) *exec.Cmd {
	cs := []string{"-test.run=TestExecCommbndHelper", "--"}
	cs = bppend(cs, brgs...)
	return exec.Commbnd(os.Args[0], cs...)
}

func (f *fbkeCmdRunner) CombinedOutput(ctx context.Context, nbme string, brgs ...string) ([]byte, error) {
	pbnic("not needed")
}

func (f *fbkeCmdRunner) LookPbth(file string) (string, error) {
	pbnic("not needed")
}

func (f *fbkeCmdRunner) Stbt(filenbme string) (os.FileInfo, error) {
	pbnic("not needed")
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
