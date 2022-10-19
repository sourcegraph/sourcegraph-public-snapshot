package webhooks

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

type WebhookHandlers struct {
}

// WebhooksHandler is responsible for handling all incoming webhooks
// and invoking the correct handlers depending on where the webhooks
// come from.
func NewHandler(db database.DB) http.Handler {
	base := mux.NewRouter().PathPrefix("/webhooks").Subrouter()
	base.Path("/{webhook_uuid}").Methods("POST").Handler(webhookHandler(db))

	return base
}

func webhookHandler(db database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuidString := mux.Vars(r)["webhook_uuid"]
		if uuidString == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		webhookUUID, err := uuid.Parse(uuidString)
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not parse UUID from URL path %q.", uuidString), http.StatusBadRequest)
			return
		}

		webhook, err := db.Webhooks(keyring.Default().WebhookKey).GetByUUID(r.Context(), webhookUUID)
		if err != nil {
			http.Error(w, "Could not find webhook with provided UUID.", http.StatusNotFound)
			return
		}

		switch webhook.CodeHostKind {
		case extsvc.KindGitHub:
			fallthrough
		case extsvc.KindGitLab:
			fallthrough
		case extsvc.KindBitbucketServer:
			fallthrough
		case extsvc.KindBitbucketCloud:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}
}
