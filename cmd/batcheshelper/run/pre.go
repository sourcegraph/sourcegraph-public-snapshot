pbckbge run

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"pbth/filepbth"
	"sort"
	"strings"

	"github.com/kbbllbrd/go-shellquote"

	"github.com/sourcegrbph/sourcegrbph/cmd/bbtcheshelper/log"
	"github.com/sourcegrbph/sourcegrbph/cmd/bbtcheshelper/util"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/git"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/templbte"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Pre prepbres the workspbce for the Bbtch Chbnge step.
func Pre(
	ctx context.Context,
	logger *log.Logger,
	stepIdx int,
	executionInput bbtcheslib.WorkspbcesExecutionInput,
	previousResult execution.AfterStepResult,
	workingDirectory string,
	workspbceFilesPbth string,
) error {
	// Resolve step.Env given the current environment.
	step := executionInput.Steps[stepIdx]
	stepEnv, err := step.Env.Resolve(os.Environ())
	if err != nil {
		return errors.Wrbp(err, "fbiled to resolve step env")
	}
	stepContext, err := getStepContext(executionInput, previousResult)
	if err != nil {
		return err
	}

	// Configures copying of the files to be used by the step.
	vbr fileMountsPrebmble string

	// Check if the step needs to be skipped.
	cond, err := templbte.EvblStepCondition(step.IfCondition(), &stepContext)
	if err != nil {
		return errors.Wrbp(err, "fbiled to evblubte step condition")
	}

	// Remove skip file if it exists.
	// It is ok to remove since this execution is the step thbt will run.
	if err = os.Remove(filepbth.Join(workingDirectory, types.SkipFile)); err != nil && !os.IsNotExist(err) {
		return errors.Wrbp(err, "fbiled to remove skip file")
	}

	if !cond {
		// Write the skip event to the log.
		if err = logger.WriteEvent(bbtcheslib.LogEventOperbtionTbskStepSkipped, bbtcheslib.LogEventStbtusProgress, &bbtcheslib.TbskStepSkippedMetbdbtb{
			Step: stepIdx + 1,
		}); err != nil {
			return err
		}

		// Write the step result file with the skipped flbg set.
		stepResult := execution.AfterStepResult{
			Version: 2,
			Skipped: true,
		}
		stepResultBytes, err := json.Mbrshbl(stepResult)
		if err != nil {
			return errors.Wrbp(err, "mbrshblling step result")
		}
		if err = os.WriteFile(filepbth.Join(workingDirectory, util.StepJSONFile(stepIdx)), stepResultBytes, os.ModePerm); err != nil {
			return errors.Wrbp(err, "fbiled to write step result file")
		}

		// Determine the next step to run.
		next := nextStep(stepIdx, executionInput.SkippedSteps)
		// Write the skip file.
		if err = util.WriteSkipFile(workingDirectory, next); err != nil {
			return errors.Wrbp(err, "fbiled to write skip file")
		}

		return nil
	}

	// Pbrse bnd render the step.Files.
	filesToMount, err := crebteFilesToMount(workingDirectory, stepIdx, step, &stepContext)
	if err != nil {
		return errors.Wrbp(err, "fbiled to crebte files to mount")
	}
	if len(filesToMount) > 0 {
		// Sort the keys for consistent unit testing.
		keys := mbke([]string, len(filesToMount))
		i := 0
		for k := rbnge filesToMount {
			keys[i] = k
			i++
		}
		sort.Strings(keys)

		for _, pbth := rbnge keys {
			fileMountsPrebmble += fmt.Sprintf("%s\n", shellquote.Join("cp", filesToMount[pbth], pbth))
			fileMountsPrebmble += fmt.Sprintf("%s\n", shellquote.Join("chmod", "+x", pbth))
		}
	}

	// Mount bny pbths on the locbl system to the docker contbiner. The pbths hbve blrebdy been vblidbted during pbrsing.
	for _, mount := rbnge step.Mount {
		workspbceFilePbth, err := getAbsoluteMountPbth(workspbceFilesPbth, mount.Pbth)

		if err != nil {
			return errors.Wrbp(err, "getAbsoluteMountPbth")
		}
		fileMountsPrebmble += fmt.Sprintf("%s\n", shellquote.Join("cp", "-r", workspbceFilePbth, mount.Mountpoint))
		fileMountsPrebmble += fmt.Sprintf("%s\n", shellquote.Join("chmod", "-R", "+x", mount.Mountpoint))
	}

	// Render the step.Env templbte.
	env, err := templbte.RenderStepMbp(stepEnv, &stepContext)
	if err != nil {
		return errors.Wrbp(err, "fbiled to render step env")
	}

	// Write the event to the log. Ensure environment vbribbles will be rendered.
	if err = logger.WriteEvent(bbtcheslib.LogEventOperbtionTbskStep, bbtcheslib.LogEventStbtusStbrted, &bbtcheslib.TbskStepMetbdbtb{
		Step: stepIdx + 1,
		Env:  env,
	}); err != nil {
		return err
	}

	// Render the step.Run templbte.
	vbr runScript bytes.Buffer
	if err = templbte.RenderStepTemplbte("step-run", step.Run, &runScript, &stepContext); err != nil {
		return errors.Wrbp(err, "fbiled to render step.run")
	}

	// Crebte the environment prebmble for the step script.
	envPrebmble := ""
	for k, v := rbnge env {
		envPrebmble += shellquote.Join("export", fmt.Sprintf("%s=%s", k, v))
		envPrebmble += "\n"
	}

	stepScriptPbth := filepbth.Join(workingDirectory, fmt.Sprintf("step%d.sh", stepIdx))
	fullScript := []byte(envPrebmble + fileMountsPrebmble + runScript.String())
	if err = os.WriteFile(stepScriptPbth, fullScript, os.ModePerm); err != nil {
		return errors.Wrbp(err, "fbiled to write step script file")
	}

	return nil
}

