package conf

import (
	"net/url"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/atomicvalue"
)

// EmailVerificationRequired returns whether users must verify an email address before they
// can perform most actions on this site.
//
// It's false for sites that do not have an email sending API key set up.
func EmailVerificationRequired() bool {
	return GetTODO().EmailSmtp != nil
}

// CanSendEmail returns whether the site can send emails (e.g., to reset a password or
// invite a user to an org).
//
// It's false for sites that do not have an email sending API key set up.
func CanSendEmail() bool {
	return GetTODO().EmailSmtp != nil
}

// HasGitHubDotComToken reports whether there are any personal access tokens configured for
// github.com.
func HasGitHubDotComToken() bool {
	for _, c := range GetTODO().Github {
		u, err := url.Parse(c.Url)
		if err != nil {
			continue
		}
		hostname := strings.ToLower(u.Hostname())
		if (hostname == "github.com" || hostname == "api.github.com") && c.Token != "" {
			return true
		}
	}
	return false
}

// HasGitLabDotComToken reports whether there are any personal access tokens configured for
// github.com.
func HasGitLabDotComToken() bool {
	for _, c := range GetTODO().Gitlab {
		u, err := url.Parse(c.Url)
		if err != nil {
			continue
		}
		hostname := strings.ToLower(u.Hostname())
		if hostname == "gitlab.com" && c.Token != "" {
			return true
		}
	}
	return false
}

// GitHubEnterpriseURLs is a map of GitHub Enterprise hosts to their full URLs.
// This can be used for the purposes of generating external GitHub enterprise links.
func GitHubEnterpriseURLs() map[string]string {
	return gitHubEnterpriseURLs.Get().(map[string]string)
}

var gitHubEnterpriseURLs = atomicvalue.New()

func addWatchers() {
	Watch(func() {
		gitHubEnterpriseURLs.Set(func() interface{} {
			urls := make(map[string]string)
			for _, c := range Get().Github {
				gheURL, err := url.Parse(c.Url)
				if err != nil {
					log15.Error("error parsing GitHub config", "error", err)
				}
				urls[gheURL.Host] = strings.TrimSuffix(c.Url, "/")
			}
			return urls
		})
	})
}
