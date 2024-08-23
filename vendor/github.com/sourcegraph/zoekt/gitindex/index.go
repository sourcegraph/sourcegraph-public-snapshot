// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package gitindex provides functions for indexing Git repositories.
package gitindex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/build"
	"github.com/sourcegraph/zoekt/ignore"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	git "github.com/go-git/go-git/v5"
)

// RepoModTime returns the time of last fetch of a git repository.
func RepoModTime(dir string) (time.Time, error) {
	var last time.Time
	refDir := filepath.Join(dir, "refs")
	if _, err := os.Lstat(refDir); err == nil {
		if err := filepath.Walk(refDir,
			func(_ string, fi os.FileInfo, _ error) error {
				if !fi.IsDir() && last.Before(fi.ModTime()) {
					last = fi.ModTime()
				}
				return nil
			}); err != nil {
			return last, err
		}
	}

	// git gc compresses refs into the following file:
	for _, fn := range []string{"info/refs", "packed-refs"} {
		if fi, err := os.Lstat(filepath.Join(dir, fn)); err == nil && !fi.IsDir() && last.Before(fi.ModTime()) {
			last = fi.ModTime()
		}
	}

	return last, nil
}

// FindGitRepos finds directories holding git repositories below the
// given directory. It will find both bare and the ".git" dirs in
// non-bare repositories. It returns the full path including the dir
// passed in.
func FindGitRepos(dir string) ([]string, error) {
	arg, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	var dirs []string
	if err := filepath.Walk(arg, func(name string, fi os.FileInfo, err error) error {
		// Best-effort, ignore filepath.Walk failing
		if err != nil {
			return nil
		}

		if fi, err := os.Lstat(filepath.Join(name, ".git")); err == nil && fi.IsDir() {
			dirs = append(dirs, filepath.Join(name, ".git"))
			return filepath.SkipDir
		}

		if !strings.HasSuffix(name, ".git") || !fi.IsDir() {
			return nil
		}

		fi, err = os.Lstat(filepath.Join(name, "objects"))
		if err != nil || !fi.IsDir() {
			return nil
		}

		dirs = append(dirs, name)
		return filepath.SkipDir
	}); err != nil {
		return nil, err
	}

	return dirs, nil
}

