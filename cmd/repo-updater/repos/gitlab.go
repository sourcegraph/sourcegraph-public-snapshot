package repos

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/atomicvalue"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/conf/reposource"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var gitlabConnections = func() *atomicvalue.Value {
	c := atomicvalue.New()
	c.Set(func() interface{} {
		return []*gitlabConnection{}
	})
	return c
}()

// SyncGitLabConnections periodically syncs connections from
// the Frontend API.
func SyncGitLabConnections(ctx context.Context) {
	t := time.NewTicker(configWatchInterval)
	var lastConfig []*schema.GitLabConnection
	for range t.C {
		gitlabConf, err := conf.GitLabConfigs(ctx)
		if err != nil {
			log15.Error("unable to fetch Gitlab configs", "err", err)
			continue
		}

		var hasGitLabDotComConnection bool
		for _, c := range gitlabConf {
			u, _ := url.Parse(c.Url)
			if u != nil && (u.Hostname() == "gitlab.com" || u.Hostname() == "www.gitlab.com") {
				hasGitLabDotComConnection = true
				break
			}
		}
		if !hasGitLabDotComConnection {
			// Add a GitLab.com entry by default, to support navigating to URL paths like
			// /gitlab.com/foo/bar to auto-add that project.
			gitlabConf = append(gitlabConf, &schema.GitLabConnection{
				ProjectQuery:                []string{"none"}, // don't try to list all repositories during syncs
				Url:                         "https://gitlab.com",
				InitialRepositoryEnablement: true,
			})
		}

		if reflect.DeepEqual(gitlabConf, lastConfig) {
			continue
		}
		lastConfig = gitlabConf

		var conns []*gitlabConnection
		for _, c := range gitlabConf {
			conn, err := newGitLabConnection(c, nil)
			if err != nil {
				log15.Error("Error processing configured GitLab connection. Skipping it.", "url", c.Url, "error", err)
				continue
			}
			conns = append(conns, conn)
		}

		gitlabConnections.Set(func() interface{} {
			return conns
		})

		gitLabRepositorySyncWorker.restart()
	}
}

// getGitLabConnection returns the GitLab connection (config + API client) that is responsible for
// the repository specified by the args.
func getGitLabConnection(args protocol.RepoLookupArgs) (*gitlabConnection, error) {
	gitlabConnections := gitlabConnections.Get().([]*gitlabConnection)
	if args.ExternalRepo != nil && args.ExternalRepo.ServiceType == gitlab.ServiceType {
		// Look up by external repository spec.
		for _, conn := range gitlabConnections {
			if args.ExternalRepo.ServiceID == conn.baseURL.String() {
				return conn, nil
			}
		}
		return nil, errors.Wrap(gitlab.ErrNotFound, fmt.Sprintf("no configured GitLab connection with URL: %q", args.ExternalRepo.ServiceID))
	}

	if args.Repo != "" {
		// Look up by repository name.
		repo := strings.ToLower(string(args.Repo))
		for _, conn := range gitlabConnections {
			if strings.HasPrefix(repo, conn.baseURL.Hostname()+"/") {
				return conn, nil
			}
		}
	}

	return nil, nil
}

// GetGitLabRepositoryMock is set by tests that need to mock GetGitLabRepository.
var GetGitLabRepositoryMock func(args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error)

// GetGitLabRepository queries a configured GitLab connection endpoint for information about the
// specified repository (a.k.a. project in GitLab's naming scheme).
//
// If args.Repo refers to a repository that is not known to be on a configured GitLab connection's
// host, it returns authoritative == false.
func GetGitLabRepository(ctx context.Context, args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error) {
	if GetGitLabRepositoryMock != nil {
		return GetGitLabRepositoryMock(args)
	}

	ghrepoToRepoInfo := func(proj *gitlab.Project, conn *gitlabConnection) *protocol.RepoInfo {
		return &protocol.RepoInfo{
			Name:         gitlabProjectToRepoPath(conn, proj),
			ExternalRepo: gitlab.ExternalRepoSpec(proj, *conn.baseURL),
			Description:  proj.Description,
			Fork:         proj.ForkedFromProject != nil,
			Archived:     proj.Archived,
			VCS: protocol.VCSInfo{
				URL: conn.authenticatedRemoteURL(proj),
			},
			Links: &protocol.RepoLinks{
				Root:   proj.WebURL,
				Tree:   proj.WebURL + "/tree/{rev}/{path}",
				Blob:   proj.WebURL + "/blob/{rev}/{path}",
				Commit: proj.WebURL + "/commit/{commit}",
			},
		}
	}

	conn, err := getGitLabConnection(args)
	if err != nil {
		return nil, true, err // refers to a GitLab repo but the host is not configured
	}
	if conn == nil {
		return nil, false, nil // refers to a non-GitLab repo
	}

	if args.ExternalRepo != nil && args.ExternalRepo.ServiceType == gitlab.ServiceType {
		// Look up by external repository spec.
		id, err := strconv.Atoi(args.ExternalRepo.ID)
		if err != nil {
			return nil, true, err
		}
		proj, err := conn.client.GetProject(ctx, gitlab.GetProjectOp{ID: id})
		if proj != nil {
			repo = ghrepoToRepoInfo(proj, conn)
		}
		return repo, true, err
	}

	if args.Repo != "" {
		// Look up by repository name.
		pathWithNamespace := strings.TrimPrefix(strings.ToLower(string(args.Repo)), conn.baseURL.Hostname()+"/")
		proj, err := conn.client.GetProject(ctx, gitlab.GetProjectOp{PathWithNamespace: pathWithNamespace})
		if proj != nil {
			repo = ghrepoToRepoInfo(proj, conn)
		}
		return repo, true, err
	}

	return nil, true, fmt.Errorf("unable to look up GitLab repository (%+v)", args)
}

