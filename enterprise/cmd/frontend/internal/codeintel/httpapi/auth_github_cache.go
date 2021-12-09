package httpapi

import "fmt"

type GitHubAuthCache struct {
	cache map[string]bool
}

var githubAuthCache = &GitHubAuthCache{
	// TODO - replace with redis
	cache: map[string]bool{},
}

func (c *GitHubAuthCache) Get(key string) (bool, bool, error) {
	authorized, ok := c.cache[key]
	return authorized, ok, nil
}

func (c *GitHubAuthCache) Set(key string, authorized bool) error {
	c.cache[key] = authorized
	return nil
}

func makeGitHubAuthCacheKey(githubToken, repoName string) string {
	// TODO - hash for privacy
	return fmt.Sprintf("%s:%s", githubToken, repoName)
}
