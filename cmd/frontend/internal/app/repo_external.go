package app

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/atomicvalue"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
)

var reposListURLs = atomicvalue.New()

func init() {
	conf.Watch(func() {
		reposListURLs.Set(func() interface{} {
			urls := make(map[api.RepoURI]string)
			for _, r := range conf.Get().ReposList {
				if r.Links != nil && r.Links.Commit != "" {
					urls[api.RepoURI(r.Path)] = r.Links.Commit
				}
			}
			return urls
		})
	})
}

// serveRepoExternalCommit resolves a commit for a given repo to a redirect to
// the corresponding commit view link on an external code host.
func serveRepoExternalCommit(w http.ResponseWriter, r *http.Request) error {
	repo, err := handlerutil.GetRepo(r.Context(), mux.Vars(r))
	if err != nil {
		return errors.Wrap(err, "GetRepo")
	}
	commitID := mux.Vars(r)["commit"]

	reposListURLs := reposListURLs.Get().(map[api.RepoURI]string)
	if commitURL, ok := reposListURLs[repo.URI]; ok {
		url := strings.Replace(commitURL, "{commit}", commitID, 1)
		http.Redirect(w, r, url, http.StatusFound)
		return nil
	}

	// Check to see if there's a Phabricator entry for this repo. If so, link to Phabricator's commit view first.
	phabRepo, _ := db.Phabricator.GetByURI(r.Context(), repo.URI)
	if phabRepo != nil {
		http.Redirect(w, r, fmt.Sprintf("%s/r%s%s", phabRepo.URL, phabRepo.Callsign, commitID), http.StatusFound)
		return nil
	}

	if strings.HasPrefix(string(repo.URI), "github.com/") {
		http.Redirect(w, r, fmt.Sprintf("https://%s/commit/%s", repo.URI, commitID), http.StatusFound)
		return nil
	}

	host := strings.Split(string(repo.URI), "/")[0]
	if gheURL, ok := conf.GitHubEnterpriseURLs()[host]; ok {
		http.Redirect(w, r, fmt.Sprintf("%s%s/commit/%s", gheURL, strings.TrimPrefix(string(repo.URI), host), commitID), http.StatusFound)
		return nil
	}

	if repo.ExternalRepo != nil {
		info, err := repoupdater.DefaultClient.RepoLookup(r.Context(), protocol.RepoLookupArgs{ExternalRepo: repo.ExternalRepo})
		if err != nil {
			return err
		}
		if info.Repo != nil && info.Repo.Links != nil && info.Repo.Links.Commit != "" {
			url := strings.Replace(info.Repo.Links.Commit, "{commit}", commitID, -1)
			http.Redirect(w, r, url, http.StatusFound)
			return nil
		}
	}

	w.WriteHeader(http.StatusNotFound)
	return nil
}
