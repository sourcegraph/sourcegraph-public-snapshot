package ui

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sourcegraph/mux"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/router"
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

	defList, err := apiclient.Search.SearchTokens(ctx, &opt)
	if err != nil {
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
		Total   int32
		Results []payloads.TokenSearchResult
	}{
		Total:   defList.Total,
		Results: results,
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

	opt.RepoRev = sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: mux.Vars(r)["Repo"]},
		Rev:      mux.Vars(r)["Rev"],
	}

	vcsEntryList, err := apiclient.Search.SearchText(ctx, &opt)
	if err != nil {
		return err
	}

	results := make([]payloads.TextSearchResult, len(vcsEntryList.SearchResults))
	// Retrieve the corresponding tokenized TreeEntry for each VCS search result.
	for i, vcsEntry := range vcsEntryList.SearchResults {
		matchEntryString := string(vcsEntry.Match)
		matchEntryLines := strings.Split(matchEntryString, "\n")

		results[i] = payloads.TextSearchResult{
			File:      vcsEntry.File,
			StartLine: vcsEntry.StartLine,
			EndLine:   vcsEntry.EndLine,
			Lines:     matchEntryLines,
		}
	}

	return e.Encode(&struct {
		Results []payloads.TextSearchResult
	}{
		Results: results,
	})
}
