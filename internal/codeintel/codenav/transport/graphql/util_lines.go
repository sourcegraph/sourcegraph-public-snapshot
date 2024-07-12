package graphql

import (
	"context"
	"io"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/byteutils"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type LinesGetter interface {
	Get(ctx context.Context, repo api.RepoName, commit api.CommitID, path string, startLine, endLine int) ([]byte, error)
}

type cacheKey struct {
	repo     api.RepoName
	revision api.CommitID
	path     string
}

type cacheValue struct {
	contents []byte
	index    byteutils.LineIndex
}

type cachedLinesGetter struct {
	mu             sync.RWMutex
	cache          map[cacheKey]cacheValue
	maxCachedBytes int
	freeBytes      int
	gitserver      gitserver.Client
}

var _ LinesGetter = (*cachedLinesGetter)(nil)

func newCachedLinesGetter(gitserver gitserver.Client, size int) *cachedLinesGetter {
	return &cachedLinesGetter{
		cache:          make(map[cacheKey]cacheValue),
		maxCachedBytes: size,
		freeBytes:      size,
		gitserver:      gitserver,
	}
}

func (c *cachedLinesGetter) Get(ctx context.Context, repo api.RepoName, commit api.CommitID, path string, startLine, endLine int) ([]byte, error) {
	key := cacheKey{repo, commit, path}

	c.mu.RLock()
	if value, ok := c.cache[key]; ok {
		c.mu.RUnlock()
		start, end := value.index.LinesRange(startLine, endLine)
		return value.contents[start:end], nil
	}
	c.mu.RUnlock()

	r, err := c.gitserver.NewFileReader(ctx, repo, commit, path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	contents, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	index := byteutils.NewLineIndex(contents)
	start, end := index.LinesRange(startLine, endLine)
	lines := contents[start:end]

	if len(contents) > c.maxCachedBytes {
		// Don't both trying to fit it in the cache
		return lines, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Make room for the file in the cache. This cache doesn't need to be high
	// performance -- just randomly delete things until we have room.
	for k, v := range c.cache {
		if c.freeBytes >= len(contents) {
			break
		}
		delete(c.cache, k)
		c.freeBytes += len(v.contents)
	}

	c.cache[key] = cacheValue{contents, index}
	c.freeBytes -= len(contents)

	return lines, nil
}
