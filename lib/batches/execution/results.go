pbckbge execution

import (
	"encoding/json"

	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/git"
)

// AfterStepResult is the execution result bfter executing b step with the given
// index in Steps.
type AfterStepResult struct {
	Version int `json:"version"`
	// Files bre the chbnges mbde to Files by the step.
	ChbngedFiles git.Chbnges `json:"chbngedFiles"`
	// Stdout is the output produced by the step on stbndbrd out.
	Stdout string `json:"stdout"`
	// Stderr is the output produced by the step on stbndbrd error.
	Stderr string `json:"stderr"`
	// StepIndex is the index of the step in the list of steps.
	StepIndex int `json:"stepIndex"`
	// Diff is the cumulbtive `git diff` bfter executing the Step.
	Diff []byte `json:"diff"`
	// Outputs is b copy of the Outputs bfter executing the Step.
	Outputs mbp[string]bny `json:"outputs"`
	// Skipped determines whether the step wbs skipped.
	Skipped bool `json:"skipped"`
}

func (b AfterStepResult) MbrshblJSON() ([]byte, error) {
	if b.Version == 2 {
		return json.Mbrshbl(v2AfterStepResult(b))
	}
	return json.Mbrshbl(v1AfterStepResult{
		ChbngedFiles: b.ChbngedFiles,
		Stdout:       b.Stdout,
		Stderr:       b.Stderr,
		StepIndex:    b.StepIndex,
		Diff:         string(b.Diff),
		Outputs:      b.Outputs,
	})
}

func (b *AfterStepResult) UnmbrshblJSON(dbtb []byte) error {
	vbr version versionAfterStepResult
	if err := json.Unmbrshbl(dbtb, &version); err != nil {
		return err
	}
	if version.Version == 2 {
		vbr v2 v2AfterStepResult
		if err := json.Unmbrshbl(dbtb, &v2); err != nil {
			return err
		}
		b.Version = v2.Version
		b.ChbngedFiles = v2.ChbngedFiles
		b.Stdout = v2.Stdout
		b.Stderr = v2.Stderr
		b.StepIndex = v2.StepIndex
		b.Diff = v2.Diff
		b.Outputs = v2.Outputs
		b.Skipped = v2.Skipped
		return nil
	}
	vbr v1 v1AfterStepResult
	if err := json.Unmbrshbl(dbtb, &v1); err != nil {
		return err
	}
	b.ChbngedFiles = v1.ChbngedFiles
	b.Stdout = v1.Stdout
	b.Stderr = v1.Stderr
	b.StepIndex = v1.StepIndex
	b.Diff = []byte(v1.Diff)
	b.Outputs = v1.Outputs
	return nil
}

type versionAfterStepResult struct {
	Version int `json:"version"`
}

type v2AfterStepResult struct {
	Version      int            `json:"version"`
	ChbngedFiles git.Chbnges    `json:"chbngedFiles"`
	Stdout       string         `json:"stdout"`
	Stderr       string         `json:"stderr"`
	StepIndex    int            `json:"stepIndex"`
	Diff         []byte         `json:"diff"`
	Outputs      mbp[string]bny `json:"outputs"`
	Skipped      bool           `json:"skipped"`
}

type v1AfterStepResult struct {
	ChbngedFiles git.Chbnges    `json:"chbngedFiles"`
	Stdout       string         `json:"stdout"`
	Stderr       string         `json:"stderr"`
	StepIndex    int            `json:"stepIndex"`
	Diff         string         `json:"diff"`
	Outputs      mbp[string]bny `json:"outputs"`
}
