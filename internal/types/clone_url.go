package types

import (
	"fmt"
	"net/url"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/phabricator"
	"github.com/sourcegraph/sourcegraph/schema"
)

// RepoCloneURL builds a cloneURL for the given repo based on the
// external service configuration.
// If authentication information is found in the configuration,
// it returns an authenticated URL for the selected code host.
func RepoCloneURL(kind, config string, repo *Repo) (string, error) {
	parsed, err := extsvc.ParseConfig(kind, config)
	if err != nil {
		return "", errors.Wrap(err, "loading service configuration")
	}

	switch t := parsed.(type) {
	case *schema.AWSCodeCommitConnection:
		return awsCodeCloneURL(repo.Metadata.(*awscodecommit.Repository), t), nil
	case *schema.BitbucketServerConnection:
		return bitbucketServerCloneURL(repo.Metadata.(*bitbucketserver.Repo), t), nil
	case *schema.BitbucketCloudConnection:
		return bitbucketCloudCloneURL(repo.Metadata.(*bitbucketcloud.Repo), t), nil
	case *schema.GitHubConnection:
		return githubCloneURL(repo.Metadata.(*github.Repository), t)
	case *schema.GitLabConnection:
		return gitlabCloneURL(repo.Metadata.(*gitlab.Project), t), nil
	case *schema.GitoliteConnection:
		return repo.Metadata.(*gitolite.Repo).URL, nil
	case *schema.PerforceConnection:
		return perforceCloneURL(repo.Metadata.(map[string]interface{}), t), nil
	case *schema.PhabricatorConnection:
		return phabricatorCloneURL(repo.Metadata.(*phabricator.Repo), t), nil
	case *schema.OtherExternalServiceConnection:
		return otherCloneURL(repo, t)
	default:
		return "", fmt.Errorf("unknown external service kind %q", kind)
	}
}

func awsCodeCloneURL(repo *awscodecommit.Repository, cfg *schema.AWSCodeCommitConnection) string {
	u, err := url.Parse(repo.HTTPCloneURL)
	if err != nil {
		log15.Warn("Error adding authentication to AWS CodeCommit repository Git remote URL.", "url", repo.HTTPCloneURL, "error", err)
		return repo.HTTPCloneURL
	}

	username := cfg.GitCredentials.Username
	password := cfg.GitCredentials.Password

	u.User = url.UserPassword(username, password)
	return u.String()
}

func bitbucketServerCloneURL(repo *bitbucketserver.Repo, cfg *schema.BitbucketServerConnection) string {
	var cloneURL string
	for _, l := range repo.Links.Clone {
		if l.Name == "ssh" && cfg.GitURLType == "ssh" {
			cloneURL = l.Href
			break
		}
		if l.Name == "http" {
			var password string
			if cfg.Token != "" {
				password = cfg.Token // prefer personal access token
			} else {
				password = cfg.Password
			}
			cloneURL = setUserinfoBestEffort(l.Href, cfg.Username, password)
			// No break, so that we fallback to http in case of ssh missing
			// with GitURLType == "ssh"
		}
	}

	return cloneURL
}

// bitbucketCloudCloneURL returns the repository's Git remote URL with the configured
// Bitbucket Cloud app password inserted in the URL userinfo.
func bitbucketCloudCloneURL(repo *bitbucketcloud.Repo, cfg *schema.BitbucketCloudConnection) string {
	if cfg.GitURLType == "ssh" {
		return fmt.Sprintf("git@%s:%s.git", cfg.Url, repo.FullName)
	}

	fallbackURL := (&url.URL{
		Scheme: "https",
		Host:   cfg.Url,
		Path:   "/" + repo.FullName,
	}).String()

	httpsURL, err := repo.Links.Clone.HTTPS()
	if err != nil {
		log15.Warn("Error adding authentication to Bitbucket Cloud repository Git remote URL.", "url", repo.Links.Clone, "error", err)
		return fallbackURL
	}
	u, err := url.Parse(httpsURL)
	if err != nil {
		log15.Warn("Error adding authentication to Bitbucket Cloud repository Git remote URL.", "url", httpsURL, "error", err)
		return fallbackURL
	}

	u.User = url.UserPassword(cfg.Username, cfg.AppPassword)
	return u.String()
}

func githubCloneURL(repo *github.Repository, cfg *schema.GitHubConnection) (string, error) {
	if cfg.GitURLType == "ssh" {
		baseURL, err := url.Parse(cfg.Url)
		if err != nil {
			return "", err
		}
		baseURL = extsvc.NormalizeBaseURL(baseURL)
		originalHostname := baseURL.Hostname()
		url := fmt.Sprintf("git@%s:%s.git", originalHostname, repo.NameWithOwner)
		return url, nil
	}

	if cfg.Token == "" {
		return repo.URL, nil
	}
	u, err := url.Parse(repo.URL)
	if err != nil {
		log15.Warn("Error adding authentication to GitHub repository Git remote URL.", "url", repo.URL, "error", err)
		return repo.URL, nil
	}
	u.User = url.User(cfg.Token)
	return u.String(), nil
}

