package app

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/prefetch"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
)

func serveUI(w http.ResponseWriter, r *http.Request) error {
	var header http.Header
	var data struct {
		tmpl.Common
		Body      template.HTML
		Stores    *json.RawMessage
		FetchURLs []string
	}

	if _, ok := w.(http.Flusher); !ok {
		return errors.New("cannot prefetch, HTTP response flushing not supported")
	}
	if parseBool(os.Getenv("SG_DISABLE_PREFETCH")) {
		var err error
		data.FetchURLs, err = prefetch.FetchURLsForRequest(r)
		if err != nil {
			log15.Warn("error occured while generating prefetch URLs", "err", err)
		}
	}

	err := tmpl.Exec(r, w, "ui.html", http.StatusOK, header, &data)
	if err != nil {
		return err
	}
	// Flush rendered HTML to make sure the browser can start rendering the
	// page without waiting for prefetching to finish.
	w.(http.Flusher).Flush()

	err = prefetch.ResolveFetches(w, r, data.FetchURLs)
	if err != nil {
		log15.Error("prefecting data endpoints failed", "err", err)
	}
	return nil
}

func parseBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}
