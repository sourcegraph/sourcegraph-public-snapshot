package ui

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/sourcegraph/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/sourcecode"
	"sourcegraph.com/sourcegraph/sourcegraph/ui/payloads"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/htmlutil"
)

func serveTokenSearch(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)
	e := json.NewEncoder(w)

	var opt sourcegraph.TokenSearchOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	opt.RepoRev = sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: mux.Vars(r)["Repo"]},
		Rev:      mux.Vars(r)["Rev"],
	}

	defList := &sourcegraph.DefList{}

	resolvedRev, dataVer, err := handlerutil.ResolveSrclibDataVersion(ctx, sourcegraph.TreeEntrySpec{RepoRev: opt.RepoRev})
	if err == nil {
		opt.RepoRev = resolvedRev

		// Only search if there is a srclib data version (otherwise
		// there will be no token results).
		defList, err = cl.Search.SearchTokens(ctx, &opt)
		if err != nil {
			return err
		}
	} else if err != nil && grpc.Code(err) != codes.NotFound {
		// Continue with no results if not found; otherwise return err.
		return err
	}

	results := make([]payloads.TokenSearchResult, len(defList.Defs))
	for i, def := range defList.Defs {
		def.DocHTML = htmlutil.SanitizeForPB(def.DocHTML.HTML)
		qualifiedName := sourcecode.DefQualifiedNameAndType(def, "dep")
		qualifiedName = sourcecode.OverrideStyleViaRegexpFlags(qualifiedName)
		results[i] = payloads.TokenSearchResult{
			Def:           def,
			QualifiedName: htmlutil.SanitizeForPB(string(qualifiedName)),
			URL:           router.Rel.URLToDef(def.DefKey).String(),
		}
	}

	return e.Encode(&struct {
		Total             int32
		Results           []payloads.TokenSearchResult
		SrclibDataVersion *sourcegraph.SrclibDataVersion
	}{
		Total:             defList.Total,
		Results:           results,
		SrclibDataVersion: dataVer,
	})
}

func serveTextSearch(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)
	e := json.NewEncoder(w)

	var opt sourcegraph.TextSearchOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	_, repoRev, _, err := handlerutil.GetRepoAndRev(ctx, mux.Vars(r))
	if err != nil {
		return err
	}
	opt.RepoRev = repoRev

	vcsEntryList, err := cl.Search.SearchText(ctx, &opt)
	if err != nil {
		return err
	}

	results := make([]payloads.TextSearchResult, len(vcsEntryList.SearchResults))
	for i, vcsEntry := range vcsEntryList.SearchResults {
		results[i] = payloads.TextSearchResult{
			File:      vcsEntry.File,
			StartLine: vcsEntry.StartLine,
			EndLine:   vcsEntry.EndLine,
			Contents:  string(vcsEntry.Match),
		}
	}

	return e.Encode(&struct {
		Total   int32
		Results []payloads.TextSearchResult
	}{
		Total:   vcsEntryList.Total,
		Results: results,
	})
}
