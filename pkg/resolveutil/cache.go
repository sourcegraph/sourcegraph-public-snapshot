package resolveutil

import (
	"encoding/json"
	"errors"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/groupcache/lru"
	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/cache"
)

var importPathMappingCache = cache.TTL(cache.Sync(lru.New(200000)), time.Hour)
var importPathMappingCacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "repos",
	Name:      "import_path_cache",
	Help:      "Counts cache hits and misses for custom import path mapping of Go packages.",
}, []string{"type"})

func init() {
	prometheus.MustRegister(importPathMappingCacheCounter)
}

// ResolveInfo holds information about a resolved Go package
// import path.
//
// See docstring on ResolveCustomImportPath for details about
// each field of this struct.
type ResolveInfo struct {
	CanonicalImportPath string
	RepoURI             string
	CloneURL            string
}

var resolveImportPath = ResolveImportPath

// ResolveCustomImportPath resolves a custom Go package import path to its
// canonical import path, repo URI and clone URL.
//
// We define the "canonical import path" of a package as "${repoURI}/path/to/pkg/dir",
// where repoURI is the base of the clone URL of the repo containing that package.
//
// Why do we resolve to "canonical import path"?
// Each Go package has 3 identifiers: import path, clone URL, and source link.
// The import path and clone URL are specified in the 'go-import' meta tag, while
// the source link is specified in the 'go-source' meta tag. For instance, consider
// the log15 package:
//
// <meta name="go-import" content="gopkg.in/inconshreveable/log15.v2 git https://gopkg.in/inconshreveable/log15.v2">
// <meta name="go-source" content="gopkg.in/inconshreveable/log15.v2 _ https://github.com/inconshreveable/log15/tree/v2.11{/dir} https://github.com/inconshreveable/log15/blob/v2.11{/dir}/{file}#L{line}">
//
// Its identifiers are:
//     import path: gopkg.in/inconshreveable/log15.v2
//     clone url:   https://gopkg.in/inconshreveable/log15.v2
//     source link: https://github.com/inconshreveable/log15/tree/v2.11{/dir}
//     repo URI:    gopkg.in/inconshreveable/log15.v2 [= cloneURL - scheme]
//
// The source link is not always specified and its spec is loosely enforced. So, within
// Sourcegraph we only rely on import path and clone URL.
//
// Additionally, srclib only gets one piece of information when indexing a repo, and that is
// its repoURI, which is the clone URL without the scheme. srclib constructs a clone URL
// from the repoURI to clone the repo.
//
// Subsequently, srclib-go also gets only the repoURI, and cannot obtain the import path
// of a package while graphing it. So, srclib-go assumes that a def's enclosing package is
// at the import path "${repoURI}/path/to/pkg/dir". We call this the "CanonicalImportPath"
// and it will differ from the actual import path if the repo has a custom import path.
// So, in Sourcegraph and srclib, we work with only the RepoURI and CanonicalImportPath
// of a Go package.
func ResolveCustomImportPath(importPath string) (info *ResolveInfo, err error) {
	if importPath == "" {
		err = errors.New("got empty import path")
		return
	}
	if infoJSON, found := importPathMappingCache.Get(importPath); found {
		importPathMappingCacheCounter.WithLabelValues("hit").Inc()
		if g, ok := infoJSON.([]byte); ok {
			info = &ResolveInfo{}
			err = json.Unmarshal(g, info)
			return
		}
	}

	target, err := resolveImportPath(importPath)
	if err != nil {
		importPathMappingCacheCounter.WithLabelValues("resolve-error").Inc()
		return nil, err
	}

	u, err := url.Parse(target.ToRepoCloneURL)
	if err != nil {
		importPathMappingCacheCounter.WithLabelValues("url-error").Inc()
		return nil, err
	}
	u.Path = strings.TrimSuffix(u.Path, ".git")

	info = &ResolveInfo{
		CanonicalImportPath: target.ToUnit,
		RepoURI:             filepath.Join(u.Host, u.Path),
		CloneURL:            target.ToRepoCloneURL,
	}

	infoJSON, err := json.Marshal(info)
	if err != nil {
		importPathMappingCacheCounter.WithLabelValues("marshal-error").Inc()
		return nil, err
	}
	importPathMappingCache.Add(importPath, infoJSON)
	importPathMappingCacheCounter.WithLabelValues("miss").Inc()
	return
}
