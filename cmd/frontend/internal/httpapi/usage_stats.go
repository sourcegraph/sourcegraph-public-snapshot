package httpapi

import (
	"net/http"
)


func usageStatsDownloadServe(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusNotFound)
	return nil
}
