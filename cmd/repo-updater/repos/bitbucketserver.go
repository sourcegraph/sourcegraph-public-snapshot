package repos

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
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
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var bitbucketServerConnections = func() *atomicvalue.Value {
	c := atomicvalue.New()
	c.Set(func() interface{} {
		return []*bitbucketServerConnection{}
	})
	return c
}()

// SyncBitbucketServerConnections periodically syncs connections from
// the Frontend API.
func SyncBitbucketServerConnections(ctx context.Context) {
	t := time.NewTicker(configWatchInterval)
	var lastConfig []*schema.BitbucketServerConnection
	for range t.C {
		config, err := conf.BitbucketServerConfigs(ctx)
		if err != nil {
			log15.Error("unable to fetch Bitbucket Server configs", "err", err)
			continue
		}

		if reflect.DeepEqual(config, lastConfig) {
			continue
		}
		lastConfig = config

		var conns []*bitbucketServerConnection
		for _, c := range config {
			conn, err := newBitbucketServerConnection(c, nil)
			if err != nil {
				log15.Error("Error processing configured Bitbucket Server connection. Skipping it.", "url", c.Url, "error", err)
				continue
			}
			conns = append(conns, conn)
		}

		bitbucketServerConnections.Set(func() interface{} {
			return conns
		})

		bitbucketServerWorker.restart()
	}
}

// getBitbucketServerConnection returns the BitbucketServer connection (config + API client) that is responsible for
// the repository specified by the args.
func getBitbucketServerConnection(args protocol.RepoLookupArgs) (*bitbucketServerConnection, error) {
	conns := bitbucketServerConnections.Get().([]*bitbucketServerConnection)

	if args.ExternalRepo != nil && args.ExternalRepo.ServiceType == bitbucketserver.ServiceType {
		// Look up by external repository spec.
		for _, conn := range conns {
			if args.ExternalRepo.ServiceID == conn.client.URL.String() {
				return conn, nil
			}
		}
		return nil, errors.Errorf("no configured Bitbucket Server connection with URL: %q", args.ExternalRepo.ServiceID)
	}

	if args.Repo != "" {
		// Look up by repository name.
		repo := strings.ToLower(string(args.Repo))
		for _, conn := range conns {
			// TODO should this be based on RepositoryPathPattern?
			if strings.HasPrefix(repo, conn.client.URL.Hostname()+"/") {
				return conn, nil
			}
		}
	}

	return nil, nil
}

func bitbucketServerRepoInfo(config *schema.BitbucketServerConnection, repo *bitbucketserver.Repo) *protocol.RepoInfo {
	host, err := url.Parse(config.Url)
	if err != nil {
		// This should never happen
		log15.Error("malformed bitbucket config, invalid URL", "url", config.Url, "error", err)
		return nil
	}
	host = NormalizeBaseURL(host)

	// Name
	project := "UNKNOWN"
	if repo.Project != nil {
		project = repo.Project.Key
	}
	repoName := reposource.BitbucketServerRepoName(config.RepositoryPathPattern, host.Hostname(), project, repo.Slug)

	// Clone URL
	var cloneURL string
	for _, l := range repo.Links.Clone {
		if l.Name == "ssh" && config.GitURLType == "ssh" {
			cloneURL = l.Href
			break
		}
		if l.Name == "http" {
			var password string
			if config.Token != "" {
				password = config.Token // prefer personal access token
			} else {
				password = config.Password
			}
			cloneURL = setUserinfoBestEffort(l.Href, config.Username, password)
			// No break, so that we fallback to http in case of ssh missing
			// with GitURLType == "ssh"
		}
	}

	// Repo Links
	var links *protocol.RepoLinks
	for _, l := range repo.Links.Self {
		root := strings.TrimSuffix(l.Href, "/browse")
		links = &protocol.RepoLinks{
			Root:   l.Href,
			Tree:   root + "/browse/{path}?at={rev}",
			Blob:   root + "/browse/{path}?at={rev}",
			Commit: root + "/commits/{commit}",
		}
		break
	}

	return &protocol.RepoInfo{
		Name: repoName,
		ExternalRepo: &api.ExternalRepoSpec{
			ID:          project + "/" + repo.Slug,
			ServiceType: bitbucketserver.ServiceType,
			ServiceID:   host.String(),
		},
		Description: repo.Name,
		Fork:        repo.Origin != nil,
		VCS: protocol.VCSInfo{
			URL: cloneURL,
		},
		Links: links,
	}
}

// bitbucketServerRepoInfoSuffix matches out {projectKey}/{repoSlug} at the
// end of a string.
var bitbucketServerRepoInfoSuffix = regexp.MustCompile(`([^/]+)/([^/]+)/?$`)