// setTemplates fills in URL templates for known git hosting
// sites.
func setTemplates(repo *zoekt.Repository, u *url.URL, typ string) error {
	if u.Scheme == "ssh+git" {
		u.Scheme = "https"
		u.User = nil
	}

	repo.URL = u.String()
	switch typ {
	case "gitiles":
		// eg. https://gerrit.googlesource.com/gitiles/+/master/tools/run_dev.sh#20
		repo.CommitURLTemplate = u.String() + "/+/{{.Version}}"
		repo.FileURLTemplate = u.String() + "/+/{{.Version}}/{{.Path}}"
		repo.LineFragmentTemplate = "#{{.LineNumber}}"
	case "github":
		// eg. https://github.com/hanwen/go-fuse/blob/notify/genversion.sh#L10
		repo.CommitURLTemplate = u.String() + "/commit/{{.Version}}"
		repo.FileURLTemplate = u.String() + "/blob/{{.Version}}/{{.Path}}"
		repo.LineFragmentTemplate = "#L{{.LineNumber}}"
	case "cgit":
		// http://git.savannah.gnu.org/cgit/lilypond.git/tree/elisp/lilypond-mode.el?h=dev/philh&id=b2ca0fefe3018477aaca23b6f672c7199ba5238e#n100
		repo.CommitURLTemplate = u.String() + "/commit/?id={{.Version}}"
		repo.FileURLTemplate = u.String() + "/tree/{{.Path}}/?id={{.Version}}"
		repo.LineFragmentTemplate = "#n{{.LineNumber}}"
	case "gitweb":
		// https://gerrit.libreoffice.org/gitweb?p=online.git;a=blob;f=Makefile.am;h=cfcfd7c36fbae10e269653dc57a9b68c92d4c10b;hb=848145503bf7b98ce4a4aa0a858a0d71dd0dbb26#l10
		repo.FileURLTemplate = u.String() + ";a=blob;f={{.Path}};hb={{.Version}}"
		repo.CommitURLTemplate = u.String() + ";a=commit;h={{.Version}}"
		repo.LineFragmentTemplate = "#l{{.LineNumber}}"
	case "source.bazel.build":
		// https://source.bazel.build/bazel/+/57bc201346e61c62a921c1cbf32ad24f185c10c9
		// https://source.bazel.build/bazel/+/57bc201346e61c62a921c1cbf32ad24f185c10c9:tools/cpp/BUILD.empty;l=10
		repo.CommitURLTemplate = u.String() + "/+/{{.Version}}"
		repo.FileURLTemplate = u.String() + "/+/{{.Version}}:{{.Path}}"
		repo.LineFragmentTemplate = ";l={{.LineNumber}}"
	case "bitbucket-server":
		// https://<bitbucketserver-host>/projects/<project>/repos/<repo>/commits/5be7ca73b898bf17a08e607918accfdeafe1e0bc
		// https://<bitbucketserver-host>/projects/<project>/repos/<repo>/browse/<file>?at=5be7ca73b898bf17a08e607918accfdeafe1e0bc
		repo.CommitURLTemplate = u.String() + "/commits/{{.Version}}"
		repo.FileURLTemplate = u.String() + "/{{.Path}}?at={{.Version}}"
		repo.LineFragmentTemplate = "#{{.LineNumber}}"
	case "gitlab":
		// https://gitlab.com/gitlab-org/omnibus-gitlab/-/commit/b152c864303dae0e55377a1e2c53c9592380ffed
		// https://gitlab.com/gitlab-org/omnibus-gitlab/-/blob/aad04155b3f6fc50ede88aedaee7fc624d481149/files/gitlab-config-template/gitlab.rb.template
		repo.CommitURLTemplate = u.String() + "/-/commit/{{.Version}}"
		repo.FileURLTemplate = u.String() + "/-/blob/{{.Version}}/{{.Path}}"
		repo.LineFragmentTemplate = "#L{{.LineNumber}}"
	case "gitea":
		repo.CommitURLTemplate = u.String() + "/commit/{{.Version}}"
		// NOTE The `display=source` query parameter is required to disable file rendering.
		// Since line numbers are disabled in rendered files, you wouldn't be able to jump to
		// a line without `display=source`. This is supported since gitea 1.17.0.
		// When /src/{{.Version}} is used it will redirect to /src/commit/{{.Version}},
		// but the query  parameters are obmitted.
		repo.FileURLTemplate = u.String() + "/src/commit/{{.Version}}/{{.Path}}?display=source"
		repo.LineFragmentTemplate = "#L{{.LineNumber}}"
	default:
		return fmt.Errorf("URL scheme type %q unknown", typ)
	}
	return nil
}

// getCommit returns a tree object for the given reference.
func getCommit(repo *git.Repository, prefix, ref string) (*object.Commit, error) {
	sha1, err := repo.ResolveRevision(plumbing.Revision(ref))
	// ref might be a branch name (e.g. "master") add branch prefix and try again.
	if err != nil {
		sha1, err = repo.ResolveRevision(plumbing.Revision(filepath.Join(prefix, ref)))
	}
	if err != nil {
		return nil, err
	}

	commitObj, err := repo.CommitObject(*sha1)
	if err != nil {
		return nil, err
	}
	return commitObj, nil
}

func configLookupRemoteURL(cfg *config.Config, key string) string {
	rc := cfg.Remotes[key]
	if rc == nil || len(rc.URLs) == 0 {
		return ""
	}
	return rc.URLs[0]
}

var sshRelativeURLRegexp = regexp.MustCompile(`^([^@]+)@([^:]+):(.*)$`)

