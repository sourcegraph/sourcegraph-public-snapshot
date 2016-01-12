package fs

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"strings"

	"github.com/kr/fs"
	"golang.org/x/net/context"
	"golang.org/x/tools/godoc/vfs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/rwvfs"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

// repos is a local filesystem-backed implementation of the Repos
// store.
type repos struct{}

var _ store.Repos = (*repos)(nil)

const timeFormat = time.RFC3339Nano

func (s *repos) Get(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
	dir := dirForRepo(repo)
	reposVFS := rwvfs.OS(reposAbsPath(ctx))
	if !isGitRepoDir(reposVFS, dir) {
		return nil, &store.RepoNotFoundError{Repo: repo}
	}
	return s.newRepo(ctx, dir)
}

func (s *repos) GetPerms(ctx context.Context, repo string) (*sourcegraph.RepoPermissions, error) {
	return &sourcegraph.RepoPermissions{Read: true, Write: true, Admin: true}, nil
}

func (s *repos) List(ctx context.Context, opt *sourcegraph.RepoListOptions) ([]*sourcegraph.Repo, error) {
	if opt == nil {
		opt = &sourcegraph.RepoListOptions{}
	}

	var repos []*sourcegraph.Repo
	reposVFS := rwvfs.OS(reposAbsPath(ctx))
	if _, err := reposVFS.Stat("/"); os.IsNotExist(err) {
		err := reposVFS.Mkdir("/")
		if err != nil {
			return nil, err
		}
	}
	w := fs.WalkFS(".", rwvfs.Walkable(reposVFS))
	for w.Step() {
		if err := w.Err(); err != nil {
			return nil, err
		}
		fi := w.Stat()
		if fi.IsDir() || fi.Mode()&os.ModeSymlink != 0 {
			if isGitRepoDir(reposVFS, w.Path()) {
				w.SkipDir()
				repo, err := s.newRepo(ctx, w.Path())
				if err != nil {
					return nil, err
				}
				repos = append(repos, repo)
			}
		}
	}

	return reposFilter(repos, opt), nil
}

func isGitRepoDir(reposVFS rwvfs.FileSystem, path string) bool {
	// Non-bare repo
	if fi, err := reposVFS.Stat(filepath.Join(path, ".git")); err == nil && fi.IsDir() {
		return true
	}

	// Bare repo
	fi1, err1 := reposVFS.Stat(filepath.Join(path, "HEAD"))
	fi2, err2 := reposVFS.Stat(filepath.Join(path, "config"))
	fi3, err3 := reposVFS.Stat(filepath.Join(path, "refs"))
	if err1 == nil && fi1.Mode().IsRegular() && err2 == nil && fi2.Mode().IsRegular() && err3 == nil && fi3.IsDir() {
		return true
	}

	return false
}

func (s *repos) newRepo(ctx context.Context, dir string) (*sourcegraph.Repo, error) {
	dir = strings.TrimPrefix(filepath.Clean(dir), string(filepath.Separator))
	uri := repoForDir(dir)
	repo := &sourcegraph.Repo{
		URI:  uri,
		Name: filepath.Base(uri),
		VCS:  "git",
	}

	fs := rwvfs.OS(reposAbsPath(ctx))

	switch repo.VCS {
	case "git":
		var err error
		repo.DefaultBranch, err = readGitDefaultBranch(fs, dir)
		if err != nil {
			log.Printf("warning: failed to determine default branch for git repo at %s: %s. (Assuming default branch 'master'.)\n", dir, err)
			repo.DefaultBranch = "master"
		}

	case "hg":
		// TODO(sqs): un-hardcode
		repo.DefaultBranch = "default"
	}

	gitConfig, err := s.getGitConfig(ctx, fs, dir)
	if err != nil {
		log.Printf("warning: failed to read config for git repo at %s: %s", dir, err)
	}
	if gitConfig != nil {
		repo.Description = gitConfig.Sourcegraph.Description
		repo.Language = gitConfig.Sourcegraph.Language
		repo.Private = gitConfig.Sourcegraph.Private

		parseTime := func(dest **pbtypes.Timestamp, value string) {
			if value != "" {
				t, err := time.Parse(timeFormat, value)
				if err == nil {
					ts := pbtypes.NewTimestamp(t)
					*dest = &ts
				} else {
					log.Printf("warning: failed to parse time %q: %s", value, err)
				}
			}
		}
		parseTime(&repo.CreatedAt, gitConfig.Sourcegraph.CreatedAt)
		parseTime(&repo.UpdatedAt, gitConfig.Sourcegraph.UpdatedAt)
		parseTime(&repo.PushedAt, gitConfig.Sourcegraph.PushedAt)

		if origin := gitConfig.Remote["origin"]; origin != nil {
			repo.Mirror = origin.Mirror

			if repo.Mirror {
				if origin.URL == "" {
					log.Printf("warning: failed to determine clone URL for git repo at %s: %s.\n", dir, err)
				} else if strings.HasPrefix(origin.URL, "http") {
					repo.HTTPCloneURL = origin.URL
				} else if strings.HasPrefix(origin.URL, "file:") || strings.HasPrefix(origin.URL, "/") {
					// no-op; leave blank
				} else {
					repo.SSHCloneURL = origin.URL
				}
			}
		}
	}

	if !repo.Mirror {
		// The clone URL for a repo stored locally is set to the repo's path at the current host.
		repo.HTTPCloneURL = conf.AppURL(ctx).ResolveReference(router.Rel.URLToRepo(uri)).String()
		if conf.SSHURL(ctx) != nil {
			repo.SSHCloneURL = fmt.Sprintf("%s/%s", conf.SSHURL(ctx).String(), uri)
		}
	}

	return repo, nil
}

