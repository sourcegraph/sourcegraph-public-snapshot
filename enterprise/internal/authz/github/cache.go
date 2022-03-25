package github

import (
	"encoding/json"

	"github.com/gregjones/httpcache"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

const cacheVersion = "v1"

type cachedGroup struct {
	// Org login, required
	Org string
	// Team slug, if empty implies group is an org
	Team string

	// Repositories entities associated with this group has access to.
	//
	// This should ONLY be populated on a USER-centric sync, but may be appended to if
	// already populated.
	//
	// If nil, a repo-centric sync should treat this cache as unpopulated and fill in this
	// value.
	Repositories []extsvc.RepoID
	// Users associated with this group
	//
	// This should ONLY be populated on a REPO-centric sync, but maybe to appended to if
	// already populated.
	//
	// If nil, a user-centric sync should treat this cache as unpopulated and fill in this
	// value.
	Users []extsvc.AccountID
}

func (g *cachedGroup) key() string {
	key := cacheVersion + "/" + g.Org
	if g.Team != "" {
		key += "/" + g.Team
	}
	return key
}

type cachedGroups struct {
	cache httpcache.Cache
}

// setGroup stores the given group in the cache.
func (c *cachedGroups) setGroup(group cachedGroup) error {
	bytes, err := json.Marshal(&group)
	if err != nil {
		return err
	}
	c.cache.Set(group.key(), bytes)
	return nil
}

// getGroup attempts to retrive the given org, team group from cache.
//
// It always returns a valid cachedGroup even if it fails to retrieve a group from cache.
func (c *cachedGroups) getGroup(org string, team string) (cachedGroup, bool) {
	rawGroup := cachedGroup{Org: org, Team: team}
	bytes, ok := c.cache.Get(rawGroup.key())
	if !ok {
		return rawGroup, ok
	}
	var cached cachedGroup
	if err := json.Unmarshal(bytes, &cached); err != nil {
		return rawGroup, false
	}
	return cached, ok
}

// invalidateGroup deletes the given group from the cache and invalidates the cached values
// within the given group.
func (c *cachedGroups) invalidateGroup(group *cachedGroup) {
	c.cache.Delete(group.key())
	group.Repositories = nil
	group.Users = nil
}