func setTemplatesFromConfig(desc *zoekt.Repository, repoDir string) error {
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return err
	}

	cfg, err := repo.Config()
	if err != nil {
		return err
	}

	sec := cfg.Raw.Section("zoekt")

	webURLStr := sec.Options.Get("web-url")
	webURLType := sec.Options.Get("web-url-type")

	if webURLType != "" && webURLStr != "" {
		webURL, err := url.Parse(webURLStr)
		if err != nil {
			return err
		}
		if err := setTemplates(desc, webURL, webURLType); err != nil {
			return err
		}
	} else if webURLStr != "" {
		desc.URL = webURLStr
	}

	name := sec.Options.Get("name")
	if name != "" {
		desc.Name = name
	} else {
		remoteURL := configLookupRemoteURL(cfg, "origin")
		if remoteURL == "" {
			return nil
		}
		if sm := sshRelativeURLRegexp.FindStringSubmatch(remoteURL); sm != nil {
			user := sm[1]
			host := sm[2]
			path := sm[3]

			remoteURL = fmt.Sprintf("ssh+git://%s@%s/%s", user, host, path)
		}

		u, err := url.Parse(remoteURL)
		if err != nil {
			return err
		}
		if err := SetTemplatesFromOrigin(desc, u); err != nil {
			return err
		}
	}

	id, _ := strconv.ParseUint(sec.Options.Get("repoid"), 10, 32)
	desc.ID = uint32(id)

	if desc.RawConfig == nil {
		desc.RawConfig = map[string]string{}
	}
	for _, o := range sec.Options {
		desc.RawConfig[o.Key] = o.Value
	}

	// Ranking info.

	// Github:
	traction := 0
	for _, s := range []string{"github-stars", "github-forks", "github-watchers", "github-subscribers"} {
		f, err := strconv.Atoi(sec.Options.Get(s))
		if err == nil {
			traction += f
		}
	}

	if strings.Contains(desc.Name, "googlesource.com/") && traction == 0 {
		// Pretend everything on googlesource.com has 1000
		// github stars.
		traction = 1000
	}

	if traction > 0 {
		l := math.Log(float64(traction))
		desc.Rank = uint16((1.0 - 1.0/math.Pow(1+l, 0.6)) * 10000)
	}

	return nil
}

// SetTemplatesFromOrigin fills in templates based on the origin URL.
func SetTemplatesFromOrigin(desc *zoekt.Repository, u *url.URL) error {
	desc.Name = filepath.Join(u.Host, strings.TrimSuffix(u.Path, ".git"))

	if strings.HasSuffix(u.Host, ".googlesource.com") {
		return setTemplates(desc, u, "gitiles")
	} else if u.Host == "github.com" {
		u.Path = strings.TrimSuffix(u.Path, ".git")
		return setTemplates(desc, u, "github")
	} else {
		return fmt.Errorf("unknown git hosting site %q", u)
	}
}

// The Options structs controls details of the indexing process.
type Options struct {
	// The repository to be indexed.
	RepoDir string

	// If set, follow submodule links. This requires RepoCacheDir to be set.
	Submodules bool

	// If set, skip indexing if the existing index shard is newer
	// than the refs in the repository.
	Incremental bool

	// Don't error out if some branch is missing
	AllowMissingBranch bool

	// Specifies the root of a Repository cache. Needed for submodule indexing.
	RepoCacheDir string

	// Indexing options.
	BuildOptions build.Options

	// Prefix of the branch to index, e.g. `remotes/origin`.
	BranchPrefix string

	// List of branch names to index, e.g. []string{"HEAD", "stable"}
	Branches []string

	// DeltaShardNumberFallbackThreshold defines an upper limit (inclusive) on the number of preexisting shards
	// that can exist before attempting another delta build. If the number of preexisting shards exceeds this threshold,
	// then a normal build will be performed instead.
	//
	// If DeltaShardNumberFallbackThreshold is 0, then this fallback behavior is disabled:
	// a delta build will always be performed regardless of the number of preexisting shards.
	DeltaShardNumberFallbackThreshold uint64
}

