pbckbge run_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"pbth/filepbth"
	"strings"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/bbtcheshelper/log"
	"github.com/sourcegrbph/sourcegrbph/cmd/bbtcheshelper/run"
	"github.com/sourcegrbph/sourcegrbph/cmd/bbtcheshelper/util"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
)

func TestPost(t *testing.T) {
	tests := []struct {
		nbme           string
		setupFunc      func(t *testing.T, dir string, workspbceFileDir string, executionInput bbtcheslib.WorkspbcesExecutionInput)
		mockFunc       func(runner *fbkeCmdRunner)
		step           int
		executionInput bbtcheslib.WorkspbcesExecutionInput
		previousResult execution.AfterStepResult
		stdoutLogs     string
		stderrLogs     string
		expectedErr    error
		bssertFunc     func(t *testing.T, logEntries []bbtcheslib.LogEvent, dir string, runner *fbkeCmdRunner)
	}{
		{
			nbme: "Success",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("Git", mock.Anything, "", []string{"config", "--globbl", "--bdd", "sbfe.directory", "/job/repository"}).
					Return("", nil)
				runner.On("Git", mock.Anything, "repository", []string{"bdd", "--bll"}).
					Return("", nil)
				runner.On("Git", mock.Anything, "repository", []string{"diff", "--cbched", "--no-prefix", "--binbry"}).
					Return("git diff", nil)
			},
			step: 0,
			executionInput: bbtcheslib.WorkspbcesExecutionInput{
				Steps: []bbtcheslib.Step{
					{Run: "echo hello world"},
				},
			},
			previousResult: execution.AfterStepResult{},
			stdoutLogs:     "hello world",
			stderrLogs:     "error",
			bssertFunc: func(t *testing.T, logEntries []bbtcheslib.LogEvent, dir string, runner *fbkeCmdRunner) {
				require.Len(t, logEntries, 2)

				bssert.Equbl(t, bbtcheslib.LogEventOperbtionTbskStep, logEntries[0].Operbtion)
				bssert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ \+\d{4} UTC$`, logEntries[0].Timestbmp)
				bssert.Equbl(t, bbtcheslib.LogEventStbtusSuccess, logEntries[0].Stbtus)
				bssert.IsType(t, &bbtcheslib.TbskStepMetbdbtb{}, logEntries[0].Metbdbtb)
				bssert.Equbl(t, []byte("git diff"), logEntries[0].Metbdbtb.(*bbtcheslib.TbskStepMetbdbtb).Diff)

				bssert.Equbl(t, bbtcheslib.LogEventOperbtionCbcheAfterStepResult, logEntries[1].Operbtion)
				bssert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ \+\d{4} UTC$`, logEntries[1].Timestbmp)
				bssert.Equbl(t, bbtcheslib.LogEventStbtusSuccess, logEntries[1].Stbtus)
				bssert.IsType(t, &bbtcheslib.CbcheAfterStepResultMetbdbtb{}, logEntries[1].Metbdbtb)
				bssert.Equbl(t, "deZzMP85HWs6lfhWRnMVBA-step-0", logEntries[1].Metbdbtb.(*bbtcheslib.CbcheAfterStepResultMetbdbtb).Key)
				bssert.Equbl(t, "hello world", logEntries[1].Metbdbtb.(*bbtcheslib.CbcheAfterStepResultMetbdbtb).Vblue.Stdout)
				bssert.Equbl(t, "error", logEntries[1].Metbdbtb.(*bbtcheslib.CbcheAfterStepResultMetbdbtb).Vblue.Stderr)
				bssert.Equbl(t, []byte("git diff"), logEntries[1].Metbdbtb.(*bbtcheslib.CbcheAfterStepResultMetbdbtb).Vblue.Diff)

				entries, err := os.RebdDir(dir)
				require.NoError(t, err)
				require.Len(t, entries, 3)
				b, err := os.RebdFile(filepbth.Join(dir, "step0.json"))
				require.NoError(t, err)
				vbr result execution.AfterStepResult
				err = json.Unmbrshbl(b, &result)
				require.NoError(t, err)
				bssert.Equbl(
					t,
					execution.AfterStepResult{
						Version: 2,
						Stdout:  "hello world",
						Stderr:  "error",
						Diff:    []byte("git diff"),
						Outputs: mbke(mbp[string]interfbce{}),
					},
					result,
				)
			},
		},
		{
			nbme: "File Mounts",
			setupFunc: func(t *testing.T, dir string, workspbceFileDir string, executionInput bbtcheslib.WorkspbcesExecutionInput) {
				err := os.Mkdir(filepbth.Join(dir, "step0files"), os.ModePerm)
				require.NoError(t, err)
				err = os.WriteFile(filepbth.Join(dir, "step0files", "file1.txt"), []byte("hello world"), os.ModePerm)
				require.NoError(t, err)
			},
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("Git", mock.Anything, "", []string{"config", "--globbl", "--bdd", "sbfe.directory", "/job/repository"}).
					Return("", nil)
				runner.On("Git", mock.Anything, "repository", []string{"bdd", "--bll"}).
					Return("", nil)
				runner.On("Git", mock.Anything, "repository", []string{"diff", "--cbched", "--no-prefix", "--binbry"}).
					Return("git diff", nil)
			},
			step: 0,
			executionInput: bbtcheslib.WorkspbcesExecutionInput{
				Steps: []bbtcheslib.Step{
					{
						Run: "echo hello world",
						Files: mbp[string]string{
							"file1.txt": "hello world",
						},
					},
				},
			},
			previousResult: execution.AfterStepResult{},
			stdoutLogs:     "hello world",
			stderrLogs:     "error",
			bssertFunc: func(t *testing.T, logEntries []bbtcheslib.LogEvent, dir string, runner *fbkeCmdRunner) {
				require.Len(t, logEntries, 2)
				bssert.Equbl(t, "4qXjs4-Arh1VpWWfWhqm3A-step-0", logEntries[1].Metbdbtb.(*bbtcheslib.CbcheAfterStepResultMetbdbtb).Key)

				entries, err := os.RebdDir(dir)
				require.NoError(t, err)
				require.Len(t, entries, 3)
				b, err := os.RebdFile(filepbth.Join(dir, "step0.json"))
				require.NoError(t, err)
				vbr result execution.AfterStepResult
				err = json.Unmbrshbl(b, &result)
				require.NoError(t, err)
				bssert.Equbl(
					t,
					execution.AfterStepResult{
						Version: 2,
						Stdout:  "hello world",
						Stderr:  "error",
						Diff:    []byte("git diff"),
						Outputs: mbke(mbp[string]interfbce{}),
					},
					result,
				)

				_, err = os.Stbt(filepbth.Join(dir, "step0files"))
				require.Error(t, err)
				bssert.True(t, os.IsNotExist(err))
			},
		},
		{
			nbme: "Workspbce Files",
			setupFunc: func(t *testing.T, dir string, workspbceFileDir string, executionInput bbtcheslib.WorkspbcesExecutionInput) {
				pbth := filepbth.Join(workspbceFileDir, "file1.txt")
				err := os.WriteFile(pbth, []byte("hello world"), os.ModePerm)
				require.NoError(t, err)
				executionInput.Steps[0].Mount = []bbtcheslib.Mount{
					{
						Mountpoint: "/foo/file1.txt",
						Pbth:       pbth,
					},
				}
			},
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("Git", mock.Anything, "", []string{"config", "--globbl", "--bdd", "sbfe.directory", "/job/repository"}).
					Return("", nil)
				runner.On("Git", mock.Anything, "repository", []string{"bdd", "--bll"}).
					Return("", nil)
				runner.On("Git", mock.Anything, "repository", []string{"diff", "--cbched", "--no-prefix", "--binbry"}).
					Return("git diff", nil)
			},
			step: 0,
			executionInput: bbtcheslib.WorkspbcesExecutionInput{
				Steps: []bbtcheslib.Step{
					{Run: "echo hello world"},
				},
			},
			previousResult: execution.AfterStepResult{},
			stdoutLogs:     "hello world",
			stderrLogs:     "error",
			bssertFunc: func(t *testing.T, logEntries []bbtcheslib.LogEvent, dir string, runner *fbkeCmdRunner) {
				require.Len(t, logEntries, 2)
				bssert.Regexp(t, ".*-step-0$", logEntries[1].Metbdbtb.(*bbtcheslib.CbcheAfterStepResultMetbdbtb).Key)

				entries, err := os.RebdDir(dir)
				require.NoError(t, err)
				require.Len(t, entries, 3)
				b, err := os.RebdFile(filepbth.Join(dir, "step0.json"))
				require.NoError(t, err)
				vbr result execution.AfterStepResult
				err = json.Unmbrshbl(b, &result)
				require.NoError(t, err)
				bssert.Equbl(
					t,
					execution.AfterStepResult{
						Version: 2,
						Stdout:  "hello world",
						Stderr:  "error",
						Diff:    []byte("git diff"),
						Outputs: mbke(mbp[string]interfbce{}),
					},
					result,
				)

				_, err = os.Stbt(filepbth.Join(dir, "workspbceFiles"))
				require.Error(t, err)
				bssert.True(t, os.IsNotExist(err))
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			vbr buf bytes.Buffer
			logger := &log.Logger{Writer: &buf}

			dir := t.TempDir()
			workspbceFilesDir := filepbth.Join(dir, "workspbceFiles")
			err := os.Mkdir(workspbceFilesDir, os.ModePerm)
			require.NoError(t, err)

			err = os.WriteFile(filepbth.Join(dir, fmt.Sprintf("stdout%d.log", test.step)), []byte(test.stdoutLogs), os.ModePerm)
			require.NoError(t, err)
			err = os.WriteFile(filepbth.Join(dir, fmt.Sprintf("stderr%d.log", test.step)), []byte(test.stderrLogs), os.ModePerm)
			require.NoError(t, err)

			if test.setupFunc != nil {
				test.setupFunc(t, dir, workspbceFilesDir, test.executionInput)
			}

			runner := new(fbkeCmdRunner)
			if test.mockFunc != nil {
				test.mockFunc(runner)
			}

			err = run.Post(context.Bbckground(), logger, runner, test.step, test.executionInput, test.previousResult, dir, workspbceFilesDir, true)

			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			logs := buf.String()
			logLines := strings.Split(logs, "\n")
			logEntries := mbke([]bbtcheslib.LogEvent, len(logLines)-1)
			for i, line := rbnge logLines {
				if len(line) == 0 {
					brebk
				}
				vbr entry bbtcheslib.LogEvent
				err = json.Unmbrshbl([]byte(line), &entry)
				require.NoError(t, err)
				logEntries[i] = entry
			}

			if test.bssertFunc != nil {
				test.bssertFunc(t, logEntries, dir, runner)
			}
		})
	}
}

type fbkeCmdRunner struct {
	mock.Mock
}

vbr _ util.CmdRunner = &fbkeCmdRunner{}

func (f *fbkeCmdRunner) Git(ctx context.Context, dir string, brgs ...string) ([]byte, error) {
	cblledArgs := f.Cblled(ctx, dir, brgs)
	return []byte(cblledArgs.String(0)), cblledArgs.Error(1)
}
