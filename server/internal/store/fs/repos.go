package fs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func CreateRepo(ctx context.Context, repo *sourcegraph.Repo) error {
	dir := absolutePathForRepo(ctx, repo.URI)

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		return grpc.Errorf(codes.AlreadyExists, "repo %s already exists", repo.URI)
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// TODO: Doing this `git init --bare` followed by a later
	//       RefreshVCS results in non-standard default branches to
	//       not be set. To fix that, either use git clone, or follow
	//       up with a `git ls-remote` and parse out HEAD.
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("creating repository %s failed with output:\n%s", repo.URI, string(out))
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
			cmd := exec.Command(c[0], c[1:]...)
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("configuring mirrored repository %s (origin clone URL %s) failed with %v:\n%s", repo.URI, repo.CloneURL(), err, string(out))
			}
		}
	}

	return nil
}

func DeleteRepo(ctx context.Context, repo string) error {
	dir := absolutePathForRepo(ctx, repo)
	if dir == absolutePathForRepo(ctx, "") {
		return errors.New("Repos.Delete needs at least one path element")
	}
	return os.RemoveAll(dir)
}

// absolutePathForRepo returns the absolute path for the given repo. It is
// guaranteed that the returned path be clean, for example:
//
//  reposAbsPath(ctx) == "example.com/foo/bar"
//  absolutePathForRepo(ctx, "../../.././x/./y/././..") == "example.com/foo/bar/x"
//
func absolutePathForRepo(ctx context.Context, repo string) string {
	// Clean the path of any relative parts.
	if !strings.HasPrefix(repo, "/") {
		repo = "/" + repo
	}
	repo = path.Clean(repo)[1:]

	return filepath.Join(reposAbsPath(ctx), repo)
}