// GetBitbucketServerRepository queries a configured BitbucketServer connection endpoint for information about the
// specified repository (a.k.a. project in BitbucketServer's naming scheme).
//
// If args.Repo refers to a repository that is not known to be on a configured BitbucketServer connection's
// host, it returns authoritative == false.
func GetBitbucketServerRepository(ctx context.Context, args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error) {
	conn, err := getBitbucketServerConnection(args)
	if err != nil {
		return nil, true, err // refers to a BitbucketServer repo but the host is not configured
	}
	if conn == nil {
		return nil, false, nil // refers to a non-BitbucketServer repo
	}

	if args.ExternalRepo != nil && args.ExternalRepo.ServiceType == bitbucketserver.ServiceType {
		// Look up by external repository spec. Expect {projectKey}/{repoSlug}
		i := strings.Index(args.ExternalRepo.ID, "/")
		if i < 0 || i == len(args.ExternalRepo.ID)-1 {
			return nil, true, errors.Errorf("malformed bitbucket server ID: %q", args.ExternalRepo.ID)
		}
		projectKey, repoSlug := args.ExternalRepo.ID[:i], args.ExternalRepo.ID[i+1:]
		repo, err := conn.client.Repo(ctx, projectKey, repoSlug)
		if err != nil {
			return nil, true, err
		}
		if conn.config.ExcludePersonalRepositories && repo.IsPersonalRepository() {
			return nil, true, &vcs.RepoNotExistError{Repo: api.RepoName(repoSlug)}
		}
		return bitbucketServerRepoInfo(conn.config, repo), true, nil
	}

	if args.Repo != "" {
		// Look up by repository name. Expect suffix {projectKey}/{repoSlug}
		// TODO shouldn't we use RepositoryPathPattern?
		match := bitbucketServerRepoInfoSuffix.FindStringSubmatch(string(args.Repo))
		if len(match) == 0 {
			return nil, true, errors.Errorf("malformed bitbucket server repo URL: %q", args.Repo)
		}
		projectKey, repoSlug := match[0], match[1]
		repo, err := conn.client.Repo(ctx, projectKey, repoSlug)
		if err != nil {
			return nil, true, err
		}
		if conn.config.ExcludePersonalRepositories && repo.IsPersonalRepository() {
			return nil, true, &vcs.RepoNotExistError{Repo: args.Repo}
		}
		return bitbucketServerRepoInfo(conn.config, repo), true, nil
	}

	return nil, true, fmt.Errorf("unable to look up Bitbucket Server repository (%+v)", args)
}

