package app

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveRepoBadges(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	rc, err := handlerutil.GetRepoCommon(r, nil)
	if err != nil {
		return err
	}

	repoSpec := rc.Repo.RepoSpec()
	badges, err := apiclient.RepoBadges.ListBadges(ctx, &repoSpec)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "repo/badges.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		Badges []*sourcegraph.Badge

		tmpl.Common
	}{
		RepoCommon: *rc,
		Badges:     badges.Badges,
	})
}

func serveRepoCounters(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	rc, err := handlerutil.GetRepoCommon(r, nil)
	if err != nil {
		return err
	}

	repoSpec := rc.Repo.RepoSpec()
	counters, err := apiclient.RepoBadges.ListCounters(ctx, &repoSpec)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "repo/counters.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		Counters []*sourcegraph.Counter

		tmpl.Common
	}{
		RepoCommon: *rc,
		Counters:   counters.Counters,
	})
}

func serveRepoBadge(w http.ResponseWriter, r *http.Request) error {
	s := handlerutil.APIClient(r)

	_, _, _, err := handlerutil.GetRepoAndRev(r, s.Repos)
	if err != nil {
		return err
	}

	vars := mux.Vars(r)
	badge := vars["Badge"]
	format := vars["Format"]

	var subject, status string

	switch badge {
	case "docs-examples":
		subject = "Usage examples"
	case "dependencies":
		subject = "Dependencies"
		// TODO(sqs): calculate
	case "dependents":
		subject = "Dependents"
	case "xrefs":
		subject = "xrefs"
	case "funcs":
		subject = "# funcs"
	case "top-func":
		subject = "Top func"
		status = "n/a"
	case "library-users":
		subject = "Users"
		status = "n/a"
	case "authors":
		subject = "Authors"
		status = "n/a"
	case "status":
		// TODO(new-arch): implement this status check when we re-add repository
		// builds and queues in new-arch. This dummy implementation here is just
		// so we can maintain the existing code that generates the badge text
		// and color.
		subject = "Sourcegraph"
		status = "Status"
	default:
		return &errcode.HTTPErr{http.StatusNotFound, errors.New("bad badge name")}
	}

	if status == "" {
		status = "Sourcegraph"
	}
	color := "blue"

	w.Header().Add("cache-control", "max-age=120")
	http.Redirect(w, r, shieldsURL(subject, status, color, format), http.StatusSeeOther)

	return nil
}

// serveRepoCounter wraps doRepoCounter and skips tracking the API
// call if it was initiated from the repo stats/counters page (beacuse
// we don't want to increment the counters when someone is just
// looking at the counter images on the stats/counters page).
func serveRepoCounter(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	repoSpec, err := sourcegraph.UnmarshalRepoSpec(mux.Vars(r))
	if err != nil {
		return err
	}

	if _, err := s.Repos.Get(ctx, &repoSpec); err != nil {
		return err
	}

	if _, noRecord := r.URL.Query()["no-record"]; !noRecord {
		if _, err := s.RepoBadges.RecordHit(ctx, &repoSpec); err != nil {
			return err
		}
	}

	return doServeRepoCounter(w, r, repoSpec)
}

func doServeRepoCounter(w http.ResponseWriter, r *http.Request, repo sourcegraph.RepoSpec) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	v := mux.Vars(r)
	counter := v["Counter"]
	format := v["Format"]

	var subject string
	var since *pbtypes.Timestamp

	switch counter {
	case "views-24h":
		subject = "views today"
		ts := pbtypes.NewTimestamp(time.Now().Add(-1 * 24 * time.Hour).In(time.UTC))
		since = &ts
	case "views":
		subject = "views"
	default:
		return &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("bad counter name")}
	}

	count, err := s.RepoBadges.CountHits(ctx, &sourcegraph.RepoBadgesCountHitsOp{
		Repo:  repo,
		Since: since,
	})
	if err != nil {
		return err
	}

	w.Header().Add("Cache-Control", "max-age=0, private, no-cache, must-revalidate")
	w.Header().Add("X-Proxy-No-Cache", "1") // don't proxy-cache counter or else it won't count everyone

	http.Redirect(w, r, shieldsURL(subject, fmt.Sprint(count.Hits), "blue", format), http.StatusSeeOther)
	return nil
}

func shieldsURL(subject, status, color, format string) string {
	return fmt.Sprintf("https://img.shields.io/badge/%s-%s-%s.%s", urlPathEscape(subject), urlPathEscape(status), urlPathEscape(color), urlPathEscape(format))
}

func urlPathEscape(s string) string {
	return strings.Replace(s, "/", "%2F", -1)
}
