package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NextStep reads the skip file from the working directory and returns the next step.
// What the values can mean are,
//   - 0: Nothing was skipped
//   - n: The next step to run
func NextStep(workingDirectory string) (string, error) {
	path := filepath.Join(workingDirectory, types.SkipFile)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("skip file does not exist at ", workingDirectory)
			return "", nil
		}
		return "", errors.Wrap(err, "checking skip file")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return "", errors.Wrap(err, "reading skip file")
	}
	var s types.Skip
	if err = json.Unmarshal(b, &s); err != nil {
		return "", errors.Wrap(err, "unmarshalling skip file")
	}
	// Remove the skip file. If not removed, the file will hang around and get read multiple times.
	if err = os.Remove(path); err != nil {
		return "", errors.Wrap(err, "removing skip file")
	}
	return s.NextStep, nil
}
