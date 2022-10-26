package webhooks

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

type WebhookHandlers struct {
}

// NewHandler is responsible for handling all incoming webhooks
// and invoking the correct handlers depending on where the webhooks
// come from.
func NewHandler(logger log.Logger, db database.DB, gh *GitHubWebhook) http.Handler {
	logger = logger.Scoped("webhookHandler", "handler used to route webhooks")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidString := mux.Vars(r)["webhook_uuid"]
		if uuidString == "" {
			http.Error(w, "missing uuid", http.StatusBadRequest)
			return
		}

		webhookUUID, err := uuid.Parse(uuidString)
		if err != nil {
			logger.Error("Error while parsing Webhook UUID", log.Error(err))
			http.Error(w, fmt.Sprintf("Could not parse UUID from URL path %q.", uuidString), http.StatusBadRequest)
			return
		}

		webhook, err := db.Webhooks(keyring.Default().WebhookKey).GetByUUID(r.Context(), webhookUUID)
		if err != nil {
			logger.Error("Error while fetching webhook by UUID", log.Error(err))
			http.Error(w, "Could not find webhook with provided UUID.", http.StatusNotFound)
			return
		}
		SetWebhookID(r.Context(), webhook.ID)

		var secret string
		if webhook.Secret != nil {
			secret, err = webhook.Secret.Decrypt(r.Context())
			if err != nil {
				logger.Error("Error while decrypting webhook secret", log.Error(err))
				http.Error(w, "Could not decrypt webhook secret.", http.StatusInternalServerError)
				return
			}
		}

		switch webhook.CodeHostKind {
		case extsvc.KindGitHub:
			handleGitHubWebHook(w, r, webhook.CodeHostURN, secret, gh)
			return
		case extsvc.KindGitLab:
			if secret != "" {
				_, err := gitlabValidatePayload(r, secret)
				if err != nil {
					http.Error(w, "Could not validate payload with secret.", http.StatusBadRequest)
					return
				}
			}
		case extsvc.KindBitbucketServer:
			// TODO: handle Bitbucket Server secret verification
		case extsvc.KindBitbucketCloud:
			// TODO: handle Bitbucket Cloud secret verification
		}

		http.Error(w, fmt.Sprintf("webhooks not implemented for code host kind %q", webhook.CodeHostKind), http.StatusNotImplemented)
	})
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

func handleGitHubWebHook(w http.ResponseWriter, r *http.Request, urn extsvc.CodeHostBaseURL, secret string, gh *GitHubWebhook) {
	if secret == "" {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error while reading request body.", http.StatusInternalServerError)
			return
		}

		gh.HandleWebhook(w, r, urn, payload)
		return
	}

	payload, err := github.ValidatePayload(r, []byte(secret))
	if err != nil {
		http.Error(w, "Could not validate payload with secret.", http.StatusBadRequest)
		return
	}

	gh.HandleWebhook(w, r, urn, payload)
}
