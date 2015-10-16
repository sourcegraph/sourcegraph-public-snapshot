package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sourcegraph/mux"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
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

	var buildInfoMessage string
	buildInfo, err := apiclient.Builds.GetRepoBuildInfo(ctx, &sourcegraph.BuildsGetRepoBuildInfoOp{Repo: opt.RepoRev})
	if err != nil {
		buildInfoMessage = fmt.Sprintf("No build found for %q.", opt.RepoRev.URI)
	} else if buildInfo.Exact == nil {
		buildInfoMessage = fmt.Sprintf("Showing definition results from %d commits behind the latest build.", buildInfo.CommitsBehind)
	}

	return e.Encode(&struct {
		Total     int32
		Results   []payloads.TokenSearchResult
		BuildInfo string
	}{
		Total:     defList.Total,
		Results:   results,
		BuildInfo: buildInfoMessage,
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
