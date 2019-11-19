package gitlab

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

type cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, b []byte)
	Delete(key string)
}

// userProjCacheKey returns the key for caching the value describing if the given GitLab user has
// access to the given GitLab project. This key must be unique among *all* OAuthAuthzProvider
// cache keys, including those for repo visibility (see projVisibilityCacheKey).
func userProjCacheKey(gitlabAccountID string, gitlabProjID int) string {
	return fmt.Sprintf("userRepo:%s:%d", gitlabAccountID, gitlabProjID) // GitLab account IDs cannot have ':'
}

type userRepoCacheVal struct {
	// Read is whether or not the repository can be read by the user specified in the key
	Read bool

	TTL time.Duration
}

func cacheGetUserRepo(c cache, gitlabAccountID string, gitlabProjID int, ttl time.Duration) (v userRepoCacheVal, exists bool) {
	k := userProjCacheKey(gitlabAccountID, gitlabProjID)
	b, exists := c.Get(k)
	if !exists {
		return userRepoCacheVal{}, false
	}
	err := json.Unmarshal(b, &v)
	if err != nil {
		c.Delete(k)
		return userRepoCacheVal{}, false
	}
	if v.TTL != ttl {
		c.Delete(k)
		return userRepoCacheVal{}, false
	}
	return v, true
}

func cacheSetUserRepo(c cache, gitlabAccountID string, gitlabProjID int, v userRepoCacheVal) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.Set(userProjCacheKey(gitlabAccountID, gitlabProjID), b)
	return nil
}

// projVisibilityCacheKey returns the key for caching the value describing the visibility of the
// GitLab project (public, internal, private). This key must be unique among *all*
// OAuthAuthzProvider cache keys, including those for user-project access (see
// userProjCacheKey).
func projVisibilityCacheKey(gitlabProjID int) string {
	return fmt.Sprintf("visibility:%d", gitlabProjID)
}

type repoVisibilityCacheVal struct {
	Visibility gitlab.Visibility
	TTL        time.Duration
}

func cacheGetRepoVisibility(c cache, gitlabProjID int, ttl time.Duration) (v repoVisibilityCacheVal, exists bool) {
	k := projVisibilityCacheKey(gitlabProjID)
	b, exists := c.Get(k)
	if !exists {
		return repoVisibilityCacheVal{}, false
	}
	err := json.Unmarshal(b, &v)
	if err != nil {
		c.Delete(k)
		return repoVisibilityCacheVal{}, false
	}
	if v.TTL != ttl {
		c.Delete(k)
		return repoVisibilityCacheVal{}, false
	}
	return v, true
}

func cacheSetRepoVisibility(c cache, gitlabProjID int, v repoVisibilityCacheVal) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.Set(projVisibilityCacheKey(gitlabProjID), b)
	return nil
}
