package runner

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NextStep reads the skip file from the working directory and returns the next step.
// What the values can mean are,
//   - 0: Nothing was skipped
//   - n: The next step to run
func NextStep(workingDirectory string) (string, error) {
	path := filepath.Join(workingDirectory, types.SkipFile)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", errors.Wrap(err, "reading skip file")
	}
	var s types.Skip
	if err = json.Unmarshal(b, &s); err != nil {
		return "", errors.Wrap(err, "unmarshalling skip file")
	}
	// Try to remove the skip file.
	defer os.Remove(path)
	return s.NextStep, nil
}
