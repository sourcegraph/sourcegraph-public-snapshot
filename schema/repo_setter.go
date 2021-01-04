package schema

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

func (c *GitLabConnection) SetRepos(all bool, repos []string) error {
	if all {
		c.ProjectQuery = []string{"projects?membership=true&archived=no"}
		c.Projects = nil
		return nil
	} else {
		c.ProjectQuery = []string{"none"}
	}
	c.Projects = []*GitLabProject{}
	for _, repo := range repos {
		c.Projects = append(c.Projects, &GitLabProject{
			Name: repo,
		})
	}
	return nil
}
