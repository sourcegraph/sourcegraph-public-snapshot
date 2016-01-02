package ui

import (
	"encoding/json"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/sourcegraph/mux"

	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sourcecode"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/htmlutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveTokenSearch(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)
	e := json.NewEncoder(w)

	var opt sourcegraph.TokenSearchOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	opt.RepoRev = sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: mux.Vars(r)["Repo"]},
		Rev:      mux.Vars(r)["Rev"],
	}

	defList := &sourcegraph.DefList{}

	resolvedRev, dataVer, err := handlerutil.ResolveSrclibDataVersion(ctx, apiclient, sourcegraph.TreeEntrySpec{RepoRev: opt.RepoRev})
	if err == nil {
		opt.RepoRev = resolvedRev

		// Only search if there is a srclib data version (otherwise
		// there will be no token results).
		defList, err = apiclient.Search.SearchTokens(ctx, &opt)
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
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)
	e := json.NewEncoder(w)

	var opt sourcegraph.TextSearchOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	_, repoRev, _, err := handlerutil.GetRepoAndRev(r, apiclient.Repos)
	if err != nil {
		return err
	}
	opt.RepoRev = repoRev

	vcsEntryList, err := apiclient.Search.SearchText(ctx, &opt)
	if err != nil {
		return err
	}

	results := make([]payloads.TextSearchResult, len(vcsEntryList.SearchResults))
	for i, vcsEntry := range vcsEntryList.SearchResults {
		matchEntryString := string(vcsEntry.Match)
		matchEntryLines := strings.Split(matchEntryString, "\n")

		sanitizedMatchEntryLines := make([]*pbtypes.HTML, len(matchEntryLines))
		for j, line := range matchEntryLines {
			sanitizedMatchEntryLines[j] = htmlutil.SanitizeForPB(line)
		}

		results[i] = payloads.TextSearchResult{
			File:      vcsEntry.File,
			StartLine: vcsEntry.StartLine,
			EndLine:   vcsEntry.EndLine,
			Lines:     sanitizedMatchEntryLines,
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
