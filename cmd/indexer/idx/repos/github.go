package repos

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-github/github"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
)

var (
	updateIntervalEnv = env.Get("REPOSITORY_SYNC_PERIOD", "60", "The number of seconds to wait in-between syncing repositories with the code host")

	// GitHub.com config
	ghcAccessToken = env.Get("GITHUB_PERSONAL_ACCESS_TOKEN", "", "personal access token for GitHub.com. All requests will use this token to access the Github API. If set, this will be used to sync private GitHub repositories to Sourcegraph Server.")

	// GitHub Enterprise config
	gheURL         = env.Get("GITHUB_ENTERPRISE_URL", "", "URL to a GitHub Enterprise instance. If non-empty, repositories are synced from this instance periodically. Note: this environment variable must be set to the same value in the gitserver process.")
	gheCert        = env.Get("GITHUB_ENTERPRISE_CERT", "", "TLS certificate of GitHub Enterprise instance, if not part of the standard certificate chain")
	gheAccessToken = env.Get("GITHUB_ENTERPRISE_TOKEN", "", "Access token used to authenticate GitHub Enterprise API requests")
	gheParsedURL   *url.URL
)

func init() {
	var err error
	gheParsedURL, err = url.Parse(gheURL)
	if err != nil {
		log.Fatalf("Couldn't parse GitHub Enterprise URL: %s", err)
	}
}

// RunRepositorySyncWorker runs the worker that syncs repositories from the GitHub Enterprise instance to Sourcegraph
func RunRepositorySyncWorker(ctx context.Context) error {
	updateIntervalParsed, err := strconv.Atoi(updateIntervalEnv)
	if err != nil {
		return err
	}
	if updateIntervalParsed == 0 {
		return errors.New("Update interval is 0 (set REPOSITORY_SYNC_PERIOD to a non-zero value or omit it)")
	}
	updateInterval := time.Duration(updateIntervalParsed) * time.Second

	var clients []*github.Client
	if gheURL != "" {
		c, err := githubEnterpriseClient(ctx, gheURL, gheCert, gheAccessToken)
		if err != nil {
			return err
		}
		clients = append(clients, c)
	}
	if ghcAccessToken != "" {
		config := &githubutil.Config{Context: ctx}
		clients = append(clients, config.AuthedClient(ghcAccessToken))
	}
	if len(clients) == 0 {
		return nil
	}

	for {
		for _, c := range clients {
			err := update(ctx, c)
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
		if ok := certPool.AppendCertsFromPEM([]byte(gheCert)); !ok {
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

	return config.AuthedClient(gheAccessToken), nil
}

// update ensures that all public and private repositories owned by the authenticated user
// exist in the repository table. It adds each repository with a URI of the form
// "${GITHUB_CLIENT_HOSTNAME}/${GITHUB_REPO_FULL_NAME}".
func update(ctx context.Context, client *github.Client) error {
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

	for i, ghRepo := range repos {
		if ghRepo.FullName == nil {
			continue
		}

		hostPart := client.BaseURL.Host
		if hostPart == "api.github.com" {
			hostPart = "github.com"
		}
		uri := fmt.Sprintf("%s/%s", hostPart, *ghRepo.FullName)
		repo, err := backend.Repos.GetByURI(ctx, uri)
		if err != nil {
			log15.Warn("Could not ensure repository existence", "uri", uri, "error", err)
			continue
		}
		// Run a fetch kick-off an update or a clone if the repo doesn't already exist.
		cmd := gitserver.DefaultClient.Command("git", "fetch")
		cmd.Repo = repo
		err = cmd.Run(ctx)
		if err != nil {
			log15.Warn("Could not ensure repository cloned", "uri", uri, "error", err)
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
