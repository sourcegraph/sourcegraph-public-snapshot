package fs

import (
	"fmt"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/gitserver"
)

func CreateRepo(ctx context.Context, repo *sourcegraph.Repo) error {
	// TODO: Doing this `git init --bare` followed by a later
	//       RefreshVCS results in non-standard default branches to
	//       not be set. To fix that, either use git clone, or follow
	//       up with a `git ls-remote` and parse out HEAD.
	if err := gitserver.Init(repo.URI); err != nil {
		return err
	}

	if repo.Mirror {
		// Configure mirror repo but do not clone it (since that would
		// block this call). The repo may be cloned with
		// MirrorRepos.RefreshVCSData (which is called when the repo
		// is loaded in the app).
		mirrorCmds := [][]string{
			{"git", "remote", "add", "origin", "--", repo.CloneURL().String()},
			{"git", "config", "remote.origin.fetch", "+refs/*:refs/*"},
			{"git", "config", "remote.origin.mirror", "true"},
		}
		for _, c := range mirrorCmds {
			cmd := gitserver.Command(c[0], c[1:]...)
			cmd.Repo = repo.URI
			out, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("configuring mirrored repository %s (origin clone URL %s) failed with %v:\n%s", repo.URI, repo.CloneURL(), err, string(out))
			}
		}
	}

	return nil
}

func DeleteRepo(ctx context.Context, repo string) error {
	return gitserver.Remove(repo)
}
