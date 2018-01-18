package app

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

// githubEnterpriseURLs is a map of GitHub Enterprise hosts to their full URLs.
// This is used for the purposes of generating external GitHub enterprise links.
var githubEnterpriseURLs = make(map[string]string)
var reposListURLs = make(map[string]string)

func init() {
	githubConf := conf.Get().Github
	for _, c := range githubConf {
		gheURL, err := url.Parse(c.Url)
		if err != nil {
			log15.Error("error parsing GitHub config", "error", err)
		}
		githubEnterpriseURLs[gheURL.Host] = strings.TrimSuffix(c.Url, "/")
	}
	reposList := conf.Get().ReposList
	for _, r := range reposList {
		if r.Links != nil && r.Links.Commit != "" {
			reposListURLs[r.Path] = r.Links.Commit
		}
	}
}

// serveRepoExternalCommit resolves a commit for a given repo to a redirect to
// the corresponding commit view link on an external code host.
func serveRepoExternalCommit(w http.ResponseWriter, r *http.Request) error {
	repo, err := handlerutil.GetRepo(r.Context(), mux.Vars(r))
	if err != nil {
		return errors.Wrap(err, "GetRepo")
	}
	commitID := mux.Vars(r)["commit"]

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

	if strings.HasPrefix(repo.URI, "github.com/") {
		http.Redirect(w, r, fmt.Sprintf("https://%s/commit/%s", repo.URI, commitID), http.StatusFound)
		return nil
	}

	host := strings.Split(repo.URI, "/")[0]
	if gheURL, ok := githubEnterpriseURLs[host]; ok {
		http.Redirect(w, r, fmt.Sprintf("%s%s/commit/%s", gheURL, strings.TrimPrefix(repo.URI, host), commitID), http.StatusFound)
		return nil
	}

	w.WriteHeader(http.StatusNotFound)
	return nil
}