func readGitDefaultBranch(fs vfs.FileSystem, dir string) (string, error) {
	// TODO(sqs): move this to go-vcs
	var headPath string
	if _, err := fs.Stat(filepath.Join(dir, ".git")); err == nil {
		headPath = filepath.Join(dir, ".git", "HEAD") // non-bare repo
	} else if os.IsNotExist(err) {
		headPath = filepath.Join(dir, "HEAD") // bare repo
	} else {
		return "", err
	}
	data, err := vfs.ReadFile(fs, headPath)
	if err != nil {
		return "", err
	}
	data = bytes.TrimPrefix(data, []byte("ref: refs/heads/"))
	data = bytes.TrimSpace(data)
	return string(data), nil
}

func (s *repos) Create(ctx context.Context, repo *sourcegraph.Repo) error {
	if repo.VCS != "git" {
		return grpc.Errorf(codes.Unimplemented, "only git repos are supported in repo creation")
	}

	if repo.Mirror {
		if repo.HTTPCloneURL == "" && repo.SSHCloneURL == "" {
			return store.ErrRepoNeedsCloneURL
		}
	}

	dir := absolutePathForRepo(ctx, repo.URI)
	if dir == absolutePathForRepo(ctx, "") {
		return grpc.Errorf(codes.InvalidArgument, "repo must have at least one path component")
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		return grpc.Errorf(codes.AlreadyExists, "repo %s already exists", repo.URI)
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	if err := s.createBareGitRepo(ctx, repo, dir); err != nil {
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
			cmd := exec.Command(c[0], c[1:]...)
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("configuring mirrored %s repository %s (origin clone URL %s) failed with %v:\n%s", repo.VCS, repo.URI, repo.CloneURL(), err, string(out))
			}
		}
	}

	return nil
}

func (s *repos) createBareGitRepo(ctx context.Context, repo *sourcegraph.Repo, dir string) error {
	// TODO: Doing this `git init --bare` followed by a later
	//       RefreshVCS results in non-standard default branches to
	//       not be set. To fix that, either use git clone, or follow
	//       up with a `git ls-remote` and parse out HEAD.
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("creating %s repository %s failed with output:\n%s", repo.VCS, repo.URI, string(out))
	}

	if repo.Private {
		if err := s.setGitConfig(ctx, dir, "sourcegraph.private", "true"); err != nil {
			return err
		}
	}
	if repo.Description != "" {
		if err := s.setGitConfig(ctx, dir, "sourcegraph.description", repo.Description); err != nil {
			return err
		}
	}
	if repo.Language != "" {
		if err := s.setGitConfig(ctx, dir, "sourcegraph.language", repo.Language); err != nil {
			return err
		}
	}
	if repo.CreatedAt != nil {
		if err := s.setGitConfig(ctx, dir, "sourcegraph.createdat", repo.CreatedAt.Time().Format(timeFormat)); err != nil {
			return err
		}
	}

	return nil
}

func (s *repos) Update(ctx context.Context, op *store.RepoUpdate) error {
	dir := absolutePathForRepo(ctx, op.Repo.URI)

	if op.Description != "" {
		if err := s.setGitConfig(ctx, dir, "sourcegraph.description", strings.TrimSpace(op.Description)); err != nil {
			return err
		}
	}

	if op.Language != "" {
		if err := s.setGitConfig(ctx, dir, "sourcegraph.language", strings.TrimSpace(op.Language)); err != nil {
			return err
		}
	}

	if op.UpdatedAt != nil {
		if err := s.setGitConfig(ctx, dir, "sourcegraph.updatedat", op.UpdatedAt.Format(timeFormat)); err != nil {
			return err
		}
	}

	if op.PushedAt != nil {
		if err := s.setGitConfig(ctx, dir, "sourcegraph.pushedat", op.PushedAt.Format(timeFormat)); err != nil {
			return err
		}
	}

	return nil
}

func (s *repos) Delete(ctx context.Context, repo string) error {
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

// dirForRepo returns the directory (relative to the VFS's root) where
// the specified repo is located.
func dirForRepo(repoURI string) string {
	// TODO for windows support this will have to be able to handle the
	// `/` in the URI's Path
	return path.Clean(repoURI)
}

// repoForDir returns the repository URI given the directory inside
// the VFS (relative to the VFS root, like "a/b") where the repo is
// located.
func repoForDir(dir string) string {
	return strings.TrimPrefix(filepath.ToSlash(path.Clean(dir)), "/")
}

// checkGitArg returns an error if arg could be a command-line flag,
// to avoid CLI injection.
func checkGitArg(arg string) error {
	arg = strings.TrimSpace(arg)
	if strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "/") {
		return fmt.Errorf("invalid git arg %q", arg)
	}
	return nil
}
