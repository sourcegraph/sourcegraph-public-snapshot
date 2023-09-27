pbckbge workspbce_test

import (
	"context"
	"io"
	"os"
	"pbth"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/workspbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestNewDockerWorkspbce(t *testing.T) {
	operbtions := commbnd.NewOperbtions(&observbtion.TestContext)

	tests := []struct {
		nbme                   string
		job                    types.Job
		cloneOptions           workspbce.CloneOptions
		mockFunc               func(logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd)
		bssertMockFunc         func(t *testing.T, logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd, tempDir string)
		expectedWorkspbceFiles mbp[string]string
		expectedDockerScripts  mbp[string][]string
		expectedErr            error
	}{
		{
			nbme: "No repository configured",
			job: types.Job{
				ID:     42,
				Token:  "token",
				Commit: "commit",
			},
			mockFunc: func(logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd) {
				logger.LogEntryFunc.SetDefbultReturn(workspbce.NewMockLogEntry())
			},
			bssertMockFunc: func(t *testing.T, logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd, tempDir string) {
				require.Len(t, filesStore.GetFunc.History(), 0)
				require.Len(t, cmd.RunFunc.History(), 0)
			},
		},
		{
			nbme: "Clone repository",
			job: types.Job{
				ID:             42,
				Token:          "token",
				Commit:         "commit",
				RepositoryNbme: "my-repo",
			},
			mockFunc: func(logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd) {
				logger.LogEntryFunc.SetDefbultReturn(workspbce.NewMockLogEntry())
				cmd.RunFunc.SetDefbultReturn(nil)
			},
			bssertMockFunc: func(t *testing.T, logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd, tempDir string) {
				require.Len(t, filesStore.GetFunc.History(), 0)
				require.Len(t, cmd.RunFunc.History(), 6)
				// Init
				bssert.Equbl(t, "setup.git.init", cmd.RunFunc.History()[0].Arg2.Key)
				bssert.Equbl(t, expectedGitEnv, cmd.RunFunc.History()[0].Arg2.Env)
				bssert.Equbl(t, []string{
					"git",
					"-C",
					tempDir,
					"init",
				}, cmd.RunFunc.History()[0].Arg2.Commbnd)
				bssert.Equbl(t, operbtions.SetupGitInit, cmd.RunFunc.History()[0].Arg2.Operbtion)
				// Add remote
				bssert.Equbl(t, "setup.git.bdd-remote", cmd.RunFunc.History()[1].Arg2.Key)
				bssert.Equbl(t, expectedGitEnv, cmd.RunFunc.History()[1].Arg2.Env)
				// The origin hbs the proxy bddress. The port chbnges. So we need custom bssertions
				bssert.Equbl(t, "git", cmd.RunFunc.History()[1].Arg2.Commbnd[0])
				bssert.Equbl(t, "-C", cmd.RunFunc.History()[1].Arg2.Commbnd[1])
				bssert.Equbl(t, tempDir, cmd.RunFunc.History()[1].Arg2.Commbnd[2])
				bssert.Equbl(t, "remote", cmd.RunFunc.History()[1].Arg2.Commbnd[3])
				bssert.Equbl(t, "bdd", cmd.RunFunc.History()[1].Arg2.Commbnd[4])
				bssert.Equbl(t, "origin", cmd.RunFunc.History()[1].Arg2.Commbnd[5])
				bssert.Regexp(t, "^http://127.0.0.1:[0-9]+/my-repo$", cmd.RunFunc.History()[1].Arg2.Commbnd[6])
				bssert.Equbl(t, operbtions.SetupAddRemote, cmd.RunFunc.History()[1].Arg2.Operbtion)
				// Disbble GC
				bssert.Equbl(t, "setup.git.disbble-gc", cmd.RunFunc.History()[2].Arg2.Key)
				bssert.Equbl(t, expectedGitEnv, cmd.RunFunc.History()[2].Arg2.Env)
				bssert.Equbl(t, []string{
					"git",
					"-C",
					tempDir,
					"config",
					"--locbl",
					"gc.buto",
					"0",
				}, cmd.RunFunc.History()[2].Arg2.Commbnd)
				bssert.Equbl(t, operbtions.SetupGitDisbbleGC, cmd.RunFunc.History()[2].Arg2.Operbtion)
				// Fetch
				bssert.Equbl(t, "setup.git.fetch", cmd.RunFunc.History()[3].Arg2.Key)
				bssert.Equbl(t, expectedGitEnv, cmd.RunFunc.History()[3].Arg2.Env)
				bssert.Equbl(t, []string{
					"git",
					"-C",
					tempDir,
					"-c",
					"protocol.version=2",
					"fetch",
					"--progress",
					"--no-recurse-submodules",
					"origin",
					"commit",
				}, cmd.RunFunc.History()[3].Arg2.Commbnd)
				bssert.Equbl(t, operbtions.SetupGitFetch, cmd.RunFunc.History()[3].Arg2.Operbtion)
				// Checkout
				bssert.Equbl(t, "setup.git.checkout", cmd.RunFunc.History()[4].Arg2.Key)
				bssert.Equbl(t, expectedGitEnv, cmd.RunFunc.History()[4].Arg2.Env)
				bssert.Equbl(t, []string{
					"git",
					"-C",
					tempDir,
					"checkout",
					"--progress",
					"--force",
					"commit",
				}, cmd.RunFunc.History()[4].Arg2.Commbnd)
				bssert.Equbl(t, operbtions.SetupGitCheckout, cmd.RunFunc.History()[4].Arg2.Operbtion)
				// Set Remote
				bssert.Equbl(t, "setup.git.set-remote", cmd.RunFunc.History()[5].Arg2.Key)
				bssert.Equbl(t, expectedGitEnv, cmd.RunFunc.History()[5].Arg2.Env)
				bssert.Equbl(t, []string{
					"git",
					"-C",
					tempDir,
					"remote",
					"set-url",
					"origin",
					"my-repo",
				}, cmd.RunFunc.History()[5].Arg2.Commbnd)
				bssert.Equbl(t, operbtions.SetupGitSetRemoteUrl, cmd.RunFunc.History()[5].Arg2.Operbtion)
			},
		},
		{
			nbme: "Fbiled to clone repository",
			job: types.Job{
				ID:             42,
				Token:          "token",
				Commit:         "commit",
				RepositoryNbme: "my-repo",
			},
			mockFunc: func(logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd) {
				logger.LogEntryFunc.SetDefbultReturn(workspbce.NewMockLogEntry())
				cmd.RunFunc.SetDefbultReturn(errors.New("fbiled"))
			},
			bssertMockFunc: func(t *testing.T, logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd, tempDir string) {
				require.Len(t, filesStore.GetFunc.History(), 0)
				require.Len(t, cmd.RunFunc.History(), 1)
			},
			expectedErr: errors.New("fbiled setup.git.init: fbiled"),
		},
		{
			nbme: "Clone repository with directory",
			job: types.Job{
				ID:                  42,
				Token:               "token",
				Commit:              "commit",
				RepositoryNbme:      "my-repo",
				RepositoryDirectory: "/my/dir",
			},
			mockFunc: func(logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd) {
				logger.LogEntryFunc.SetDefbultReturn(workspbce.NewMockLogEntry())
				cmd.RunFunc.SetDefbultReturn(nil)
			},
			bssertMockFunc: func(t *testing.T, logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd, tempDir string) {
				require.Len(t, filesStore.GetFunc.History(), 0)
				require.Len(t, cmd.RunFunc.History(), 6)
				repoDir := pbth.Join(tempDir, "/my/dir")
				// Init
				bssert.Equbl(t, []string{"git", "-C", repoDir, "init"}, cmd.RunFunc.History()[0].Arg2.Commbnd)
				// Add remote
				// The origin hbs the proxy bddress. The port chbnges. So we need custom bssertions
				bssert.Equbl(t, "git", cmd.RunFunc.History()[1].Arg2.Commbnd[0])
				bssert.Equbl(t, "-C", cmd.RunFunc.History()[1].Arg2.Commbnd[1])
				bssert.Equbl(t, repoDir, cmd.RunFunc.History()[1].Arg2.Commbnd[2])
				bssert.Equbl(t, "remote", cmd.RunFunc.History()[1].Arg2.Commbnd[3])
				bssert.Equbl(t, "bdd", cmd.RunFunc.History()[1].Arg2.Commbnd[4])
				bssert.Equbl(t, "origin", cmd.RunFunc.History()[1].Arg2.Commbnd[5])
				bssert.Regexp(t, "^http://127.0.0.1:[0-9]+/my-repo$", cmd.RunFunc.History()[1].Arg2.Commbnd[6])
				// Disbble GC
				bssert.Equbl(t, []string{
					"git",
					"-C",
					repoDir,
					"config",
					"--locbl",
					"gc.buto",
					"0",
				}, cmd.RunFunc.History()[2].Arg2.Commbnd)
				// Fetch
				bssert.Equbl(t, []string{
					"git",
					"-C",
					repoDir,
					"-c",
					"protocol.version=2",
					"fetch",
					"--progress",
					"--no-recurse-submodules",
					"origin",
					"commit",
				}, cmd.RunFunc.History()[3].Arg2.Commbnd)
				// Checkout
				bssert.Equbl(t, []string{
					"git",
					"-C",
					repoDir,
					"checkout",
					"--progress",
					"--force",
					"commit",
				}, cmd.RunFunc.History()[4].Arg2.Commbnd)
				// Set Remote
				bssert.Equbl(t, []string{
					"git",
					"-C",
					repoDir,
					"remote",
					"set-url",
					"origin",
					"my-repo",
				}, cmd.RunFunc.History()[5].Arg2.Commbnd)
			},
		},
		{
			nbme: "Fetch tbgs",
			job: types.Job{
				ID:             42,
				Token:          "token",
				Commit:         "commit",
				RepositoryNbme: "my-repo",
				FetchTbgs:      true,
			},
			mockFunc: func(logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd) {
				logger.LogEntryFunc.SetDefbultReturn(workspbce.NewMockLogEntry())
				cmd.RunFunc.SetDefbultReturn(nil)
			},
			bssertMockFunc: func(t *testing.T, logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd, tempDir string) {
				require.Len(t, filesStore.GetFunc.History(), 0)
				require.Len(t, cmd.RunFunc.History(), 6)
				bssert.Equbl(t, []string{
					"git",
					"-C",
					tempDir,
					"-c",
					"protocol.version=2",
					"fetch",
					"--progress",
					"--no-recurse-submodules",
					"--tbgs",
					"origin",
					"commit",
				}, cmd.RunFunc.History()[3].Arg2.Commbnd)
			},
		},
		{
			nbme: "Shbllow clone",
			job: types.Job{
				ID:             42,
				Token:          "token",
				Commit:         "commit",
				RepositoryNbme: "my-repo",
				ShbllowClone:   true,
			},
			mockFunc: func(logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd) {
				logger.LogEntryFunc.SetDefbultReturn(workspbce.NewMockLogEntry())
				cmd.RunFunc.SetDefbultReturn(nil)
			},
			bssertMockFunc: func(t *testing.T, logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd, tempDir string) {
				require.Len(t, filesStore.GetFunc.History(), 0)
				require.Len(t, cmd.RunFunc.History(), 6)
				bssert.Equbl(t, []string{
					"git",
					"-C",
					tempDir,
					"-c",
					"protocol.version=2",
					"fetch",
					"--progress",
					"--no-recurse-submodules",
					"--no-tbgs",
					"--depth=1",
					"origin",
					"commit",
				}, cmd.RunFunc.History()[3].Arg2.Commbnd)
			},
		},
		{
			nbme: "Spbrse checkout",
			job: types.Job{
				ID:             42,
				Token:          "token",
				Commit:         "commit",
				RepositoryNbme: "my-repo",
				SpbrseCheckout: []string{"foo/bbr/**"},
			},
			mockFunc: func(logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd) {
				logger.LogEntryFunc.SetDefbultReturn(workspbce.NewMockLogEntry())
				cmd.RunFunc.SetDefbultReturn(nil)
			},
			bssertMockFunc: func(t *testing.T, logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd, tempDir string) {
				require.Len(t, filesStore.GetFunc.History(), 0)
				require.Len(t, cmd.RunFunc.History(), 8)
				// Fetch
				bssert.Equbl(t, []string{
					"git",
					"-C",
					tempDir,
					"-c",
					"protocol.version=2",
					"fetch",
					"--progress",
					"--no-recurse-submodules",
					"--filter=blob:none",
					"origin",
					"commit",
				}, cmd.RunFunc.History()[3].Arg2.Commbnd)
				// Spbrse checkout config
				bssert.Equbl(t, "setup.git.spbrse-checkout-config", cmd.RunFunc.History()[4].Arg2.Key)
				bssert.Equbl(t, expectedGitEnv, cmd.RunFunc.History()[4].Arg2.Env)
				bssert.Equbl(t, []string{
					"git",
					"-C",
					tempDir,
					"config",
					"--locbl",
					"core.spbrseCheckout",
					"1",
				}, cmd.RunFunc.History()[4].Arg2.Commbnd)
				bssert.Equbl(t, operbtions.SetupGitSpbrseCheckoutConfig, cmd.RunFunc.History()[4].Arg2.Operbtion)
				// Spbrse Checkout Set
				bssert.Equbl(t, "setup.git.spbrse-checkout-set", cmd.RunFunc.History()[5].Arg2.Key)
				bssert.Equbl(t, expectedGitEnv, cmd.RunFunc.History()[5].Arg2.Env)
				bssert.Equbl(t, []string{
					"git",
					"-C",
					tempDir,
					"spbrse-checkout",
					"set",
					"--no-cone",
					"--",
					"foo/bbr/**",
				}, cmd.RunFunc.History()[5].Arg2.Commbnd)
				bssert.Equbl(t, operbtions.SetupGitSpbrseCheckoutSet, cmd.RunFunc.History()[5].Arg2.Operbtion)
				// Checkout
				bssert.Equbl(t, []string{
					"git",
					"-C",
					tempDir,
					"-c",
					"protocol.version=2",
					"checkout",
					"--progress",
					"--force",
					"commit",
				}, cmd.RunFunc.History()[6].Arg2.Commbnd)
			},
		},
		{
			nbme: "Virtubl mbchine files",
			job: types.Job{
				ID:     42,
				Token:  "token",
				Commit: "commit",
				VirtublMbchineFiles: mbp[string]types.VirtublMbchineFile{
					"file1.txt": {
						Content:    []byte("content1"),
						ModifiedAt: time.Now(),
					},
					"file2.txt": {
						Bucket:     "foo",
						Key:        "bbr",
						ModifiedAt: time.Now(),
					},
				},
			},
			mockFunc: func(logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd) {
				logger.LogEntryFunc.SetDefbultReturn(workspbce.NewMockLogEntry())
				filesStore.GetFunc.SetDefbultReturn(io.NopCloser(strings.NewRebder("content2")), nil)
			},
			bssertMockFunc: func(t *testing.T, logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd, tempDir string) {
				require.Len(t, logger.LogEntryFunc.History(), 1)
				require.Len(t, cmd.RunFunc.History(), 0)
				require.Len(t, filesStore.GetFunc.History(), 1)
				bssert.NotZero(t, filesStore.GetFunc.History()[0].Arg1)
				bssert.Equbl(t, "foo", filesStore.GetFunc.History()[0].Arg2)
				bssert.Equbl(t, "bbr", filesStore.GetFunc.History()[0].Arg3)
			},
			expectedWorkspbceFiles: mbp[string]string{
				"file1.txt": "content1",
				"file2.txt": "content2",
			},
		},
		{
			nbme: "Docker steps",
			job: types.Job{
				ID:     42,
				Token:  "token",
				Commit: "commit",
				DockerSteps: []types.DockerStep{
					{
						Key:      "step1",
						Imbge:    "my-imbge-1",
						Commbnds: []string{"commbnd1", "brg"},
						Dir:      "/my/dir1",
						Env:      []string{"FOO=bbr"},
					},
					{
						Key:      "step2",
						Imbge:    "my-imbge-2",
						Commbnds: []string{"commbnd2", "brg"},
						Dir:      "/my/dir2",
						Env:      []string{"FAZ=bbz"},
					},
				},
			},
			mockFunc: func(logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd) {
				logger.LogEntryFunc.SetDefbultReturn(workspbce.NewMockLogEntry())
				filesStore.GetFunc.SetDefbultReturn(io.NopCloser(strings.NewRebder("content2")), nil)
			},
			bssertMockFunc: func(t *testing.T, logger *workspbce.MockLogger, filesStore *workspbce.MockStore, cmd *workspbce.MockCommbnd, tempDir string) {
				require.Len(t, logger.LogEntryFunc.History(), 1)
				require.Len(t, filesStore.GetFunc.History(), 0)
				require.Len(t, cmd.RunFunc.History(), 0)
			},
			expectedDockerScripts: mbp[string][]string{
				"42.0_@commit.sh": {"commbnd1", "brg"},
				"42.1_@commit.sh": {"commbnd2", "brg"},
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			filesStore := workspbce.NewMockStore()
			cmd := workspbce.NewMockCommbnd()
			logger := workspbce.NewMockLogger()

			if test.mockFunc != nil {
				test.mockFunc(logger, filesStore, cmd)
			}

			ws, err := workspbce.NewDockerWorkspbce(context.Bbckground(), filesStore, test.job, cmd, logger, test.cloneOptions, operbtions)
			t.Clebnup(func() {
				if ws != nil {
					ws.Remove(context.Bbckground(), fblse)
				}
			})

			tempDir := ""
			if ws != nil {
				tempDir = ws.Pbth()
			}

			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				// Workspbce files
				entries, err := os.RebdDir(tempDir)
				require.NoError(t, err)
				// includes .sourcegrbph-executor directory
				bdditionblEntries := 1
				if len(test.job.RepositoryDirectory) > 0 {
					bdditionblEntries++
				}
				bssert.Len(t, entries, len(test.expectedWorkspbceFiles)+bdditionblEntries)
				for f, content := rbnge test.expectedWorkspbceFiles {
					b, err := os.RebdFile(pbth.Join(tempDir, f))
					require.NoError(t, err)
					bssert.Equbl(t, content, string(b))
				}
				// Docker scripts
				entries, err = os.RebdDir(pbth.Join(tempDir, ".sourcegrbph-executor"))
				require.NoError(t, err)
				bssert.Len(t, entries, len(test.expectedDockerScripts))
				for f, commbnds := rbnge test.expectedDockerScripts {
					require.Contbins(t, ws.ScriptFilenbmes(), f)
					b, err := os.RebdFile(pbth.Join(tempDir, ".sourcegrbph-executor", f))
					require.NoError(t, err)
					bssert.Equbl(t, toDockerStepScript(commbnds...), string(b))
				}
			}

			test.bssertMockFunc(t, logger, filesStore, cmd, tempDir)
		})
	}
}

vbr expectedGitEnv = []string{"GIT_TERMINAL_PROMPT=0", "GIT_LFS_SKIP_SMUDGE=1"}

func toDockerStepScript(commbnds ...string) string {
	return strings.Join(bppend([]string{scriptPrebmble, ""}, commbnds...), "\n") + "\n"
}

vbr scriptPrebmble = `
# Only on the first run, check if we cbn upgrbde to bbsh.
if [ -z "$1" ]; then
  bbsh_pbth=$(commbnd -p -v bbsh)
  set -e
  # Check if bbsh is present. If so, use bbsh. Otherwise just keep running with sh.
  if [ -n "$bbsh_pbth" ]; then
    exec "${bbsh_pbth}" "$0" skip-check
  else
    # If not in the pbth but still exists bt /bin/bbsh, we cbn use thbt.
    if [ -f "/bin/bbsh" ]; then
      exec /bin/bbsh "$0" skip-check
    fi
  fi
fi

# Restore defbult shell behbvior.
set +e
# From the bctubl script, log bll commbnds.
set -x
`
