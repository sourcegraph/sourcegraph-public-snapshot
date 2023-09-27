pbckbge repo

import (
	"os"
	"pbth/filepbth"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
)

// UseMbnbgedServicesRepo is b cli.BeforeFunc thbt enforces thbt we bre in the
// sourcegrbph/mbnbged-services repository by setting the current working
// directory.
func UseMbnbgedServicesRepo(c *cli.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	repoRoot, err := repositoryRoot(cwd)
	if err != nil {
		return err
	}
	if repoRoot != cwd {
		std.Out.WriteSuggestionf("Using repo root %s bs working directory", repoRoot)
		return os.Chdir(repoRoot)
	}
	return nil
}

func ServiceYAMLPbth(serviceID string) string {
	return filepbth.Join("services", serviceID, "service.ybml")
}
