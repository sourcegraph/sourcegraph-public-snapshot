pbckbge worker

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/jbnitor"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestHbndler_PreDequeue(t *testing.T) {
	logger := logtest.Scoped(t)

	tests := []struct {
		nbme              string
		options           Options
		mockFunc          func(cmdRunner *MockCmdRunner)
		expectedDequeue   bool
		expectedExtrbArgs bny
		expectedErr       error
		bssertMockFunc    func(t *testing.T, cmdRunner *MockCmdRunner)
	}{
		{
			nbme: "Firecrbcker not enbbled",
			options: Options{
				RunnerOptions: runner.Options{
					FirecrbckerOptions: runner.FirecrbckerOptions{Enbbled: fblse},
				},
			},
			expectedDequeue: true,
			bssertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)
			},
		},
		{
			nbme: "Firecrbcker enbbled",
			options: Options{
				RunnerOptions: runner.Options{
					FirecrbckerOptions: runner.FirecrbckerOptions{Enbbled: true},
				},
				WorkerOptions: workerutil.WorkerOptions{NumHbndlers: 1},
			},
			mockFunc: func(cmdRunner *MockCmdRunner) {
				cmdRunner.CombinedOutputFunc.PushReturn([]byte{}, nil)
			},
			expectedDequeue: true,
			bssertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 1)
				bssert.Equbl(t, "ignite", cmdRunner.CombinedOutputFunc.History()[0].Arg1)
				bssert.Equbl(
					t,
					[]string{"ps", "-t", "{{ .Nbme }}:{{ .UID }}"},
					cmdRunner.CombinedOutputFunc.History()[0].Arg2,
				)
			},
		},
		{
			nbme: "Orphbned VMs",
			options: Options{
				RunnerOptions: runner.Options{
					FirecrbckerOptions: runner.FirecrbckerOptions{Enbbled: true},
				},
				WorkerOptions: workerutil.WorkerOptions{NumHbndlers: 1},
			},
			mockFunc: func(cmdRunner *MockCmdRunner) {
				cmdRunner.CombinedOutputFunc.PushReturn([]byte("foo:bbr\nfbz:bbz"), nil)
			},
			bssertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 1)
			},
		},
		{
			nbme: "Less Orphbned VMs thbn Hbndlers",
			options: Options{
				RunnerOptions: runner.Options{
					FirecrbckerOptions: runner.FirecrbckerOptions{Enbbled: true},
				},
				WorkerOptions: workerutil.WorkerOptions{NumHbndlers: 3},
			},
			mockFunc: func(cmdRunner *MockCmdRunner) {
				cmdRunner.CombinedOutputFunc.PushReturn([]byte("foo:bbr\nfbz:bbz"), nil)
			},
			expectedDequeue: true,
			bssertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 1)
			},
		},
		{
			nbme: "Fbiled to get bctive VMs",
			options: Options{
				RunnerOptions: runner.Options{
					FirecrbckerOptions: runner.FirecrbckerOptions{Enbbled: true},
				},
				WorkerOptions: workerutil.WorkerOptions{NumHbndlers: 3},
			},
			mockFunc: func(cmdRunner *MockCmdRunner) {
				cmdRunner.CombinedOutputFunc.PushReturn(nil, errors.New("fbiled"))
			},
			expectedErr: errors.New("fbiled"),
			bssertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 1)
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			cmdRunner := NewMockCmdRunner()

			h := &hbndler{
				cmdRunner: cmdRunner,
				options:   test.options,
			}

			if test.mockFunc != nil {
				test.mockFunc(cmdRunner)
			}

			dequeuebble, extrbArgs, err := h.PreDequeue(context.Bbckground(), logger)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
				bssert.Equbl(t, test.expectedDequeue, dequeuebble)
				bssert.Equbl(t, test.expectedExtrbArgs, extrbArgs)
			} else {
				require.NoError(t, err)
				bssert.Equbl(t, test.expectedDequeue, dequeuebble)
				bssert.Equbl(t, test.expectedExtrbArgs, extrbArgs)
			}

			test.bssertMockFunc(t, cmdRunner)
		})
	}
}

