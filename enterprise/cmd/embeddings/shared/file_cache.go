package shared

import (
	"context"
	"fmt"

	lru "github.com/hashicorp/golang-lru"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func readCachedFile(ctx context.Context, cache *lru.Cache, gitserverClient gitserver.Client, repoName api.RepoName, revision api.CommitID, fileName string) ([]byte, error) {
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
}