func expandBranches(repo *git.Repository, bs []string, prefix string) ([]string, error) {
	var result []string
	for _, b := range bs {
		// Sourcegraph: We disable resolving refs. We want to return the exact ref
		// requested so we can match it up.
		if b == "HEAD" && false {
			ref, err := repo.Head()
			if err != nil {
				return nil, err
			}

			result = append(result, strings.TrimPrefix(ref.Name().String(), prefix))
			continue
		}

		if strings.Contains(b, "*") {
			iter, err := repo.Branches()
			if err != nil {
				return nil, err
			}

			defer iter.Close()
			for {
				ref, err := iter.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					return nil, err
				}

				name := ref.Name().Short()
				if matched, err := filepath.Match(b, name); err != nil {
					return nil, err
				} else if !matched {
					continue
				}

				result = append(result, strings.TrimPrefix(name, prefix))
			}
			continue
		}

		result = append(result, b)
	}

	return result, nil
}

// IndexGitRepo indexes the git repository as specified by the options.
// The returned bool indicates whether the index was updated as a result. This
// can be informative if doing incremental indexing.
func IndexGitRepo(opts Options) (bool, error) {
	return indexGitRepo(opts, gitIndexConfig{})
}

// indexGitRepo indexes the git repository as specified by the options and the provided gitIndexConfig.
// The returned bool indicates whether the index was updated as a result. This
// can be informative if doing incremental indexing.
func indexGitRepo(opts Options, config gitIndexConfig) (bool, error) {
	prepareDeltaBuild := prepareDeltaBuild
	if config.prepareDeltaBuild != nil {
		prepareDeltaBuild = config.prepareDeltaBuild
	}

	prepareNormalBuild := prepareNormalBuild
	if config.prepareNormalBuild != nil {
		prepareNormalBuild = config.prepareNormalBuild
	}

	// Set max thresholds, since we use them in this function.
	opts.BuildOptions.SetDefaults()
	if opts.RepoDir == "" {
		return false, fmt.Errorf("gitindex: must set RepoDir")
	}

	opts.BuildOptions.RepositoryDescription.Source = opts.RepoDir
	repo, err := git.PlainOpen(opts.RepoDir)
	if err != nil {
		return false, fmt.Errorf("git.PlainOpen: %w", err)
	}

	if err := setTemplatesFromConfig(&opts.BuildOptions.RepositoryDescription, opts.RepoDir); err != nil {
		log.Printf("setTemplatesFromConfig(%s): %s", opts.RepoDir, err)
	}

	branches, err := expandBranches(repo, opts.Branches, opts.BranchPrefix)
	if err != nil {
		return false, fmt.Errorf("expandBranches: %w", err)
	}
	for _, b := range branches {
		commit, err := getCommit(repo, opts.BranchPrefix, b)
		if err != nil {
			if opts.AllowMissingBranch && err.Error() == "reference not found" {
				continue
			}

			return false, fmt.Errorf("getCommit(%q, %q): %w", opts.BranchPrefix, b, err)
		}

		opts.BuildOptions.RepositoryDescription.Branches = append(opts.BuildOptions.RepositoryDescription.Branches, zoekt.RepositoryBranch{
			Name:    b,
			Version: commit.Hash.String(),
		})

		if when := commit.Committer.When; when.After(opts.BuildOptions.RepositoryDescription.LatestCommitDate) {
			opts.BuildOptions.RepositoryDescription.LatestCommitDate = when
		}
	}

	if opts.Incremental && opts.BuildOptions.IncrementalSkipIndexing() {
		return false, nil
	}

	// branch => (path, sha1) => repo.
	var repos map[fileKey]BlobLocation

	// fileKey => branches
	var branchMap map[fileKey][]string

	// Branch => Repo => SHA1
	var branchVersions map[string]map[string]plumbing.Hash

	// set of file paths that have been changed or deleted since
	// the last indexed commit
	//
	// These only have an effect on delta builds
	var changedOrRemovedFiles []string

	if opts.BuildOptions.IsDelta {
		repos, branchMap, branchVersions, changedOrRemovedFiles, err = prepareDeltaBuild(opts, repo)
		if err != nil {
			log.Printf("delta build: falling back to normal build since delta build failed, repository=%q, err=%s", opts.BuildOptions.RepositoryDescription.Name, err)
			opts.BuildOptions.IsDelta = false
		}
	}

	if !opts.BuildOptions.IsDelta {
		repos, branchMap, branchVersions, err = prepareNormalBuild(opts, repo)
		if err != nil {
			return false, fmt.Errorf("preparing normal build: %w", err)
		}
	}

	reposByPath := map[string]BlobLocation{}
	for key, location := range repos {
		reposByPath[key.SubRepoPath] = location
	}

	opts.BuildOptions.SubRepositories = map[string]*zoekt.Repository{}
	for path, location := range reposByPath {
		tpl := opts.BuildOptions.RepositoryDescription
		if path != "" {
			tpl = zoekt.Repository{URL: location.URL.String()}
			if err := SetTemplatesFromOrigin(&tpl, location.URL); err != nil {
				log.Printf("setTemplatesFromOrigin(%s, %s): %s", path, location.URL, err)
			}
		}
		opts.BuildOptions.SubRepositories[path] = &tpl
	}

	for _, br := range opts.BuildOptions.RepositoryDescription.Branches {
		for path, repo := range opts.BuildOptions.SubRepositories {
			id := branchVersions[br.Name][path]
			repo.Branches = append(repo.Branches, zoekt.RepositoryBranch{
				Name:    br.Name,
				Version: id.String(),
			})
		}
	}

	builder, err := build.NewBuilder(opts.BuildOptions)
	if err != nil {
		return false, fmt.Errorf("build.NewBuilder: %w", err)
	}

	var ranks repoPathRanks
	var meanRank float64
	if opts.BuildOptions.DocumentRanksPath != "" {
		data, err := os.ReadFile(opts.BuildOptions.DocumentRanksPath)
		if err != nil {
			return false, err
		}

		err = json.Unmarshal(data, &ranks)
		if err != nil {
			return false, err
		}

		// Compute the mean rank for this repository. Note: we overwrite the rank
		// mean that's stored in the document ranks file, since that currently
		// represents a global mean rank across repos, which is not what we want.
		numRanks := len(ranks.Paths)
		if numRanks > 0 {
			for _, rank := range ranks.Paths {
				meanRank += rank
			}
			ranks.MeanRank = meanRank / float64(numRanks)
		}
	}

	// we don't need to check error, since we either already have an error, or
	// we returning the first call to builder.Finish.
	defer builder.Finish() // nolint:errcheck

	for _, f := range changedOrRemovedFiles {
		builder.MarkFileAsChangedOrRemoved(f)
	}

	var names []string
	fileKeys := map[string][]fileKey{}
	totalFiles := 0

	for key := range repos {
		n := key.FullPath()
		fileKeys[n] = append(fileKeys[n], key)
		names = append(names, n)
		totalFiles++
	}

	sort.Strings(names)
	names = uniq(names)

	log.Printf("attempting to index %d total files", totalFiles)
	for _, name := range names {
		keys := fileKeys[name]

		for _, key := range keys {
			doc, err := createDocument(key, repos, branchMap, ranks, opts.BuildOptions)
			if err != nil {
				return false, err
			}

			if err := builder.Add(doc); err != nil {
				return false, fmt.Errorf("error adding document with name %s: %w", key.FullPath(), err)
			}
		}
	}

	return true, builder.Finish()
}

