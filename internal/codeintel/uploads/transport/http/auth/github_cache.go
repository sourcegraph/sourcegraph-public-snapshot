package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

type GitHubAuthCache struct {
	cache *rcache.RedisCache
}

var githubAuthCache = &GitHubAuthCache{
	cache: rcache.NewWithTTL(redispool.Cache, "codeintel.github-authz:", 60 /* seconds */),
}

func (c *GitHubAuthCache) Get(ctx context.Context, key string) (authorized bool, _ bool) {
	b, ok := c.cache.Get(ctx, key)
	if !ok {
		return false, false
	}

	err := json.Unmarshal(b, &authorized)
	return authorized, err == nil
}

func (c *GitHubAuthCache) Set(ctx context.Context, key string, authorized bool) {
	b, _ := json.Marshal(authorized)
	c.cache.Set(ctx, key, b)
}

func makeGitHubAuthCacheKey(githubToken, repoName string) string {
	key := sha256.Sum256([]byte(githubToken + ":" + repoName))
	return hex.EncodeToString(key[:])
}
