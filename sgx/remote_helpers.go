package sgx

import (
	"fmt"

	srclib "sourcegraph.com/sourcegraph/srclib/cli"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// getRemoteRepo gets the remote repository that corresponds to the
// local repository (from srclib.OpenLocalRepo).
func getRemoteRepo(cl *sourcegraph.Client) (*sourcegraph.Repo, error) {
	lrepo, err := srclib.OpenLocalRepo()
	if err != nil {
		return nil, err
	}
	if lrepo.CloneURL == "" {
		return nil, srclib.ErrNoVCSCloneURL
	}
	uri := lrepo.URI()
	if uri == "" {
		return nil, fmt.Errorf("local repo URI is malformed: %s", lrepo.CloneURL)
	}

	rrepo, err := cl.Repos.Get(cli.Ctx, &sourcegraph.RepoSpec{URI: uri})
	if err != nil {
		return nil, fmt.Errorf("repo %s: %s", uri, err)
	}
	return rrepo, nil
}
