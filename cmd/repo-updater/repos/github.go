package repos

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-github/github"
	log15 "gopkg.in/inconshreveable/log15.v2"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// configAndClient binds together a GitHub config and the authenticated GitHub client created from that config.
// This is a KLUDGE to enable public repository cloning with an authenticated client,
// This data structure would probably be obviated by better separation of responsibility between the
// different packages involved in indexing.
type configAndClient struct {
	config schema.GitHubConnection
	client *github.Client
}

var (
	githubConf = conf.Get().Github
)

// RunGitHubRepositorySyncWorker runs the worker that syncs repositories from the GitHub Enterprise instance to Sourcegraph
func RunGitHubRepositorySyncWorker(ctx context.Context) error {
	if updateIntervalConf == 0 {
		return errors.New("Update interval is 0 (set repoListUpdateInterval to a non-zero value or omit it)")
	}
	updateInterval := time.Duration(updateIntervalConf) * time.Second

	configs := githubConf
	var clients []configAndClient
	for _, c := range configs {
		u, err := url.Parse(c.Url)
		if err != nil {
			return fmt.Errorf("error processing GitHub config URL %s: %s", c.Url, err)
		}
		if u.Hostname() == "github.com" {
			config := &githubutil.Config{Context: ctx}
			cc := configAndClient{config: c}
			if c.Token != "" {
				cc.client = config.AuthedClient(c.Token)
			} else {
				cc.client = config.UnauthedClient()
			}
			clients = append(clients, cc)
		} else {
			cl, err := githubEnterpriseClient(ctx, c.Url, c.Certificate, c.Token)
			if err != nil {
				return err
			}
			clients = append(clients, configAndClient{config: c, client: cl})
		}
	}

	if len(clients) == 0 {
		return nil
	}
	for {
		for _, c := range clients {
			// update explicitly listed repositories
			var explicitRepos []*github.Repository
			for _, ownerAndRepoString := range c.config.Repos {
				ownerAndRepo := strings.Split(ownerAndRepoString, "/")
				if len(ownerAndRepo) != 2 {
					log15.Error("Could not update public GitHub repository, name must be owner/repo format", "repo", ownerAndRepoString)
					continue
				}
				repo, _, err := c.client.Repositories.Get(ctx, ownerAndRepo[0], ownerAndRepo[1])
				if err != nil {
					log15.Error("Could not update public GitHubrepository", "error", err)
					continue
				}
				explicitRepos = append(explicitRepos, repo)
			}
			updateGitHubRepos(ctx, c.client, explicitRepos)

			// update implicit repositories (repositories owned by an organization to which the user who created
			// the personal access token belongs)
			err := updateForClient(ctx, c.client)
			if err != nil {
				log15.Error("Could not update repositories", "error", err)
			}
		}
		time.Sleep(updateInterval)
	}
}

func githubEnterpriseClient(ctx context.Context, gheURL, cert, accessToken string) (*github.Client, error) {
	gheAPIURL := fmt.Sprintf("%s/api/v3/", gheURL)
	baseURL, err := url.Parse(gheAPIURL)
	if err != nil {
		return nil, err
	}

	var transport http.RoundTripper
	if cert != "" {
		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM([]byte(cert)); !ok {
			return nil, errors.New("Invalid certificate value")
		}
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: certPool},
		}
	}

	config := &githubutil.Config{
		BaseURL:   baseURL,
		Context:   ctx,
		Transport: transport,
	}

	return config.AuthedClient(accessToken), nil
}

// updateForClient ensures that all public and private repositories owned by the authenticated user
// are updated.
func updateForClient(ctx context.Context, client *github.Client) error {
	var repos []*github.Repository
	var err error
	if client.BaseURL.Host == "api.github.com" {
		repos, err = fetchAllGitHubRepos(ctx, client)
	} else {
		repos, err = fetchAllGitHubEnterpriseRepos(ctx, client)
	}
	if err != nil {
		return fmt.Errorf("could not list repositories: %s", err)
	}
	return updateGitHubRepos(ctx, client, repos)
}

// update ensures that all provided repositories exist in the repository table.
// It adds each repository with a URI of the form
// "${GITHUB_CLIENT_HOSTNAME}/${GITHUB_REPO_FULL_NAME}".
func updateGitHubRepos(ctx context.Context, client *github.Client, repos []*github.Repository) error {
	for i, ghRepo := range repos {
		hostPart := client.BaseURL.Host
		if hostPart == "api.github.com" {
			hostPart = "github.com"
		}
		uri := fmt.Sprintf("%s/%s", hostPart, ghRepo.GetFullName())

		repo, err := sourcegraph.InternalClient.ReposCreateIfNotExists(ctx, uri, ghRepo.GetDescription(), ghRepo.GetFork(), ghRepo.GetPrivate())
		if err != nil {
			log15.Warn("Could not ensure repository exists", "uri", uri, "error", err)
			continue
		}
		// Run a fetch kick-off an update or a clone if the repo doesn't already exist.
		cmd := gitserver.DefaultClient.Command("git", "fetch")
		cmd.Repo = repo
		err = cmd.Run(ctx)
		if err != nil {
			log15.Warn("Could not ensure repository cloned", "uri", uri, "error", err)
			continue
		}

		// Every 100 repos we clone, wait a bit to prevent overloading gitserver.
		if i > 0 && i%100 == 0 {
			log15.Info(fmt.Sprintf("%d out of %d repositories updated. Waiting for a moment.", i, len(repos)))
			time.Sleep(1 * time.Minute)
		}
	}

	return nil
}

// fetchAllGitHubRepos returns all repos that belong to the org of the provided client's authenticated user.
func fetchAllGitHubRepos(ctx context.Context, client *github.Client) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	orgs, _, err := client.Organizations.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}
	for _, org := range orgs {
		if org.Login == nil {
			return nil, errors.New("org login required")
		}
		var resp *github.Response
		for page := 1; page != 0; page = resp.NextPage {
			var repos []*github.Repository
			var err error
			repos, resp, err = client.Repositories.ListByOrg(ctx, *org.Login, &github.RepositoryListByOrgOptions{
				ListOptions: github.ListOptions{Page: page, PerPage: 100},
			})
			if err != nil {
				return nil, err
			}
			allRepos = append(allRepos, repos...)
		}
	}
	return allRepos, nil
}

// fetchAllGitHubEnterpriseRepos returns all repos that exists on a GitHub instance and
// all of the private repositories for the provided client's authenticated user.
func fetchAllGitHubEnterpriseRepos(ctx context.Context, client *github.Client) (allRepos []*github.Repository, err error) {
	repoMap := make(map[string]*github.Repository)
	// Add all public repositories
	since := 0
	for {
		repos, _, err := client.Repositories.ListAll(ctx, &github.RepositoryListAllOptions{Since: since})
		if err != nil {
			return nil, err
		}
		if len(repos) == 0 {
			break
		}
		for _, repo := range repos {
			if repo.FullName == nil {
				continue
			}
			repoMap[*repo.FullName] = repo
		}
		since = *repos[len(repos)-1].ID
	}
	// Add all private repositories corresponding to the access token user
	for page := 1; ; {
		repos, resp, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{
			ListOptions: github.ListOptions{Page: page, PerPage: 100},
		})
		if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			if repo.FullName == nil {
				continue
			}
			repoMap[*repo.FullName] = repo
		}
		page = resp.NextPage
		if page == 0 {
			break
		}
	}

	for _, repo := range repoMap {
		allRepos = append(allRepos, repo)
	}
	return allRepos, nil
}
