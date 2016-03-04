package app

import (
	"net/http"
	"net/url"

	"github.com/sourcegraph/mux"

	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

func serveDefShare(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)
	dc, rc, vc, err := handlerutil.GetDefCommon(ctx, mux.Vars(r), nil)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "def/share.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		payloads.DefCommon
		tmpl.Common
	}{
		RepoCommon:    *rc,
		RepoRevCommon: *vc,
		DefCommon:     *dc,
	})
}

func serveRepoTreeShare(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)
	tc, rc, vc, err := handlerutil.GetTreeEntryCommon(ctx, mux.Vars(r), nil)
	if err != nil {
		return err
	}

	sourceboxURL := router.Rel.URLToSourceboxFile(tc.EntrySpec, "js")
	sourceboxURL.RawQuery = r.URL.RawQuery
	sourceboxJSONURL := router.Rel.URLToSourceboxFile(tc.EntrySpec, "json")
	sourceboxJSONURL.RawQuery = r.URL.RawQuery

	return tmpl.Exec(r, w, "repo/tree/share.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		handlerutil.TreeEntryCommon
		SourceboxURL     *url.URL
		SourceboxJSONURL *url.URL
		tmpl.Common
	}{
		RepoCommon:       *rc,
		RepoRevCommon:    *vc,
		TreeEntryCommon:  *tc,
		SourceboxURL:     sourceboxURL,
		SourceboxJSONURL: sourceboxJSONURL,
	})
}
