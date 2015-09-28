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
	"regexp"

	"strings"

	"github.com/kr/fs"
	"golang.org/x/net/context"
	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/rwvfs"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
)

// Repos is a local filesystem-backed implementation of the
// base Repos.
type Repos struct{}

var _ store.Repos = (*Repos)(nil)

func (s *Repos) Get(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
	dir := dirForRepo(repo)
	if fi, err := os.Stat(filepath.Join(reposAbsPath(ctx), dir)); os.IsNotExist(err) {
		return nil, &store.RepoNotFoundError{Repo: repo}
	} else if err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return nil, &os.PathError{Op: "Repos.Get", Path: dir, Err: os.ErrInvalid}
	}
	return s.newRepo(ctx, dir)
}

func (s *Repos) GetPerms(ctx context.Context, repo string) (*sourcegraph.RepoPermissions, error) {
	return &sourcegraph.RepoPermissions{Read: true, Write: true, Admin: true}, nil
}

func (s *Repos) List(ctx context.Context, opt *sourcegraph.RepoListOptions) ([]*sourcegraph.Repo, error) {
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
	skip := opt.ListOptions.Offset()
	limit := opt.ListOptions.Limit()
	for w.Step() {
		if err := w.Err(); err != nil {
			return nil, err
		}
		fi := w.Stat()
		if fi.IsDir() || fi.Mode()&os.ModeSymlink != 0 {
			if isGitRepoDir(reposVFS, w.Path()) {
				repo, err := s.newRepo(ctx, w.Path())
				if err != nil {
					return nil, err
				}
				w.SkipDir()

				if !repoSatisfiesOpts(repo, opt) {
					continue
				}

				// Paginate.
				if skip == 0 {
					repos = append(repos, repo)
					limit--
				} else {
					skip--
				}
				if limit == 0 {
					break
				}
			}
		}
	}

	return repos, nil
}

