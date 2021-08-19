package extsvc

import (
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type CodeHost struct {
	ServiceID   string
	ServiceType string
	BaseURL     *url.URL
}

// Known public code hosts and their URLs
var (
	GitHubDotComURL = mustParseURL("https://github.com")
	GitHubDotCom    = NewCodeHost(GitHubDotComURL, TypeGitHub)

	GitLabDotComURL = mustParseURL("https://gitlab.com")
	GitLabDotCom    = NewCodeHost(GitLabDotComURL, TypeGitLab)

	MavenURL    = &url.URL{Host: "maven"}
	JVMPackages = NewCodeHost(MavenURL, TypeJVMPackages)

	PublicCodeHosts = []*CodeHost{
		GitHubDotCom,
		GitLabDotCom,
		JVMPackages,
	}
)

func NewCodeHost(baseURL *url.URL, serviceType string) *CodeHost {
	return &CodeHost{
		ServiceID:   NormalizeBaseURL(baseURL).String(),
		ServiceType: serviceType,
		BaseURL:     baseURL,
	}
}

// IsHostOfRepo returns true if the repository belongs to given code host.
func IsHostOfRepo(c *CodeHost, repo *api.ExternalRepoSpec) bool {
	return c.ServiceID == repo.ServiceID && c.ServiceType == repo.ServiceType
}

// IsHostOfAccount returns true if the account belongs to given code host.
func IsHostOfAccount(c *CodeHost, account *Account) bool {
	return c.ServiceID == account.ServiceID && c.ServiceType == account.ServiceType
}

// NormalizeBaseURL modifies the input and returns a normalized form of the a base URL with insignificant
// differences (such as in presence of a trailing slash, or hostname case) eliminated. Its return value should be
// used for the (ExternalRepoSpec).ServiceID field (and passed to XyzExternalRepoSpec) instead of a non-normalized
// base URL.
func NormalizeBaseURL(baseURL *url.URL) *url.URL {
	baseURL.Host = strings.ToLower(baseURL.Host)
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	return baseURL
}

// CodeHostOf returns the CodeHost of the given repo, if any. A correct repo name will have three
// parts separated by a "/":
// 1. Codehost URL
// 2. Repo Owner
// 3. Repo Name
//
// For example: github.com/sourcegraph/sourcegraph
//
// If "name" does not adhere to this format or the Codehost URL does not match the list of
// "codehosts" given as the argument to CodeHostOf, it will return nil, otherwise it retuns the
// matching codehost from the given list.
func CodeHostOf(name api.RepoName, codehosts ...*CodeHost) *CodeHost {
	repoNameParts := strings.Split(string(name), "/")
	if len(repoNameParts) != 3 {
		return nil
	}

	for _, c := range codehosts {
		if strings.EqualFold(repoNameParts[0], c.BaseURL.Hostname()) {
			return c
		}
	}
	return nil
}

func mustParseURL(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}
