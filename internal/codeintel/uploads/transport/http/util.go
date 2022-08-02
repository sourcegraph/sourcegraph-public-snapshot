package http

import (
	"net/http"

	"github.com/sourcegraph/log"
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
	logger := log.Scoped("codeintel.uploads.transport.http", "")
	if statusCode >= 500 {
		logger.Error(logMessage, log.Error(err))
	}

	if w != nil {
		logger.Error(logMessage, log.Error(err))
	}
}
