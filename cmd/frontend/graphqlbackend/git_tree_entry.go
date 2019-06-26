package graphqlbackend

import (
	"context"
	"fmt"
	neturl "net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf/reposource"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// gitTreeEntryResolver resolves an entry in a Git tree in a repository. The entry can be any Git
// object type that is valid in a tree.
type gitTreeEntryResolver struct {
	commit *GitCommitResolver

	path string      // this tree entry's path (relative to the root)
	stat os.FileInfo // this tree entry's file info

	isRecursive bool // whether entries is populated recursively (otherwise just current level of hierarchy)
}

func (r *gitTreeEntryResolver) Path() string { return r.path }
func (r *gitTreeEntryResolver) Name() string { return path.Base(r.path) }

func (r *gitTreeEntryResolver) ToGitTree() (*gitTreeEntryResolver, bool) { return r, true }
func (r *gitTreeEntryResolver) ToGitBlob() (*gitTreeEntryResolver, bool) { return r, true }

func (r *gitTreeEntryResolver) Commit() *GitCommitResolver { return r.commit }

func (r *gitTreeEntryResolver) Repository() *repositoryResolver { return r.commit.repo }

func (r *gitTreeEntryResolver) IsRecursive() bool { return r.isRecursive }

func (r *gitTreeEntryResolver) URL(ctx context.Context) (string, error) {
	if submodule := r.Submodule(); submodule != nil {
		repoName, err := cloneURLToRepoName(ctx, submodule.URL())
		if err != nil {
			log15.Error("Failed to resolve submodule repository name from clone URL", "cloneURL", submodule.URL(), "err", err)
			return "", nil
		}
		return "/" + repoName + "@" + submodule.Commit(), nil
	}
	url, err := r.commit.repoRevURL()
	if err != nil {
		return "", err
	}
	return r.urlPath(url)
}

func (r *gitTreeEntryResolver) CanonicalURL() (string, error) {
	url, err := r.commit.canonicalRepoRevURL()
	if err != nil {
		return "", err
	}
	return r.urlPath(url)
}

func (r *gitTreeEntryResolver) urlPath(prefix string) (string, error) {
	if r.IsRoot() {
		return prefix, nil
	}

	u, err := neturl.Parse(prefix)
	if err != nil {
		return "", err
	}

	typ := "blob"
	if r.IsDirectory() {
		typ = "tree"
	}

	u.Path = path.Join(u.Path, "-", typ, r.path)
	return u.String(), nil
}

func (r *gitTreeEntryResolver) IsDirectory() bool { return r.stat.Mode().IsDir() }

func (r *gitTreeEntryResolver) ExternalURLs(ctx context.Context) ([]*externallink.Resolver, error) {
	return externallink.FileOrDir(ctx, r.commit.repo.repo, r.commit.inputRevOrImmutableRev(), r.path, r.stat.Mode().IsDir())
}

func (r *gitTreeEntryResolver) Submodule() *gitSubmoduleResolver {
	if submoduleInfo, ok := r.stat.Sys().(git.Submodule); ok {
		return &gitSubmoduleResolver{submodule: submoduleInfo}
	}
	return nil
}

func cloneURLToRepoName(ctx context.Context, cloneURL string) (string, error) {
	repoName, err := reposourceCloneURLToRepoName(ctx, cloneURL)
	if err != nil {
		return "", err
	}
	if repoName == "" {
		return "", fmt.Errorf("No matching code host found for %s", cloneURL)
	}
	return string(repoName), nil
}

// reposourceCloneURLToRepoName maps a Git clone URL (format documented here:
// https://git-scm.com/docs/git-clone#_git_urls_a_id_urls_a) to the corresponding repo name if there
// exists a code host configuration that matches the clone URL. Implicitly, it includes a code host
// configuration for github.com, even if one is not explicitly specified. Returns the empty string and nil
// error if a matching code host could not be found. This function does not actually check the code
// host to see if the repository actually exists.
func reposourceCloneURLToRepoName(ctx context.Context, cloneURL string) (repoName api.RepoName, err error) {
	if repoName := reposource.CustomCloneURLToRepoName(cloneURL); repoName != "" {
		return repoName, nil
	}

	var repoSources []reposource.RepoSource

	// The following code makes serial database calls.
	// Ideally these could be done in parallel, but the table is small
	// and I don't think real world perf is going to be bad.
	// It is also unclear to me if deterministic order is important here (it seems like it might be),
	// so if this is parallalized in the future, consider whether order is important.

	githubs, err := db.ExternalServices.ListGitHubConnections(ctx)
	if err != nil {
		return "", err
	}
	for _, c := range githubs {
		repoSources = append(repoSources, reposource.GitHub{GitHubConnection: c})
	}

	gitlabs, err := db.ExternalServices.ListGitLabConnections(ctx)
	if err != nil {
		return "", err
	}
	for _, c := range gitlabs {
		repoSources = append(repoSources, reposource.GitLab{GitLabConnection: c})
	}

	bitbuckets, err := db.ExternalServices.ListBitbucketServerConnections(ctx)
	if err != nil {
		return "", err
	}
	for _, c := range bitbuckets {
		repoSources = append(repoSources, reposource.BitbucketServer{BitbucketServerConnection: c})
	}

	awscodecommits, err := db.ExternalServices.ListAWSCodeCommitConnections(ctx)
	if err != nil {
		return "", err
	}
	for _, c := range awscodecommits {
		repoSources = append(repoSources, reposource.AWS{AWSCodeCommitConnection: c})
	}

	gitolites, err := db.ExternalServices.ListGitoliteConnections(ctx)
	if err != nil {
		return "", err
	}
	for _, c := range gitolites {
		repoSources = append(repoSources, reposource.Gitolite{GitoliteConnection: c})
	}

	// Fallback for github.com
	repoSources = append(repoSources, reposource.GitHub{
		GitHubConnection: &schema.GitHubConnection{Url: "https://github.com"},
	})
	for _, ch := range repoSources {
		repoName, err := ch.CloneURLToRepoName(cloneURL)
		if err != nil {
			return "", err
		}
		if repoName != "" {
			return repoName, nil
		}
	}

	return "", nil
}

func createFileInfo(path string, isDir bool) os.FileInfo {
	return fileInfo{path: path, isDir: isDir}
}

func (r *gitTreeEntryResolver) IsSingleChild(ctx context.Context, args *gitTreeEntryConnectionArgs) (bool, error) {
	if !r.IsDirectory() {
		return false, nil
	}
	cachedRepo, err := backend.CachedGitRepo(ctx, r.commit.repo.repo)
	if err != nil {
		return false, err
	}
	entries, err := git.ReadDir(ctx, *cachedRepo, api.CommitID(r.commit.OID()), filepath.Dir(r.path), false)
	if err != nil {
		return false, err
	}
	return len(entries) == 1, nil
}

type fileInfo struct {
	path  string
	isDir bool
}

func (f fileInfo) Name() string { return f.path }
func (f fileInfo) Size() int64  { return 0 }
func (f fileInfo) IsDir() bool  { return f.isDir }
func (f fileInfo) Mode() os.FileMode {
	if f.IsDir() {
		return os.ModeDir
	}
	return 0
}
func (f fileInfo) ModTime() time.Time { return time.Now() }
func (f fileInfo) Sys() interface{}   { return interface{}(nil) }