var gitLabRepositorySyncWorker = &worker{
	work: func(ctx context.Context, shutdown chan struct{}) {
		gitlabConnections := gitlabConnections.Get().([]*gitlabConnection)
		if len(gitlabConnections) == 0 {
			return
		}
		for _, c := range gitlabConnections {
			go func(c *gitlabConnection) {
				for {
					if rateLimitRemaining, rateLimitReset, ok := c.client.RateLimit.Get(); ok && rateLimitRemaining < 50 {
						wait := rateLimitReset + 10*time.Second
						log15.Warn("GitLab API rate limit is almost exhausted. Waiting until rate limit is reset.", "wait", rateLimitReset, "rateLimitRemaining", rateLimitRemaining)
						time.Sleep(wait)
					}
					updateGitLabProjects(ctx, c)
					gitlabUpdateTime.WithLabelValues(c.baseURL.String()).Set(float64(time.Now().Unix()))
					select {
					case <-shutdown:
						return
					case <-time.After(GetUpdateInterval()):
					}
				}
			}(c)
		}
	},
}

// RunGitLabRepositorySyncWorker runs the worker that syncs projects from configured GitLab instances to
// Sourcegraph.
func RunGitLabRepositorySyncWorker(ctx context.Context) {
	gitLabRepositorySyncWorker.start(ctx)
}

func gitlabProjectToRepoPath(conn *gitlabConnection, proj *gitlab.Project) api.RepoName {
	return reposource.GitLabRepoName(conn.config.RepositoryPathPattern, conn.baseURL.Hostname(), proj.PathWithNamespace)
}

// updateGitLabProjects ensures that all provided repositories exist in the repository table.
func updateGitLabProjects(ctx context.Context, conn *gitlabConnection) {
	projs, err := conn.listAllProjects(ctx)
	if err != nil {
		log15.Error("failed to list some gitlab projects", "error", err.Error())
	}

	repoChan := make(chan repoCreateOrUpdateRequest)
	defer close(repoChan)
	go createEnableUpdateRepos(ctx, fmt.Sprintf("gitlab:%s", conn.config.Token), repoChan)
	for _, proj := range projs {
		repoChan <- repoCreateOrUpdateRequest{
			RepoCreateOrUpdateRequest: api.RepoCreateOrUpdateRequest{
				RepoName:     gitlabProjectToRepoPath(conn, proj),
				ExternalRepo: gitlab.ExternalRepoSpec(proj, *conn.baseURL),
				Description:  proj.Description,
				Fork:         proj.ForkedFromProject != nil,
				Archived:     proj.Archived,
				Enabled:      conn.config.InitialRepositoryEnablement,
			},
			URL: conn.authenticatedRemoteURL(proj),
		}
	}
}

func newGitLabConnection(config *schema.GitLabConnection, cf httpcli.Factory) (*gitlabConnection, error) {
	baseURL, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}
	baseURL = NormalizeBaseURL(baseURL)

	if cf == nil {
		cf = NewHTTPClientFactory()
	}

	var opts []httpcli.Opt
	if config.Certificate != "" {
		pool, err := newCertPool(config.Certificate)
		if err != nil {
			return nil, err
		}
		opts = append(opts, httpcli.NewCertPoolOpt(pool))
	}

	cli, err := cf.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	exclude := make(map[string]bool, len(config.Exclude))
	for _, r := range config.Exclude {
		if r.Name != "" {
			exclude[r.Name] = true
		}

		if r.Id != 0 {
			exclude[strconv.Itoa(r.Id)] = true
		}
	}

	return &gitlabConnection{
		config:  config,
		exclude: exclude,
		baseURL: baseURL,
		client:  gitlab.NewClientProvider(baseURL, cli).GetPATClient(config.Token, ""),
	}, nil
}

type gitlabConnection struct {
	config  *schema.GitLabConnection
	exclude map[string]bool
	baseURL *url.URL // URL with path /api/v4 (no trailing slash)
	client  *gitlab.Client
}