type repoPathRanks struct {
	MeanRank float64            `json:"mean_reference_count"`
	Paths    map[string]float64 `json:"paths"`
}

// rank returns the rank for a given path. It uses these rules:
//   - If we have a concrete rank for this file, always use it
//   - If there's no rank, and it's a low priority file like a test, then use rank 0
//   - Otherwise use the mean rank of this repository, to avoid giving it a big disadvantage
func (r repoPathRanks) rank(path string) float64 {
	if rank, ok := r.Paths[path]; ok {
		return rank
	} else if build.IsLowPriority(path) {
		return 0.0
	} else {
		return r.MeanRank
	}
}

func newIgnoreMatcher(tree *object.Tree) (*ignore.Matcher, error) {
	ignoreFile, err := tree.File(ignore.IgnoreFile)
	if err == object.ErrFileNotFound {
		return &ignore.Matcher{}, nil
	}
	if err != nil {
		return nil, err
	}
	content, err := ignoreFile.Contents()
	if err != nil {
		return nil, err
	}
	return ignore.ParseIgnoreFile(strings.NewReader(content))
}

// prepareDeltaBuildFunc is a function that calculates the necessary metadata for preparing
// a build.Builder instance for generating a delta build.
type prepareDeltaBuildFunc func(options Options, repository *git.Repository) (repos map[fileKey]BlobLocation, branchMap map[fileKey][]string, branchVersions map[string]map[string]plumbing.Hash, changedOrDeletedPaths []string, err error)

