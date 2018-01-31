package repos

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"

	"github.com/pkg/errors"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
)

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

		if op.Enabled {
			// If newly added, the repository will have been set to enabled upon creation above. Explicitly enqueue a
			// clone/update now so that those occur in order of most recently pushed.
			isCloned, err := gitserver.DefaultClient.IsRepoCloned(ctx, createdRepo.URI)
			if err != nil {
				log15.Warn("Error creating/checking local mirror repository for remote source repository", "repo", createdRepo.URI, "error", err)
				return
			}
			if !isCloned {
				cloned++
				log15.Debug("fetching repo", "repo", createdRepo.URI, "cloned", isCloned)
				err := gitserver.DefaultClient.EnqueueRepoUpdate(ctx, gitserver.Repo{Name: createdRepo.URI, URL: op.URL})
				if err != nil {
					log15.Warn("Error enqueueing Git clone/update for repository", "repo", op.RepoURI, "error", err)
					return
				}
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
