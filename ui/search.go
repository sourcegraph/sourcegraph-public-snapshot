package ui

import (
	"encoding/json"
	"fmt"
	"net/http"

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

	repoURI, query := mux.Vars(r)["Repo"], r.URL.Query()["q"][0]
	rawQuery := fmt.Sprintf("%s %s", repoURI, query)
	opt := sourcegraph.SearchOptions{
		Defs:  true,
		Query: rawQuery,
		ListOptions: sourcegraph.ListOptions{
			PerPage: 100,
		},
	}

	defList, err := apiclient.Search.Search(ctx, &opt)
	if err != nil {
		return err
	}

	results := make([]payloads.TokenSearchResult, len(defList.Defs))
	for i, def := range defList.Defs {
		def.DocHTML = htmlutil.SanitizeForPB(def.DocHTML.HTML)
		qualifiedName := sourcecode.DefQualifiedNameAndType(def, "scope")
		qualifiedName = sourcecode.OverrideStyleViaRegexpFlags(qualifiedName)
		results[i] = payloads.TokenSearchResult{
			Def:           def,
			QualifiedName: htmlutil.SanitizeForPB(string(qualifiedName)),
			URL:           router.Rel.URLToDef(def.DefKey).String(),
		}
	}

	return e.Encode(&struct {
		Results []payloads.TokenSearchResult
	}{
		Results: results,
	})
}
