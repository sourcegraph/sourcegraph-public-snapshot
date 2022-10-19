package webhooks

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

type WebhookHandlers struct {
}

func WebhooksHandler(db database.DB, gh *GitHubWebhook) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuidString := mux.Vars(r)["webhook_uuid"]
		if uuidString == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		webhookUUID, err := uuid.Parse(uuidString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		webhook, err := db.Webhooks(keyring.Default().WebhookKey).GetByUUID(r.Context(), webhookUUID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch webhook.CodeHostKind {
		case extsvc.KindGitHub:
			gh.ServeHTTP(w, r)
			break
		case extsvc.KindGitLab:
		case extsvc.KindBitbucketServer:
		case extsvc.KindBitbucketCloud:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}
}