// authenticatedRemoteURL returns the GitLab projects's Git remote URL with the
// configured GitLab personal access token inserted in the URL userinfo.
func gitlabCloneURL(repo *gitlab.Project, cfg *schema.GitLabConnection) string {
	if cfg.GitURLType == "ssh" {
		return repo.SSHURLToRepo // SSH authentication must be provided out-of-band
	}
	if cfg.Token == "" {
		return repo.HTTPURLToRepo
	}
	u, err := url.Parse(repo.HTTPURLToRepo)
	if err != nil {
		log15.Warn("Error adding authentication to GitLab repository Git remote URL.", "url", repo.HTTPURLToRepo, "error", err)
		return repo.HTTPURLToRepo
	}
	// Any username works; "git" is not special.
	u.User = url.UserPassword("git", cfg.Token)
	return u.String()
}

// perforceCloneURL composes a clone URL for a Perforce depot based on
// given information. e.g.
// perforce://admin:password@ssl:111.222.333.444:1666//Sourcegraph/
func perforceCloneURL(repo map[string]interface{}, cfg *schema.PerforceConnection) string {
	cloneURL := url.URL{
		Scheme: "perforce",
		User:   url.UserPassword(cfg.P4User, cfg.P4Passwd),
		Host:   cfg.P4Port,
		Path:   repo["depot"].(string),
	}
	return cloneURL.String()
}

func phabricatorCloneURL(repo *phabricator.Repo, _ *schema.PhabricatorConnection) string {
	var external []*phabricator.URI
	builtin := make(map[string]*phabricator.URI)

	for _, u := range repo.URIs {
		if u.Disabled || u.Normalized == "" {
			continue
		} else if u.BuiltinIdentifier != "" {
			builtin[u.BuiltinProtocol+"+"+u.BuiltinIdentifier] = u
		} else {
			external = append(external, u)
		}
	}

	var name string
	if len(external) > 0 {
		name = external[0].Normalized
	}

	var cloneURL string
	for _, alt := range [...]struct {
		protocol, identifier string
	}{ // Ordered by priority.
		{"https", "shortname"},
		{"https", "callsign"},
		{"https", "id"},
		{"ssh", "shortname"},
		{"ssh", "callsign"},
		{"ssh", "id"},
	} {
		if u, ok := builtin[alt.protocol+"+"+alt.identifier]; ok {
			cloneURL = u.Effective
			// TODO(tsenart): Authenticate the cloneURL with the user's
			// VCS password once we have that setting in the config. The
			// Conduit token can't be used for cloning.
			// cloneURL = setUserinfoBestEffort(cloneURL, conn.VCSPassword, "")

			if name == "" {
				name = u.Normalized
			}
		}
	}

	if cloneURL == "" {
		log15.Warn("unable to construct clone URL for repo", "name", name, "phabricator_id", repo.PHID)
	}

	return cloneURL
}

func otherCloneURL(repo *Repo, cfg *schema.OtherExternalServiceConnection) (string, error) {
	if cfg.Url == "" {
		return repo.URI, nil
	}

	pattern := cfg.RepositoryPathPattern
	if pattern == "" {
		pattern = "{base}/{repo}"
	}

	base, err := url.Parse(cfg.Url)
	if err != nil {
		return "", err
	}

	for _, name := range cfg.Repos {
		// normalize the repo name for comparison with the unescaped repo.Name
		uName, err := url.Parse(name)
		if err != nil {
			return "", err
		}

		if reposource.OtherRepoName(pattern, cfg.Url, uName.Path) == string(repo.Name) {
			u, err := base.Parse(uName.Path)
			if err != nil {
				return "", err
			}
			return u.String(), nil
		}
	}

	return repo.URI, nil
}

// setUserinfoBestEffort adds the username and password to rawurl. If user is
// not set in rawurl, username is used. If password is not set and there is a
// user, password is used. If anything fails, the original rawurl is returned.
func setUserinfoBestEffort(rawurl, username, password string) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		return rawurl
	}

	passwordSet := password != ""

	// Update username and password if specified in rawurl
	if u.User != nil {
		if u.User.Username() != "" {
			username = u.User.Username()
		}
		if p, ok := u.User.Password(); ok {
			password = p
			passwordSet = true
		}
	}

	if username == "" {
		return rawurl
	}

	if passwordSet {
		u.User = url.UserPassword(username, password)
	} else {
		u.User = url.User(username)
	}

	return u.String()
}
