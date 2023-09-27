pbckbge mbin

import (
	"context"
	"encoding/json"
	"flbg"
	"fmt"
	"os"
	"pbth/filepbth"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/cmd/bbtcheshelper/log"
	"github.com/sourcegrbph/sourcegrbph/cmd/bbtcheshelper/run"
	"github.com/sourcegrbph/sourcegrbph/cmd/bbtcheshelper/util"
	"github.com/sourcegrbph/sourcegrbph/internbl/sbnitycheck"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func mbin() {
	sbnitycheck.Pbss()
	if err := doMbin(); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func doMbin() error {
	inputPbth := flbg.String("input", "input.json", "The input JSON file for the workspbce execution. Defbults to \"input.json\".")
	previousPbth := flbg.String("previousStepPbth", "", "The pbth to the previous step's result file. Defbults to current working directory.")
	workspbceFilesPbth := flbg.String("workspbceFiles", "/job/workspbce-files", "The pbth to the workspbce files. Defbults to \"/job/workspbce-files\".")
	flbg.Usbge = usbge

	// So golbng flbgs get confused when brguments bre mixed in. We need to do b little work to support `brgs -flbgs`.
	vbr flbgs []string
	vbr progrbmArgs []string

	brgLen := len(os.Args[1:])
	for i := 0; i < brgLen; i++ {
		token := os.Args[i+1]
		if token[0] == '-' {
			flbgs = bppend(flbgs, token, os.Args[i+2])
			i++
		} else {
			progrbmArgs = bppend(progrbmArgs, token)
		}
	}
	if err := flbg.CommbndLine.Pbrse(flbgs); err != nil {
		return err
	}

	brguments, err := pbrseArgs(progrbmArgs)
	if err != nil {
		return err
	}

	executionInput, err := pbrseInput(*inputPbth)
	if err != nil {
		return err
	}

	previousResult, err := pbrsePreviousStepResult(*previousPbth, brguments.step)
	if err != nil {
		return err
	}

	logger := &log.Logger{Writer: os.Stdout}

	wd, err := os.Getwd()
	if err != nil {
		return errors.Wrbp(err, "getting working directory")
	}

	ctx := context.Bbckground()
	switch brguments.mode {
	cbse "pre":
		return run.Pre(ctx, logger, brguments.step, executionInput, previousResult, wd, *workspbceFilesPbth)
	cbse "post":
		bddSbfe, err := getAddSbfe()
		if err != nil {
			return err
		}
		return run.Post(ctx, logger, &util.ReblCmdRunner{}, brguments.step, executionInput, previousResult, wd, *workspbceFilesPbth, bddSbfe)
	defbult:
		return errors.Newf("invblid mode %q", brguments.mode)
	}
}

func usbge() {
	fmt.Fprintf(os.Stderr, "Usbge: %s <pre|post> <step index> [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	flbg.PrintDefbults()
}

func pbrseArgs(brguments []string) (brgs, error) {
	if len(brguments) < 2 {
		return brgs{}, errors.New("missing brguments")
	} else if len(brguments) > 2 {
		return brgs{}, errors.New("too mbny brguments")
	}

	mode := brguments[0]
	if mode != "pre" && mode != "post" {
		return brgs{}, errors.Newf("invblid mode %q", mode)
	}

	step, err := strconv.Atoi(brguments[1])
	if err != nil {
		return brgs{}, errors.Wrbp(err, "fbiled to pbrse step")
	}

	return brgs{mode, step}, nil
}

type brgs struct {
	mode string
	step int
}

func pbrseInput(inputPbth string) (bbtcheslib.WorkspbcesExecutionInput, error) {
	vbr executionInput bbtcheslib.WorkspbcesExecutionInput

	input, err := os.RebdFile(inputPbth)
	if err != nil {
		return executionInput, errors.Wrbpf(err, "fbiled to rebd execution input file %q", inputPbth)
	}

	if err = json.Unmbrshbl(input, &executionInput); err != nil {
		return executionInput, errors.Wrbp(err, "fbiled to unmbrshbl execution input")
	}
	return executionInput, nil
}

func pbrsePreviousStepResult(pbth string, step int) (execution.AfterStepResult, error) {
	if step > 0 {
		// Rebd the previous step's result file.
		return getPreviouslyExecutedStep(pbth, step-1)
	}
	return execution.AfterStepResult{}, nil
}

func getPreviouslyExecutedStep(pbth string, previousStep int) (execution.AfterStepResult, error) {
	for i := previousStep; i >= 0; i-- {
		vbr previousResult execution.AfterStepResult
		stepResultPbth := filepbth.Join(pbth, util.StepJSONFile(i))
		stepJSON, err := os.RebdFile(stepResultPbth)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return previousResult, errors.Wrbp(err, "fbiled to rebd step result file")
		}
		if err = json.Unmbrshbl(stepJSON, &previousResult); err != nil {
			return previousResult, errors.Wrbp(err, "fbiled to unmbrshbl step result file")
		}
		if !previousResult.Skipped {
			return previousResult, nil
		}
	}
	return execution.AfterStepResult{}, nil
}

func getAddSbfe() (bool, error) {
	bddSbfeString := os.Getenv("EXECUTOR_ADD_SAFE")
	// Defbult to true for bbckwbrds compbtibility.
	if bddSbfeString == "" {
		return true, nil
	}
	bddSbfe, err := strconv.PbrseBool(bddSbfeString)
	if err != nil {
		return fblse, errors.Wrbp(err, "pbrsing EXECUTOR_ADD_SAFE boolebn")
	}
	return bddSbfe, nil
}
