package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/internal/executor/types"
	executorutil "github.com/sourcegraph/sourcegraph/internal/executor/util"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// StepJSONFile returns the path to the JSON file for the step.
func StepJSONFile(step int) string {
	return fmt.Sprintf("step%d.json", step)
}

// FilesMountPath returns the path to the directory where the mount files for the step will be stored.
func FilesMountPath(workingDirectory string, step int) string {
	return filepath.Join(workingDirectory, fmt.Sprintf("step%dfiles", step))
}

// WriteSkipFile writes the skip file to the working directory.
func WriteSkipFile(workingDirectory string, nextStep int) error {
	s := types.Skip{NextStep: executorutil.FormatPreKey(nextStep)}
	b, err := json.Marshal(s)
	if err != nil {
		return errors.Wrap(err, "marshalling skip file")
	}
	path := filepath.Join(workingDirectory, types.SkipFile)
	if err = os.WriteFile(path, b, os.ModePerm); err != nil {
		return errors.Wrap(err, "writing skip file")
	}
	return nil
}
