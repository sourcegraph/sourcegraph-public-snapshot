package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution/cache"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func execPost(ctx context.Context, stepIdx int, executionInput batcheslib.WorkspacesExecutionInput, previousResult execution.AfterStepResult) error {
	step := executionInput.Steps[stepIdx]

	// Generate the diff.
	if _, err := runGitCmd(context.Background(), "git", "add", "--all"); err != nil {
		return errors.Wrap(err, "git add --all failed")
	}
	diff, err := runGitCmd(context.Background(), "git", "diff", "--cached", "--no-prefix", "--binary")
	if err != nil {
		return errors.Wrap(err, "git diff --cached --no-prefix --binary failed")
	}

	// Read the stdout of the current step.
	stdout, err := os.ReadFile(fmt.Sprintf("stdout%d.log", stepIdx))
	if err != nil {
		return errors.Wrap(err, "failed to read stdout file")
	}
	// Read the stderr of the current step.
	stderr, err := os.ReadFile(fmt.Sprintf("stderr%d.log", stepIdx))
	if err != nil {
		return errors.Wrap(err, "failed to read stderr file")
	}

	// Build the step result.
	stepResult := execution.AfterStepResult{
		Stdout:    string(stdout),
		Stderr:    string(stderr),
		StepIndex: stepIdx,
		Diff:      string(diff),
		// Those will be set below.
		Outputs: make(map[string]interface{}),
	}

	// Render the step outputs.
	changes, err := git.ChangesInDiff([]byte(previousResult.Diff))
	if err != nil {
		return errors.Wrap(err, "failed to get changes in diff")
	}
	outputs := previousResult.Outputs
	if outputs == nil {
		outputs = make(map[string]any)
	}
	stepContext := template.StepContext{
		BatchChange: executionInput.BatchChangeAttributes,
		Repository: template.Repository{
			Name:        executionInput.Repository.Name,
			Branch:      executionInput.Branch.Name,
			FileMatches: executionInput.SearchResultPaths,
		},
		Outputs: outputs,
		Steps: template.StepsContext{
			Path:    executionInput.Path,
			Changes: changes,
		},
		PreviousStep: previousResult,
		Step:         stepResult,
	}

	// Render and evaluate outputs.
	if err := batcheslib.SetOutputs(step.Outputs, outputs, &stepContext); err != nil {
		return errors.Wrap(err, "setting outputs")
	}
	for k, v := range outputs {
		stepResult.Outputs[k] = v
	}

	// Serialize the step result to disk.
	cntnt, err := json.Marshal(stepResult)
	if err != nil {
		return errors.Wrap(err, "marshalling step result")
	}
	if err := os.WriteFile(fmt.Sprintf("step%d.json", stepIdx), cntnt, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to write step result file")
	}

	key := cache.KeyForWorkspace(
		&executionInput.BatchChangeAttributes,
		batcheslib.Repository{
			ID:          executionInput.Repository.ID,
			Name:        executionInput.Repository.Name,
			BaseRef:     executionInput.Branch.Name,
			BaseRev:     executionInput.Branch.Target.OID,
			FileMatches: executionInput.SearchResultPaths,
		},
		executionInput.Path,
		os.Environ(),
		executionInput.OnlyFetchWorkspace,
		executionInput.Steps,
		stepIdx,
		nil, // todo: should not be nil.
	)

	k, err := key.Key()
	if err != nil {
		return errors.Wrap(err, "failed to compute cache key")
	}

	metadata := &batcheslib.CacheAfterStepResultMetadata{
		Key:   k,
		Value: stepResult,
	}
	e := batcheslib.LogEvent{Operation: batcheslib.LogEventOperationCacheAfterStepResult, Status: batcheslib.LogEventStatusSuccess, Metadata: metadata}
	e.Timestamp = time.Now().UTC().Truncate(time.Millisecond)
	err = json.NewEncoder(os.Stdout).Encode(e)
	if err != nil {
		return errors.Wrap(err, "failed to encode after step result event")
	}

	return nil
}

func runGitCmd(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = "repository"

	return cmd.Output()
}
