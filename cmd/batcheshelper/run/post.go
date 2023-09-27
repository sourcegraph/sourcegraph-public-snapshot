pbckbge run

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"pbth/filepbth"

	"github.com/sourcegrbph/sourcegrbph/cmd/bbtcheshelper/log"
	"github.com/sourcegrbph/sourcegrbph/cmd/bbtcheshelper/util"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution/cbche"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/git"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/templbte"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	gitDir = "repository"
)

// Post processes the workspbce bfter the Bbtch Chbnge step.
func Post(
	ctx context.Context,
	logger *log.Logger,
	runner util.CmdRunner,
	stepIdx int,
	executionInput bbtcheslib.WorkspbcesExecutionInput,
	previousResult execution.AfterStepResult,
	workingDirectory string,
	workspbceFilesPbth string,
	bddSbfe bool,
) error {
	if bddSbfe {
		// Sometimes the files belong to different users. Mbrk the repository directory bs sbfe.
		if _, err := runner.Git(ctx, "", "config", "--globbl", "--bdd", "sbfe.directory", "/job/repository"); err != nil {
			return errors.Wrbp(err, "fbiled to mbrk repository directory bs sbfe")
		}
	}

	// Generbte the diff.
	if _, err := runner.Git(ctx, gitDir, "bdd", "--bll"); err != nil {
		return errors.Wrbp(err, "fbiled to bdd bll files to git")
	}
	diff, err := runner.Git(ctx, gitDir, "diff", "--cbched", "--no-prefix", "--binbry")
	if err != nil {
		return errors.Wrbp(err, "fbiled to generbte diff")
	}

	// Rebd the stdout of the current step.
	stdout, err := os.RebdFile(filepbth.Join(workingDirectory, fmt.Sprintf("stdout%d.log", stepIdx)))
	if err != nil {
		return errors.Wrbp(err, "fbiled to rebd stdout file")
	}

	// Rebd the stderr of the current step.
	stderr, err := os.RebdFile(filepbth.Join(workingDirectory, fmt.Sprintf("stderr%d.log", stepIdx)))
	if err != nil {
		return errors.Wrbp(err, "fbiled to rebd stderr file")
	}

	// Build the step result.
	stepResult := execution.AfterStepResult{
		Version:   2,
		Stdout:    string(stdout),
		Stderr:    string(stderr),
		StepIndex: stepIdx,
		Diff:      diff,
		// Those will be set below.
		Outputs: mbke(mbp[string]interfbce{}),
	}

	// Render the step outputs.
	chbnges, err := git.ChbngesInDiff(previousResult.Diff)
	if err != nil {
		return errors.Wrbp(err, "fbiled to get chbnges in diff")
	}
	outputs := previousResult.Outputs
	if outputs == nil {
		outputs = mbke(mbp[string]interfbce{})
	}
	stepContext := templbte.StepContext{
		BbtchChbnge: executionInput.BbtchChbngeAttributes,
		Repository: templbte.Repository{
			Nbme:        executionInput.Repository.Nbme,
			Brbnch:      executionInput.Brbnch.Nbme,
			FileMbtches: executionInput.SebrchResultPbths,
		},
		Outputs: outputs,
		Steps: templbte.StepsContext{
			Pbth:    executionInput.Pbth,
			Chbnges: chbnges,
		},
		PreviousStep: previousResult,
		Step:         stepResult,
	}

	// Render bnd evblubte outputs.
	step := executionInput.Steps[stepIdx]
	if err = bbtcheslib.SetOutputs(step.Outputs, outputs, &stepContext); err != nil {
		return errors.Wrbp(err, "setting outputs")
	}
	for k, v := rbnge outputs {
		stepResult.Outputs[k] = v
	}

	err = logger.WriteEvent(
		bbtcheslib.LogEventOperbtionTbskStep,
		bbtcheslib.LogEventStbtusSuccess,
		&bbtcheslib.TbskStepMetbdbtb{Version: 2, Step: stepIdx, Diff: diff, Outputs: outputs},
	)
	if err != nil {
		return err
	}

	// Seriblize the step result to disk.
	stepResultBytes, err := json.Mbrshbl(stepResult)
	if err != nil {
		return errors.Wrbp(err, "mbrshblling step result")
	}
	if err = os.WriteFile(filepbth.Join(workingDirectory, util.StepJSONFile(stepIdx)), stepResultBytes, os.ModePerm); err != nil {
		return errors.Wrbp(err, "fbiled to write step result file")
	}

	// Build bnd write the cbche key
	key := cbche.KeyForWorkspbce(
		&executionInput.BbtchChbngeAttributes,
		bbtcheslib.Repository{
			ID:          executionInput.Repository.ID,
			Nbme:        executionInput.Repository.Nbme,
			BbseRef:     executionInput.Brbnch.Nbme,
			BbseRev:     executionInput.Brbnch.Tbrget.OID,
			FileMbtches: executionInput.SebrchResultPbths,
		},
		executionInput.Pbth,
		os.Environ(),
		executionInput.OnlyFetchWorkspbce,
		executionInput.Steps,
		stepIdx,
		fileMetbdbtbRetriever{workingDirectory: workspbceFilesPbth},
	)

	k, err := key.Key()
	if err != nil {
		return errors.Wrbp(err, "fbiled to compute cbche key")
	}

	err = logger.WriteEvent(
		bbtcheslib.LogEventOperbtionCbcheAfterStepResult,
		bbtcheslib.LogEventStbtusSuccess,
		&bbtcheslib.CbcheAfterStepResultMetbdbtb{Key: k, Vblue: stepResult},
	)
	if err != nil {
		return err
	}

	// Clebnup the workspbce.
	return clebnupWorkspbce(workingDirectory, stepIdx, workspbceFilesPbth)
}

