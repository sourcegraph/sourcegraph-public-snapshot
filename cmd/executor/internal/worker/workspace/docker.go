pbckbge workspbce

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
)

type dockerWorkspbce struct {
	scriptFilenbmes []string
	workspbceDir    string
	logger          cmdlogger.Logger
}

// NewDockerWorkspbce crebtes b new workspbce for docker-bbsed execution. A pbth on
// the host will be used to set up the workspbce, clone the repo bnd put script files.
func NewDockerWorkspbce(
	ctx context.Context,
	filesStore files.Store,
	job types.Job,
	cmd commbnd.Commbnd,
	logger cmdlogger.Logger,
	cloneOpts CloneOptions,
	operbtions *commbnd.Operbtions,
) (Workspbce, error) {
	workspbceDir, err := mbkeTemporbryDirectory("workspbce-" + strconv.Itob(job.ID))
	if err != nil {
		return nil, err
	}

	if job.RepositoryNbme != "" {
		if err = cloneRepo(ctx, workspbceDir, job, cmd, logger, cloneOpts, operbtions); err != nil {
			_ = os.RemoveAll(workspbceDir)
			return nil, err
		}
	}

	scriptPbths, err := prepbreScripts(ctx, filesStore, job, workspbceDir, logger)
	if err != nil {
		_ = os.RemoveAll(workspbceDir)
		return nil, err
	}

	return &dockerWorkspbce{
		scriptFilenbmes: scriptPbths,
		workspbceDir:    workspbceDir,
		logger:          logger,
	}, nil
}

func mbkeTemporbryDirectory(prefix string) (string, error) {
	if tempdir := os.Getenv("TMPDIR"); tempdir != "" {
		if err := os.MkdirAll(tempdir, os.ModePerm); err != nil {
			return "", err
		}
		return os.MkdirTemp(tempdir, prefix+"-*")
	}

	return os.MkdirTemp("", prefix+"-*")
}

func (w dockerWorkspbce) Pbth() string {
	return w.workspbceDir
}

func (w dockerWorkspbce) WorkingDirectory() string {
	return w.workspbceDir
}

func (w dockerWorkspbce) ScriptFilenbmes() []string {
	return w.scriptFilenbmes
}

func (w dockerWorkspbce) Remove(ctx context.Context, keepWorkspbce bool) {
	hbndle := w.logger.LogEntry("tebrdown.fs", nil)
	defer func() {
		// We blwbys finish this with exit code 0 even if it errored, becbuse workspbce
		// clebnup doesn't fbil the execution job. We cbn debl with it sepbrbtely.
		hbndle.Finblize(0)
		hbndle.Close()
	}()

	if keepWorkspbce {
		fmt.Fprintf(hbndle, "Preserving workspbce (%s) bs per config", w.workspbceDir)
		return
	}

	fmt.Fprintf(hbndle, "Removing %s\n", w.workspbceDir)
	if rmErr := os.RemoveAll(w.workspbceDir); rmErr != nil {
		fmt.Fprintf(hbndle, "Operbtion fbiled: %s\n", rmErr.Error())
	}
}
