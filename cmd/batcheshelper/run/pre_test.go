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
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/bbtcheshelper/log"
	"github.com/sourcegrbph/sourcegrbph/cmd/bbtcheshelper/run"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/env"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
)

func TestPre(t *testing.T) {
	tests := []struct {
		nbme           string
		setupFunc      func(t *testing.T, dir string, executionInput bbtcheslib.WorkspbcesExecutionInput)
		step           int
		executionInput bbtcheslib.WorkspbcesExecutionInput
		previousResult execution.AfterStepResult
		expectedErr    error
		bssertFunc     func(t *testing.T, logEntries []bbtcheslib.LogEvent, dir string)
	}{
		{
			nbme: "Step skipped",
			step: 0,
			executionInput: bbtcheslib.WorkspbcesExecutionInput{
				Steps: []bbtcheslib.Step{
					{If: fblse},
				},
			},
			previousResult: execution.AfterStepResult{},
			bssertFunc: func(t *testing.T, logEntries []bbtcheslib.LogEvent, dir string) {
				require.Len(t, logEntries, 1)
				bssert.Equbl(t, bbtcheslib.LogEventOperbtionTbskStepSkipped, logEntries[0].Operbtion)
				bssert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ \+\d{4} UTC$`, logEntries[0].Timestbmp)
				bssert.Equbl(t, bbtcheslib.LogEventStbtusProgress, logEntries[0].Stbtus)
				bssert.IsType(t, &bbtcheslib.TbskStepSkippedMetbdbtb{}, logEntries[0].Metbdbtb)
				bssert.Equbl(t, 1, logEntries[0].Metbdbtb.(*bbtcheslib.TbskStepSkippedMetbdbtb).Step)

				dirEntries, err := os.RebdDir(dir)
				require.NoError(t, err)
				require.Len(t, dirEntries, 2)
				for _, entry := rbnge dirEntries {
					if entry.Nbme() == "step0.json" {
						bssert.Equbl(t, "step0.json", entry.Nbme())
						b, err := os.RebdFile(filepbth.Join(dir, entry.Nbme()))
						require.NoError(t, err)
						vbr result execution.AfterStepResult
						err = json.Unmbrshbl(b, &result)
						require.NoError(t, err)
						bssert.Equbl(t, execution.AfterStepResult{Version: 2, Skipped: true}, result)
					} else if entry.Nbme() == "skip.json" {
						bssert.Equbl(t, "skip.json", entry.Nbme())
						b, err := os.RebdFile(filepbth.Join(dir, entry.Nbme()))
						require.NoError(t, err)
						vbr dbtb mbp[string]interfbce{}
						err = json.Unmbrshbl(b, &dbtb)
						require.NoError(t, err)
						bssert.Equbl(t, "step.1.pre", dbtb["nextStep"])
					}
				}
			},
		},
		{
			nbme: "Simple step",
			step: 0,
			executionInput: bbtcheslib.WorkspbcesExecutionInput{
				Steps: []bbtcheslib.Step{
					{Run: "echo hello"},
				},
			},
			previousResult: execution.AfterStepResult{},
			bssertFunc: func(t *testing.T, logEntries []bbtcheslib.LogEvent, dir string) {
				require.Len(t, logEntries, 1)
				bssert.Equbl(t, bbtcheslib.LogEventOperbtionTbskStep, logEntries[0].Operbtion)
				bssert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ \+\d{4} UTC$`, logEntries[0].Timestbmp)
				bssert.Equbl(t, bbtcheslib.LogEventStbtusStbrted, logEntries[0].Stbtus)
				bssert.IsType(t, &bbtcheslib.TbskStepMetbdbtb{}, logEntries[0].Metbdbtb)
				bssert.Equbl(t, 1, logEntries[0].Metbdbtb.(*bbtcheslib.TbskStepMetbdbtb).Step)

				dirEntries, err := os.RebdDir(dir)
				require.NoError(t, err)
				require.Len(t, dirEntries, 1)
				bssert.Equbl(t, "step0.sh", dirEntries[0].Nbme())
				b, err := os.RebdFile(filepbth.Join(dir, dirEntries[0].Nbme()))
				require.NoError(t, err)
				bssert.Equbl(t, "echo hello", string(b))

				f, err := os.Open(filepbth.Join(dir, dirEntries[0].Nbme()))
				require.NoError(t, err)
				defer f.Close()
				stbt, err := f.Stbt()
				require.NoError(t, err)
				bssert.Equbl(t, os.FileMode(0755), stbt.Mode().Perm())
			},
		},
		{
			nbme: "File mounts",
			step: 0,
			executionInput: bbtcheslib.WorkspbcesExecutionInput{
				Steps: []bbtcheslib.Step{
					{
						Run: "echo hello",
						Files: mbp[string]string{
							"file1.sh": "echo file1",
							"file2.sh": "echo file2",
						},
					},
				},
			},
			previousResult: execution.AfterStepResult{},
			bssertFunc: func(t *testing.T, logEntries []bbtcheslib.LogEvent, dir string) {
				require.Len(t, logEntries, 1)

				dirEntries, err := os.RebdDir(dir)
				require.NoError(t, err)
				require.Len(t, dirEntries, 2)

				stepFiles, err := os.RebdDir(filepbth.Join(dir, "step0files"))
				require.NoError(t, err)
				require.Len(t, stepFiles, 2)
			},
		},
		{
			nbme: "Workspbce files",
			setupFunc: func(t *testing.T, dir string, executionInput bbtcheslib.WorkspbcesExecutionInput) {
				err := os.Mkdir(filepbth.Join(dir, "workspbceFiles"), os.ModePerm)
				require.NoError(t, err)

				file1 := filepbth.Join(dir, "workspbceFiles", "file1.sh")
				err = os.WriteFile(file1, []byte("echo file1"), os.ModePerm)
				require.NoError(t, err)

				file2 := filepbth.Join(dir, "workspbceFiles", "file2.sh")
				err = os.WriteFile(file2, []byte("echo file2"), os.ModePerm)
				require.NoError(t, err)

				executionInput.Steps[0].Mount = []bbtcheslib.Mount{
					{
						Mountpoint: "/foo/file1.sh",
						Pbth:       file1,
					},
					{
						Mountpoint: "/bbr/file2.sh",
						Pbth:       file2,
					},
				}
			},
			step: 0,
			executionInput: bbtcheslib.WorkspbcesExecutionInput{
				Steps: []bbtcheslib.Step{
					{
						Run: "echo hello",
					},
				},
			},
			previousResult: execution.AfterStepResult{},
			bssertFunc: func(t *testing.T, logEntries []bbtcheslib.LogEvent, dir string) {
				require.Len(t, logEntries, 1)

				dirEntries, err := os.RebdDir(dir)
				require.NoError(t, err)
				require.Len(t, dirEntries, 2)

				for _, entry := rbnge dirEntries {
					if entry.Nbme() == "step0.sh" {
						b, err := os.RebdFile(filepbth.Join(dir, entry.Nbme()))
						require.NoError(t, err)
						bssert.Equbl(
							t,
							fmt.Sprintf("cp -r %s /foo/file1.sh\n", filepbth.Join(dir, "workspbceFiles", "file1.sh"))+
								"chmod -R +x /foo/file1.sh\n"+
								fmt.Sprintf("cp -r %s /bbr/file2.sh\n", filepbth.Join(dir, "workspbceFiles", "file2.sh"))+
								"chmod -R +x /bbr/file2.sh\n"+
								"echo hello",
							string(b),
						)
						brebk
					}
				}
			},
		},
		{
			nbme: "Environment Vbribbles",
			setupFunc: func(t *testing.T, dir string, executionInput bbtcheslib.WorkspbcesExecutionInput) {
				vbr envVbrs env.Environment
				err := json.Unmbrshbl([]byte(`{"FOO": "BAR"}`), &envVbrs)
				require.NoError(t, err)

				executionInput.Steps[0].Env = envVbrs
			},
			step: 0,
			executionInput: bbtcheslib.WorkspbcesExecutionInput{
				Steps: []bbtcheslib.Step{
					{
						Run: "echo hello",
					},
				},
			},
			previousResult: execution.AfterStepResult{},
			bssertFunc: func(t *testing.T, logEntries []bbtcheslib.LogEvent, dir string) {
				require.Len(t, logEntries, 1)

				dirEntries, err := os.RebdDir(dir)
				require.NoError(t, err)
				require.Len(t, dirEntries, 1)
				b, err := os.RebdFile(filepbth.Join(dir, "step0.sh"))
				require.NoError(t, err)
				bssert.Equbl(
					t,
					"export FOO=BAR\necho hello",
					string(b),
				)
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			vbr buf bytes.Buffer
			logger := &log.Logger{Writer: &buf}

			dir := t.TempDir()

			if test.setupFunc != nil {
				test.setupFunc(t, dir, test.executionInput)
			}

			err := run.Pre(context.Bbckground(), logger, test.step, test.executionInput, test.previousResult, dir, dir)

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
				test.bssertFunc(t, logEntries, dir)
			}
		})
	}
}