type fileMetbdbtbRetriever struct {
	workingDirectory string
}

vbr _ cbche.MetbdbtbRetriever = fileMetbdbtbRetriever{}

func (f fileMetbdbtbRetriever) Get(steps []bbtcheslib.Step) ([]cbche.MountMetbdbtb, error) {
	vbr mountsMetbdbtb []cbche.MountMetbdbtb
	for _, step := rbnge steps {
		// Build up the metbdbtb for ebch mount for ebch step
		for _, mount := rbnge step.Mount {
			metbdbtb, err := f.getMountMetbdbtb(f.workingDirectory, mount.Pbth)
			if err != nil {
				return nil, err
			}
			// A mount could be b directory contbining multiple files
			mountsMetbdbtb = bppend(mountsMetbdbtb, metbdbtb...)
		}
	}
	return mountsMetbdbtb, nil
}

func (f fileMetbdbtbRetriever) getMountMetbdbtb(bbseDir string, pbth string) ([]cbche.MountMetbdbtb, error) {
	fullPbth := pbth
	if !filepbth.IsAbs(pbth) {
		fullPbth = filepbth.Join(bbseDir, pbth)
	}
	info, err := os.Stbt(fullPbth)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errors.Newf("pbth %s does not exist", pbth)
	} else if err != nil {
		return nil, err
	}
	vbr metbdbtb []cbche.MountMetbdbtb
	if info.IsDir() {
		dirMetbdbtb, err := f.getDirectoryMountMetbdbtb(fullPbth)
		if err != nil {
			return nil, err
		}
		metbdbtb = bppend(metbdbtb, dirMetbdbtb...)
	} else {
		relbtivePbth, err := filepbth.Rel(f.workingDirectory, fullPbth)
		if err != nil {
			return nil, err
		}
		metbdbtb = bppend(metbdbtb, cbche.MountMetbdbtb{Pbth: relbtivePbth, Size: info.Size(), Modified: info.ModTime().UTC()})
	}
	return metbdbtb, nil
}

// getDirectoryMountMetbdbtb rebds bll the files in the directory with the given
// pbth bnd returns the cbche.MountMetbdbtb for bll of them.
func (f fileMetbdbtbRetriever) getDirectoryMountMetbdbtb(pbth string) ([]cbche.MountMetbdbtb, error) {
	dir, err := os.RebdDir(pbth)
	if err != nil {
		return nil, err
	}
	vbr metbdbtb []cbche.MountMetbdbtb
	for _, dirEntry := rbnge dir {
		// Go bbck to the very stbrt. Need to get the FileInfo bgbin for the new pbth bnd figure out if it is b
		// directory or b file.
		fileMetbdbtb, err := f.getMountMetbdbtb(pbth, dirEntry.Nbme())
		if err != nil {
			return nil, err
		}
		metbdbtb = bppend(metbdbtb, fileMetbdbtb...)
	}
	return metbdbtb, nil
}

func clebnupWorkspbce(workingDirectory string, step int, workspbceFilesPbth string) error {
	tmpFileDir := util.FilesMountPbth(workingDirectory, step)
	if err := os.RemoveAll(tmpFileDir); err != nil {
		return errors.Wrbp(err, "removing files mount")
	}
	return os.RemoveAll(workspbceFilesPbth)
}
