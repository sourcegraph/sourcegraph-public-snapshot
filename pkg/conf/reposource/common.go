package reposource

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// repoSource is a wrapper around a repository source (typically a code host config) that provides a
// method to map clone URLs to repo names using only the configuration (i.e., no network requests).
type repoSource interface {
	// cloneURLToRepoName maps a Git clone URL (format documented here:
	// https://git-scm.com/docs/git-clone#_git_urls_a_id_urls_a) to the expected repo name for the
	// repository on the code host.  It does not actually check if the repository exists in the code
	// host. It merely does the mapping based on the rules set in the code host config.
	//
	// If the clone URL does not correspond to a repository that could exist on the code host, the
	// empty string is returned and err is nil. If there is an unrelated error, an error is
	// returned.
	cloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error)
}

// CloneURLToRepoName maps a Git clone URL (format documented here:
// https://git-scm.com/docs/git-clone#_git_urls_a_id_urls_a) to the corresponding repo name if there
// exists a code host configuration that matches the clone URL. Implicitly, it includes a code host
// configuration for github.com, even if one is not explicitly specified. Returns the empty string and nil
// error if a matching code host could not be found. This function does not actually check the code
// host to see if the repository actually exists.
func CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error) {
	cfg := conf.Get()

	if repoName := customCloneURLToRepoName(cloneURL); repoName != "" {
		return repoName, nil
	}

	repoSources := make([]repoSource, 0, len(cfg.Github)+
		len(cfg.Gitlab)+
		len(cfg.BitbucketServer)+
		len(cfg.AwsCodeCommit)+
		2+ /* for repos.list and fallback GitHub.com */
		len(cfg.Gitolite))

	for _, c := range cfg.Github {
		repoSources = append(repoSources, GitHub{c})
	}
	for _, c := range cfg.Gitlab {
		repoSources = append(repoSources, GitLab{c})
	}
	for _, c := range cfg.BitbucketServer {
		repoSources = append(repoSources, BitbucketServer{c})
	}
	for _, c := range cfg.AwsCodeCommit {
		repoSources = append(repoSources, AWS{c})
	}
	repoSources = append(repoSources, getReposListInstance())
	for _, c := range cfg.Gitolite {
		repoSources = append(repoSources, Gitolite{c})
	}
	// Fallback for github.com
	repoSources = append(repoSources, GitHub{&schema.GitHubConnection{Url: "https://github.com"}})
	for _, ch := range repoSources {
		repoName, err := ch.cloneURLToRepoName(cloneURL)
		if err != nil {
			return "", err
		}
		if repoName != "" {
			return repoName, nil
		}
	}

	return "", nil
}

// NormalizeBaseURL modifies the input and returns a normalized form of the base URL with insignificant
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

// parseCloneURL parses a git clone URL into a URL struct. It supports the SCP-style git@host:path
// syntax that is common among code hosts.
func parseCloneURL(cloneURL string) (*url.URL, error) {
	if strings.HasPrefix(cloneURL, "https://") || strings.HasPrefix(cloneURL, "http://") || strings.HasPrefix(cloneURL, "ssh://") || strings.HasPrefix(cloneURL, "git://") || strings.HasPrefix(cloneURL, "rsync://") || strings.HasPrefix(cloneURL, "file://") {
		return url.Parse(cloneURL)
	}

	// Support SCP-style syntax
	u, err := url.Parse("fake://" + strings.Replace(cloneURL, ":", "/", 1))
	if err != nil {
		return nil, err
	}
	u.Scheme = ""
	u.Path = strings.TrimPrefix(u.Path, "/")
	return u, nil
}

// parseURLs parses the clone URL and repository host base URL into structs. It also returns a
// boolean indicating whether the hostnames of the URLs match.
func parseURLs(cloneURL, baseURL string) (parsedCloneURL, parsedBaseURL *url.URL, equalHosts bool, err error) {
	if baseURL != "" {
		parsedBaseURL, err = url.Parse(baseURL)
		if err != nil {
			return nil, nil, false, fmt.Errorf("Error parsing baseURL: %s", err)
		}
		parsedBaseURL = NormalizeBaseURL(parsedBaseURL)
	}

	parsedCloneURL, err = parseCloneURL(cloneURL)
	if err != nil {
		return nil, nil, false, fmt.Errorf("Error parsing cloneURL: %s", err)
	}

	return parsedCloneURL, parsedBaseURL, parsedBaseURL != nil && parsedBaseURL.Hostname() == parsedCloneURL.Hostname(), nil
}