// prepareNormalBuildFunc is a function that calculates the necessary metadata for preparing
// a build.Builder instance for generating a normal build.
type prepareNormalBuildFunc func(options Options, repository *git.Repository) (repos map[fileKey]BlobLocation, branchMap map[fileKey][]string, branchVersions map[string]map[string]plumbing.Hash, err error)

type gitIndexConfig struct {
	// prepareDeltaBuild, if not nil, is the function that is used to calculate the metadata that will be used to
	// prepare the build.Builder instance for generating a delta build.
	//
	// If prepareDeltaBuild is nil, gitindex.prepareDeltaBuild will be used instead.
	prepareDeltaBuild prepareDeltaBuildFunc

	// prepareNormalBuild, if not nil, is the function that is used to calculate the metadata that will be used to
	// prepare the build.Builder instance for generating a normal build.
	//
	// If prepareNormalBuild is nil, gitindex.prepareNormalBuild will be used instead.
	prepareNormalBuild prepareNormalBuildFunc
}

func prepareDeltaBuild(options Options, repository *git.Repository) (repos map[fileKey]BlobLocation, branchMap map[fileKey][]string, branchVersions map[string]map[string]plumbing.Hash, changedOrDeletedPaths []string, err error) {
	if options.Submodules {
		return nil, nil, nil, nil, fmt.Errorf("delta builds currently don't support submodule indexing")
	}

	// discover what commits we indexed during our last build
	existingRepository, _, ok, err := options.BuildOptions.FindRepositoryMetadata()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get repository metadata: %w", err)
	}

	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("no existing shards found for repository")
	}

	if options.DeltaShardNumberFallbackThreshold > 0 {
		// HACK: For our interim compaction strategy, we force a full normal index once
		// the number of shards on disk for this repository exceeds the provided threshold.
		//
		// This strategy obviously isn't optimal (as an example: we currently can't differentiate
		// between "normal" and "delta" shards, so repositories like the gigarepo that generate a large number of shards per
		// build would be disproportionately affected by this), but it'll allow us to continue experimenting on real workloads
		// while we create a better compaction strategy).

		oldShards := options.BuildOptions.FindAllShards()
		if uint64(len(oldShards)) > options.DeltaShardNumberFallbackThreshold {
			return nil, nil, nil, nil, fmt.Errorf("number of existing shards (%d) > requested shard threshold (%d)", len(oldShards), options.DeltaShardNumberFallbackThreshold)
		}
	}

	// Check to see if the set of branch names is consistent with what we last indexed.
	// If it isn't consistent, that we can't proceed with a delta build (and the caller should fall back to a
	// normal one).

	if !build.BranchNamesEqual(existingRepository.Branches, options.BuildOptions.RepositoryDescription.Branches) {
		var existingBranchNames []string
		for _, b := range existingRepository.Branches {
			existingBranchNames = append(existingBranchNames, b.Name)
		}

		var optionsBranchNames []string
		for _, b := range options.BuildOptions.RepositoryDescription.Branches {
			optionsBranchNames = append(optionsBranchNames, b.Name)
		}

		existingBranchList := strings.Join(existingBranchNames, ", ")
		optionsBranchList := strings.Join(optionsBranchNames, ", ")

		return nil, nil, nil, nil, fmt.Errorf("requested branch set in build options (%q) != branch set found on disk (%q) - branch set must be the same for delta shards", optionsBranchList, existingBranchList)
	}

	// Check if the build options hash does not match the repository metadata's hash
	// If it does not match then one or more index options has changed and will require a normal build instead of a delta build
	if options.BuildOptions.GetHash() != existingRepository.IndexOptions {
		return nil, nil, nil, nil, fmt.Errorf("one or more index options previously stored for repository %s (ID: %d) does not match the index options for this requested build; These index option updates are incompatible with delta build. new index options: %+v", existingRepository.Name, existingRepository.ID, options.BuildOptions.HashOptions())
	}

	// branch => (path, sha1) => repo.
	repos = map[fileKey]BlobLocation{}

	// fileKey => branches
	branchMap = map[fileKey][]string{}

	// branch name -> git worktree at most current commit
	branchToCurrentTree := make(map[string]*object.Tree, len(options.Branches))

	for _, b := range options.Branches {
		commit, err := getCommit(repository, options.BranchPrefix, b)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("getting last current commit for branch %q: %w", b, err)
		}

		tree, err := commit.Tree()
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("getting current git tree for branch %q: %w", b, err)
		}

		branchToCurrentTree[b] = tree
	}

	rawURL := options.BuildOptions.RepositoryDescription.URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("parsing repository URL %q: %w", rawURL, err)
	}

	// TODO: Support repository submodules for delta builds
	// For this prototype, we are ignoring repository submodules, which means that we can use the same
	// blob location for all files
	hackSharedBlobLocation := BlobLocation{
		Repo: repository,
		URL:  u,
	}

	// loop over all branches, calculate the diff between our
	// last indexed commit and the current commit, and add files mentioned in the diff
	for _, branch := range existingRepository.Branches {
		lastIndexedCommit, err := getCommit(repository, "", branch.Version)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("getting last indexed commit for branch %q: %w", branch.Name, err)
		}

		lastIndexedTree, err := lastIndexedCommit.Tree()
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("getting lasted indexed git tree for branch %q: %w", branch.Name, err)
		}

		changes, err := object.DiffTreeWithOptions(context.Background(), lastIndexedTree, branchToCurrentTree[branch.Name], &object.DiffTreeOptions{DetectRenames: false})
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("generating changeset for branch %q: %w", branch.Name, err)
		}

		for i, c := range changes {
			oldFile, newFile, err := c.Files()
			if err != nil {
				return nil, nil, nil, nil, fmt.Errorf("change #%d: getting files before and after change: %w", i, err)
			}

			if newFile != nil {
				// note: newFile.Name could be a path that isn't relative to the repository root - using the
				// change's Name field is the only way that @ggilmore saw to get the full path relative to the root
				newFileRelativeRootPath := c.To.Name

				// TODO@ggilmore: HACK - remove once ignore files are supported in delta builds
				if newFileRelativeRootPath == ignore.IgnoreFile {
					return nil, nil, nil, nil, fmt.Errorf("%q file is not yet supported in delta builds", ignore.IgnoreFile)
				}

				// either file is added or renamed, so we need to add the new version to the build
				file := fileKey{Path: newFileRelativeRootPath, ID: newFile.Hash}
				repos[file] = hackSharedBlobLocation
				branchMap[file] = append(branchMap[file], branch.Name)
			}

			if oldFile == nil {
				// file added - nothing more to do
				continue
			}

			// Note: oldFile.Name could be a path that isn't relative to the repository root - using the
			// change's "Name" field is the only way that ggilmore saw to get the full path relative to the root
			oldFileRelativeRootPath := c.From.Name

			if oldFileRelativeRootPath == ignore.IgnoreFile {
				return nil, nil, nil, nil, fmt.Errorf("%q file is not yet supported in delta builds", ignore.IgnoreFile)
			}

			// The file is either modified or deleted. So, we need to add ALL versions
			// of the old file (across all branches) to the build.
			for b, currentTree := range branchToCurrentTree {
				f, err := currentTree.File(oldFileRelativeRootPath)
				if err != nil {
					// the file doesn't exist in this branch
					if errors.Is(err, object.ErrFileNotFound) {
						continue
					}

					return nil, nil, nil, nil, fmt.Errorf("getting hash for file %q in branch %q: %w", oldFile.Name, b, err)
				}

				file := fileKey{Path: oldFileRelativeRootPath, ID: f.ID()}
				repos[file] = hackSharedBlobLocation
				branchMap[file] = append(branchMap[file], b)
			}

			changedOrDeletedPaths = append(changedOrDeletedPaths, oldFileRelativeRootPath)
		}
	}

	// we need to de-duplicate the branch map before returning it - it's possible for the same
	// branch to have been added multiple times if a file has been modified across multiple commits

	for file, branches := range branchMap {
		sort.Strings(branches)
		branchMap[file] = uniq(branches)
	}

	// we also need to de-duplicate the list of changed or deleted file paths, it's also possible to have duplicates
	// for the same reasoning as above

	sort.Strings(changedOrDeletedPaths)
	changedOrDeletedPaths = uniq(changedOrDeletedPaths)

	return repos, branchMap, nil, changedOrDeletedPaths, nil
}

