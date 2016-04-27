package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui"
	"sourcegraph.com/sourcegraph/sourcegraph/conf/feature"
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

	var statusCode int
	if !feature.Features.JSServer || parseBool(os.Getenv("SG_DISABLE_JSSERVER")) || parseBool(r.URL.Query().Get("disable_jsserver")) {
		ctx = ui.DisabledReactPrerendering(ctx)
		statusCode = http.StatusOK
	}

	ctx, cancel := context.WithTimeout(ctx, 2500*time.Millisecond*time.Duration(ciFactor))
	defer cancel()

	res, err := ui.RenderRouter(ctx, r, nil)
	if err != nil {
		switch err {
		case context.DeadlineExceeded:
			log15.Warn("Context deadline exceeded for rendering React component, returning early", "URL", r.URL)
			statusCode = http.StatusAccepted
		default:
			// TODO Return err so it appropriately triggers a response with a 500 status.
			log15.Warn("Error rendering React component on the server", "err", err, "URL", r.URL)
		}
	}

	var header http.Header
	var data struct {
		tmpl.Common
		Head   *ui.Head
		Body   template.HTML
		Stores *json.RawMessage

		ErrorTitle     string
		ErrorDebugInfo string
	}

	if res != nil {
		statusCode = res.StatusCode
		data.Stores = &res.Stores
		data.Head = &res.Head

		if strings.HasPrefix(res.ContentType, "text/html") {
			data.Body = template.HTML(res.Body)
		} else if res.StatusCode >= 300 && res.StatusCode <= 399 {
			// Nothing to do; we set the Location header below.
		} else if res.Body == "" && res.StatusCode >= 400 {
			data.ErrorTitle = fmt.Sprintf("HTTP %d %s", res.StatusCode, http.StatusText(res.StatusCode))
			if handlerutil.DebugMode(r) {
				data.ErrorDebugInfo = res.Error
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

	if statusCode == 0 || statusCode == 500 {
		// TODO Return a http.StatusInternalServerError response instead of pretending everything went ok.
		statusCode = http.StatusAccepted
	}

	return tmpl.Exec(r, w, "ui.html", statusCode, header, &data)
}

func parseBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}
