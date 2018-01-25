package repos

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

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

func createEnableUpdateRepos(ctx context.Context, repoSlice []api.RepoCreateOrUpdateRequest, repoChan <-chan api.RepoCreateOrUpdateRequest) {
	if repoSlice != nil && repoChan != nil {
		panic("unexpected args")
	}

	cloned := 0
	do := func(op api.RepoCreateOrUpdateRequest) {
		createdRepo, err := api.InternalClient.ReposCreateIfNotExists(ctx, op)
		if err != nil {
			log15.Warn("Could not ensure repository exists", "repo", op.RepoURI, "error", err)
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
				err := gitserver.DefaultClient.EnqueueRepoUpdate(ctx, createdRepo.URI)
				if err != nil {
					log15.Warn("Error enqueueing Git clone/update for repository", "repo", op.RepoURI, "error", err)
					return
				}

				// Every 100 repos we clone, wait a bit to prevent overloading gitserver.
				if cloned > 0 && cloned%100 == 0 {
					log15.Info(fmt.Sprintf("%d repositories cloned so far. Waiting for a moment.", cloned))
					time.Sleep(1 * time.Minute)
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
