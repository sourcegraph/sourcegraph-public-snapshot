package shared

import (
	"context"
	"fmt"

	lru "github.com/hashicorp/golang-lru"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const FILE_CACHE_MAX_ENTRIES = 128

func readCachedFileFn(gitserverClient gitserver.Client) (readFileFn, error) {
	cache, err := lru.New(FILE_CACHE_MAX_ENTRIES)
	if err != nil {
		return nil, errors.Wrap(err, "creating file cache")
	}

	return func(ctx context.Context, repoName api.RepoName, revision api.CommitID, fileName string) ([]byte, error) {
		cacheKey := fmt.Sprintf("%s:%s:%s", repoName, revision, fileName)
		var content []byte
		var err error
		if cachedContent, ok := cache.Get(cacheKey); ok {
			content = cachedContent.([]byte)
		} else {
			content, err = gitserverClient.ReadFile(ctx, authz.DefaultSubRepoPermsChecker, repoName, revision, fileName)
			if err != nil {
				return nil, err
			}
			cache.Add(cacheKey, content)
		}

		return content, err
	}, nil
}
