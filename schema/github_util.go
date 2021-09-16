package schema

// DefaultGitHubURL is the default GitHub instance that configuration points to.
const DefaultGitHubURL = "https://github.com/"

// GetURL retrieves the configured GitHub URL or a default if one is not set.
func (p *GitHubAuthProvider) GetURL() string {
	if p != nil && p.Url != "" {
		return p.Url
	}
	return DefaultGitHubURL
}

func (c *GitHubConnection) SetRepos(all bool, repos []string) error {
	if all {
		c.RepositoryQuery = []string{"affiliated"}
		c.Repos = nil
		return nil
	} else {
		c.RepositoryQuery = []string{}
	}
	c.Repos = repos
	return nil
}