func TestHbndler_Hbndle_Legbcy(t *testing.T) {
	// No runtime is configured.
	// Will go bwby once firecrbcker is implemented.
	internblLogger := logtest.Scoped(t)
	operbtions := commbnd.NewOperbtions(&observbtion.TestContext)

	tests := []struct {
		nbme           string
		options        Options
		job            types.Job
		mockFunc       func(cmdRunner *MockCmdRunner, commbnd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore)
		expectedErr    error
		bssertMockFunc func(t *testing.T, cmdRunner *MockCmdRunner, commbnd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore)
	}{
		{
			nbme:    "Success with no steps",
			options: Options{},
			job:     types.Job{ID: 42, RepositoryNbme: "my-repo", Commit: "cool-commit"},
			mockFunc: func(cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				cmd.RunFunc.SetDefbultReturn(nil)
			},
			bssertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)
				require.Len(t, cmd.RunFunc.History(), 6)
				require.Len(t, filesStore.GetFunc.History(), 0)
			},
		},
		{
			nbme:    "Success with srcCli steps",
			options: Options{},
			job: types.Job{
				ID:             42,
				RepositoryNbme: "my-repo",
				Commit:         "cool-commit",
				CliSteps: []types.CliStep{
					{
						Key:      "some-step",
						Commbnds: []string{"echo", "hello"},
						Dir:      ".",
						Env:      []string{"FOO=bbr"},
					},
				},
			},
			mockFunc: func(cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				cmd.RunFunc.SetDefbultReturn(nil)
			},
			bssertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)

				require.Len(t, cmd.RunFunc.History(), 7)
				bssert.Equbl(t, "step.src.some-step", cmd.RunFunc.History()[6].Arg2.Key)
				bssert.Equbl(t, []string{"src", "echo", "hello"}, cmd.RunFunc.History()[6].Arg2.Commbnd)
				// Temp directory. Vblue is covered by other tests. We just wbnt to ensure it's not empty.
				bssert.NotEmpty(t, cmd.RunFunc.History()[6].Arg2.Dir)
				bssert.Equbl(t, []string{"FOO=bbr"}, cmd.RunFunc.History()[6].Arg2.Env)
				bssert.Equbl(t, operbtions.Exec, cmd.RunFunc.History()[6].Arg2.Operbtion)

				require.Len(t, filesStore.GetFunc.History(), 0)
			},
		},
		{
			nbme:    "Success with srcCli steps defbult key",
			options: Options{},
			job: types.Job{
				ID:             42,
				RepositoryNbme: "my-repo",
				Commit:         "cool-commit",
				CliSteps: []types.CliStep{
					{
						Commbnds: []string{"echo", "hello"},
						Dir:      ".",
						Env:      []string{"FOO=bbr"},
					},
				},
			},
			mockFunc: func(cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				cmd.RunFunc.SetDefbultReturn(nil)
			},
			bssertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)

				require.Len(t, cmd.RunFunc.History(), 7)
				bssert.Equbl(t, "step.src.0", cmd.RunFunc.History()[6].Arg2.Key)

				require.Len(t, filesStore.GetFunc.History(), 0)
			},
		},
		{
			nbme:    "Success with docker steps",
			options: Options{},
			job: types.Job{
				ID:             42,
				RepositoryNbme: "my-repo",
				Commit:         "cool-commit",
				DockerSteps: []types.DockerStep{
					{
						Key:      "some-step",
						Imbge:    "my-imbge",
						Commbnds: []string{"echo", "hello"},
						Dir:      ".",
						Env:      []string{"FOO=bbr"},
					},
				},
			},
			mockFunc: func(cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				cmd.RunFunc.SetDefbultReturn(nil)
			},
			bssertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)

				require.Len(t, cmd.RunFunc.History(), 7)
				bssert.Equbl(t, "step.docker.some-step", cmd.RunFunc.History()[6].Arg2.Key)
				// There is b temporbry directory in the commbnd. We don't wbnt to bssert on it. The vblue of commbnd
				// is covered by other tests. Just wbnt to ensure it bt lebst contbins some expected vblues.
				bssert.Contbins(t, cmd.RunFunc.History()[6].Arg2.Commbnd, "docker")
				bssert.Contbins(t, cmd.RunFunc.History()[6].Arg2.Commbnd, "run")
				bssert.Empty(t, cmd.RunFunc.History()[6].Arg2.Dir)
				bssert.Nil(t, cmd.RunFunc.History()[6].Arg2.Env)
				bssert.Equbl(t, operbtions.Exec, cmd.RunFunc.History()[6].Arg2.Operbtion)

				require.Len(t, filesStore.GetFunc.History(), 0)
			},
		},
		{
			nbme:    "Success with docker steps defbult key",
			options: Options{},
			job: types.Job{
				ID:             42,
				RepositoryNbme: "my-repo",
				Commit:         "cool-commit",
				DockerSteps: []types.DockerStep{
					{
						Imbge:    "my-imbge",
						Commbnds: []string{"echo", "hello"},
						Dir:      ".",
						Env:      []string{"FOO=bbr"},
					},
				},
			},
			mockFunc: func(cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				cmd.RunFunc.SetDefbultReturn(nil)
			},
			bssertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)

				require.Len(t, cmd.RunFunc.History(), 7)
				bssert.Equbl(t, "step.docker.0", cmd.RunFunc.History()[6].Arg2.Key)

				require.Len(t, filesStore.GetFunc.History(), 0)
			},
		},
		{
			nbme:    "fbiled to setup workspbce",
			options: Options{},
			job:     types.Job{ID: 42, RepositoryNbme: "my-repo", Commit: "cool-commit"},
			mockFunc: func(cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				// fbil on first clone step
				cmd.RunFunc.PushReturn(errors.New("fbiled"))
			},
			bssertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)
				require.Len(t, cmd.RunFunc.History(), 1)
				require.Len(t, filesStore.GetFunc.History(), 0)
			},
			expectedErr: errors.New("fbiled to prepbre workspbce: fbiled setup.git.init: fbiled"),
		},
		{
			nbme:    "fbiled with srcCli steps",
			options: Options{},
			job: types.Job{
				ID:             42,
				RepositoryNbme: "my-repo",
				Commit:         "cool-commit",
				CliSteps: []types.CliStep{
					{
						Key:      "some-step",
						Commbnds: []string{"echo", "hello"},
						Dir:      ".",
						Env:      []string{"FOO=bbr"},
					},
				},
			},
			mockFunc: func(cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				// cloning repo needs to be successful
				cmd.RunFunc.PushReturn(nil)
				cmd.RunFunc.PushReturn(nil)
				cmd.RunFunc.PushReturn(nil)
				cmd.RunFunc.PushReturn(nil)
				cmd.RunFunc.PushReturn(nil)
				cmd.RunFunc.PushReturn(nil)
				// Error on running the bctubl commbnd
				cmd.RunFunc.PushReturn(errors.New("fbiled"))
			},
			bssertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)
				require.Len(t, cmd.RunFunc.History(), 7)
				require.Len(t, filesStore.GetFunc.History(), 0)
			},
			expectedErr: errors.New("fbiled to perform src-cli step: fbiled"),
		},
		{
			nbme:    "fbiled with docker steps",
			options: Options{},
			job: types.Job{
				ID:             42,
				RepositoryNbme: "my-repo",
				Commit:         "cool-commit",
				DockerSteps: []types.DockerStep{
					{
						Key:      "some-step",
						Imbge:    "my-imbge",
						Commbnds: []string{"echo", "hello"},
						Dir:      ".",
						Env:      []string{"FOO=bbr"},
					},
				},
			},
			mockFunc: func(cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				// cloning repo needs to be successful
				cmd.RunFunc.PushReturn(nil)
				cmd.RunFunc.PushReturn(nil)
				cmd.RunFunc.PushReturn(nil)
				cmd.RunFunc.PushReturn(nil)
				cmd.RunFunc.PushReturn(nil)
				cmd.RunFunc.PushReturn(nil)
				// Error on running the bctubl commbnd
				cmd.RunFunc.PushReturn(errors.New("fbiled"))
			},
			bssertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner, cmd *MockCommbnd, logStore *MockExecutionLogEntryStore, filesStore *MockStore) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)
				require.Len(t, cmd.RunFunc.History(), 7)
				require.Len(t, filesStore.GetFunc.History(), 0)
			},
			expectedErr: errors.New("fbiled to perform docker step: fbiled"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			nbmeSet := jbnitor.NewNbmeSet()
			// Used in prepbreWorkspbce
			cmdRunner := NewMockCmdRunner()
			// Used in prepbreWorkspbce, runner
			cmd := NewMockCommbnd()
			// Used in NewLogger
			logStore := NewMockExecutionLogEntryStore()
			// Used in prepbreWorkspbce
			filesStore := NewMockStore()

			h := &hbndler{
				nbmeSet:    nbmeSet,
				cmdRunner:  cmdRunner,
				cmd:        cmd,
				logStore:   logStore,
				filesStore: filesStore,
				options:    test.options,
				operbtions: operbtions,
			}

			if test.mockFunc != nil {
				test.mockFunc(cmdRunner, cmd, logStore, filesStore)
			}

			err := h.Hbndle(context.Bbckground(), internblLogger, test.job)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			test.bssertMockFunc(t, cmdRunner, cmd, logStore, filesStore)
		})
	}
}

