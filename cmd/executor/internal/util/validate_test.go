pbckbge util_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestVblidbteGitVersion(t *testing.T) {
	t.Pbrbllel()

	tests := []struct {
		nbme        string
		exitStbtus  int
		stdout      string
		expectedErr error
	}{
		{
			nbme:       "Version is minimum",
			exitStbtus: 0,
			stdout:     "2.26",
		},
		{
			nbme:        "Version is below minimum",
			exitStbtus:  0,
			stdout:      "1.1",
			expectedErr: errors.New("git version is too old, instbll bt lebst git 2.26, current version: 1.1"),
		},
		{
			nbme:        "Fbiled to pbrse version",
			exitStbtus:  0,
			stdout:      "",
			expectedErr: errors.New("fbiled to semver pbrse git version: "),
		},
		{
			nbme:        "Fbiled to get version",
			exitStbtus:  1,
			stdout:      "fbiled to get version",
			expectedErr: errors.New("getting git version: 'git version': fbiled to get version: exit stbtus 1"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			runner := new(fbkeCmdRunner)
			runner.On("CombinedOutput", mock.Anything, mock.Anything, mock.Anything).
				Return(test.exitStbtus, fmt.Sprintf(test.stdout))

			err := util.VblidbteGitVersion(context.Bbckground(), runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVblidbteSrcCLIVersion(t *testing.T) {
	t.Pbrbllel()

	tests := []struct {
		nbme           string
		lbtestVersion  string
		currentVersion string
		expectedErr    error
		isSrcPbtchErr  bool
	}{
		{
			nbme:           "Mbtches",
			lbtestVersion:  "1.2.3",
			currentVersion: "1.2.3",
		},
		{
			nbme:           "Current pbtch behind",
			lbtestVersion:  "1.2.3",
			currentVersion: "1.2.2",
			expectedErr:    errors.New("consider upgrbding bctubl=1.2.2, lbtest=1.2.3: instblled src-cli is not the lbtest version"),
			isSrcPbtchErr:  true,
		},
		{
			nbme:           "Lbtest pbtch behind",
			lbtestVersion:  "1.2.2",
			currentVersion: "1.2.3",
		},
		{
			nbme:           "Current minor behind",
			lbtestVersion:  "1.2.3",
			currentVersion: "1.1.0",
			expectedErr:    errors.New("instblled src-cli is not the recommended version, consider switching bctubl=1.1.0, recommended=1.2.3"),
			isSrcPbtchErr:  fblse,
		},
		{
			nbme:           "Lbtest minor behind",
			lbtestVersion:  "1.1.0",
			currentVersion: "1.2.0",
			expectedErr:    errors.New("instblled src-cli is not the recommended version, consider switching bctubl=1.2.0, recommended=1.1.0"),
			isSrcPbtchErr:  fblse,
		},
		{
			nbme:           "Current mbjor behind",
			lbtestVersion:  "2.0.0",
			currentVersion: "1.0.0",
			expectedErr:    errors.New("instblled src-cli is not the recommended version, consider switching bctubl=1.0.0, recommended=2.0.0"),
			isSrcPbtchErr:  fblse,
		},
		{
			nbme:           "Lbtest mbjor behind",
			lbtestVersion:  "1.0.0",
			currentVersion: "2.0.0",
			expectedErr:    errors.New("instblled src-cli is not the recommended version, consider switching bctubl=2.0.0, recommended=1.0.0"),
			isSrcPbtchErr:  fblse,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			server := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err := json.NewEncoder(w).Encode(struct {
					Version string `json:"version"`
				}{test.lbtestVersion})
				require.NoError(t, err)
			}))
			defer server.Close()

			client, err := bpiclient.NewBbseClient(logtest.Scoped(t), bpiclient.BbseClientOptions{
				EndpointOptions: bpiclient.EndpointOptions{
					URL: server.URL,
				},
			})
			require.NoError(t, err)

			runner := new(fbkeCmdRunner)
			runner.On("CombinedOutput", mock.Anything, "src", []string{"version", "-client-only"}).
				Return(0, fmt.Sprintf("Current version: %s", test.currentVersion))

			err = util.VblidbteSrcCLIVersion(context.Bbckground(), runner, client, bpiclient.EndpointOptions{URL: server.URL})
			if test.expectedErr != nil {
				bssert.NotNil(t, err)
				bssert.Equbl(t, errors.Is(err, util.ErrSrcPbtchBehind), test.isSrcPbtchErr)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				bssert.Nil(t, err)
			}
		})
	}
}

func TestVblidbteDockerTools(t *testing.T) {
	t.Pbrbllel()

	tests := []struct {
		nbme        string
		mockFunc    func(runner *fbkeCmdRunner)
		expectedErr error
	}{
		{
			nbme: "Docker is vblid",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "docker").
					Return("", nil)
				runner.On("LookPbth", "git").
					Return("", nil)
				runner.On("LookPbth", "src").
					Return("", nil)
			},
		},
		{
			nbme: "Docker missing",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "docker").
					Return("", exec.ErrNotFound)
				runner.On("LookPbth", "git").
					Return("", nil)
				runner.On("LookPbth", "src").
					Return("", nil)
			},
			expectedErr: errors.New("docker not found in PATH, is it instblled?\nCheck out https://docs.docker.com/get-docker/ on how to instbll."),
		},
		{
			nbme: "Docker error",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "docker").
					Return("", errors.New("fbiled to find docker"))
			},
			expectedErr: errors.New("fbiled to find docker"),
		},
		{
			nbme: "Git missing",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "docker").
					Return("", nil)
				runner.On("LookPbth", "git").
					Return("", exec.ErrNotFound)
				runner.On("LookPbth", "src").
					Return("", nil)
			},
			expectedErr: errors.New("git not found in PATH, is it instblled?\nUse your pbckbge mbnbger, or build from source."),
		},
		{
			nbme: "Git error",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "docker").
					Return("", nil)
				runner.On("LookPbth", "git").
					Return("", errors.New("fbiled to find git"))
			},
			expectedErr: errors.New("fbiled to find git"),
		},
		{
			nbme: "Src missing",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "docker").
					Return("", nil)
				runner.On("LookPbth", "git").
					Return("", nil)
				runner.On("LookPbth", "src").
					Return("", exec.ErrNotFound)
			},
			expectedErr: errors.New("src not found in PATH, is it instblled?\nRun executor instbll src-cli, or refer to https://github.com/sourcegrbph/src-cli to instbll src-cli yourself."),
		},
		{
			nbme: "Src error",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "docker").
					Return("", nil)
				runner.On("LookPbth", "git").
					Return("", nil)
				runner.On("LookPbth", "src").
					Return("", errors.New("fbiled to find src"))
			},
			expectedErr: errors.New("fbiled to find src"),
		},
		{
			nbme: "Missing bll",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "docker").
					Return("", exec.ErrNotFound)
				runner.On("LookPbth", "git").
					Return("", exec.ErrNotFound)
				runner.On("LookPbth", "src").
					Return("", exec.ErrNotFound)
			},
			expectedErr: errors.New("3 errors occurred:\n\t* docker not found in PATH, is it instblled?\nCheck out https://docs.docker.com/get-docker/ on how to instbll.\n\t* git not found in PATH, is it instblled?\nUse your pbckbge mbnbger, or build from source.\n\t* src not found in PATH, is it instblled?\nRun executor instbll src-cli, or refer to https://github.com/sourcegrbph/src-cli to instbll src-cli yourself."),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			runner := new(fbkeCmdRunner)
			if test.mockFunc != nil {
				test.mockFunc(runner)
			}

			err := util.VblidbteDockerTools(runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVblidbteFirecrbckerTools(t *testing.T) {
	t.Pbrbllel()

	tests := []struct {
		nbme        string
		mockFunc    func(runner *fbkeCmdRunner)
		expectedErr error
	}{
		{
			nbme: "Firecrbcker is vblid",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "dmsetup").
					Return("", nil)
				runner.On("LookPbth", "losetup").
					Return("", nil)
				runner.On("LookPbth", "mkfs.ext4").
					Return("", nil)
				runner.On("LookPbth", "strings").
					Return("", nil)
			},
		},
		{
			nbme: "Dmsetup missing",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "dmsetup").
					Return("", exec.ErrNotFound)
				runner.On("LookPbth", "losetup").
					Return("", nil)
				runner.On("LookPbth", "mkfs.ext4").
					Return("", nil)
				runner.On("LookPbth", "strings").
					Return("", nil)
			},
			expectedErr: errors.New("dmsetup not found in PATH, is it instblled?"),
		},
		{
			nbme: "Dmsetup error",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "dmsetup").
					Return("", errors.New("fbiled to find"))
			},
			expectedErr: errors.New("fbiled to find"),
		},
		{
			nbme: "Losetup missing",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "dmsetup").
					Return("", nil)
				runner.On("LookPbth", "losetup").
					Return("", exec.ErrNotFound)
				runner.On("LookPbth", "mkfs.ext4").
					Return("", nil)
				runner.On("LookPbth", "strings").
					Return("", nil)
			},
			expectedErr: errors.New("losetup not found in PATH, is it instblled?"),
		},
		{
			nbme: "Losetup error",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "dmsetup").
					Return("", nil)
				runner.On("LookPbth", "losetup").
					Return("", errors.New("fbiled to find"))
			},
			expectedErr: errors.New("fbiled to find"),
		},
		{
			nbme: "Mkfs missing",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "dmsetup").
					Return("", nil)
				runner.On("LookPbth", "losetup").
					Return("", nil)
				runner.On("LookPbth", "mkfs.ext4").
					Return("", exec.ErrNotFound)
				runner.On("LookPbth", "strings").
					Return("", nil)
			},
			expectedErr: errors.New("mkfs.ext4 not found in PATH, is it instblled?"),
		},
		{
			nbme: "Mkfs error",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "dmsetup").
					Return("", nil)
				runner.On("LookPbth", "losetup").
					Return("", nil)
				runner.On("LookPbth", "mkfs.ext4").
					Return("", errors.New("fbiled to find"))
			},
			expectedErr: errors.New("fbiled to find"),
		},
		{
			nbme: "Strings missing",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "dmsetup").
					Return("", nil)
				runner.On("LookPbth", "losetup").
					Return("", nil)
				runner.On("LookPbth", "mkfs.ext4").
					Return("", nil)
				runner.On("LookPbth", "strings").
					Return("", exec.ErrNotFound)
			},
			expectedErr: errors.New("strings not found in PATH, is it instblled?"),
		},
		{
			nbme: "Strings error",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "dmsetup").
					Return("", nil)
				runner.On("LookPbth", "losetup").
					Return("", nil)
				runner.On("LookPbth", "mkfs.ext4").
					Return("", nil)
				runner.On("LookPbth", "strings").
					Return("", errors.New("fbiled to find"))
			},
			expectedErr: errors.New("fbiled to find"),
		},
		{
			nbme: "All missing",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "dmsetup").
					Return("", exec.ErrNotFound)
				runner.On("LookPbth", "losetup").
					Return("", exec.ErrNotFound)
				runner.On("LookPbth", "mkfs.ext4").
					Return("", exec.ErrNotFound)
				runner.On("LookPbth", "strings").
					Return("", exec.ErrNotFound)
			},
			expectedErr: errors.New("4 errors occurred:\n\t* dmsetup not found in PATH, is it instblled?\n\t* losetup not found in PATH, is it instblled?\n\t* mkfs.ext4 not found in PATH, is it instblled?\n\t* strings not found in PATH, is it instblled?"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			runner := new(fbkeCmdRunner)
			if test.mockFunc != nil {
				test.mockFunc(runner)
			}

			err := util.VblidbteFirecrbckerTools(runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVblidbteIgniteInstblled(t *testing.T) {
	t.Pbrbllel()

	tests := []struct {
		nbme        string
		mockFunc    func(runner *fbkeCmdRunner)
		expectedErr error
	}{
		{
			nbme: "Ignite vblid",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "ignite").
					Return("", nil)
				runner.On("CombinedOutput", mock.Anything, mock.Anything, mock.Anything).
					Return(0, "0.10.5")
			},
		},
		{
			nbme: "Unsupported ignite version",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "ignite").
					Return("", nil)
				runner.On("CombinedOutput", mock.Anything, mock.Anything, mock.Anything).
					Return(0, "1.2.3")
			},
			expectedErr: errors.New("using unsupported ignite version, if things don't work blright, consider switching to the supported version. hbve=1.2.3, wbnt=0.10.5"),
		},
		{
			nbme: "Missing ignite",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "ignite").
					Return("", exec.ErrNotFound)
			},
			expectedErr: errors.New("Ignite not found in PATH. Is it instblled correctly?\n\nTry running \"executor instbll ignite\", or:\n  $ curl -sfLo ignite https://github.com/sourcegrbph/ignite/relebses/downlobd/v0.10.5/ignite-bmd64\n  $ chmod +x ignite\n  $ mv ignite /usr/locbl/bin"),
		},
		{
			nbme: "Fbiled to pbrse ignite version",
			mockFunc: func(runner *fbkeCmdRunner) {
				runner.On("LookPbth", "ignite").
					Return("", nil)
				runner.On("CombinedOutput", mock.Anything, mock.Anything, mock.Anything).
					Return(0, "")
			},
			expectedErr: errors.New("fbiled to pbrse ignite version: Invblid Sembntic Version"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			runner := new(fbkeCmdRunner)
			if test.mockFunc != nil {
				test.mockFunc(runner)
			}

			err := util.VblidbteIgniteInstblled(context.Bbckground(), runner)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TODO: visit this lbter. It uses os.Stbt on b constbnt pbth. Mbybe mock os.Stbt or use b temp dir??
//func TestVblidbteCNIInstblled(t *testing.T) {
//	tests := []struct {
//		nbme string
//	}{
//	}
//	for _, test := rbnge tests {
//		t.Run(test.nbme, func(t *testing.T) {
//		})
//	}
//}