func prepareNormalBuild(options Options, repository *git.Repository) (repos map[fileKey]BlobLocation, branchMap map[fileKey][]string, branchVersions map[string]map[string]plumbing.Hash, err error) {
	var repoCache *RepoCache
	if options.Submodules {
		repoCache = NewRepoCache(options.RepoCacheDir)
	}

	// branch => (path, sha1) => repo.
	repos = map[fileKey]BlobLocation{}

	// fileKey => branches
	branchMap = map[fileKey][]string{}

	// Branch => Repo => SHA1
	branchVersions = map[string]map[string]plumbing.Hash{}

	branches, err := expandBranches(repository, options.Branches, options.BranchPrefix)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("expandBranches: %w", err)
	}

	for _, b := range branches {
		commit, err := getCommit(repository, options.BranchPrefix, b)
		if err != nil {
			if options.AllowMissingBranch && err.Error() == "reference not found" {
				continue
			}

			return nil, nil, nil, fmt.Errorf("getCommit: %w", err)
		}

		tree, err := commit.Tree()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("commit.Tree: %w", err)
		}

		ig, err := newIgnoreMatcher(tree)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("newIgnoreMatcher: %w", err)
		}

		files, subVersions, err := TreeToFiles(repository, tree, options.BuildOptions.RepositoryDescription.URL, repoCache)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("TreeToFiles: %w", err)
		}
		for k, v := range files {
			if ig.Match(k.Path) {
				continue
			}
			repos[k] = v
			branchMap[k] = append(branchMap[k], b)
		}

		branchVersions[b] = subVersions
	}

	return repos, branchMap, branchVersions, nil
}

