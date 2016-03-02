package ui

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/mux"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

func serveDefExamples(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	query := struct {
		sourcegraph.DefListExamplesOptions
		FallbackRepoURI string
	}{}
	err := schemaDecoder.Decode(&query, r.URL.Query())
	if err != nil {
		return err
	}

	dc, _, _, err := handlerutil.GetDefCommon(r, nil)
	if err != nil {
		return err
	}
	def := dc.Def
	spec := sourcegraph.DefSpec{
		Repo:     def.Repo,
		CommitID: def.CommitID,
		Unit:     def.Unit,
		UnitType: def.UnitType,
		Path:     def.Path,
	}
	examples, err := cl.Defs.ListExamples(ctx, &sourcegraph.DefsListExamplesOp{
		Def: spec,
		Rev: mux.Vars(r)["Rev"],
		Opt: &(query.DefListExamplesOptions),
	})
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(examples.Examples)
}
