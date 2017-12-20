// Package originmap maps Sourcegraph repository URIs to repository
// origins (i.e., clone URLs). It accepts external customization via
// the ORIGIN_MAP environment variable.
//
// It always includes the mapping
// "github.com/!https://github.com/%.git" (github.com ->
// https://github.com/%.git)
package server

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

type prefixAndOrgin struct {
	Prefix string
	Origin string
}

var originMapEnv = conf.Get().GitOriginMap
var gitoliteHostsEnv = conf.Get().GitoliteHosts
var githubConf = conf.Get().Github
var reposListConf = conf.Get().ReposList

// DEPRECATED in favor of GITHUB_CONFIG:
var githubEnterpriseURLEnv = conf.Get().GithubEnterpriseURL

var originMap []prefixAndOrgin
var gitoliteHostMap []prefixAndOrgin

// reposListOriginMap is a mapping from repo URI (path) to repo origin (clone URL).
var reposListOriginMap = make(map[string]string)

func init() {
	var err error
	originMap, err = parse(originMapEnv, 1)
	if err != nil {
		log.Fatal(err)
	}

	gitoliteHostMap, err = parse(gitoliteHostsEnv, 0)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range gitoliteHostMap {
		originMap = append(originMap, prefixAndOrgin{Prefix: entry.Prefix, Origin: entry.Origin + ":%"})
	}

	// Add origin map for repos.list configuration.
	for _, c := range reposListConf {
		reposListOriginMap[c.Path] = c.Url
	}

	// Add origin map for GitHub Enterprise instance of the form "${HOSTNAME}/!git@${HOSTNAME}:%.git"
	//
	// TODO: remove after removing deprecated GITHUB_ENERPRISE config.
	if githubEnterpriseURLEnv != "" {
		gheURL, err := url.Parse(githubEnterpriseURLEnv)
		if err != nil {
			log.Fatal(err)
		}
		originMap = append(originMap, prefixAndOrgin{Prefix: gheURL.Hostname() + "/", Origin: fmt.Sprintf("git@%s:%%.git", gheURL.Hostname())})
	}

	// Add origin map for GitHub Enterprise instances of the form "${HOSTNAME}/!git@${HOSTNAME}:%.git"
	for _, c := range githubConf {
		ghURL, err := url.Parse(c.Url)
		if err != nil {
			log.Fatal(err)
		}
		// Clone via SSH if this GitHub Enterprise has a self-signed certificate provided.
		// Otherwise git will run into issues with cloning over HTTPS using an invalid certificate.
		if c.Certificate != "" {
			originMap = append(originMap, prefixAndOrgin{Prefix: ghURL.Hostname() + "/", Origin: fmt.Sprintf("git@%s:%%.git", ghURL.Hostname())})
		} else {
			var auth string
			if c.Token != "" {
				auth = c.Token + "@"
			}
			originMap = append(originMap, prefixAndOrgin{Prefix: ghURL.Hostname() + "/", Origin: fmt.Sprintf("%s://%s%s/%%.git", ghURL.Scheme, auth, ghURL.Hostname())})
		}
	}

	addGitHubDefaults()
}

func addGitHubDefaults() {
	// Note: We add several variants here specifically for reverse, so that if
	// a user-provided clone URL is passed to reverse, it still functions as
	// expected. For the case of OriginMap, the first one is returned (i.e. the
	// order below matters).
	originMap = append(originMap, prefixAndOrgin{Prefix: "github.com/", Origin: "https://github.com/%.git"})
	originMap = append(originMap, prefixAndOrgin{Prefix: "github.com/", Origin: "http://github.com/%.git"})
	originMap = append(originMap, prefixAndOrgin{Prefix: "github.com/", Origin: "ssh://git@github.com:%.git"})
	originMap = append(originMap, prefixAndOrgin{Prefix: "github.com/", Origin: "git://git@github.com:%.git"})
	originMap = append(originMap, prefixAndOrgin{Prefix: "github.com/", Origin: "git@github.com:%.git"})

	originMap = append(originMap, prefixAndOrgin{Prefix: "bitbucket.org/", Origin: "https://bitbucket.org/%.git"})
	originMap = append(originMap, prefixAndOrgin{Prefix: "bitbucket.org/", Origin: "git@bitbucket.org:%.git"})
}

// OriginMap maps the repo URI to the repository origin (clone URL). Returns empty string if no mapping was found.
func OriginMap(repoURI string) string {
	if origin, ok := reposListOriginMap[repoURI]; ok {
		return origin
	}
	for _, entry := range originMap {
		if strings.HasPrefix(repoURI, entry.Prefix) {
			return strings.Replace(entry.Origin, "%", strings.TrimPrefix(repoURI, entry.Prefix), 1)
		}
	}
	return ""
}

// reverse maps the repository origin (clone URL) to the repo URI. Returns empty string of no mapping was found.
func reverse(cloneURL string) string {
	for uri, origin := range reposListOriginMap {
		if cloneURL == origin {
			return uri
		}
	}
	for _, entry := range originMap {
		s := strings.Split(entry.Origin, "%")
		originPrefix := s[0]
		if strings.HasPrefix(cloneURL, originPrefix) {
			originSuffix := s[1]
			repo := strings.TrimSuffix(strings.TrimPrefix(cloneURL, originPrefix), originSuffix)
			return entry.Prefix + repo
		}
	}
	return ""
}

func parse(raw string, placeholderCount int) (originMap []prefixAndOrgin, err error) {
	for _, e := range strings.Fields(raw) {
		p := strings.Split(e, "!")
		if len(p) != 2 {
			return nil, fmt.Errorf("invalid entry: %s", e)
		}
		if strings.Count(p[1], "%") != placeholderCount {
			return nil, fmt.Errorf("invalid entry: %s", e)
		}
		originMap = append(originMap, prefixAndOrgin{Prefix: p[0], Origin: p[1]})
	}
	return
}
