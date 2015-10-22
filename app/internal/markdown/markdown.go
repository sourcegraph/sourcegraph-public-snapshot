package markdown

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"

	"src.sourcegraph.com/sourcegraph/app/internal"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func init() {
	internal.Handlers[router.Markdown] = serveMarkdown
}

// TODO(slimsag): put this into a gRPC-exposed service which also handles
// mentions.

func serveMarkdown(w http.ResponseWriter, r *http.Request) error {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	// Use Markdown service to render the markdown.
	resp, err := apiclient.Markdown.Render(ctx, &sourcegraph.MarkdownRenderOp{
		Markdown: data,
		Opt: sourcegraph.MarkdownOpt{
			EnableCheckboxes: true,
		},
	})
	if err != nil {
		return err
	}

	// Serialize for rendering.
	html := &pbtypes.HTML{HTML: string(resp.Rendered)}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(html)
}
