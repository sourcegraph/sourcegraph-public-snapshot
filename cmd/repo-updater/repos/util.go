package repos

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/pkg/errors"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
)

// normalizeBaseURL modifies the input and returns a normalized form of the a base URL with insignificant
// differences (such as in presence of a trailing slash, or hostname case) eliminated. Its return value should be
// used for the (ExternalRepoSpec).ServiceID field (and passed to XyzExternalRepoSpec) instead of a non-normalized
// base URL.
func normalizeBaseURL(baseURL *url.URL) *url.URL {
	baseURL.Host = strings.ToLower(baseURL.Host)
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	return baseURL
}

// transportWithCertTrusted returns an http.Transport that trusts the provided PEM cert, or http.DefaultTransport
// if it is empty.
func transportWithCertTrusted(cert string) (http.RoundTripper, error) {
	if cert == "" {
		return http.DefaultTransport, nil
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM([]byte(cert)); !ok {
		return nil, errors.New("invalid certificate value")
	}
	return &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: certPool},
	}, nil
}

type repoCreateOrUpdateRequest struct {
	api.RepoCreateOrUpdateRequest
	URL string // the repository's Git remote URL
}

func createEnableUpdateRepos(ctx context.Context, repoSlice []repoCreateOrUpdateRequest, repoChan <-chan repoCreateOrUpdateRequest) {
	if repoSlice != nil && repoChan != nil {
		panic("unexpected args")
	}

	cloned := 0
	do := func(op repoCreateOrUpdateRequest) {
		createdRepo, err := api.InternalClient.ReposCreateIfNotExists(ctx, op.RepoCreateOrUpdateRequest)
		if err != nil {
			log15.Warn("Error creating or updating repository", "repo", op.RepoURI, "error", err)
			return
		}

		if createdRepo.Enabled {
			// If newly added, the repository will have been set to enabled upon creation above. Explicitly enqueue a
			// clone/update now so that those occur in order of most recently pushed.
			isCloned, err := gitserver.DefaultClient.IsRepoCloned(ctx, createdRepo.URI)
			if err != nil {
				log15.Warn("Error creating/checking local mirror repository for remote source repository", "repo", createdRepo.URI, "error", err)
				return
			}
			if !isCloned {
				cloned++
			}
			log15.Debug("fetching repo", "repo", createdRepo.URI, "cloned", isCloned)
			err = gitserver.DefaultClient.EnqueueRepoUpdate(ctx, gitserver.Repo{Name: createdRepo.URI, URL: op.URL})
			if err != nil {
				log15.Warn("Error enqueueing Git clone/update for repository", "repo", op.RepoURI, "error", err)
				return
			}
		}
	}

	for _, repo := range repoSlice {
		do(repo)
	}
	for repo := range repoChan {
		do(repo)
	}
}

// addPasswordBestEffort adds the password to rawurl if the user is
// specified. If anything fails, the original rawurl is returned.
func addPasswordBestEffort(rawurl, password string) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		return rawurl
	}
	if u.User == nil || u.User.Username() == "" {
		return rawurl
	}
	u.User = url.UserPassword(u.User.Username(), password)
	return u.String()
}

// atomicValue manages an atomic value.
type atomicValue struct {
	mu    sync.RWMutex
	value interface{}
}

// get returns the current value.
func (a *atomicValue) get() interface{} {
	a.mu.RLock()
	v := a.value
	a.mu.RUnlock()
	return v
}

// set changes the value to the result of f. The mutex is held for the entire
// duration of the function call.
func (a *atomicValue) set(f func() interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.value = f()
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
