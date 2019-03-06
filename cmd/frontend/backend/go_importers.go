package backend

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/golang/groupcache/lru"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"golang.org/x/net/context/ctxhttp"
)

var MockCountGoImporters func(ctx context.Context, repo api.RepoName) (int, error)

var (
	goImportersCountCacheMu sync.Mutex
	goImportersCountCache   = lru.New(1000) // 1000 is arbitrarily chosen

	countGoImportersHTTPClient *http.Client // mockable in tests
)

// CountGoImporters returns the number of Go importers for the repository's Go subpackages. This is
// a special case used only on Sourcegraph.com for repository badges.
//
// TODO: The import path is not always the same as the repository name.
func CountGoImporters(ctx context.Context, repo api.RepoName) (count int, err error) {
	if MockCountGoImporters != nil {
		return MockCountGoImporters(ctx, repo)
	}

	if !envvar.SourcegraphDotComMode() {
		// Avoid confusing users by exposing this on self-hosted instances, because it relies on the
		// public godoc.org API.
		return 0, errors.New("counting Go importers is not supported on self-hosted instances")
	}

	goImportersCountCacheMu.Lock()
	v, ok := goImportersCountCache.Get(repo)
	goImportersCountCacheMu.Unlock()
	if ok {
		return v.(int), nil // cache hit
	}

	defer func() {
		if err == nil {
			// Store in cache.
			goImportersCountCacheMu.Lock()
			defer goImportersCountCacheMu.Unlock()
			goImportersCountCache.Add(repo, count)
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second) // avoid tying up resources unduly
	defer cancel()

	// Find all (possible) Go packages in the repository.
	goPackages, err := listGoPackagesInRepoImprecise(ctx, repo)
	if err != nil {
		return 0, err
	}
	const maxSubpackages = 50 // arbitrary limit to avoid overloading api.godoc.org
	if len(goPackages) > maxSubpackages {
		goPackages = goPackages[:maxSubpackages]
	}

	// Count importers for each of the repository's Go packages.
	for _, pkg := range goPackages {
		// Assumes the import path is the same as the repo name - not always true!
		response, err := ctxhttp.Get(ctx, countGoImportersHTTPClient, "https://api.godoc.org/importers/"+string(pkg))
		if err != nil {
			return 0, err
		}
		var result struct {
			Results []struct {
				Path string
			}
		}
		bytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return 0, err
		}
		err = json.Unmarshal(bytes, &result)
		if err != nil {
			return 0, err
		}
		count += len(result.Results)
	}

	return count, nil
}

// listGoPackagesInRepo returns a list of import paths for all (probable) Go packages in the
// repository. It computes the list based solely on the repository name (as a prefix) and filenames
// in the repository; it does not parse or build the Go files to determine the list precisely.
func listGoPackagesInRepoImprecise(ctx context.Context, repoName api.RepoName) ([]string, error) {
	if !envvar.SourcegraphDotComMode() {
		// 🚨 SECURITY: Avoid leaking information about private repositories that the viewer is not
		// allowed to access.
		return nil, errors.New("listGoPackagesInRepo is only supported on Sourcegraph.com for public repositories")
	}

	repo, err := Repos.GetByName(ctx, repoName)
	if err != nil {
		return nil, err
	}
	gitRepo, err := CachedGitRepo(ctx, repo)
	if err != nil {
		return nil, err
	}
	commitID, err := git.ResolveRevision(ctx, *gitRepo, nil, "HEAD", nil)
	if err != nil {
		return nil, err
	}
	fis, err := git.ReadDir(ctx, *gitRepo, commitID, "", true)
	if err != nil {
		return nil, err
	}

	subpaths := map[string]struct{}{} // all non-vendor/internal/hidden dirs containing *.go files
	for _, fi := range fis {
		if name := fi.Name(); filepath.Ext(name) == ".go" {
			dir := path.Dir(name)
			if isPossibleExternallyImportableGoPackageDir(dir) {
				subpaths[dir] = struct{}{}
			}
		}
	}

	importPaths := make([]string, 0, len(subpaths))
	for subpath := range subpaths {
		importPaths = append(importPaths, string(repo.Name)+"/"+subpath)
	}
	sort.Strings(importPaths)
	return importPaths, nil
}

func isPossibleExternallyImportableGoPackageDir(dirPath string) bool {
	components := strings.Split(dirPath, "/")
	for _, c := range components {
		if (strings.HasPrefix(c, ".") && len(c) > 1) || strings.HasPrefix(c, "_") || c == "vendor" || c == "internal" || c == "testdata" {
			return false
		}
	}
	return true
}
