package schema

var DefaultBitbucketCloudURL = "https://bitbucket.org"

// GetURL retrieves the configured GitHub URL or a default if one is not set.
func (p *BitbucketCloudAuthProvider) GetURL() string {
	if p != nil && p.Url != "" {
		return p.Url
	}
	return DefaultBitbucketCloudURL
}
