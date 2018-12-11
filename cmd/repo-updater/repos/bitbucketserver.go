package repos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/atomicvalue"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/conf/reposource"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/time/rate"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var bitbucketServerConnections = atomicvalue.New()

func init() {
	bitbucketServerConnections.Set(func() interface{} {
		return []*bitbucketServerConnection{}
	})

	go func() {
		t := time.NewTicker(configWatchInterval)
		var lastConfig []*schema.BitbucketServerConnection
		for range t.C {
			config, err := conf.BitbucketServerConfigs(context.Background())
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
				conn, err := newBitbucketServerConnection(c)
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
	}()
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
						log15.Error("Bitbucket worker cannot reserve requests. Is the maximum burst size lower than the reservation size?", "reservation_size", rateLimitReservationSize, "max_burst_size", rateLimitMaxBurstRequests)
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
					case <-time.After(getUpdateInterval()):
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
	repoChan := make(chan repoCreateOrUpdateRequest)
	defer close(repoChan)
	sourceID := conn.config.Token
	if sourceID == "" {
		sourceID = conn.config.Username
	}
	go createEnableUpdateRepos(ctx, fmt.Sprintf("bitbucket:%s", sourceID), repoChan)
	for r := range conn.listAllRepos(ctx) {
		if r.State != "AVAILABLE" {
			continue
		}

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

// These fields define the self-imposed Bitbucket rate limit (since Bitbucket Server does
// not have a concept of rate limiting in HTTP response headers).
//
// See https://godoc.org/golang.org/x/time/rate#Limiter for an explanation of these fields.
//
// The limits chosen here are based on the following logic: Bitbucket Cloud restricts
// "List all repositories" requests (which are a good portion of our requests) to 1,000/hr,
// and they restrict "List a user or team's repositories" requests (which are roughly equal
// to our repository lookup requests) to 1,000/hr. We peform a list repositories request
// for every 100 repositories on Bitbucket every 1m by default, so for someone with 20,000
// Bitbucket repositories we need 20,000/100 requests per minute (1200/hr) + overhead for
// repository lookup requests by users. So we use a generous 7,200/hr here until we hear
// from someone that these values do not work well for them.
const (
	rateLimitRequestsPerSecond = 2 // 120/min or 7200/hr
	rateLimitMaxBurstRequests  = 500

	// rateLimitReservationSize is the minimum number of requests (tokens in the bucket)
	// that must be available for us to perform work without waiting first.
	//
	// We choose this number because each reservation lets us list 100 repositories, so
	// having at least 250 lets us list 20,000 repositories and still have 50 API requests
	// left-over to serve users.
	rateLimitReservationSize = 250
)

func newBitbucketServerConnection(config *schema.BitbucketServerConnection) (*bitbucketServerConnection, error) {
	baseURL, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}
	baseURL = NormalizeBaseURL(baseURL)

	transport, err := cachedTransportWithCertTrusted(config.Certificate)
	if err != nil {
		return nil, err
	}

	return &bitbucketServerConnection{
		config: config,
		client: &bitbucketserver.Client{
			URL:      baseURL,
			Token:    config.Token,
			Username: config.Username,
			Password: config.Password,
			HTTPClient: &http.Client{
				Transport: bitbucketserver.WithRequestCounter(transport),
			},
			RateLimit: rate.NewLimiter(rateLimitRequestsPerSecond, rateLimitMaxBurstRequests),
		},
	}, nil
}

type bitbucketServerConnection struct {
	config *schema.BitbucketServerConnection
	client *bitbucketserver.Client
}

func (c *bitbucketServerConnection) listAllRepos(ctx context.Context) <-chan *bitbucketserver.Repo {
	perPage := 100
	ch := make(chan *bitbucketserver.Repo, perPage)
	go func() {
		defer close(ch)

		// First we list one page of recent repos, so that we clone them first
		repos, _, err := c.client.RecentRepos(ctx, &bitbucketserver.PageToken{Limit: perPage})
		if err != nil {
			log15.Warn("failed to list recent repos for Bitbucket Server", "url", c.client.URL, "error", err)
		}
		recent := map[int]bool{}
		for _, r := range repos {
			if c.config.ExcludePersonalRepositories && r.IsPersonalRepository() {
				continue
			}
			recent[r.ID] = true
			ch <- r
		}

		// Then we list all repos, taking care not to repeat repos we have
		// already sent via recent.
		page := &bitbucketserver.PageToken{Limit: perPage}
		for page.HasMore() {
			repos, page, err = c.client.Repos(ctx, page)
			if err != nil {
				log15.Error("failed when listing Bitbucket Server repos", "url", c.client.URL, "error", err)
				return
			}
			for _, r := range repos {
				if c.config.ExcludePersonalRepositories && r.IsPersonalRepository() {
					continue
				}
				if !recent[r.ID] {
					ch <- r
				}
			}
		}
	}()

	return ch
}
