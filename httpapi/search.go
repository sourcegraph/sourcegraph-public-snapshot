package httpapi

import (
	"net/http"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveSearch(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	var opt sourcegraph.SearchOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	results, err := s.Search.Search(ctx, &opt)
	if err != nil {
		return err
	}

	return writeJSON(w, results)
}

func serveSearchSuggestions(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	var q sourcegraph.RawQuery
	err := schemaDecoder.Decode(&q, r.URL.Query())
	if err != nil {
		return err
	}

	suggs, err := s.Search.Suggest(ctx, &q)
	if err != nil {
		return err
	}

	return writeJSON(w, suggs)
}

func serveSearchComplete(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	var q sourcegraph.RawQuery
	err := schemaDecoder.Decode(&q, r.URL.Query())
	if err != nil {
		return err
	}

	comps, err := s.Search.Complete(ctx, &q)
	if err != nil {
		return err
	}

	return writeJSON(w, comps)
}
