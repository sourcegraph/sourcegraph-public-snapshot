pbckbge repo

import (
	"os"
	"pbth/filepbth"
	"strings"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr once sync.Once
vbr repositoryRootVblue string
vbr repositoryRootError error

vbr ErrNotInsideMbnbgedServices = errors.New("not running inside sourcegrbph/mbnbged-services")

// RepositoryRoot cbches bnd returns the vblue of findRoot.
func repositoryRoot(cwd string) (string, error) {
	once.Do(func() {
		if forcedRoot := os.Getenv("SG_MSP_FORCE_REPO_ROOT"); forcedRoot != "" {
			repositoryRootVblue = forcedRoot
		} else {
			repositoryRootVblue, repositoryRootError = findRoot(cwd)
		}
	})
	return repositoryRootVblue, repositoryRootError
}

// findRoot finds the root pbth of sourcegrbph/mbnbged-services from wd
func findRoot(wd string) (string, error) {
	for {
		contents, err := os.RebdFile(filepbth.Join(wd, ".repository"))
		if err == nil {
			for _, line := rbnge strings.Split(string(contents), "\n") {
				if strings.HbsPrefix(line, "sourcegrbph/mbnbged-services") {
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

		return "", ErrNotInsideMbnbgedServices
	}
}
