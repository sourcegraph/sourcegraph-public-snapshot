package repos

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-github/github"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// githubConfig binds together a GitHub connection and the authenticated
// GitHub client created from that config. This enables reuse of a single client
// for a given config.
type githubConfig struct {
	conn   schema.GitHubConnection
	client *github.Client
}

var (
	githubConns = conf.Get().Github
)

// RunGitHubRepositorySyncWorker runs the worker that syncs repositories from the GitHub Enterprise instance to Sourcegraph
func RunGitHubRepositorySyncWorker(ctx context.Context) error {
	var configs []githubConfig
	for _, c := range githubConns {
		u, err := url.Parse(c.Url)
		if err != nil {
			return fmt.Errorf("error processing GitHub config URL %s: %s", c.Url, err)
		}
		if u.Hostname() == "github.com" {
			cfg := githubConfig{conn: c}
			clientConf := &githubutil.Config{Context: ctx}
			if c.Token != "" {
				cfg.client = clientConf.AuthedClient(c.Token)
			} else {
				cfg.client = clientConf.UnauthedClient()
			}
			configs = append(configs, cfg)
		} else {
			cl, err := githubEnterpriseClient(ctx, c.Url, c.Certificate, c.Token)
			if err != nil {
				return err
			}
			configs = append(configs, githubConfig{conn: c, client: cl})
		}
	}

	if len(configs) == 0 {
		return nil
	}
	for _, c := range configs {
		go func(c githubConfig) {
			for {
				err := updateForConfig(ctx, c.conn, c.client)
				if err != nil {
					log15.Warn("error updating GitHub repos", "url", c.conn.Url, "err", err)
				}
				time.Sleep(updateInterval)
			}
		}(c)
	}
	select {}
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

// updateForConfig ensures that all public and private repositories listed by the given config
func updateForConfig(ctx context.Context, conn schema.GitHubConnection, client *github.Client) error {
	initialEnablement := conn.InitialRepositoryEnablement || conn.PreemptivelyClone
	if conn.PreemptivelyClone {
		log15.Info("The site config element github[].preemptivelyClone is deprecated. Use initialRepositoryEnablement instead.")
	}

	// update explicitly listed repositories
	var explicitRepos []*github.Repository
	for _, ownerAndRepoString := range conn.Repos {
		ownerAndRepo := strings.Split(ownerAndRepoString, "/")
		if len(ownerAndRepo) != 2 {
			log15.Error("Could not update public GitHub repository, name must be owner/repo format", "repo", ownerAndRepoString)
			continue
		}
		repo, _, err := client.Repositories.Get(ctx, ownerAndRepo[0], ownerAndRepo[1])
		if err != nil {
			log15.Error("Could not update public GitHub repository", "error", err)
			continue
		}
		explicitRepos = append(explicitRepos, repo)
	}
	updateGitHubRepos(ctx, client, explicitRepos, initialEnablement)

	// update implicit repositories (repositories owned by an organization to which the user who created
	// the personal access token belongs)
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
	updateGitHubRepos(ctx, client, repos, initialEnablement)
	return nil
}

// update ensures that all provided repositories exist in the repository table.
// It adds each repository with a URI of the form
// "${GITHUB_CLIENT_HOSTNAME}/${GITHUB_REPO_FULL_NAME}".
func updateGitHubRepos(ctx context.Context, client *github.Client, repos []*github.Repository, initialEnablement bool) {
	// Sort repos by most recently pushed, so we prioritize adding (and possibly enabling/cloning) repos first that
	// are more likely to be important.
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].GetPushedAt().Time.After(repos[j].GetPushedAt().Time)
	})

	cloned := 0
	for _, ghRepo := range repos {
		hostPart := client.BaseURL.Host
		if hostPart == "api.github.com" {
			hostPart = "github.com"
		}
		uri := api.RepoURI(fmt.Sprintf("%s/%s", hostPart, ghRepo.GetFullName()))

		repo, err := api.InternalClient.ReposCreateIfNotExists(ctx, uri, ghRepo.GetDescription(), ghRepo.GetFork(), initialEnablement)
		if err != nil {
			log15.Warn("Could not ensure repository exists", "uri", uri, "error", err)
			continue
		}

		if initialEnablement {
			// If newly added, the repository will have been set to enabled upon creation above. Explicitly enqueue a
			// clone/update now so that those occur in order of most recently pushed.
			isCloned, err := gitserver.DefaultClient.IsRepoCloned(ctx, repo.URI)
			if err != nil {
				log15.Warn("Could not ensure repository cloned", "uri", uri, "error", err)
				continue
			}
			if !isCloned {
				cloned++
				log15.Debug("fetching GitHub repo", "repo", repo.URI, "cloned", isCloned)
				err := gitserver.DefaultClient.EnqueueRepoUpdate(ctx, repo.URI)
				if err != nil {
					log15.Warn("Could not ensure repository updated", "uri", uri, "error", err)
					continue
				}

				// Every 100 repos we clone, wait a bit to prevent overloading gitserver.
				if cloned > 0 && cloned%100 == 0 {
					log15.Info(fmt.Sprintf("%d repositories cloned so far (out of %d repositories total, not all of which need cloning). Waiting for a moment.", cloned, len(repos)))
					time.Sleep(1 * time.Minute)
				}
			}
		}
	}
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
