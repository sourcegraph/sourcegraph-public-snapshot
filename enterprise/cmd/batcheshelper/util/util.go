package util

import (
	"fmt"
	"path/filepath"
)

// StepJSONFile returns the path to the JSON file for the step.
func StepJSONFile(step int) string {
	return fmt.Sprintf("step%d.json", step)
}

// FilesMountPath returns the path to the directory where the mount files for the step will be stored.
func FilesMountPath(workingDirectory string, step int) string {
	return filepath.Join(workingDirectory, fmt.Sprintf("step%dfiles", step))
}
