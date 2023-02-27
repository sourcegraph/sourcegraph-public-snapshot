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

func (c *CodeHost) IsPackageHost() bool {
	switch c.ServiceType {
	case TypeNpmPackages, TypeJVMPackages, TypeGoModules, TypePythonPackages, TypeRustPackages, TypeRubyPackages:
		return true
	}
	return false
}

// Known public code hosts and their URLs
var (
	GitHubDotComURL = mustParseURL("https://github.com")
	GitHubDotCom    = NewCodeHost(GitHubDotComURL, TypeGitHub)

	GitLabDotComURL = mustParseURL("https://gitlab.com")
	GitLabDotCom    = NewCodeHost(GitLabDotComURL, TypeGitLab)

	BitbucketOrgURL = mustParseURL("https://bitbucket.org")

	MavenURL    = &url.URL{Host: "maven"}
	JVMPackages = NewCodeHost(MavenURL, TypeJVMPackages)

	NpmURL      = &url.URL{Host: "npm"}
	NpmPackages = NewCodeHost(NpmURL, TypeNpmPackages)

	GoURL     = &url.URL{Host: "go"}
	GoModules = NewCodeHost(GoURL, TypeGoModules)

	PythonURL      = &url.URL{Host: "python"}
	PythonPackages = NewCodeHost(PythonURL, TypePythonPackages)

	RustURL      = &url.URL{Host: "crates"}
	RustPackages = NewCodeHost(RustURL, TypeRustPackages)

	RubyURL      = &url.URL{Host: "rubygems"}
	RubyPackages = NewCodeHost(RubyURL, TypeRubyPackages)

	PublicCodeHosts = []*CodeHost{
		GitHubDotCom,
		GitLabDotCom,
		JVMPackages,
		NpmPackages,
		GoModules,
		PythonPackages,
		RustPackages,
		RubyPackages,
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

// CodeHostOf returns the CodeHost of the given repo, if any. If CodeHostOf fails to find a match
// from the list of "codehosts" in the argument, it will return nil. Otherwise it retuns the
// matching codehost from the given list.
func CodeHostOf(name api.RepoName, codehosts ...*CodeHost) *CodeHost {
	// We do not want to fail in case the name includes query parameteres with a "/" in it. As a
	// result we only want to retrieve the first substring delimited by a "/" and verify if it is a
	// valid CodeHost URL.
	//
	// This means that the following check will let repo names like github.com/sourcegraph
	// pass. This function is only reponsible for identifying the CodeHost from a repo name and not
	// if the repo name points to a valid repo.
	repoNameParts := strings.SplitN(string(name), "/", 2)

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