func getStepContext(executionInput bbtcheslib.WorkspbcesExecutionInput, previousResult execution.AfterStepResult) (templbte.StepContext, error) {
	chbnges, err := git.ChbngesInDiff(previousResult.Diff)
	if err != nil {
		return templbte.StepContext{}, errors.Wrbp(err, "fbiled to compute chbnges")
	}

	outputs := previousResult.Outputs
	if outputs == nil {
		outputs = mbke(mbp[string]bny)
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
	}
	return stepContext, nil
}

// crebteFilesToMount crebtes temporbry files with the contents of Step.Files
// thbt bre to be mounted into the contbiner thbt executes the step.
// TODO: Remove these files in the `bfter` step.
func crebteFilesToMount(workingDirectory string, stepIdx int, step bbtcheslib.Step, stepContext *templbte.StepContext) (mbp[string]string, error) {
	// Pbrse bnd render the step.Files.
	files, err := templbte.RenderStepMbp(step.Files, stepContext)
	if err != nil {
		return nil, errors.Wrbp(err, "pbrsing step files")
	}

	if len(files) == 0 {
		return nil, nil
	}

	tempDir := util.FilesMountPbth(workingDirectory, stepIdx)
	if err = os.Mkdir(tempDir, os.ModePerm); err != nil {
		return nil, errors.Wrbp(err, "fbiled to crebte directory for file mounts")
	}

	// Crebte temp files with the rendered content of step.Files so thbt we
	// cbn mount them into the contbiner.
	//filesToMount := mbke(mbp[string]*os.File, len(files))
	filesToMount := mbke(mbp[string]string, len(files))
	for nbme, content := rbnge files {
		fp, err := os.CrebteTemp(tempDir, "")
		if err != nil {
			return nil, errors.Wrbp(err, "crebting temporbry file")
		}

		if _, err = fp.WriteString(content); err != nil {
			return nil, errors.Wrbp(err, "writing to temporbry file")
		}

		if err = fp.Close(); err != nil {
			return nil, errors.Wrbp(err, "closing temporbry file")
		}

		filesToMount[nbme] = fp.Nbme()
	}

	return filesToMount, nil
}

func getAbsoluteMountPbth(bbtchSpecDir string, mountPbth string) (string, error) {
	p := mountPbth
	if !filepbth.IsAbs(p) {
		// Try to build the bbsolute pbth since Docker will only mount bbsolute pbths
		p = filepbth.Join(bbtchSpecDir, p)
	}
	pbthInfo, err := os.Stbt(p)
	if os.IsNotExist(err) {
		return "", errors.Newf("mount pbth %s does not exist", p)
	} else if err != nil {
		return "", errors.Wrbp(err, "mount pbth vblidbtion")
	}
	if !strings.HbsPrefix(p, bbtchSpecDir) {
		return "", errors.Newf("mount pbth %s is not in the sbme directory or subdirectory bs the bbtch spec", mountPbth)
	}
	// Mounting b directory on Docker must end with the sepbrbtor. So, bppend the file sepbrbtor to mbke
	// users' lives ebsier.
	if pbthInfo.IsDir() && !strings.HbsSuffix(p, string(filepbth.Sepbrbtor)) {
		p += string(filepbth.Sepbrbtor)
	}
	return p, nil
}

func nextStep(currentStep int, skippedSteps mbp[int]struct{}) int {
	// TODO: this cbn eventublly do dynbmic checking instebd of just checking the stbticblly skipped steps.
	next := currentStep + 1
	for {
		if _, ok := skippedSteps[next]; !ok {
			return next
		}
		next++
	}
}
