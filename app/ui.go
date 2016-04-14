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
	"time"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

var ciFactor = func() int {
	if os.Getenv("CI") != "" {
		return 4
	}
	return 1
}()

func serveUI(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)

	if v := os.Getenv("SG_DISABLE_JSSERVER"); v != "" {
		ctx = ui.DisabledReactPrerendering(ctx)
	}

	ctx, cancel := context.WithTimeout(ctx, 2500*time.Millisecond*time.Duration(ciFactor))
	defer cancel()

	var statusCode int
	res, err := ui.RenderRouter(ctx, r, nil)
	if err != nil {
		switch err {
		case context.DeadlineExceeded:
			statusCode = 202
		default:
			return err
		}
	}

	var body template.HTML
	var stores *json.RawMessage
	var head *ui.Head
	var header http.Header
	if res != nil {
		statusCode = res.StatusCode
		stores = &res.Stores
		head = &res.Head

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

		header = make(http.Header)
		header.Set("content-type", res.ContentType)
		if res.RedirectLocation != "" {
			header.Set("location", res.RedirectLocation)
		}
	}

	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	return tmpl.Exec(r, w, "ui.html", statusCode, header, &struct {
		tmpl.Common
		Head   *ui.Head
		Body   interface{}
		Stores *json.RawMessage
	}{
		Head:   head,
		Body:   body,
		Stores: stores,
	})
}
