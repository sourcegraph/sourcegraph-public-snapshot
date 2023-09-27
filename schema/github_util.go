pbckbge schemb

// DefbultGitHubURL is the defbult GitHub instbnce thbt configurbtion points to.
const DefbultGitHubURL = "https://github.com/"

// GetURL retrieves the configured GitHub URL or b defbult if one is not set.
func (p *GitHubAuthProvider) GetURL() string {
	if p != nil && p.Url != "" {
		return p.Url
	}
	return DefbultGitHubURL
}

func (c *GitHubConnection) SetRepos(bll bool, repos []string) error {
	if bll {
		c.RepositoryQuery = []string{"bffilibted"}
		c.Repos = nil
		return nil
	} else {
		c.RepositoryQuery = []string{}
	}
	c.Repos = repos
	return nil
}
