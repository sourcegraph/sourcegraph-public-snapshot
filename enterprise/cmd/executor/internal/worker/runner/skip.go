package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NextStep reads the skip file from the working directory and returns the next step.
// What the values can mean are,
//   - 0: Nothing was skipped
//   - n: The next step to run
func NextStep(workingDirectory string) (int, error) {
	path := filepath.Join(workingDirectory, "skip.json")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("skip file does not exist at ", workingDirectory)
			return 0, nil
		}
		return 0, errors.Wrap(err, "checking skip file")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return 0, errors.Wrap(err, "reading skip file")
	}
	var s skip
	if err = json.Unmarshal(b, &s); err != nil {
		return 0, errors.Wrap(err, "unmarshalling skip file")
	}
	return s.NextStep, nil
}

type skip struct {
	NextStep int `json:"nextStep"`
}
