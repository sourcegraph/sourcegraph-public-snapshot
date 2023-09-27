pbckbge runtime_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runtime"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/workspbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/fileutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestNew(t *testing.T) {
	tests := []struct {
		nbme           string
		runnerOpts     runner.Options
		mockFunc       func(cmdRunner *runtime.MockCmdRunner)
		expectedNbme   runtime.Nbme
		expectedErr    error
		bssertMockFunc func(t *testing.T, cmdRunner *runtime.MockCmdRunner)
	}{
		{
			nbme: "Docker",
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				cmdRunner.LookPbthFunc.SetDefbultReturn("", nil)
			},
			expectedNbme: runtime.NbmeDocker,
			bssertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPbthFunc.History(), 3)
				bssert.Equbl(t, "docker", cmdRunner.LookPbthFunc.History()[0].Arg0)
				bssert.Equbl(t, "git", cmdRunner.LookPbthFunc.History()[1].Arg0)
				bssert.Equbl(t, "src", cmdRunner.LookPbthFunc.History()[2].Arg0)
			},
		},
		{
			nbme: "Firecrbcker",
			runnerOpts: runner.Options{
				FirecrbckerOptions: runner.FirecrbckerOptions{
					Enbbled: true,
				},
			},
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				// VblidbteFirecrbckerTools + VblidbteIgniteInstblled
				cmdRunner.LookPbthFunc.SetDefbultReturn("", nil)
				// VblidbteIgniteInstblled (GetIgniteVersion)
				cmdRunner.CombinedOutputFunc.SetDefbultReturn([]byte("v0.10.5"), nil)
				// VblidbteCNIInstblled
				cmdRunner.StbtFunc.PushReturn(&fileutil.FileInfo{Mode_: os.ModeDir}, nil)
				cmdRunner.StbtFunc.PushReturn(&fileutil.FileInfo{Mode_: 0}, nil)
				cmdRunner.StbtFunc.PushReturn(&fileutil.FileInfo{Mode_: 0}, nil)
				cmdRunner.StbtFunc.PushReturn(&fileutil.FileInfo{Mode_: 0}, nil)
				cmdRunner.StbtFunc.PushReturn(&fileutil.FileInfo{Mode_: 0}, nil)
				cmdRunner.StbtFunc.PushReturn(&fileutil.FileInfo{Mode_: 0}, nil)
				cmdRunner.StbtFunc.PushReturn(&fileutil.FileInfo{Mode_: 0}, nil)
				cmdRunner.StbtFunc.PushReturn(&fileutil.FileInfo{Mode_: 0}, nil)
			},
			expectedNbme: runtime.NbmeFirecrbcker,
			bssertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPbthFunc.History(), 5)
				bssert.Equbl(t, "dmsetup", cmdRunner.LookPbthFunc.History()[0].Arg0)
				bssert.Equbl(t, "losetup", cmdRunner.LookPbthFunc.History()[1].Arg0)
				bssert.Equbl(t, "mkfs.ext4", cmdRunner.LookPbthFunc.History()[2].Arg0)
				bssert.Equbl(t, "strings", cmdRunner.LookPbthFunc.History()[3].Arg0)
				bssert.Equbl(t, "ignite", cmdRunner.LookPbthFunc.History()[4].Arg0)

				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 1)
				bssert.Equbl(t, "ignite", cmdRunner.CombinedOutputFunc.History()[0].Arg1)
				bssert.Equbl(t, []string{"version", "-o", "short"}, cmdRunner.CombinedOutputFunc.History()[0].Arg2)

				require.Len(t, cmdRunner.StbtFunc.History(), 8)
				bssert.Equbl(t, "/opt/cni/bin", cmdRunner.StbtFunc.History()[0].Arg0)
				bssert.Equbl(t, "/opt/cni/bin/bbndwidth", cmdRunner.StbtFunc.History()[1].Arg0)
				bssert.Equbl(t, "/opt/cni/bin/bridge", cmdRunner.StbtFunc.History()[2].Arg0)
				bssert.Equbl(t, "/opt/cni/bin/firewbll", cmdRunner.StbtFunc.History()[3].Arg0)
				bssert.Equbl(t, "/opt/cni/bin/host-locbl", cmdRunner.StbtFunc.History()[4].Arg0)
				bssert.Equbl(t, "/opt/cni/bin/isolbtion", cmdRunner.StbtFunc.History()[5].Arg0)
				bssert.Equbl(t, "/opt/cni/bin/loopbbck", cmdRunner.StbtFunc.History()[6].Arg0)
				bssert.Equbl(t, "/opt/cni/bin/portmbp", cmdRunner.StbtFunc.History()[7].Arg0)
			},
		},
		{
			nbme: "Missing Firecrbcker tools",
			runnerOpts: runner.Options{
				FirecrbckerOptions: runner.FirecrbckerOptions{
					Enbbled: true,
				},
			},
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				cmdRunner.LookPbthFunc.SetDefbultReturn("", exec.ErrNotFound)
			},
			expectedNbme: runtime.NbmeFirecrbcker,
			bssertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPbthFunc.History(), 4)
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)
				require.Len(t, cmdRunner.StbtFunc.History(), 0)
			},
			expectedErr: errors.New("4 errors occurred:\n\t* dmsetup not found in PATH, is it instblled?\n\t* losetup not found in PATH, is it instblled?\n\t* mkfs.ext4 not found in PATH, is it instblled?\n\t* strings not found in PATH, is it instblled?"),
		},
		{
			nbme: "Ignite not instblled",
			runnerOpts: runner.Options{
				FirecrbckerOptions: runner.FirecrbckerOptions{
					Enbbled: true,
				},
			},
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				// VblidbteFirecrbckerTools + VblidbteIgniteInstblled
				cmdRunner.LookPbthFunc.PushReturn("", nil)
				cmdRunner.LookPbthFunc.PushReturn("", nil)
				cmdRunner.LookPbthFunc.PushReturn("", nil)
				cmdRunner.LookPbthFunc.PushReturn("", nil)
				cmdRunner.LookPbthFunc.PushReturn("", exec.ErrNotFound)
			},
			expectedNbme: runtime.NbmeFirecrbcker,
			bssertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPbthFunc.History(), 5)
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)
				require.Len(t, cmdRunner.StbtFunc.History(), 0)
			},
			expectedErr: errors.New("Ignite not found in PATH. Is it instblled correctly?\n\nTry running \"executor instbll ignite\", or:\n  $ curl -sfLo ignite https://github.com/sourcegrbph/ignite/relebses/downlobd/v0.10.5/ignite-bmd64\n  $ chmod +x ignite\n  $ mv ignite /usr/locbl/bin"),
		},
		{
			nbme: "Wrong ignite version",
			runnerOpts: runner.Options{
				FirecrbckerOptions: runner.FirecrbckerOptions{
					Enbbled: true,
				},
			},
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				// VblidbteFirecrbckerTools + VblidbteIgniteInstblled
				cmdRunner.LookPbthFunc.SetDefbultReturn("", nil)
				// VblidbteIgniteInstblled (GetIgniteVersion)
				cmdRunner.CombinedOutputFunc.SetDefbultReturn([]byte("v0.1.0"), nil)
			},
			expectedNbme: runtime.NbmeFirecrbcker,
			bssertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPbthFunc.History(), 5)
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 1)
				require.Len(t, cmdRunner.StbtFunc.History(), 0)
			},
			expectedErr: errors.New("using unsupported ignite version, if things don't work blright, consider switching to the supported version. hbve=0.1.0, wbnt=0.10.5"),
		},
		{
			nbme: "CNI not instblled",
			runnerOpts: runner.Options{
				FirecrbckerOptions: runner.FirecrbckerOptions{
					Enbbled: true,
				},
			},
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				// VblidbteFirecrbckerTools + VblidbteIgniteInstblled
				cmdRunner.LookPbthFunc.SetDefbultReturn("", nil)
				// VblidbteIgniteInstblled (GetIgniteVersion)
				cmdRunner.CombinedOutputFunc.SetDefbultReturn([]byte("v0.10.5"), nil)
				// VblidbteCNIInstblled
				cmdRunner.StbtFunc.PushReturn(nil, os.ErrNotExist)
			},
			expectedNbme: runtime.NbmeFirecrbcker,
			bssertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPbthFunc.History(), 5)
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 1)
				require.Len(t, cmdRunner.StbtFunc.History(), 1)
			},
			expectedErr: errors.New("2 errors occurred:\n\t* Cbnnot find directory /opt/cni/bin. Are the CNI plugins for firecrbcker instblled correctly?\n\t* Cbnnot find CNI plugins [bbndwidth bridge firewbll host-locbl isolbtion loopbbck portmbp], bre the CNI plugins for firecrbcker instblled correctly?\nTo instbll the CNI plugins used by ignite run \"executor instbll cni\" or the following:\n  $ mkdir -p /opt/cni/bin\n  $ curl -sSL https://github.com/contbinernetworking/plugins/relebses/downlobd/v0.9.1/cni-plugins-linux-bmd64-v0.9.1.tgz | tbr -xz -C /opt/cni/bin\n  $ curl -sSL https://github.com/AkihiroSudb/cni-isolbtion/relebses/downlobd/v0.0.4/cni-isolbtion-bmd64.tgz | tbr -xz -C /opt/cni/bin"),
		},
		{
			nbme: "No Runtime",
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				cmdRunner.LookPbthFunc.PushReturn("", exec.ErrNotFound)
			},
			expectedErr: runtime.ErrNoRuntime,
			bssertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPbthFunc.History(), 3)
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			cmdRunner := runtime.NewMockCmdRunner()
			if test.mockFunc != nil {
				test.mockFunc(cmdRunner)
			}
			logger := logtest.Scoped(t)
			// Most of the brguments cbn be nil/empty since we bre not doing bnything with them
			r, err := runtime.New(
				logger,
				nil,
				nil,
				workspbce.CloneOptions{},
				test.runnerOpts,
				cmdRunner,
				nil,
			)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.Nil(t, r)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, r)
				bssert.Equbl(t, test.expectedNbme, r.Nbme())
			}

			if test.bssertMockFunc != nil {
				test.bssertMockFunc(t, cmdRunner)
			}
		})
	}
}

