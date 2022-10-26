package webhooks

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"

	gh "github.com/google/go-github/v43/github"
	"github.com/inconshreveable/log15"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type Registerer interface {
	Register(webhook *GitHubWebhook)
}

// WebhookHandler is a handler for a webhook event, the 'event' param could be any of the event types
// permissible based on the event type(s) the handler was registered against. If you register a handler
// for many event types, you should do a type switch within your handler.
// Handlers are responsible for fetching the necessary credentials to perform their associated tasks.
type WebhookHandler func(ctx context.Context, db database.DB, codeHostURN extsvc.CodeHostBaseURL, event any) error

// GitHubWebhook is responsible for handling incoming http requests for github webhooks
// and routing to any registered WebhookHandlers, events are routed by their event type,
// passed in the X-Github-Event header
type GitHubWebhook struct {
	DB database.DB

	mu       sync.RWMutex
	handlers map[string][]WebhookHandler
}

func (h *GitHubWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log15.Error("Error parsing github webhook event", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// get external service and validate webhook payload signature
	extSvc, err := h.getExternalService(r, body)
	if err != nil {
		log15.Error("Could not find valid external service for webhook", "error", err)
		http.Error(w, "External service not found", http.StatusInternalServerError)
		return
	}

	SetExternalServiceID(r.Context(), extSvc.ID)

	rawConfig, err := extSvc.Config.Decrypt(r.Context())
	if err != nil {
		log15.Error("Could not decode external service config", "error", err)
		http.Error(w, "Invalid external service config", http.StatusInternalServerError)
		return
	}

	config := &schema.GitHubConnection{}
	err = json.Unmarshal([]byte(rawConfig), config)
	if err != nil {
		log15.Error("Could not decode external service config", "error", err)
		http.Error(w, "Invalid external service config", http.StatusInternalServerError)
		return
	}

	codeHostURN, err := extsvc.NewCodeHostBaseURL(config.Url)
	if err != nil {
		log15.Error("Could not parse code host URL from config", "error", err)
		http.Error(w, "Invalid code host URL", http.StatusInternalServerError)
		return
	}

	h.HandleWebhook(w, r, codeHostURN, body)
}

func (h *GitHubWebhook) HandleWebhook(w http.ResponseWriter, r *http.Request, codeHostURN extsvc.CodeHostBaseURL, requestBody []byte) {
	// 🚨 SECURITY: now that the payload and shared secret have been validated,
	// we can use an internal actor on the context.
	ctx := actor.WithInternalActor(r.Context())

	// parse event
	eventType := gh.WebHookType(r)
	e, err := gh.ParseWebHook(eventType, requestBody)
	if err != nil {
		log15.Error("Error parsing github webhook event", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// match event handlers
	err = h.Dispatch(ctx, eventType, codeHostURN, e)
	if err != nil {
		log15.Error("Error handling github webhook event", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Dispatch accepts an event for a particular event type and dispatches it
// to the appropriate stack of handlers, if any are configured.
func (h *GitHubWebhook) Dispatch(ctx context.Context, eventType string, codeHostURN extsvc.CodeHostBaseURL, e any) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	g := errgroup.Group{}
	for _, handler := range h.handlers[eventType] {
		// capture the handler variable within this loop
		handler := handler
		g.Go(func() error {
			return handler(ctx, h.DB, codeHostURN, e)
		})
	}
	return g.Wait()
}

// Register associates a given event type(s) with the specified handler.
// Handlers are organized into a stack and executed sequentially, so the order in
// which they are provided is significant.
func (h *GitHubWebhook) Register(handler WebhookHandler, eventTypes ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.handlers == nil {
		h.handlers = make(map[string][]WebhookHandler)
	}
	for _, eventType := range eventTypes {
		h.handlers[eventType] = append(h.handlers[eventType], handler)
	}
}

func (h *GitHubWebhook) getExternalService(r *http.Request, body []byte) (*types.ExternalService, error) {
	var (
		sig   = r.Header.Get("X-Hub-Signature")
		rawID = r.FormValue(extsvc.IDParam)
		err   error
	)

	// this should only happen on old legacy webhook configurations
	// TODO: delete this path once legacy webhooks are deprecated
	if rawID == "" {
		return h.findAndValidateExternalService(r.Context(), sig, body)
	}

	externalServiceID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return nil, err
	}
	e, err := h.DB.ExternalServices().GetByID(r.Context(), externalServiceID)
	if err != nil {
		return nil, err
	}
	c, err := e.Configuration(r.Context())
	if err != nil {
		return nil, err
	}
	gc, ok := c.(*schema.GitHubConnection)
	if !ok {
		return nil, errors.Errorf("invalid configuration, received github webhook for non-github external service: %v", externalServiceID)
	}

	if err := validateAnyConfiguredSecret(gc, sig, body); err != nil {
		return nil, errors.Wrap(err, "validating webhook payload")
	}
	return e, nil
}

// findExternalService is the slow path for validating an incoming webhook against a configured
// external service, it iterates over all configured external services and attempts to match
// the signature to the configured secret
// TODO: delete this once old style webhooks are deprecated
func (h *GitHubWebhook) findAndValidateExternalService(ctx context.Context, sig string, body []byte) (*types.ExternalService, error) {
	// 🚨 SECURITY: Try to authenticate the request with any of the stored secrets
	// in GitHub external services config.
	// If there are no secrets or no secret managed to authenticate the request,
	// we return an error to the client.
	args := database.ExternalServicesListOptions{Kinds: []string{extsvc.KindGitHub}}
	es, err := h.DB.ExternalServices().List(ctx, args)
	if err != nil {
		return nil, err
	}

	for _, e := range es {
		var c any
		c, err = e.Configuration(ctx)
		if err != nil {
			return nil, err
		}
		gc, ok := c.(*schema.GitHubConnection)
		if !ok {
			continue
		}

		if err := validateAnyConfiguredSecret(gc, sig, body); err == nil {
			return e, nil
		}
	}
	return nil, errors.Errorf("couldn't find any external service for webhook")
}

func validateAnyConfiguredSecret(c *schema.GitHubConnection, sig string, body []byte) error {
	if sig == "" {
		// No signature, this implies no secret was configured
		return nil
	}

	// 🚨 SECURITY: Try to authenticate the request with any of the stored secrets
	// If there are no secrets or no secret managed to authenticate the request,
	// we return an error to the client.
	if len(c.Webhooks) == 0 {
		return errors.Errorf("no webhooks defined")
	}

	for _, hook := range c.Webhooks {
		if hook.Secret == "" {
			continue
		}

		if err := gh.ValidateSignature(sig, body, []byte(hook.Secret)); err == nil {
			return nil
		}
	}

	// If we make it here then none of our webhook secrets were valid
	return errors.Errorf("unable to validate webhook signature")
}
