package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/opencodegraph"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/limitedgzip"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Update OpenCodeGraph JSON Schemas (assuming ../opencodegraph relative to this repository's root
// is the https://github.com/sourcegraph/opencodegraph repository):
//
// cp ../opencodegraph/lib/schema/src/opencodegraph.schema.json schema/
// cp ../opencodegraph/lib/protocol/src/opencodegraph-protocol.schema.json schema/
// sed -i 's#../../schema/src/##g' schema/opencodegraph-protocol.schema.json
// bazel run //schema:write_generated_schema

func serveOpenCodeGraph(logger log.Logger) func(w http.ResponseWriter, r *http.Request) (err error) {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		flagSet := featureflag.FromContext(r.Context())
		if !flagSet.GetBoolOr("opencodegraph", false) {
			return errors.New("OpenCodeGraph is not enabled (use the 'opencodegraph' feature flag)")
		}

		if r.Method != "POST" {
			// The URL router should not have routed to this handler if method is not POST, but just
			// in case.
			return errors.New("method must be POST")
		}

		requestSource := search.GuessSource(r)
		r = r.WithContext(trace.WithRequestSource(r.Context(), requestSource))

		if r.Header.Get("Content-Encoding") == "gzip" {
			r.Body, err = limitedgzip.WithReader(r.Body, int64(gzipFileSizeLimit))
			if err != nil {
				return errors.Wrap(err, "failed to decompress request body")
			}

			defer r.Body.Close()
		}

		method, cap, ann, err := opencodegraph.DecodeRequestMessage(json.NewDecoder(r.Body))
		if err != nil {
			return errors.Wrapf(err, "failed to decode request message")
		}

		var result any
		switch {
		case cap != nil:
			result, err = opencodegraph.AllProviders.Capabilities(r.Context(), *cap)
		case ann != nil:
			result, err = opencodegraph.AllProviders.Annotations(r.Context(), *ann)
		default:
			return errors.Newf("unrecognized OpenCodeGraph request method %q", method)
		}

		var respMsg schema.ResponseMessage
		if err == nil {
			respMsg.Result = result
		} else {
			respMsg.Error = &schema.ResponseError{
				Code:    1,
				Message: err.Error(),
			}
			logger.Error("error handling OpenCodeGraph method", log.Error(err))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(respMsg)
		return nil
	}
}
