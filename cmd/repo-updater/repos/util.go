package repos

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"strings"

	"github.com/gregjones/httpcache"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/httputil"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// NormalizeBaseURL modifies the input and returns a normalized form of the a base URL with insignificant
// differences (such as in presence of a trailing slash, or hostname case) eliminated. Its return value should be
// used for the (ExternalRepoSpec).ServiceID field (and passed to XyzExternalRepoSpec) instead of a non-normalized
// base URL.
func NormalizeBaseURL(baseURL *url.URL) *url.URL {
	baseURL.Host = strings.ToLower(baseURL.Host)
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	return baseURL
}

// cachedTransportWithCertTrusted returns an http.Transport that trusts the
// provided PEM cert, or http.DefaultTransport if it is empty. The transport
// is also using our redis backed cache.
func cachedTransportWithCertTrusted(cert string) (http.RoundTripper, error) {
	transport := http.DefaultTransport
	if cert != "" {
		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM([]byte(cert)); !ok {
			return nil, errors.New("invalid certificate value")
		}
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: certPool},
		}
	}

	return &httpcache.Transport{
		Transport:           &nethttp.Transport{RoundTripper: transport},
		Cache:               httputil.Cache,
		MarkCachedResponses: true, // so we avoid using cached rate limit info
	}, nil
}

// A repoCreateOrUpdateRequest is a RepoCreateOrUpdateRequest, from the API,
// plus a specific URL we'd like to use for it.
type repoCreateOrUpdateRequest struct {
	api.RepoCreateOrUpdateRequest
	URL string // the repository's Git remote URL
}

// createEnableUpdateRepos receives requests on the provided channel. The
// source argument should be a distinctive string identifying the configuration
// being updated, so repo-updater can detect when repositories are dropped from
// a given source.
func createEnableUpdateRepos(ctx context.Context, source string, repoChan <-chan repoCreateOrUpdateRequest) {
	newList := make(sourceRepoList)

	do := func(op repoCreateOrUpdateRequest) {
		if op.RepoCreateOrUpdateRequest.RepoURI == "" {
			log15.Warn("ignoring invalid request to create or enable repo with empty name", "source", source, "repo", op.RepoCreateOrUpdateRequest.ExternalRepo)
			return
		}
		createdRepo, err := api.InternalClient.ReposCreateIfNotExists(ctx, op.RepoCreateOrUpdateRequest)
		if err != nil {
			log15.Warn("Error creating or updating repository", "repo", op.RepoURI, "error", err)
			return
		}

		err = api.InternalClient.ReposUpdateMetadata(ctx, op.RepoURI, op.Description, op.Fork, op.Archived)
		if err != nil {
			log15.Warn("Error updating repository metadata", "repo", op.RepoURI, "error", err)
			return
		}

		newList[string(createdRepo.URI)] = configuredRepo{url: op.URL, enabled: createdRepo.Enabled}
	}
	for repo := range repoChan {
		do(repo)
	}
	repos.updateSource(source, newList)
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

// worker represents a worker that does work under some context and can be restarted.
type worker struct {
	// work is invoked to perform work under the given context. It should
	// stop and return when the given shutdown channel is closed.
	work func(ctx context.Context, shutdown chan struct{})

	shutdown chan struct{}
	context  context.Context
}

// restart restarts the worker. It only does so if the worker was previously
// started.
func (w *worker) restart() {
	if w.shutdown == nil {
		return // not yet started
	}

	// Shutdown the previously started workers.
	close(w.shutdown)

	// Note for the weary traveller: We do not wait for workers to stop, which
	// could lead to workers doing duplicate work. I (Keegan) have a sneaky
	// feeling for large installations this could be an issue.

	// Start the new workers.
	w.start(w.context)
}

// start starts the worker with the given context. The work is done in a
// separate goroutine.
func (w *worker) start(ctx context.Context) {
	shutdown := make(chan struct{})
	w.shutdown = shutdown
	w.context = ctx
	go w.work(ctx, shutdown)
}
