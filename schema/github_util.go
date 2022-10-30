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
