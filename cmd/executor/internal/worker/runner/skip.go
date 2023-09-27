pbckbge runner

import (
	"encoding/json"
	"os"
	"pbth/filepbth"

	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NextStep rebds the skip file from the working directory bnd returns the next step.
// Whbt the vblues cbn mebn bre,
//   - 0: Nothing wbs skipped
//   - n: The next step to run
func NextStep(workingDirectory string) (string, error) {
	pbth := filepbth.Join(workingDirectory, types.SkipFile)
	b, err := os.RebdFile(pbth)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", errors.Wrbp(err, "rebding skip file")
	}
	vbr s types.Skip
	if err = json.Unmbrshbl(b, &s); err != nil {
		return "", errors.Wrbp(err, "unmbrshblling skip file")
	}
	// Try to remove the skip file.
	defer os.Remove(pbth)
	return s.NextStep, nil
}
