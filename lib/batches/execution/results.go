package execution

import "github.com/sourcegraph/sourcegraph/lib/batches/git"

// AfterStepResult is the execution result after executing a step with the given
// index in Steps.
type AfterStepResult struct {
	// Files are the changes made to Files by the step.
	ChangedFiles git.Changes `json:"changedFiles"`
	// Stdout is the output produced by the step on standard out.
	Stdout string `json:"stdout"`
	// Stderr is the output produced by the step on standard error.
	Stderr string `json:"stderr"`
	// StepIndex is the index of the step in the list of steps.
	StepIndex int `json:"stepIndex"`
	// Diff is the cumulative `git diff` after executing the Step.
	Diff string `json:"diff"`
	// Outputs is a copy of the Outputs after executing the Step.
	Outputs map[string]any `json:"outputs"`
}
