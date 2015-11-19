package markdown

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"src.sourcegraph.com/sourcegraph/app/internal"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/htmlutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func init() {
	internal.Handlers[router.Markdown] = serveMarkdown
}

func serveMarkdown(w http.ResponseWriter, r *http.Request) error {
	// Read the Markdown.
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// Use Markdown service to render the markdown.
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)
	resp, err := apiclient.Markdown.Render(ctx, &sourcegraph.MarkdownRenderOp{
		Markdown: data,
		Opt: sourcegraph.MarkdownOpt{
			EnableCheckboxes: true,
		},
	})
	if err != nil {
		return err
	}

	// Sanitize the HTML.
	html := htmlutil.SanitizeForPB(string(resp.Rendered))

	// Serialize for rendering.
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(html)
}
