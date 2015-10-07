package ui

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/mux"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
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

	// The fully resolved commit ID is needed for when this RepoRevSpec is eventually
	// passed into sourcecode.sanitizeEntry via the RepoTree.Get call below.
	if opt.RepoRev.CommitID == "" {
		commit, err := apiclient.Repos.GetCommit(ctx, &opt.RepoRev)
		if err != nil {
			return err
		}
		opt.RepoRev.CommitID = string(commit.ID)
	}

	vcsEntryList, err := apiclient.Search.SearchText(ctx, &opt)
	if err != nil {
		return err
	}

	results := make([]payloads.TextSearchResult, len(vcsEntryList.SearchResults))
	// Retrieve the corresponding tokenized TreeEntry for each VCS search result.
	for i, vcsEntry := range vcsEntryList.SearchResults {
		entrySpec := sourcegraph.TreeEntrySpec{RepoRev: opt.RepoRev, Path: vcsEntry.File}
		// TODO(perf) speed this process up by converting each vcsEntry into a treeEntry directly
		// instead of making a grpc call for each result. Right now this is the best we can do,
		// and is similar to the way formatting is handled in server/local/repo_tree.Search.
		entry, err := apiclient.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: entrySpec, Opt: &sourcegraph.RepoTreeGetOptions{
			TokenizedSource:  true,
			HighlightStrings: []string{opt.Query},
			GetFileOptions: vcsclient.GetFileOptions{
				FileRange: vcsclient.FileRange{
					StartLine: int64(vcsEntry.StartLine),
					EndLine:   int64(vcsEntry.EndLine),
				},
			},
		}})
		if err != nil {
			return err
		}

		results[i] = payloads.TextSearchResult{
			TreeEntry: entry,
		}
	}

	return e.Encode(&struct {
		Results []payloads.TextSearchResult
	}{
		Results: results,
	})
}
