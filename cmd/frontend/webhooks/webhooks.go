package webhooks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/google/go-github/github"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

type WebhookHandlers struct{}

type Registerer interface {
	Register(webhook *Webhook)
}

type RegistererHandler interface {
	Registerer
	http.Handler
}

// Register associates a given event type(s) with the specified handler.
// Handlers are organized into a stack and executed sequentially, so the order in
// which they are provided is significant.
func (h *Webhook) Register(handler WebhookHandler, codeHostKind string, eventTypes ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.handlers == nil {
		h.handlers = make(map[string]map[string][]WebhookHandler)
	}
	if _, ok := h.handlers[codeHostKind]; !ok {
		h.handlers[codeHostKind] = make(map[string][]WebhookHandler)
	}
	for _, eventType := range eventTypes {
		h.handlers[codeHostKind][eventType] = append(h.handlers[codeHostKind][eventType], handler)
	}
}

// NewHandler is responsible for handling all incoming webhooks
// and invoking the correct handlers depending on where the webhooks
// come from.
func NewHandler(logger log.Logger, db database.DB, gh *Webhook) http.Handler {
	base := mux.NewRouter().PathPrefix("/.api/webhooks").Subrouter()
	base.Path("/{webhook_uuid}").Methods("POST").Handler(webhookHandler(logger, db, gh))

	return base
}

// WebhookHandler is a handler for a webhook event, the 'event' param could be any of the event types
// permissible based on the event type(s) the handler was registered against. If you register a handler
// for many event types, you should do a type switch within your handler.
// Handlers are responsible for fetching the necessary credentials to perform their associated tasks.
type WebhookHandler func(ctx context.Context, db database.DB, codeHostURN extsvc.CodeHostBaseURL, event any) error

// Webhook is responsible for handling incoming http requests for github webhooks
// and routing to any registered WebhookHandlers, events are routed by their event type,
// passed in the X-Github-Event header
type Webhook struct {
	DB database.DB

	mu sync.RWMutex
	// Mapped by codeHostKind: webhookEvent: handlers
	handlers map[string]map[string][]WebhookHandler
}

func webhookHandler(logger log.Logger, db database.DB, wh *Webhook) http.HandlerFunc {
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
			handleGitHubWebHook(w, r, webhook.CodeHostURN, secret, &GitHubWebhook{Webhook: wh})
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

// Dispatch accepts an event for a particular event type and dispatches it
// to the appropriate stack of handlers, if any are configured.
func (h *Webhook) Dispatch(ctx context.Context, eventType string, codeHostKind string, codeHostURN extsvc.CodeHostBaseURL, e any) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	g := errgroup.Group{}
	for _, handler := range h.handlers[codeHostKind][eventType] {
		// capture the handler variable within this loop
		handler := handler
		g.Go(func() error {
			return handler(ctx, h.DB, codeHostURN, e)
		})
	}
	return g.Wait()
}
