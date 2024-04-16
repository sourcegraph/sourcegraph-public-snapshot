package cloneurl

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	ghauth "github.com/sourcegraph/sourcegraph/internal/extsvc/github/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitolite"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/pagure"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/perforce"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/phabricator"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func ForEncryptableConfig(ctx context.Context, db database.DB, kind string, config *extsvc.EncryptableConfig, repo *types.Repo) (string, error) {
	parsed, err := extsvc.ParseEncryptableConfig(ctx, kind, config)
	if err != nil {
		return "", errors.Wrap(err, "loading service configuration")
	}

	return cloneURL(ctx, db, kind, parsed, repo)
}

func cloneURL(ctx context.Context, db database.DB, kind string, parsed any, repo *types.Repo) (string, error) {
	switch t := parsed.(type) {
	case *schema.AWSCodeCommitConnection:
		if r, ok := repo.Metadata.(*awscodecommit.Repository); ok {
			return awsCodeCloneURL(r, t)
		}
	case *schema.AzureDevOpsConnection:
		if r, ok := repo.Metadata.(*azuredevops.Repository); ok {
			return azureDevOpsCloneURL(r, t)
		}
	case *schema.BitbucketServerConnection:
		if r, ok := repo.Metadata.(*bitbucketserver.Repo); ok {
			return bitbucketServerCloneURL(r, t)
		}
	case *schema.BitbucketCloudConnection:
		if r, ok := repo.Metadata.(*bitbucketcloud.Repo); ok {
			return bitbucketCloudCloneURL(r, t)
		}
	case *schema.GerritConnection:
		if r, ok := repo.Metadata.(*gerrit.Project); ok {
			return gerritCloneURL(r, t)
		}
	case *schema.GitHubConnection:
		if r, ok := repo.Metadata.(*github.Repository); ok {
			return githubCloneURL(ctx, db, r, t)
		}
	case *schema.GitLabConnection:
		if r, ok := repo.Metadata.(*gitlab.Project); ok {
			return gitlabCloneURL(r, t)
		}
	case *schema.GitoliteConnection:
		if r, ok := repo.Metadata.(*gitolite.Repo); ok {
			return r.URL, nil
		}
	case *schema.PerforceConnection:
		if r, ok := repo.Metadata.(*perforce.Depot); ok {
			return perforceCloneURL(r, t), nil
		}
	case *schema.PhabricatorConnection:
		if r, ok := repo.Metadata.(*phabricator.Repo); ok {
			return phabricatorCloneURL(r, t)
		}
	case *schema.PagureConnection:
		if r, ok := repo.Metadata.(*pagure.Project); ok {
			return r.FullURL, nil
		}
	case *schema.OtherExternalServiceConnection:
		if r, ok := repo.Metadata.(*extsvc.OtherRepoMetadata); ok {
			return otherCloneURL(repo, r), nil
		}
	case *schema.GoModulesConnection:
		return string(repo.Name), nil
	case *schema.PythonPackagesConnection:
		return string(repo.Name), nil
	case *schema.RustPackagesConnection:
		return string(repo.Name), nil
	case *schema.RubyPackagesConnection:
		return string(repo.Name), nil
	case *schema.JVMPackagesConnection:
		if r, ok := repo.Metadata.(*reposource.MavenMetadata); ok {
			return r.Module.CloneURL(), nil
		}
	case *schema.NpmPackagesConnection:
		if r, ok := repo.Metadata.(*reposource.NpmMetadata); ok {
			return r.Package.CloneURL(), nil
		}
	default:
		return "", errors.Errorf("unknown external service kind %q for repo %d", kind, repo.ID)
	}
	return "", errors.Errorf("unknown repo.Metadata type %T for repo %d", repo.Metadata, repo.ID)
}

func awsCodeCloneURL(repo *awscodecommit.Repository, cfg *schema.AWSCodeCommitConnection) (string, error) {
	u, err := url.Parse(repo.HTTPCloneURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse HTTP clone URL")
	}

	username := cfg.GitCredentials.Username
	password := cfg.GitCredentials.Password

	u.User = url.UserPassword(username, password)
	return u.String(), nil
}

func azureDevOpsCloneURL(repo *azuredevops.Repository, cfg *schema.AzureDevOpsConnection) (string, error) {
	if cfg.GitURLType == "ssh" {
		return repo.SSHURL, nil
	}

	u, err := url.Parse(repo.RemoteURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse remote URL")
	}
	u.User = url.UserPassword(cfg.Username, cfg.Token)

	return u.String(), nil
}

func bitbucketServerCloneURL(repo *bitbucketserver.Repo, cfg *schema.BitbucketServerConnection) (string, error) {
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
			cloneURL = l.Href
			if cfg.Username != "" {
				u, err := url.Parse(l.Href)
				if err == nil {
					if password != "" {
						u.User = url.UserPassword(cfg.Username, password)
					} else {
						u.User = url.User(cfg.Username)
					}
					cloneURL = u.String()
				}
			}
			// No break, so that we fallback to http in case of ssh missing
			// with GitURLType == "ssh"
		}
	}

	if cloneURL == "" {
		return "", errors.New("no clone URL found in repo links")
	}

	return cloneURL, nil
}

