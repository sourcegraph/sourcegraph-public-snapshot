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
)

var (
	gheURL            = env.Get("GITHUB_ENTERPRISE_URL", "", "URL to a GitHub Enterprise instance. If non-empty, repositories are synced from this instance periodically. Note: this environment variable must be set to the same value in the gitserver process.")
	gheCert           = env.Get("GITHUB_ENTERPRISE_CERT", "", "TLS certificate of GitHub Enterprise instance, if not part of the standard certificate chain")
	gheAccessToken    = env.Get("GITHUB_ENTERPRISE_TOKEN", "", "Access token used to authenticate GitHub Enterprise API requests")
	updateIntervalEnv = env.Get("REPOSITORY_SYNC_PERIOD", "60", "The number of seconds to wait in-between syncing repositories with the code host")
	gheParsedURL      *url.URL
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
	if gheURL == "" { // If no GHE URL is set, we don't expect to index from GitHub
		return nil
	}
	updateIntervalParsed, err := strconv.Atoi(updateIntervalEnv)
	if err != nil {
		return err
	}
	if updateIntervalParsed == 0 {
		return errors.New("Update interval is 0 (set REPOSITORY_SYNC_PERIOD to a non-zero value or omit it)")
	}
	updateInterval := time.Duration(updateIntervalParsed) * time.Second

	gheAPIURL := fmt.Sprintf("%s/api/v3/", gheURL)
	baseURL, err := url.Parse(gheAPIURL)
	if err != nil {
		return err
	}
	var transport http.RoundTripper
	if gheCert != "" {
		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM([]byte(gheCert)); !ok {
			return errors.New("Invalid certificate value")
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
	client := config.AuthedClient(gheAccessToken)
	for {
		update(ctx, client)
		time.Sleep(updateInterval)
	}
}

// update ensures that all public and private repositories owned by the authenticated user
// exist in the repository table. It adds each repository with a URI of the form
// "${GITHUB_ENTERPRISE_HOSTNAME}/${GITHUB_REPO_FULL_NAME}".
func update(ctx context.Context, client *github.Client) {
	repos, err := fetchAllRepos(ctx, client)
	if err != nil {
		log15.Error("Could not list repositories", "error", err)
		return
	}

	for _, repo := range repos {
		if repo.FullName == nil {
			continue
		}

		uri := fmt.Sprintf("%s/%s", gheParsedURL.Hostname(), *repo.FullName)
		if _, err := backend.Repos.GetByURI(ctx, uri); err != nil {
			log15.Warn("Could not ensure repository existence", "uri", uri, "error", err)
		}
	}
}

func fetchAllRepos(ctx context.Context, client *github.Client) (allRepos []*github.Repository, err error) {
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
