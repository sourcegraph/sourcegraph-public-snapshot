package http

import (
	"fmt"
	"net/http"

	"github.com/inconshreveable/log15"
)

// func hasQuery(r *http.Request, name string) bool {
// 	return r.URL.Query().Get(name) != ""
// }

// func getQuery(r *http.Request, name string) string {
// 	return r.URL.Query().Get(name)
// }

// func getQueryInt(r *http.Request, name string) int {
// 	value, _ := strconv.Atoi(r.URL.Query().Get(name))
// 	return value
// }

const logPrefix = "codeintel.uploads.transport.http"

func handleErr(w http.ResponseWriter, err error, logMessage string, statusCode int) {
	if statusCode >= 500 {
		log15.Error(fmt.Sprintf("%s: %s", logPrefix, logMessage), "error", err)
	}

	if w != nil {
		http.Error(w, fmt.Sprintf("%s: %s: %s", logPrefix, logMessage, err), statusCode)
	}
}
