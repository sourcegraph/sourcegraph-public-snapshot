package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/inconshreveable/log15"
	"github.com/tomnomnom/linkheader"
)

func hasQuery(r *http.Request, name string) bool {
	return r.URL.Query().Get(name) != ""
}

func getQuery(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

func getQueryInt(r *http.Request, name string) int {
	value, _ := strconv.Atoi(r.URL.Query().Get(name))
	return value
}

func getQueryIntDefault(r *http.Request, name string, defaultValue int) int {
	value, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil {
		value = defaultValue
	}
	return value
}

func getQueryBool(r *http.Request, name string) bool {
	value, _ := strconv.ParseBool(r.URL.Query().Get(name))
	return value
}

func makeNextLink(url *url.URL, newQueryValues map[string]interface{}) string {
	q := url.Query()
	for k, v := range newQueryValues {
		q.Set(k, fmt.Sprintf("%v", v))
	}
	url.RawQuery = q.Encode()

	header := linkheader.Link{
		URL: url.String(),
		Rel: "next",
	}
	return header.String()
}

// idFromRequest returns the database id from the request URL's path. This method
// must only be called from routes containing the `id:[0-9]+` pattern, as the error
// return from ParseInt is not checked.
func idFromRequest(r *http.Request) int64 {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	return id
}

// copyAll writes the contents of r to w and logs on write failure.
func copyAll(w http.ResponseWriter, r io.Reader) {
	if _, err := io.Copy(w, r); err != nil {
		log15.Error("Failed to write payload to client", "error", err)
	}
}

// writeJSON writes the JSON-encoded payload to w and logs on write failure.
// If there is an encoding error, then a 500-level status is written to w.
func writeJSON(w http.ResponseWriter, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log15.Error("Failed to serialize result", "error", err)
		http.Error(w, fmt.Sprintf("failed to serialize result: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	copyAll(w, bytes.NewReader(data))
}

func sanitizeRoot(s string) string {
	if s == "" || s == "/" {
		return ""
	}
	if !strings.HasSuffix(s, "/") {
		s += "/"
	}
	return s
}
