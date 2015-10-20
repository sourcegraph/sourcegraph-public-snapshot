package markdown

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"src.sourcegraph.com/sourcegraph/app/internal"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/doc"
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

	rendered, err := doc.ToHTML(doc.Markdown, data)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, bytes.NewReader(rendered))
	if err != nil {
		return err
	}
	return nil
}