func createDocument(key fileKey,
	repos map[fileKey]BlobLocation,
	branchMap map[fileKey][]string,
	ranks repoPathRanks,
	opts build.Options,
) (zoekt.Document, error) {
	blob, err := repos[key].Repo.BlobObject(key.ID)
	if err != nil {
		return zoekt.Document{}, err
	}

	keyFullPath := key.FullPath()
	if blob.Size > int64(opts.SizeMax) && !opts.IgnoreSizeMax(keyFullPath) {
		return zoekt.Document{
			SkipReason:        fmt.Sprintf("file size %d exceeds maximum size %d", blob.Size, opts.SizeMax),
			Name:              key.FullPath(),
			Branches:          branchMap[key],
			SubRepositoryPath: key.SubRepoPath,
		}, nil
	}

	contents, err := blobContents(blob)
	if err != nil {
		return zoekt.Document{}, err
	}

	var pathRanks []float64
	if len(ranks.Paths) > 0 {
		// If the repository has ranking data, then store the file's rank.
		pathRank := ranks.rank(keyFullPath)
		pathRanks = []float64{pathRank}
	}

	return zoekt.Document{
		SubRepositoryPath: key.SubRepoPath,
		Name:              keyFullPath,
		Content:           contents,
		Branches:          branchMap[key],
		Ranks:             pathRanks,
	}, nil
}

func blobContents(blob *object.Blob) ([]byte, error) {
	r, err := blob.Reader()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var buf bytes.Buffer
	buf.Grow(int(blob.Size))
	_, err = buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func uniq(ss []string) []string {
	result := ss[:0]
	var last string
	for i, s := range ss {
		if i == 0 || s != last {
			result = append(result, s)
		}
		last = s
	}
	return result
}
