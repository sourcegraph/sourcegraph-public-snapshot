package execution

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/lib/batches/git"
)

// AfterStepResult is the execution result after executing a step with the given
// index in Steps.
type AfterStepResult struct {
	Version int `json:"version"`
	// Files are the changes made to Files by the step.
	ChangedFiles git.Changes `json:"changedFiles"`
	// Stdout is the output produced by the step on standard out.
	Stdout string `json:"stdout"`
	// Stderr is the output produced by the step on standard error.
	Stderr string `json:"stderr"`
	// StepIndex is the index of the step in the list of steps.
	StepIndex int `json:"stepIndex"`
	// Diff is the cumulative `git diff` after executing the Step.
	Diff []byte `json:"diff"`
	// Outputs is a copy of the Outputs after executing the Step.
	Outputs map[string]any `json:"outputs"`
	// Skipped determines whether the step was skipped.
	Skipped bool `json:"skipped"`
}

func (a AfterStepResult) MarshalJSON() ([]byte, error) {
	if a.Version == 2 {
		return json.Marshal(v2AfterStepResult(a))
	}
	return json.Marshal(v1AfterStepResult{
		ChangedFiles: a.ChangedFiles,
		Stdout:       a.Stdout,
		Stderr:       a.Stderr,
		StepIndex:    a.StepIndex,
		Diff:         string(a.Diff),
		Outputs:      a.Outputs,
	})
}

func (a *AfterStepResult) UnmarshalJSON(data []byte) error {
	var version versionAfterStepResult
	if err := json.Unmarshal(data, &version); err != nil {
		return err
	}
	if version.Version == 2 {
		var v2 v2AfterStepResult
		if err := json.Unmarshal(data, &v2); err != nil {
			return err
		}
		a.Version = v2.Version
		a.ChangedFiles = v2.ChangedFiles
		a.Stdout = v2.Stdout
		a.Stderr = v2.Stderr
		a.StepIndex = v2.StepIndex
		a.Diff = v2.Diff
		a.Outputs = v2.Outputs
		a.Skipped = v2.Skipped
		return nil
	}
	var v1 v1AfterStepResult
	if err := json.Unmarshal(data, &v1); err != nil {
		return err
	}
	a.ChangedFiles = v1.ChangedFiles
	a.Stdout = v1.Stdout
	a.Stderr = v1.Stderr
	a.StepIndex = v1.StepIndex
	a.Diff = []byte(v1.Diff)
	a.Outputs = v1.Outputs
	return nil
}

type versionAfterStepResult struct {
	Version int `json:"version"`
}

type v2AfterStepResult struct {
	Version      int            `json:"version"`
	ChangedFiles git.Changes    `json:"changedFiles"`
	Stdout       string         `json:"stdout"`
	Stderr       string         `json:"stderr"`
	StepIndex    int            `json:"stepIndex"`
	Diff         []byte         `json:"diff"`
	Outputs      map[string]any `json:"outputs"`
	Skipped      bool           `json:"skipped"`
}

type v1AfterStepResult struct {
	ChangedFiles git.Changes    `json:"changedFiles"`
	Stdout       string         `json:"stdout"`
	Stderr       string         `json:"stderr"`
	StepIndex    int            `json:"stepIndex"`
	Diff         string         `json:"diff"`
	Outputs      map[string]any `json:"outputs"`
}