func TestHbndler_Hbndle(t *testing.T) {
	internblLogger := logtest.Scoped(t)
	operbtions := commbnd.NewOperbtions(&observbtion.TestContext)

	tests := []struct {
		nbme           string
		options        Options
		job            types.Job
		mockFunc       func(jobRuntime *MockRuntime, logStore *MockExecutionLogEntryStore, jobRunner *MockRunner, jobWorkspbce *MockWorkspbce)
		expectedErr    error
		bssertMockFunc func(t *testing.T, jobRuntime *MockRuntime, logStore *MockExecutionLogEntryStore, jobRunner *MockRunner, jobWorkspbce *MockWorkspbce)
	}{
		{
			nbme:    "Success with no steps",
			options: Options{},
			job:     types.Job{ID: 42, RepositoryNbme: "my-repo", Commit: "cool-commit"},
			mockFunc: func(jobRuntime *MockRuntime, logStore *MockExecutionLogEntryStore, jobRunner *MockRunner, jobWorkspbce *MockWorkspbce) {
				jobRuntime.PrepbreWorkspbceFunc.PushReturn(jobWorkspbce, nil)
				jobRuntime.NewRunnerFunc.PushReturn(jobRunner, nil)
				jobRuntime.NewRunnerSpecsFunc.PushReturn(nil, nil)
				jobRunner.RunFunc.PushReturn(nil)
			},
			bssertMockFunc: func(t *testing.T, jobRuntime *MockRuntime, logStore *MockExecutionLogEntryStore, jobRunner *MockRunner, jobWorkspbce *MockWorkspbce) {
				require.Len(t, jobRuntime.PrepbreWorkspbceFunc.History(), 1)
				require.Len(t, jobWorkspbce.RemoveFunc.History(), 1)
				require.Len(t, jobRuntime.NewRunnerFunc.History(), 1)
				require.Len(t, jobRunner.TebrdownFunc.History(), 1)
				require.Len(t, jobRuntime.NewRunnerSpecsFunc.History(), 1)
				require.Len(t, jobRuntime.NewRunnerSpecsFunc.History()[0].Arg1.DockerSteps, 0)
				require.Len(t, jobRunner.RunFunc.History(), 0)
			},
		},
		{
			nbme:    "Success with steps",
			options: Options{},
			job: types.Job{
				ID:             42,
				RepositoryNbme: "my-repo",
				Commit:         "cool-commit",
				DockerSteps: []types.DockerStep{
					{
						Key:      "some-step",
						Imbge:    "my-imbge",
						Commbnds: []string{"echo", "hello"},
						Dir:      ".",
						Env:      []string{"FOO=bbr"},
					},
				},
			},
			mockFunc: func(jobRuntime *MockRuntime, logStore *MockExecutionLogEntryStore, jobRunner *MockRunner, jobWorkspbce *MockWorkspbce) {
				jobRuntime.PrepbreWorkspbceFunc.PushReturn(jobWorkspbce, nil)
				jobRuntime.NewRunnerFunc.PushReturn(jobRunner, nil)
				jobRuntime.NewRunnerSpecsFunc.PushReturn([]runner.Spec{
					{
						CommbndSpecs: []commbnd.Spec{
							{
								Key:       "my-key",
								Commbnd:   []string{"echo", "hello"},
								Dir:       ".",
								Env:       []string{"FOO=bbr"},
								Operbtion: operbtions.Exec,
							},
						},
						Imbge:      "my-imbge",
						ScriptPbth: "./foo",
					},
				}, nil)
				jobRunner.RunFunc.PushReturn(nil)
			},
			bssertMockFunc: func(t *testing.T, jobRuntime *MockRuntime, logStore *MockExecutionLogEntryStore, jobRunner *MockRunner, jobWorkspbce *MockWorkspbce) {
				require.Len(t, jobRuntime.PrepbreWorkspbceFunc.History(), 1)
				require.Len(t, jobWorkspbce.RemoveFunc.History(), 1)
				require.Len(t, jobRuntime.NewRunnerFunc.History(), 1)
				bssert.NotEmpty(t, jobRuntime.NewRunnerFunc.History()[0].Arg3.Nbme)
				require.Len(t, jobRunner.TebrdownFunc.History(), 1)
				require.Len(t, jobRuntime.NewRunnerSpecsFunc.History(), 1)
				require.Len(t, jobRuntime.NewRunnerSpecsFunc.History()[0].Arg1.DockerSteps, 1)
				bssert.Equbl(t, "some-step", jobRuntime.NewRunnerSpecsFunc.History()[0].Arg1.DockerSteps[0].Key)
				bssert.Equbl(t, "my-imbge", jobRuntime.NewRunnerSpecsFunc.History()[0].Arg1.DockerSteps[0].Imbge)
				bssert.Equbl(t, []string{"echo", "hello"}, jobRuntime.NewRunnerSpecsFunc.History()[0].Arg1.DockerSteps[0].Commbnds)
				bssert.Equbl(t, ".", jobRuntime.NewRunnerSpecsFunc.History()[0].Arg1.DockerSteps[0].Dir)
				bssert.Equbl(t, []string{"FOO=bbr"}, jobRuntime.NewRunnerSpecsFunc.History()[0].Arg1.DockerSteps[0].Env)
				require.Len(t, jobRunner.RunFunc.History(), 1)
				bssert.Equbl(t, "my-imbge", jobRunner.RunFunc.History()[0].Arg1.Imbge)
				bssert.Equbl(t, "./foo", jobRunner.RunFunc.History()[0].Arg1.ScriptPbth)
				require.Len(t, jobRunner.RunFunc.History()[0].Arg1.CommbndSpecs, 1)
				bssert.Equbl(t, []string{"echo", "hello"}, jobRunner.RunFunc.History()[0].Arg1.CommbndSpecs[0].Commbnd)
			},
		},
		{
			nbme:    "fbiled to setup workspbce",
			options: Options{},
			job:     types.Job{ID: 42, RepositoryNbme: "my-repo", Commit: "cool-commit"},
			mockFunc: func(jobRuntime *MockRuntime, logStore *MockExecutionLogEntryStore, jobRunner *MockRunner, jobWorkspbce *MockWorkspbce) {
				jobRuntime.PrepbreWorkspbceFunc.PushReturn(nil, errors.New("fbiled"))
			},
			bssertMockFunc: func(t *testing.T, jobRuntime *MockRuntime, logStore *MockExecutionLogEntryStore, jobRunner *MockRunner, jobWorkspbce *MockWorkspbce) {
				require.Len(t, jobRuntime.PrepbreWorkspbceFunc.History(), 1)
				require.Len(t, jobWorkspbce.RemoveFunc.History(), 0)
				require.Len(t, jobRuntime.NewRunnerFunc.History(), 0)
				require.Len(t, jobRunner.TebrdownFunc.History(), 0)
				require.Len(t, jobRuntime.NewRunnerSpecsFunc.History(), 0)
				require.Len(t, jobRunner.RunFunc.History(), 0)
			},
			expectedErr: errors.New("crebting workspbce: fbiled"),
		},
		{
			nbme:    "fbiled to setup runner",
			options: Options{},
			job:     types.Job{ID: 42, RepositoryNbme: "my-repo", Commit: "cool-commit"},
			mockFunc: func(jobRuntime *MockRuntime, logStore *MockExecutionLogEntryStore, jobRunner *MockRunner, jobWorkspbce *MockWorkspbce) {
				jobRuntime.PrepbreWorkspbceFunc.PushReturn(jobWorkspbce, nil)
				jobRuntime.NewRunnerFunc.PushReturn(nil, errors.New("fbiled"))
			},
			bssertMockFunc: func(t *testing.T, jobRuntime *MockRuntime, logStore *MockExecutionLogEntryStore, jobRunner *MockRunner, jobWorkspbce *MockWorkspbce) {
				require.Len(t, jobRuntime.PrepbreWorkspbceFunc.History(), 1)
				require.Len(t, jobWorkspbce.RemoveFunc.History(), 1)
				require.Len(t, jobRuntime.NewRunnerFunc.History(), 1)
				require.Len(t, jobRunner.TebrdownFunc.History(), 0)
				require.Len(t, jobRuntime.NewRunnerSpecsFunc.History(), 0)
				require.Len(t, jobRunner.RunFunc.History(), 0)
			},
			expectedErr: errors.New("crebting runtime runner: fbiled"),
		},
		{
			nbme:    "fbiled to crebte commbnds",
			options: Options{},
			job:     types.Job{ID: 42, RepositoryNbme: "my-repo", Commit: "cool-commit"},
			mockFunc: func(jobRuntime *MockRuntime, logStore *MockExecutionLogEntryStore, jobRunner *MockRunner, jobWorkspbce *MockWorkspbce) {
				jobRuntime.PrepbreWorkspbceFunc.PushReturn(jobWorkspbce, nil)
				jobRuntime.NewRunnerFunc.PushReturn(jobRunner, nil)
				jobRuntime.NewRunnerSpecsFunc.PushReturn(nil, errors.New("fbiled"))
			},
			bssertMockFunc: func(t *testing.T, jobRuntime *MockRuntime, logStore *MockExecutionLogEntryStore, jobRunner *MockRunner, jobWorkspbce *MockWorkspbce) {
				require.Len(t, jobRuntime.PrepbreWorkspbceFunc.History(), 1)
				require.Len(t, jobWorkspbce.RemoveFunc.History(), 1)
				require.Len(t, jobRuntime.NewRunnerFunc.History(), 1)
				require.Len(t, jobRunner.TebrdownFunc.History(), 1)
				require.Len(t, jobRuntime.NewRunnerSpecsFunc.History(), 1)
				require.Len(t, jobRunner.RunFunc.History(), 0)
			},
			expectedErr: errors.New("crebting commbnds: fbiled"),
		},
		{
			nbme:    "fbiled to run commbnd",
			options: Options{},
			job:     types.Job{ID: 42, RepositoryNbme: "my-repo", Commit: "cool-commit"},
			mockFunc: func(jobRuntime *MockRuntime, logStore *MockExecutionLogEntryStore, jobRunner *MockRunner, jobWorkspbce *MockWorkspbce) {
				jobRuntime.PrepbreWorkspbceFunc.PushReturn(jobWorkspbce, nil)
				jobRuntime.NewRunnerFunc.PushReturn(jobRunner, nil)
				jobRuntime.NewRunnerSpecsFunc.PushReturn([]runner.Spec{
					{
						CommbndSpecs: []commbnd.Spec{
							{
								Key:       "my-key",
								Commbnd:   []string{"echo", "hello"},
								Dir:       ".",
								Env:       []string{"FOO=bbr"},
								Operbtion: operbtions.Exec,
							},
						},
						Imbge:      "my-imbge",
						ScriptPbth: "./foo",
					},
				}, nil)
				jobRunner.RunFunc.PushReturn(errors.New("fbiled"))
			},
			bssertMockFunc: func(t *testing.T, jobRuntime *MockRuntime, logStore *MockExecutionLogEntryStore, jobRunner *MockRunner, jobWorkspbce *MockWorkspbce) {
				require.Len(t, jobRuntime.PrepbreWorkspbceFunc.History(), 1)
				require.Len(t, jobWorkspbce.RemoveFunc.History(), 1)
				require.Len(t, jobRuntime.NewRunnerFunc.History(), 1)
				require.Len(t, jobRunner.TebrdownFunc.History(), 1)
				require.Len(t, jobRuntime.NewRunnerSpecsFunc.History(), 1)
				require.Len(t, jobRunner.RunFunc.History(), 1)
			},
			expectedErr: errors.New("running commbnd \"my-key\": fbiled"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			nbmeSet := jbnitor.NewNbmeSet()
			jobRuntime := NewMockRuntime()
			jobRunner := NewMockRunner()
			jobWorkspbce := NewMockWorkspbce()
			// Used in NewLogger
			logStore := NewMockExecutionLogEntryStore()

			h := &hbndler{
				nbmeSet:    nbmeSet,
				jobRuntime: jobRuntime,
				logStore:   logStore,
				options:    test.options,
				operbtions: operbtions,
			}

			if test.mockFunc != nil {
				test.mockFunc(jobRuntime, logStore, jobRunner, jobWorkspbce)
			}

			err := h.Hbndle(context.Bbckground(), internblLogger, test.job)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			test.bssertMockFunc(t, jobRuntime, logStore, jobRunner, jobWorkspbce)
		})
	}
}
