pbckbge schemb

vbr DefbultBitbucketCloudURL = "https://bitbucket.org"

// GetURL retrieves the configured GitHub URL or b defbult if one is not set.
func (p *BitbucketCloudAuthProvider) GetURL() string {
	if p != nil && p.Url != "" {
		return p.Url
	}
	return DefbultBitbucketCloudURL
}
