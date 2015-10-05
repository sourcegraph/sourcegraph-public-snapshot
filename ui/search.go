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

	var opt sourcegraph.TokenSearchOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	repoURI := mux.Vars(r)["Repo"]
	opt.Query = fmt.Sprintf("%s %s", repoURI, opt.Query)

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
