package github

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	"github.com/inconshreveable/log15"
)

// allMatchingSemver is a *semver.Version that will always match for the latest GitHub, which is either the
// latest GHE or the current deployment on GitHub.com.
var allMatchingSemver = semver.MustParse("99.99.99")

// versionCacheResetTime stores the time until a version cache is reset. It's set to 6 hours.
const versionCacheResetTime = 6 * 60 * time.Minute

type versionCache struct {
	mu        sync.RWMutex
	versions  map[string]*semver.Version
	lastReset time.Time
}

var globalVersionCache *versionCache = &versionCache{
	versions: make(map[string]*semver.Version),
}

// normalizeURL will attempt to normalise rawURL.
// If there is an error parsing it, we'll just return rawURL lower cased.
func normalizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return strings.ToLower(rawURL)
	}
	parsed.Host = strings.ToLower(parsed.Host)
	if !strings.HasSuffix(parsed.Path, "/") {
		parsed.Path += "/"
	}
	return parsed.String()
}

// determineGitHubVersion returns a *semver.Version for the targetted GitHub instance by this client. When an
// error occurs, we print a warning to the logs but don't fail and return the allMatchingSemver.
func (c *V4Client) determineGitHubVersion(ctx context.Context) *semver.Version {
	url := normalizeURL(c.apiURL.String())

	globalVersionCache.mu.RLock()
	if globalVersionCache.lastReset.IsZero() || time.Now().After(globalVersionCache.lastReset.Add(versionCacheResetTime)) {
		globalVersionCache.mu.RUnlock()
		globalVersionCache.mu.Lock()

		// Clear cache and set last expiry to now.
		globalVersionCache.lastReset = time.Now()
		globalVersionCache.versions = make(map[string]*semver.Version)

		globalVersionCache.mu.Unlock()
		globalVersionCache.mu.RLock()
	}

	if version, ok := globalVersionCache.versions[url]; ok {
		globalVersionCache.mu.RUnlock()
		return version
	}

	globalVersionCache.mu.RUnlock()
	globalVersionCache.mu.Lock()
	defer globalVersionCache.mu.Unlock()
	version := c.fetchGitHubVersion(ctx)
	globalVersionCache.versions[url] = version
	return version
}

func (c *V4Client) fetchGitHubVersion(ctx context.Context) *semver.Version {
	if c.githubDotCom {
		return allMatchingSemver
	}

	var resp struct {
		InstalledVersion string `json:"installed_version"`
	}
	req, err := http.NewRequest("GET", "/meta", nil)
	if err != nil {
		log15.Warn("Failed to fetch GitHub enterprise version", "build request", "apiURL", c.apiURL, "err", err)
		return allMatchingSemver
	}
	if err = doRequest(ctx, c.apiURL, c.auth, c.rateLimitMonitor, c.httpClient, req, &resp); err != nil {
		log15.Warn("Failed to fetch GitHub enterprise version", "doRequest", "apiURL", c.apiURL, "err", err)
		return allMatchingSemver
	}
	version, err := semver.NewVersion(resp.InstalledVersion)
	if err == nil {
		return version
	}
	log15.Warn("Failed to fetch GitHub enterprise version", "parse version", "apiURL", c.apiURL, "resp.InstalledVersion", resp.InstalledVersion, "err", err)
	return allMatchingSemver
}
