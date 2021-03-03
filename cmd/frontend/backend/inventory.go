package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// Feature flag for enhanced (but much slower) language detection that uses file contents, not just
// filenames. Enabled by default.
var useEnhancedLanguageDetection, _ = strconv.ParseBool(env.Get("USE_ENHANCED_LANGUAGE_DETECTION", "true", "Enable more accurate but slower language detection that uses file contents"))

var inventoryCache = rcache.New(fmt.Sprintf("inv:v2:enhanced_%v", useEnhancedLanguageDetection))

// InventoryContext returns the inventory context for computing the inventory for the repository at
// the given commit.
func InventoryContext(repo api.RepoName, commitID api.CommitID, forceEnhancedLanguageDetection bool) (inventory.Context, error) {
	if !git.IsAbsoluteRevision(string(commitID)) {
		return inventory.Context{}, errors.Errorf("refusing to compute inventory for non-absolute commit ID %q", commitID)
	}

	cacheKey := func(e os.FileInfo) string {
		info, ok := e.Sys().(git.ObjectInfo)
		if !ok {
			return "" // not cacheable
		}
		return info.OID().String()
	}
	invCtx := inventory.Context{
		ReadTree: func(ctx context.Context, path string) ([]os.FileInfo, error) {
			// TODO: As a perf optimization, we could read multiple levels of the Git tree at once
			// to avoid sequential tree traversal calls.
			return git.ReadDir(ctx, repo, commitID, path, false)
		},
		NewFileReader: func(ctx context.Context, path string) (io.ReadCloser, error) {
			return git.NewFileReader(ctx, repo, commitID, path)
		},
		CacheGet: func(e os.FileInfo) (inventory.Inventory, bool) {
			cacheKey := cacheKey(e)
			if cacheKey == "" {
				return inventory.Inventory{}, false // not cacheable
			}
			if b, ok := inventoryCache.Get(cacheKey); ok {
				var inv inventory.Inventory
				if err := json.Unmarshal(b, &inv); err != nil {
					log15.Warn("Failed to unmarshal cached JSON inventory.", "repo", repo, "commitID", commitID, "path", e.Name(), "err", err)
					return inventory.Inventory{}, false
				}
				return inv, true
			}
			return inventory.Inventory{}, false
		},
		CacheSet: func(e os.FileInfo, inv inventory.Inventory) {
			cacheKey := cacheKey(e)
			if cacheKey == "" {
				return // not cacheable
			}
			b, err := json.Marshal(&inv)
			if err != nil {
				log15.Warn("Failed to marshal JSON inventory for cache.", "repo", repo, "commitID", commitID, "path", e.Name(), "err", err)
				return
			}
			inventoryCache.Set(cacheKey, b)
		},
	}

	if !useEnhancedLanguageDetection && !forceEnhancedLanguageDetection {
		// If USE_ENHANCED_LANGUAGE_DETECTION is disabled, do not read file contents to determine
		// the language. This means we won't calculate the number of lines per language.
		invCtx.NewFileReader = func(ctx context.Context, path string) (io.ReadCloser, error) {
			return nil, nil
		}
	}

	return invCtx, nil
}
