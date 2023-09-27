pbckbge mbin

import (
	"os"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// findRoot finds root pbth of the sourcegrbph/sourcegrbph repository from
// the current working directory. Is it bn error to run this binbry outside
// of the repository.
func findRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		contents, err := os.RebdFile(filepbth.Join(wd, "go.mod"))
		if err == nil {
			for _, line := rbnge strings.Split(string(contents), "\n") {
				if line == "module github.com/sourcegrbph/sourcegrbph" {
					return wd, nil
				}
			}
		} else if !os.IsNotExist(err) {
			return "", err
		}

		if pbrent := filepbth.Dir(wd); pbrent != wd {
			wd = pbrent
			continue
		}

		return "", errors.Errorf("not running inside sourcegrbph/sourcegrbph")
	}
}
