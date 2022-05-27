package execution

import (
	"bytes"

	"github.com/sourcegraph/sourcegraph/lib/batches/git"
)

// StepResult represents the result of a single, previously executed step.
type StepResult struct {
	// Files are the changes made to Files by the step.
	Files *git.Changes
	// Stdout is the output produced by the step on standard out.
	Stdout *bytes.Buffer
	// Stderr is the output produced by the step on standard error.
	Stderr *bytes.Buffer
}

// ModifiedFiles returns the files modified by a step.
func (r StepResult) ModifiedFiles() []string {
	if r.Files != nil {
		return r.Files.Modified
	}
	return []string{}
}

// AddedFiles returns the files added by a step.
func (r StepResult) AddedFiles() []string {
	if r.Files != nil {
		return r.Files.Added
	}
	return []string{}
}

// DeletedFiles returns the files deleted by a step.
func (r StepResult) DeletedFiles() []string {
	if r.Files != nil {
		return r.Files.Deleted
	}
	return []string{}
}

// RenamedFiles returns the new name of files that have been renamed by a step.
func (r StepResult) RenamedFiles() []string {
	if r.Files != nil {
		return r.Files.Renamed
	}
	return []string{}
}

// AfterStepResult is the ExecutionResult after executing a Step with the given
// index in Steps.
type AfterStepResult struct {
	// StepIndex is the index of the step in the list of steps.
	StepIndex int `json:"stepIndex"`
	// Diff is the cumulative `git diff` after executing the Step.
	Diff string `json:"diff"`
	// Outputs is a copy of the Outputs after executing the Step.
	Outputs map[string]any `json:"outputs"`
	// PreviousStepResult is the StepResult of the step before Step, if
	// StepIndex != 0.
	PreviousStepResult StepResult `json:"previousStepResult"`
}

// Result is the result of executing all executable steps in a workspace.
type Result struct {
	// Diff is the produced by executing all steps.
	Diff string `json:"diff"`

	// ChangedFiles are files that have been changed by all steps.
	ChangedFiles *git.Changes `json:"changedFiles"`

	// Outputs are the outputs produced by all steps.
	Outputs map[string]any `json:"outputs"`

	// Path relative to the repository's root directory in which the steps
	// have been executed.
	// No leading slashes. Root directory is blank string.
	Path string `json:"path"`
}
