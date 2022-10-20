package webhooks

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-github/github"
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
			http.Error(w, "missing uuid", http.StatusBadRequest)
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

		secret, err := webhook.Secret.Decrypt(r.Context())
		if err != nil {
			http.Error(w, "Could not decrypt webhook secret.", http.StatusInternalServerError)
			return
		}

		switch webhook.CodeHostKind {
		case extsvc.KindGitHub:
			_, err := github.ValidatePayload(r, []byte(secret))
			if err != nil {
				http.Error(w, "Could not validate payload with secret.", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNotImplemented)
		case extsvc.KindGitLab:
			_, err := gitlabValidatePayload(r, secret)
			if err != nil {
				http.Error(w, "Could not validate payload with secret.", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNotImplemented)
		case extsvc.KindBitbucketServer:
			// TODO: handle Bitbucket Server secret verification
			w.WriteHeader(http.StatusNotImplemented)
		case extsvc.KindBitbucketCloud:
			// TODO: handle Bitbucket Cloud secret verification
			w.WriteHeader(http.StatusNotImplemented)
		}

		http.Error(w, fmt.Sprintf("webhooks not implemented for code host kind %q", webhook.CodeHostKind), http.StatusNotImplemented)
	}
}

func gitlabValidatePayload(r *http.Request, secret string) ([]byte, error) {
	glSecret := r.Header.Get("X-Gitlab-Token")
	if glSecret != secret {
		return nil, errors.New("secrets don't match!")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return body, nil
}
