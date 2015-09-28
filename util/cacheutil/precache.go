package cacheutil

import (
	"log"
	"net/http"
	"path/filepath"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/sourcecode"
)

var Precache = true

// PrecacheTreeEntry pre-fetches the children of a directory for
// performance. This function should eventually be removed when we
// have VCS caching at a lower level.
func PrecacheTreeEntry(cl *sourcegraph.Client, ctx context.Context, entry *sourcegraph.TreeEntry, entrySpec sourcegraph.TreeEntrySpec) {
	if !Precache {
		return
	}

	for _, child := range entry.Entries {
		child := child

		childSpec := sourcegraph.TreeEntrySpec{
			RepoRev: entrySpec.RepoRev,
			Path:    filepath.Join(entrySpec.Path, child.Name),
		}
		opt := &sourcegraph.RepoTreeGetOptions{
			TokenizedSource: sourcecode.IsLikelyCodeFile(childSpec.Path),
		}

		go func() {
			log15.Info("prefetching tree entry", "path", childSpec.Path)
			_, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: childSpec, Opt: opt})
			if err != nil {
				log.Printf("Precaching failed for %s: %s", entrySpec.Path, err)
			}
			log15.Info("done prefetching tree entry", "path", childSpec.Path)
		}()
	}
}

var HTTPAddr string

// PrecacheRoot pre-fetches the root repository page for performance.
// This function should eventually be removed when we have VCS caching
// at a lower level.
func PrecacheRoot(repoURI string) {
	if !Precache {
		return
	}

	if HTTPAddr == "" {
		log.Printf("failing to precache directory root because HTTPAddr empty")
		return
	}
	log15.Debug("precaching directory root", "repo", repoURI, "http-addr", HTTPAddr)

	u := router.New(nil).URLTo(router.Repo, "Repo", repoURI)
	if _, err := http.DefaultClient.Get(HTTPAddr + u.String()); err != nil {
		log.Printf("failing to fetch and precache directory root: %s", err)
	}
}
