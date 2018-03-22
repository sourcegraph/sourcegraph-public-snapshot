package repos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/externalservice/bitbucketserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/atomicvalue"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// bitbucketServerServiceType is the (api.ExternalRepoSpec).ServiceType value
// for Bitbucket Server projects. The ServiceID value is the base URL to the
// Bitbucket Server instance.
const bitbucketServerServiceType = "bitbucketServer"

var bitbucketServerConnections = atomicvalue.New()

func init() {
	conf.Watch(func() {
		bitbucketServerConnections.Set(func() interface{} {
			var conns []*bitbucketServerConnection
			for _, c := range conf.Get().BitbucketServer {
				conn, err := newBitbucketServerConnection(&c)
				if err != nil {
					log15.Error("Error processing configured Bitbucket Server connection. Skipping it.", "url", c.Url, "error", err)
					continue
				}
				conns = append(conns, conn)
			}
			return conns
		})
		bitbucketServerWorker.restart()
	})
}

// getBitbucketServerConnection returns the BitbucketServer connection (config + API client) that is responsible for
// the repository specified by the args.
func getBitbucketServerConnection(args protocol.RepoLookupArgs) (*bitbucketServerConnection, error) {
	conns := bitbucketServerConnections.Get().([]*bitbucketServerConnection)

	if args.ExternalRepo != nil && args.ExternalRepo.ServiceType == bitbucketServerServiceType {
		// Look up by external repository spec.
		for _, conn := range conns {
			if args.ExternalRepo.ServiceID == conn.client.URL.String() {
				return conn, nil
			}
		}
		return nil, errors.Errorf("no configured Bitbucket Server connection with URL: %q", args.ExternalRepo.ServiceID)
	}

	if args.Repo != "" {
		// Look up by repository URI.
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
	host = normalizeBaseURL(host)

	// URI
	repositoryPathPattern := config.RepositoryPathPattern
	if repositoryPathPattern == "" {
		repositoryPathPattern = "{host}/{projectKey}/{repositorySlug}"
	}
	project := "UNKNOWN"
	if repo.Project != nil {
		project = repo.Project.Key
	}
	repoURI := api.RepoURI(strings.NewReplacer(
		"{host}", host.Hostname(),
		"{projectKey}", project,
		"{repositorySlug}", repo.Slug,
	).Replace(repositoryPathPattern))

	// Clone URL
	var cloneURL string
	for _, l := range repo.Links.Clone {
		if l.Name == "ssh" && config.GitURLType == "ssh" {
			cloneURL = l.Href
			break
		}
		if l.Name == "http" {
			// l.Href already contains the username in the URL userinfo, so just add the token or
			// password.
			var password string
			if config.Token != "" {
				password = config.Token // prefer personal access token
			} else {
				password = config.Password
			}
			cloneURL = addPasswordBestEffort(l.Href, password)
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
		URI: repoURI,
		ExternalRepo: &api.ExternalRepoSpec{
			ID:          project + "/" + repo.Slug,
			ServiceType: bitbucketServerServiceType,
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

	if args.ExternalRepo != nil && args.ExternalRepo.ServiceType == bitbucketServerServiceType {
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
		return bitbucketServerRepoInfo(conn.config, repo), true, nil
	}

	if args.Repo != "" {
		// Look up by repository URI. Expect suffix {projectKey}/{repoSlug}
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
		return bitbucketServerRepoInfo(conn.config, repo), true, nil
	}

	return nil, true, fmt.Errorf("unable to look up Bitbucket Server repository (%+v)", args)
}

var bitbucketServerWorker = &worker{
	work: func(ctx context.Context, shutdown chan struct{}) {
		for _, c := range bitbucketServerConnections.Get().([]*bitbucketServerConnection) {
			go func(c *bitbucketServerConnection) {
				for {
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
	go createEnableUpdateRepos(ctx, nil, repoChan)
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
				RepoURI:      ri.URI,
				ExternalRepo: ri.ExternalRepo,
				Description:  ri.Description,
				Fork:         ri.Fork,
				Enabled:      conn.config.InitialRepositoryEnablement,
			},
			URL: ri.VCS.URL,
		}
	}
}

func newBitbucketServerConnection(config *schema.BitbucketServerConnection) (*bitbucketServerConnection, error) {
	baseURL, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}
	baseURL = normalizeBaseURL(baseURL)

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
				Transport: transport,
			},
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
				if !recent[r.ID] {
					ch <- r
				}
			}
		}
	}()

	return ch
}