// authenticatedRemoteURL returns the GitLab projects's Git remote URL with the configured GitLab personal access
// token inserted in the URL userinfo, for repositories needing authentication.
func (c *gitlabConnection) authenticatedRemoteURL(proj *gitlab.Project) string {
	if c.config.GitURLType == "ssh" {
		return proj.SSHURLToRepo // SSH authentication must be provided out-of-band
	}
	if c.config.Token == "" || !proj.RequiresAuthentication() {
		return proj.HTTPURLToRepo
	}
	u, err := url.Parse(proj.HTTPURLToRepo)
	if err != nil {
		log15.Warn("Error adding authentication to GitLab repository Git remote URL.", "url", proj.HTTPURLToRepo, "error", err)
		return proj.HTTPURLToRepo
	}
	// Any username works; "git" is not special.
	u.User = url.UserPassword("git", c.config.Token)
	return u.String()
}

func (c *gitlabConnection) excludes(p *gitlab.Project) bool {
	return c.exclude[p.PathWithNamespace] || c.exclude[strconv.Itoa(p.ID)]
}

func (c *gitlabConnection) listAllProjects(ctx context.Context) ([]*gitlab.Project, error) {
	type batch struct {
		projs []*gitlab.Project
		err   error
	}

	ch := make(chan batch)

	var wg sync.WaitGroup

	projch := make(chan *schema.GitLabProject)
	for i := 0; i < 5; i++ { // 5 concurrent requests
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range projch {
				proj, err := c.client.GetProject(ctx, gitlab.GetProjectOp{
					ID:                p.Id,
					PathWithNamespace: p.Name,
					CommonOp:          gitlab.CommonOp{NoCache: true},
				})

				if err != nil {
					// TODO(tsenart): When implementing dry-run, reconsider alternatives to return
					// 404 errors on external service config validation.
					if gitlab.IsNotFound(err) {
						log15.Warn("skipping missing gitlab.projects entry:", "name", p.Name, "id", p.Id, "err", err)
						continue
					}
					ch <- batch{err: errors.Wrapf(err, "gitlab.projects: id: %d, name: %q", p.Id, p.Name)}
				} else {
					ch <- batch{projs: []*gitlab.Project{proj}}
				}

				time.Sleep(c.client.RateLimit.RecommendedWaitForBackgroundOp(1))
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(projch)
		for _, p := range c.config.Projects {
			select {
			case projch <- p:
			case <-ctx.Done():
				break
			}
		}
	}()

	for _, projectQuery := range c.config.ProjectQuery {
		if projectQuery == "none" {
			continue
		}

		const perPage = 100
		wg.Add(1)
		go func(projectQuery string) {
			defer wg.Done()

			url, err := projectQueryToURL(projectQuery, perPage) // first page URL
			if err != nil {
				ch <- batch{err: errors.Wrapf(err, "invalid GitLab projectQuery=%q", projectQuery)}
				return
			}

			for {
				if err := ctx.Err(); err != nil {
					ch <- batch{err: err}
					return
				}
				projects, nextPageURL, err := c.client.ListProjects(ctx, url)
				if err != nil {
					ch <- batch{err: errors.Wrapf(err, "error listing GitLab projects: url=%q", url)}
					return
				}
				ch <- batch{projs: projects}
				if nextPageURL == nil {
					return
				}
				url = *nextPageURL

				// 0-duration sleep unless nearing rate limit exhaustion
				time.Sleep(c.client.RateLimit.RecommendedWaitForBackgroundOp(1))
			}
		}(projectQuery)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	seen := make(map[int]bool)
	errs := new(multierror.Error)
	var projects []*gitlab.Project

	for b := range ch {
		if b.err != nil {
			errs = multierror.Append(errs, b.err)
			continue
		}

		for _, proj := range b.projs {
			if !seen[proj.ID] && !c.excludes(proj) {
				projects = append(projects, proj)
				seen[proj.ID] = true
			}
		}
	}

	return projects, errs.ErrorOrNil()
}

var schemeOrHostNotEmptyErr = errors.New("scheme and host should be empty")

func projectQueryToURL(projectQuery string, perPage int) (string, error) {
	// If all we have is the URL query, prepend "projects"
	if strings.HasPrefix(projectQuery, "?") {
		projectQuery = "projects" + projectQuery
	} else if projectQuery == "" {
		projectQuery = "projects"
	}

	u, err := url.Parse(projectQuery)
	if err != nil {
		return "", err
	}
	if u.Scheme != "" || u.Host != "" {
		return "", schemeOrHostNotEmptyErr
	}
	normalizeQuery(u, perPage)

	return u.String(), nil
}

func normalizeQuery(u *url.URL, perPage int) {
	q := u.Query()
	if q.Get("order_by") == "" && q.Get("sort") == "" {
		// Apply default ordering to get the likely more relevant projects first.
		q.Set("order_by", "last_activity_at")
	}
	q.Set("per_page", strconv.Itoa(perPage))
	u.RawQuery = q.Encode()
}