var bitbucketServerWorker = &worker{
	work: func(ctx context.Context, shutdown chan struct{}) {
		for _, c := range bitbucketServerConnections.Get().([]*bitbucketServerConnection) {
			go func(c *bitbucketServerConnection) {
				for {
					reservationTime := time.Now()
					r := c.client.RateLimit.ReserveN(reservationTime, rateLimitReservationSize)
					if !r.OK() {
						log15.Error("Bitbucket worker cannot reserve requests. Is the maximum burst size lower than the reservation size?", "reservation_size", rateLimitReservationSize, "max_burst_size", bitbucketserver.RateLimitMaxBurstRequests)
					}
					delay := r.Delay()
					// Since we're not actually planning to use the reservation, cancel it now.
					// We only wanted to know the delay / availability of the reservation.
					r.CancelAt(reservationTime)
					if delay > time.Second {
						log15.Warn("Bitbucket self-enforced API rate limit is almost exhausted. Waiting before doing more work", "delay", r.Delay())
					}
					time.Sleep(delay)

					updateBitbucketServerRepos(ctx, c)
					bitbucketServerUpdateTime.WithLabelValues(c.config.Url).Set(float64(time.Now().Unix()))
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

// RunBitbucketServerRepositorySyncWorker runs the worker that syncs projects from configured BitbucketServer instances to
// Sourcegraph.
func RunBitbucketServerRepositorySyncWorker(ctx context.Context) {
	bitbucketServerWorker.start(ctx)
}

// updateBitbucketServerRepos ensures that all provided repositories exist in the repository table.
func updateBitbucketServerRepos(ctx context.Context, conn *bitbucketServerConnection) {
	repos, err := conn.listAllRepos(ctx)
	if err != nil {
		log15.Error("failed to list some bitbucketserver repos", "error", err.Error())
	}

	repoChan := make(chan repoCreateOrUpdateRequest)
	defer close(repoChan)

	sourceID := conn.config.Token
	if sourceID == "" {
		sourceID = conn.config.Username
	}

	go createEnableUpdateRepos(ctx, fmt.Sprintf("bitbucket:%s", sourceID), repoChan)

	for _, r := range repos {
		ri := bitbucketServerRepoInfo(conn.config, r)
		if ri.VCS.URL == "" {
			continue
		}

		repoChan <- repoCreateOrUpdateRequest{
			RepoCreateOrUpdateRequest: api.RepoCreateOrUpdateRequest{
				RepoName:     ri.Name,
				ExternalRepo: ri.ExternalRepo,
				Description:  ri.Description,
				Fork:         ri.Fork,
				Enabled:      conn.config.InitialRepositoryEnablement,
			},
			URL: ri.VCS.URL,
		}
	}
}

// rateLimitReservationSize is the minimum number of requests (tokens in the bucket)
// that must be available for us to perform work without waiting first.
//
// We choose this number because each reservation lets us list 100
// repositories, so having at least 250 lets us list 20,000 repositories and
// still have 50 API requests left-over to serve users.
const rateLimitReservationSize = 250

func newBitbucketServerConnection(config *schema.BitbucketServerConnection, cf httpcli.Factory) (*bitbucketServerConnection, error) {
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
			exclude[strings.ToLower(r.Name)] = true
		}

		if r.Id != 0 {
			exclude[strconv.Itoa(r.Id)] = true
		}
	}

	client := bitbucketserver.NewClient(baseURL, cli)
	client.Token = config.Token
	client.Username = config.Username
	client.Password = config.Password

	return &bitbucketServerConnection{
		config:  config,
		exclude: exclude,
		client:  client,
	}, nil
}

type bitbucketServerConnection struct {
	config  *schema.BitbucketServerConnection
	exclude map[string]bool
	client  *bitbucketserver.Client
}

func (c *bitbucketServerConnection) excludes(r *bitbucketserver.Repo) bool {
	name := r.Slug
	if r.Project != nil {
		name = r.Project.Key + "/" + name
	}
	return c.exclude[strings.ToLower(name)] ||
		c.exclude[strconv.Itoa(r.ID)] ||
		c.config.ExcludePersonalRepositories && r.IsPersonalRepository()
}

func (c *bitbucketServerConnection) listAllRepos(ctx context.Context) ([]*bitbucketserver.Repo, error) {
	type batch struct {
		repos []*bitbucketserver.Repo
		err   error
	}

	ch := make(chan batch)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		repos := make([]*bitbucketserver.Repo, 0, len(c.config.Repos))
		errs := new(multierror.Error)

		for _, r := range c.config.Repos {
			ps := strings.SplitN(r.Name, "/", 2)
			if len(ps) != 2 {
				errs = multierror.Append(errs,
					errors.Errorf("bitbucketserver.repos: name=%q", r.Name))
				continue
			}

			projectKey, repoSlug := ps[0], ps[1]
			repo, err := c.client.Repo(ctx, projectKey, repoSlug)
			if err != nil {
				// TODO(tsenart): When implementing dry-run, reconsider alternatives to return
				// 404 errors on external service config validation.
				if bitbucketserver.IsNotFound(err) {
					log15.Warn("skipping missing bitbucketserver.repos entry:", "name", r.Name, "id", fmt.Sprint(r.Id), "err", err)
					continue
				}
				errs = multierror.Append(errs,
					errors.Wrapf(err, "bitbucketserver.repos: id: %d, name: %q", r.Id, r.Name))
			} else {
				repos = append(repos, repo)
			}
		}

		ch <- batch{repos: repos, err: errs.ErrorOrNil()}
	}()

	for _, q := range c.config.RepositoryQuery {
		if q == "none" {
			continue
		}

		wg.Add(1)
		go func(q string) {
			defer wg.Done()

			page := bitbucketserver.PageToken{Limit: 100}
			for page.HasMore() {
				repos, page, err := c.client.Repos(ctx, &page, q)
				if err != nil {
					ch <- batch{err: errors.Wrapf(err, "bibucketserver.repositoryQuery: item=%q, page=%+v", q, page)}
					break
				}
				ch <- batch{repos: repos}
			}
		}(q)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	seen := make(map[int]bool)
	errs := new(multierror.Error)
	var repos []*bitbucketserver.Repo

	for r := range ch {
		if r.err != nil {
			errs = multierror.Append(errs, r.err)
			continue
		}

		for _, repo := range r.repos {
			if !seen[repo.ID] && !c.excludes(repo) && repo.State == "AVAILABLE" {
				repos = append(repos, repo)
				seen[repo.ID] = true
			}
		}
	}

	return repos, errs.ErrorOrNil()
}
