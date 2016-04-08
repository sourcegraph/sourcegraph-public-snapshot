package app

import (
	"encoding/json"
	"errors"
	"fmt"
	htmlpkg "html"
	"html/template"
	"net/http"
	"os"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func serveUI(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)

	if v := os.Getenv("SG_DISABLE_REACTBRIDGE"); v != "" {
		ctx = ui.DisabledReactPrerendering(ctx)
	}

	res, err := ui.RenderRouter(ctx, r, nil)
	if err != nil {
		return err
	}

	if res.StatusCode == 0 {
		res.StatusCode = http.StatusInternalServerError
	}

	var body template.HTML
	if strings.HasPrefix(res.ContentType, "text/html") {
		body = template.HTML(res.Body)
	} else if res.StatusCode >= 300 && res.StatusCode <= 399 {
		// Nothing to do; we set the Location header below.
	} else if res.Body == "" && res.StatusCode >= 400 {
		body = template.HTML(fmt.Sprintf("<h1>Error</h1><p>HTTP %d %s</p>", res.StatusCode, http.StatusText(res.StatusCode)))
		if handlerutil.DebugMode(r) {
			body += template.HTML(fmt.Sprintf("<p><pre>%s</pre></p>", htmlpkg.EscapeString(res.Error)))
		}
	} else {
		return errors.New("ui render router response is neither text/html nor an error")

	}

	header := make(http.Header)
	header.Set("content-type", res.ContentType)
	if res.RedirectLocation != "" {
		header.Set("location", res.RedirectLocation)
	}

	return tmpl.Exec(r, w, "ui.html", res.StatusCode, header, &struct {
		tmpl.Common
		Body   interface{}
		Stores *json.RawMessage
	}{
		Body:   body,
		Stores: &res.Stores,
	})
}
