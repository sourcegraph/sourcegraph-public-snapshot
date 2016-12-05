package resolveutil

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/groupcache/lru"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/cache"
)

func resetCache() {
	importPathMappingCache = cache.TTL(cache.Sync(lru.New(500)), time.Hour)
}

// TestResolveCustomImportPath tests the behavior of ResolveCustomImportPath
// when called on some common Go package import paths.
func TestResolveCustomImportPath(t *testing.T) {
	resetCache()

	targets := map[string]*ResolvedTarget{
		"golang.org/x/net":         &ResolvedTarget{ToRepoCloneURL: "https://github.com/golang/net", ToUnit: "github.com/golang/net"},
		"golang.org/x/net/context": &ResolvedTarget{ToRepoCloneURL: "https://github.com/golang/net", ToUnit: "github.com/golang/net/context"},

		"k8s.io/kubernetes":         &ResolvedTarget{ToRepoCloneURL: "https://github.com/kubernetes/kubernetes", ToUnit: "github.com/kubernetes/kubernetes"},
		"k8s.io/kubernetes/pkg/api": &ResolvedTarget{ToRepoCloneURL: "https://github.com/kubernetes/kubernetes", ToUnit: "github.com/kubernetes/kubernetes/pkg/api"},

		"gopkg.in/inconshreveable/log15.v2": &ResolvedTarget{ToRepoCloneURL: "https://gopkg.in/inconshreveable/log15.v2", ToUnit: "gopkg.in/inconshreveable/log15.v2"},
		"azul3d.org/semver.v2":              &ResolvedTarget{ToRepoCloneURL: "https://azul3d.org/semver.v2", ToUnit: "azul3d.org/semver.v2"},

		"sourcegraph.com/sourcegraph/srclib/graph":    &ResolvedTarget{ToRepoCloneURL: "https://sourcegraph.com/sourcegraph/srclib.git", ToUnit: "sourcegraph.com/sourcegraph/srclib/graph"},
		"sourcegraph.com/sourcegraph/sourcegraph/app": &ResolvedTarget{ToRepoCloneURL: "https://sourcegraph.com/sourcegraph/sourcegraph.git", ToUnit: "sourcegraph.com/sourcegraph/sourcegraph/app"},
	}
	var calledAPI bool
	resolveImportPath = func(importPath string) (*ResolvedTarget, error) {
		calledAPI = true
		if repoRoot, ok := targets[importPath]; ok {
			return repoRoot, nil
		}
		return nil, fmt.Errorf("import path not found")
	}

	testCases := []struct {
		ImportPath string
		CalledAPI  bool
		Result     *ResolveInfo
	}{
		{"golang.org/x/net/context", true, &ResolveInfo{"github.com/golang/net/context", "github.com/golang/net", "https://github.com/golang/net"}},
		{"golang.org/x/net/context", false, &ResolveInfo{"github.com/golang/net/context", "github.com/golang/net", "https://github.com/golang/net"}},
		{"k8s.io/kubernetes/pkg/api", true, &ResolveInfo{"github.com/kubernetes/kubernetes/pkg/api", "github.com/kubernetes/kubernetes", "https://github.com/kubernetes/kubernetes"}},
		{"k8s.io/kubernetes/pkg/api", false, &ResolveInfo{"github.com/kubernetes/kubernetes/pkg/api", "github.com/kubernetes/kubernetes", "https://github.com/kubernetes/kubernetes"}},
		{"gopkg.in/inconshreveable/log15.v2", true, &ResolveInfo{"gopkg.in/inconshreveable/log15.v2", "gopkg.in/inconshreveable/log15.v2", "https://gopkg.in/inconshreveable/log15.v2"}},
		{"azul3d.org/semver.v2", true, &ResolveInfo{"azul3d.org/semver.v2", "azul3d.org/semver.v2", "https://azul3d.org/semver.v2"}},
		{"sourcegraph.com/sourcegraph/srclib/graph", true, &ResolveInfo{"sourcegraph.com/sourcegraph/srclib/graph", "sourcegraph.com/sourcegraph/srclib", "https://sourcegraph.com/sourcegraph/srclib.git"}},
		{"sourcegraph.com/sourcegraph/sourcegraph/app", true, &ResolveInfo{"sourcegraph.com/sourcegraph/sourcegraph/app", "sourcegraph.com/sourcegraph/sourcegraph", "https://sourcegraph.com/sourcegraph/sourcegraph.git"}},
		{"sourcegraph.com/sourcegraph/srclib/graph", false, &ResolveInfo{"sourcegraph.com/sourcegraph/srclib/graph", "sourcegraph.com/sourcegraph/srclib", "https://sourcegraph.com/sourcegraph/srclib.git"}},
	}
	for _, test := range testCases {
		calledAPI = false
		got, err := ResolveCustomImportPath(test.ImportPath)
		if err != nil {
			t.Fatal(err)
		}
		if test.CalledAPI != calledAPI {
			t.Errorf("calledAPI: got %v, want %v", calledAPI, test.CalledAPI)
		}
		if got == nil {
			t.Errorf("got nil result from ResolveCustomImportPath")
			continue
		}
		if !reflect.DeepEqual(got, test.Result) {
			t.Errorf("got %+v, want %+v", got, test.Result)
		}
	}
}