// bitbucketCloudCloneURL returns the repository's Git remote URL with the configured
// Bitbucket Cloud app password or workspace access token inserted in the URL userinfo.
func bitbucketCloudCloneURL(repo *bitbucketcloud.Repo, cfg *schema.BitbucketCloudConnection) (string, error) {
	if cfg.GitURLType == "ssh" {
		return fmt.Sprintf("git@%s:%s.git", cfg.Url, repo.FullName), nil
	}

	httpsURL, err := repo.Links.Clone.HTTPS()
	if err != nil {
		return "", err
	}
	u, err := url.Parse(httpsURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse HTTPS clone URL")
	}

	if cfg.AccessToken != "" {
		u.User = url.UserPassword("x-token-auth", cfg.AccessToken)
	} else {
		u.User = url.UserPassword(cfg.Username, cfg.AppPassword)
	}
	return u.String(), nil
}

func githubCloneURL(ctx context.Context, db database.DB, repo *github.Repository, cfg *schema.GitHubConnection) (string, error) {
	if cfg.GitURLType == "ssh" {
		baseURL, err := url.Parse(cfg.Url)
		if err != nil {
			return "", err
		}
		baseURL = extsvc.NormalizeBaseURL(baseURL)
		originalHostname := baseURL.Hostname()
		cloneUrl := fmt.Sprintf("git@%s:%s.git", originalHostname, repo.NameWithOwner)
		return cloneUrl, nil
	}

	if repo.URL == "" {
		return "", errors.New("empty repo.URL")
	}
	if cfg.Token == "" && cfg.GitHubAppDetails == nil {
		return repo.URL, nil
	}
	u, err := url.Parse(repo.URL)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse repo.URL")
	}

	auther, err := ghauth.FromConnection(context.Background(), cfg, db.GitHubApps(), keyring.Default().GitHubAppKey)
	if err != nil {
		return "", err
	}
	if auther.NeedsRefresh() {
		if err := auther.Refresh(ctx, httpcli.ExternalClient); err != nil {
			return "", err
		}
	}
	auther.SetURLUser(u)

	return u.String(), nil
}

// authenticatedRemoteURL returns the GitLab project's Git remote URL with the
// configured GitLab personal access token inserted in the URL userinfo.
func gitlabCloneURL(repo *gitlab.Project, cfg *schema.GitLabConnection) (string, error) {
	if cfg.GitURLType == "ssh" {
		return repo.SSHURLToRepo, nil // SSH authentication must be provided out-of-band
	}
	if cfg.Token == "" {
		return repo.HTTPURLToRepo, nil
	}
	u, err := url.Parse(repo.HTTPURLToRepo)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse repo.HTTPURLToRepo")
	}
	username := "git"
	if cfg.TokenType == "oauth" {
		username = "oauth2"
	}
	u.User = url.UserPassword(username, cfg.Token)
	return u.String(), nil
}

func gerritCloneURL(project *gerrit.Project, cfg *schema.GerritConnection) (string, error) {
	if cfg.GitURLType == "ssh" {
		return project.SSHURLToRepo, nil // SSH authentication must be provided out-of-band
	}

	// TODO: Backcompat. Remove this branch in Sourcegraph 5.5.
	if project.HTTPURLToRepo == "" {
		u, err := url.Parse(cfg.Url)
		if err != nil {
			return "", errors.Wrap(err, "failed to parse Gerrit URL")
		}
		u.User = url.UserPassword(cfg.Username, cfg.Password)

		// Gerrit encodes slashes in IDs, so need to decode them. The 'a' is for cloning with auth.
		u.Path = path.Join("a", strings.ReplaceAll(project.ID, "%2F", "/"))

		return u.String(), nil
	}

	u, err := url.Parse(project.HTTPURLToRepo)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse Gerrit project remote URL")
	}
	u.User = url.UserPassword(cfg.Username, cfg.Password)

	return u.String(), nil
}

// perforceCloneURL composes a clone URL for a Perforce depot based on
// given information. e.g.
// perforce://admin:password@ssl:111.222.333.444:1666//Sourcegraph/
func perforceCloneURL(depot *perforce.Depot, cfg *schema.PerforceConnection) string {
	cloneURL := url.URL{
		Scheme: "perforce",
		User:   url.UserPassword(cfg.P4User, cfg.P4Passwd),
		Host:   cfg.P4Port,
		Path:   depot.Depot,
	}
	return cloneURL.String()
}

func phabricatorCloneURL(repo *phabricator.Repo, _ *schema.PhabricatorConnection) (string, error) {
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
		return "", errors.New("no clone URL found")
	}

	return cloneURL, nil
}

func otherCloneURL(repo *types.Repo, m *extsvc.OtherRepoMetadata) string {
	return repo.ExternalRepo.ServiceID + m.RelativePath
}