func repoSatisfiesOpts(repo *sourcegraph.Repo, opt *sourcegraph.RepoListOptions) bool {
	if opt == nil {
		return true
	}

	if query := opt.Query; query != "" {
		ok := func() bool {
			query = strings.ToLower(query)
			uri, name := strings.ToLower(repo.URI), strings.ToLower(repo.Name)

			if query == uri || strings.HasPrefix(name, query) {
				return true
			}

			// Match any path component prefix.
			for _, pc := range strings.Split(uri, "/") {
				if strings.HasPrefix(pc, query) {
					return true
				}
			}

			return false
		}()
		if !ok {
			return false
		}
	}

	if len(opt.URIs) > 0 {
		uriMatch := false
		for _, uri := range opt.URIs {
			if strings.EqualFold(uri, repo.URI) {
				uriMatch = true
				break
			}
		}
		if !uriMatch {
			return false
		}
	}

	return true
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

func (s *Repos) newRepo(ctx context.Context, dir string) (*sourcegraph.Repo, error) {
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

	isPrivate, err := readGitIsPrivate(fs, dir)
	if err != nil {
		log.Printf("warning: failed to determine whether git repo at %s is private: %s.\n", dir, err)
	} else if isPrivate {
		repo.Private = true
	}

	isMirror, err := readGitIsMirror(fs, dir)
	if err != nil {
		log.Printf("warning: failed to determine whether git repo at %s is a mirror: %s.\n", dir, err)
	} else if isMirror {
		repo.Mirror = true
	}
	if isMirror {
		// Mirrored repos are set to clone from their original source.
		cloneURL, err := readGitOriginURL(fs, dir)
		if err != nil || cloneURL == "" {
			log.Printf("warning: failed to determine clone URL for git repo at %s: %s.\n", dir, err)
		} else if strings.HasPrefix(cloneURL, "http") {
			repo.HTTPCloneURL = cloneURL
		} else if strings.HasPrefix(cloneURL, "file:") || strings.HasPrefix(cloneURL, "/") {
			// no-op; leave blank
		} else {
			repo.SSHCloneURL = cloneURL
		}
	} else {
		// The clone URL for a repo stored locally is set to the repo's path at the current host.
		repo.HTTPCloneURL = conf.AppURL(ctx).ResolveReference(router.Rel.URLToRepo(uri)).String()
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

var originURLPat = regexp.MustCompile(`\[remote "origin"\]
\s*url\s*=\s*(.*)`)

func readGitOriginURL(fs vfs.FileSystem, dir string) (string, error) {
	// TODO(sqs): move this to go-vcs; hacky
	var configPath string
	if _, err := fs.Stat(filepath.Join(dir, ".git")); err == nil {
		configPath = filepath.Join(dir, ".git", "config") // non-bare repo
	} else if os.IsNotExist(err) {
		configPath = filepath.Join(dir, "config") // bare repo
	} else {
		return "", err
	}
	data, err := vfs.ReadFile(fs, configPath)
	if err != nil {
		return "", err
	}
	matches := originURLPat.FindStringSubmatch(string(data))
	if len(matches) == 2 {
		return matches[1], nil
	}
	return "", nil
}

func readGitIsPrivate(fs vfs.FileSystem, dir string) (bool, error) {
	// TODO: Figure out long term solution.
	var configPath string
	if _, err := fs.Stat(filepath.Join(dir, ".git")); err == nil {
		configPath = filepath.Join(dir, ".git", "config") // non-bare repo
	} else if os.IsNotExist(err) {
		configPath = filepath.Join(dir, "config") // bare repo
	} else {
		return false, err
	}
	data, err := vfs.ReadFile(fs, configPath)
	if err != nil {
		return false, err
	}
	return bytes.Contains(data, []byte("private = true")), nil
}

func readGitIsMirror(fs vfs.FileSystem, dir string) (bool, error) {
	// TODO(sqs): move this to go-vcs; hacky
	var configPath string
	if _, err := fs.Stat(filepath.Join(dir, ".git")); err == nil {
		configPath = filepath.Join(dir, ".git", "config") // non-bare repo
	} else if os.IsNotExist(err) {
		configPath = filepath.Join(dir, "config") // bare repo
	} else {
		return false, err
	}
	data, err := vfs.ReadFile(fs, configPath)
	if err != nil {
		return false, err
	}
	return bytes.Contains(data, []byte("mirror = true")), nil
}

func (s *Repos) Create(ctx context.Context, repo *sourcegraph.Repo) (*sourcegraph.Repo, error) {
	if repo.VCS != "git" {
		return nil, &sourcegraph.NotImplementedError{What: "only git is supported in Repos.Create"}
	}

	if repo.Mirror {
		if repo.HTTPCloneURL == "" && repo.SSHCloneURL == "" {
			return nil, store.ErrRepoNeedsCloneURL
		}
	}

	dir := absolutePathForRepo(ctx, repo.URI)
	if dir == absolutePathForRepo(ctx, "") {
		return nil, errors.New("Repos.Create needs at least one path element")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	// TODO: Doing this `git init --bare` followed by a later RefreshVCS results in non-standard default branches
	//       to not be set. To fix that, either use git clone, or follow up with a `git ls-remote` and parse out HEAD.

	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("creating %s repository %s failed with output:\n%s", repo.VCS, repo.URI, string(out))
	}

	if repo.Private {
		cmd := exec.Command("git", "config", "sourcegraph.private", "true")
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("making %s repository %s private failed with %v:\n%s", repo.VCS, repo.URI, err, string(out))
		}
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
				return nil, fmt.Errorf("configuring mirrored %s repository %s (origin clone URL %s) failed with %v:\n%s", repo.VCS, repo.URI, repo.CloneURL(), err, string(out))
			}
		}
	}

	return &sourcegraph.Repo{URI: repo.URI, VCS: repo.VCS, DefaultBranch: "master"}, nil
}

func (s *Repos) Delete(ctx context.Context, repo string) error {
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
	return strings.TrimPrefix(path.Clean(dir), "/")
}
