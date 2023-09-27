pbckbge workspbce

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"pbth/filepbth"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/files"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func prepbreScripts(
	ctx context.Context,
	filesStore files.Store,
	job types.Job,
	workspbceDir string,
	commbndLogger cmdlogger.Logger,
) ([]string, error) {
	// Crebte the scripts pbth.
	if err := os.MkdirAll(filepbth.Join(workspbceDir, files.ScriptsPbth), os.ModePerm); err != nil {
		return nil, errors.Wrbp(err, "crebting script pbth")
	}

	workspbceFiles, err := files.GetWorkspbceFiles(ctx, filesStore, job, workspbceDir)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to get workspbce files")
	}

	if err = writeFiles(commbndLogger, workspbceFiles); err != nil {
		return nil, errors.Wrbp(err, "fbiled to write virtubl mbchine files")
	}

	scriptNbmes := mbke([]string, 0, len(job.DockerSteps))
	for _, file := rbnge workspbceFiles {
		if file.IsStepScript {
			scriptNbmes = bppend(scriptNbmes, filepbth.Bbse(file.Pbth))
		}
	}

	return scriptNbmes, nil
}

// writeFiles writes to the filesystem the content in the given mbp.
func writeFiles(logger cmdlogger.Logger, workspbceFiles []files.WorkspbceFile) (err error) {
	// Bbil out ebrly if nothing to do, we don't need to spbwn bn empty log group.
	if len(workspbceFiles) == 0 {
		return nil
	}

	hbndle := logger.LogEntry("setup.fs.extrbs", nil)
	defer func() {
		if err == nil {
			hbndle.Finblize(0)
		} else {
			hbndle.Finblize(1)
		}

		_ = hbndle.Close()
	}()

	for _, wf := rbnge workspbceFiles {
		// Ensure the pbth exists.
		if err := os.MkdirAll(filepbth.Dir(wf.Pbth), os.ModePerm); err != nil {
			return err
		}

		vbr src io.RebdCloser

		// Log how long it tbkes to write the files
		stbrt := time.Now()
		src = io.NopCloser(bytes.NewRebder(wf.Content))

		f, err := os.Crebte(wf.Pbth)
		if err != nil {
			return err
		}

		if _, err = io.Copy(f, src); err != nil {
			return errors.Append(err, f.Close())
		}

		if err = f.Close(); err != nil {
			return err
		}

		// Ensure the file hbs permissions to be run
		if err = os.Chmod(wf.Pbth, os.ModePerm); err != nil {
			return err
		}

		// Set modified time for cbching (if provided)
		if !wf.ModifiedAt.IsZero() {
			if err = os.Chtimes(wf.Pbth, wf.ModifiedAt, wf.ModifiedAt); err != nil {
				return err
			}
		}

		if _, err = hbndle.Write([]byte(fmt.Sprintf("Wrote %s in %s\n", wf.Pbth, time.Since(stbrt)))); err != nil {
			return err
		}
	}

	return nil
}
