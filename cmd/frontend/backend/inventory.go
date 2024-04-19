package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"golang.org/x/sync/semaphore"
	"io"
	"io/fs"
	"strconv"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Feature flag for enhanced (but much slower) language detection that uses file contents, not just
// filenames. Enabled by default.
var useEnhancedLanguageDetection, _ = strconv.ParseBool(env.Get("USE_ENHANCED_LANGUAGE_DETECTION", "true", "Enable more accurate but slower language detection that uses file contents"))

var inventoryCache = rcache.New(fmt.Sprintf("inv:v2:enhanced_%v", useEnhancedLanguageDetection))

type semaphoredReadCloser struct {
	readCloser io.ReadCloser
	semaphore  *semaphore.Weighted
	ctx        context.Context
}

func (s *semaphoredReadCloser) Read(p []byte) (int, error) {
	trc, ctx := trace.New(s.ctx, "Wait for semaphore")
	err := s.semaphore.Acquire(ctx, 1)
	trc.End()
	if err != nil {
		return 0, err
	}
	defer s.semaphore.Release(1)
	return s.readCloser.Read(p)
}

func (s *semaphoredReadCloser) Close() error {
	return s.readCloser.Close()
}

// InventoryContext returns the inventory context for computing the inventory for the repository at
// the given commit.
func InventoryContext(logger log.Logger, repo api.RepoName, gsClient gitserver.Client, commitID api.CommitID, forceEnhancedLanguageDetection bool) (inventory.Context, error) {
	if !gitdomain.IsAbsoluteRevision(string(commitID)) {
		return inventory.Context{}, errors.Errorf("refusing to compute inventory for non-absolute commit ID %q", commitID)
	}

	cacheKey := func(e fs.FileInfo) string {
		info, ok := e.Sys().(gitdomain.ObjectInfo)
		if !ok {
			return "" // not cacheable
		}
		return info.OID().String()
	}

	// todo bahrmichael: explain reasoning for picked values, and make them configurable
	gitServerSemaphore := semaphore.NewWeighted(5)
	cacheSemaphore := semaphore.NewWeighted(10)

	logger = logger.Scoped("InventoryContext").
		With(log.String("repo", string(repo)), log.String("commitID", string(commitID)))
	invCtx := inventory.Context{
		ReadTree: func(ctx context.Context, path string) ([]fs.FileInfo, error) {
			if err := gitServerSemaphore.Acquire(ctx, 1); err != nil {
				return nil, err
			}
			defer gitServerSemaphore.Release(1)
			// TODO: As a perf optimization, we could read multiple levels of the Git tree at once
			// to avoid sequential tree traversal calls.
			return gsClient.ReadDir(ctx, repo, commitID, path, false)
		},
		NewFileReader: func(ctx context.Context, path string) (io.ReadCloser, error) {
			reader, err := gsClient.NewFileReader(ctx, repo, commitID, path)
			if err != nil {
				return nil, err
			}
			return &semaphoredReadCloser{readCloser: reader, ctx: ctx, semaphore: gitServerSemaphore}, nil
		},
		CacheGet: func(ctx context.Context, e fs.FileInfo) (inventory.Inventory, bool) {
			cacheKey := cacheKey(e)
			if cacheKey == "" {
				return inventory.Inventory{}, false // not cacheable
			}

			if err := cacheSemaphore.Acquire(ctx, 1); err != nil {
				return inventory.Inventory{}, false
			}
			defer cacheSemaphore.Release(1)

			if b, ok := inventoryCache.Get(cacheKey); ok {
				var inv inventory.Inventory
				if err := json.Unmarshal(b, &inv); err != nil {
					logger.Warn("Failed to unmarshal cached JSON inventory.", log.String("path", e.Name()), log.Error(err))
					return inventory.Inventory{}, false
				}
				return inv, true
			}
			return inventory.Inventory{}, false
		},
		CacheSet: func(ctx context.Context, e fs.FileInfo, inv inventory.Inventory) {
			cacheKey := cacheKey(e)
			if cacheKey == "" {
				return // not cacheable
			}
			b, err := json.Marshal(&inv)
			if err != nil {
				logger.Warn("Failed to marshal JSON inventory for cache.", log.String("path", e.Name()), log.Error(err))
				return
			}

			if err := cacheSemaphore.Acquire(ctx, 1); err != nil {
				return
			}
			defer cacheSemaphore.Release(1)
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