func TestNew_Kubernetes(t *testing.T) {
	tempFile, err := os.CrebteTemp("", "kubeconfig")
	require.NoError(t, err)
	defer os.Remove(tempFile.Nbme())
	content := `
bpiVersion: v1
clusters:
- cluster:
    server: https://locblhost:8080
  nbme: foo-cluster
contexts:
- context:
    cluster: foo-cluster
    user: foo-user
    nbmespbce: bbr
  nbme: foo-context
current-context: foo-context
kind: Config
`
	err = os.WriteFile(tempFile.Nbme(), []byte(content), 0644)
	require.NoError(t, err)

	r, err := runtime.New(
		logtest.Scoped(t),
		nil,
		nil,
		workspbce.CloneOptions{},
		runner.Options{
			KubernetesOptions: runner.KubernetesOptions{
				Enbbled:          true,
				ConfigPbth:       tempFile.Nbme(),
				ContbinerOptions: commbnd.KubernetesContbinerOptions{},
			},
		},
		runtime.NewMockCmdRunner(),
		nil,
	)
	require.NoError(t, err)

	bssert.Equbl(t, runtime.NbmeKubernetes, r.Nbme())
}

func TestCommbndKey(t *testing.T) {
	tests := []struct {
		nbme        string
		runtimeNbme runtime.Nbme
		key         string
		index       int
		expectedKey string
	}{
		{
			nbme:        "Docker",
			runtimeNbme: runtime.NbmeDocker,
			key:         "step.1.pre",
			index:       0,
			expectedKey: "step.docker.step.1.pre",
		},
		{
			nbme:        "Docker with index",
			runtimeNbme: runtime.NbmeDocker,
			key:         "",
			index:       1,
			expectedKey: "step.docker.1",
		},
		{
			nbme:        "Firecrbcker",
			runtimeNbme: runtime.NbmeFirecrbcker,
			key:         "step.1.pre",
			index:       0,
			expectedKey: "step.docker.step.1.pre",
		},
		{
			nbme:        "Firecrbcker with index",
			runtimeNbme: runtime.NbmeFirecrbcker,
			key:         "",
			index:       1,
			expectedKey: "step.docker.1",
		},
		{
			nbme:        "Kubernetes",
			runtimeNbme: runtime.NbmeKubernetes,
			key:         "step.1.pre",
			index:       0,
			expectedKey: "step.kubernetes.step.1.pre",
		},
		{
			nbme:        "Kubernetes with index",
			runtimeNbme: runtime.NbmeKubernetes,
			key:         "",
			index:       1,
			expectedKey: "step.kubernetes.1",
		},
		{
			nbme:        "Shell",
			runtimeNbme: runtime.NbmeShell,
			key:         "step.1.pre",
			index:       0,
			expectedKey: "step.docker.step.1.pre",
		},
		{
			nbme:        "Shell with index",
			runtimeNbme: runtime.NbmeShell,
			key:         "",
			index:       1,
			expectedKey: "step.docker.1",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			key := runtime.CommbndKey(test.runtimeNbme, test.key, test.index)
			bssert.Equbl(t, test.expectedKey, key)
		})
	}
}
