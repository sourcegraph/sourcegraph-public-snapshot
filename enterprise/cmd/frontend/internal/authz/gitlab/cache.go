package gitlab

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
)

type cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, b []byte)
	Delete(key string)
}

func userRepoCacheKey(gitlabAccountID string, gitlabProjID int) string {
	return fmt.Sprintf("userRepo:%s:%d", gitlabAccountID, gitlabProjID) // GitLab account IDs cannot have ':'
}

type userRepoCacheVal struct {
	// Read is whether or not the repository can be read by the user specified in the key
	Read bool

	TTL time.Duration
}

func cacheGetUserRepo(c cache, gitlabAccountID string, gitlabProjID int, ttl time.Duration) (v userRepoCacheVal, exists bool) {
	k := userRepoCacheKey(gitlabAccountID, gitlabProjID)
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
	c.Set(userRepoCacheKey(gitlabAccountID, gitlabProjID), b)
	return nil
}

func repoVisibilityCacheKey(gitlabProjID int) string {
	return fmt.Sprintf("visibility:%d", gitlabProjID)
}

type repoVisibilityCacheVal struct {
	Visibility gitlab.Visibility
	TTL        time.Duration
}

func cacheGetRepoVisibility(c cache, gitlabProjID int, ttl time.Duration) (v repoVisibilityCacheVal, exists bool) {
	k := repoVisibilityCacheKey(gitlabProjID)
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
	c.Set(repoVisibilityCacheKey(gitlabProjID), b)
	return nil
}
