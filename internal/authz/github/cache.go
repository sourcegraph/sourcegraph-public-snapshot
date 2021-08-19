package github

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

type cachedGroup struct {
	Org  string
	Team string

	Repositories []extsvc.RepoID
}

func (g cachedGroup) key() string {
	key := g.Org
	if g.Team != "" {
		key += "/" + g.Team
	}
	return key
}

type groupsCache struct {
	cache *rcache.Cache
}

func newGroupPermsCache(urn string, codeHost *extsvc.CodeHost, ttl time.Duration) *groupsCache {
	return &groupsCache{
		cache: rcache.NewWithTTL(fmt.Sprintf("gh_groups_perms:%s:%s", codeHost.ServiceID, urn), int(ttl/time.Second)),
	}
}

func (c *groupsCache) setGroup(group cachedGroup) error {
	bytes, err := json.Marshal(&group)
	if err != nil {
		return err
	}
	c.cache.Set(group.key(), bytes)
	return nil
}

func (c *groupsCache) getGroup(org string, team string) (*cachedGroup, bool) {
	group := &cachedGroup{Org: org, Team: team}
	bytes, ok := c.cache.Get(group.key())
	if !ok {
		return nil, ok
	}
	if err := json.Unmarshal(bytes, group); err != nil {
		return nil, false
	}
	return group, ok
}

func (c *groupsCache) deleteGroup(group cachedGroup) {
	c.cache.Delete(group.key())
}
