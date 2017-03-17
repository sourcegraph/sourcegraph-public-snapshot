package httpapi

import (
	"encoding/json"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

func serveRepoCreate(w http.ResponseWriter, r *http.Request) error {
	// legacy support for Chrome extension
	var data struct {
		Op struct {
			New struct {
				URI string
			}
		}
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return err
	}
	if _, err := backend.Repos.GetByURI(r.Context(), data.Op.New.URI); err != nil {
		return err
	}
	w.Write([]byte("OK"))
	return nil
}
