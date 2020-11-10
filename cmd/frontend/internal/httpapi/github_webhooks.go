package httpapi

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	gh "github.com/google/go-github/v28/github"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

type GithubWebhook struct {
	Repos repos.Store

	mu       sync.RWMutex
	handlers map[string][]WebhookHandler
}

func (h GithubWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log15.Error("Error parsing github webhook event", "error", err)
		http.Error(w, err.Error(), 400)
		return
	}

	// get external service and validate webhook payload signature
	extSvc, err := h.getExternalService(r, body)
	if err != nil {
		log15.Error("Could not find valid external service for webhook", "error", err)
		http.Error(w, "External service not found", 404)
		return
	}

	// parse event
	eventType := gh.WebHookType(r)
	e, err := gh.ParseWebHook(gh.WebHookType(r), body)
	if err != nil {
		log15.Error("Error parsing github webhook event", "error", err)
		http.Error(w, err.Error(), 400)
		return
	}

	// match event handlers
	err = h.Dispatch(r.Context(), eventType, extSvc, e)
	if err != nil {
		log15.Error("Error handling github webhook event", "error", err)
		http.Error(w, err.Error(), 500)
		return
	}
}

// Dispatch accepts an event for a particular event type and dispatches it
// to the appropriate stack of handlers, if any are configured.
func (h *GithubWebhook) Dispatch(ctx context.Context, eventType string, extSvc *repos.ExternalService, e interface{}) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	g := errgroup.Group{}
	for _, handler := range h.handlers[eventType] {
		// capture the handler variable within this loop
		handler := handler
		g.Go(func() error {
			return handler(ctx, extSvc, e)
		})
	}
	return g.Wait()
}

// Register associates a given event type(s) with the specified handler.
// Handlers are organized into a stack and executed sequentially, so the order in
// which they are provided is significant.
func (h *GithubWebhook) Register(handler WebhookHandler, eventTypes ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.handlers == nil {
		h.handlers = make(map[string][]WebhookHandler)
	}
	for _, eventType := range eventTypes {
		h.handlers[eventType] = append(h.handlers[eventType], handler)
	}
}

type WebhookHandler func(ctx context.Context, extSvc *repos.ExternalService, event interface{}) error

func (h *GithubWebhook) getExternalService(r *http.Request, body []byte) (*repos.ExternalService, error) {
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
	e, err := h.Repos.GetExternalService(r.Context(), externalServiceID)
	if err != nil {
		return nil, err
	}
	c, err := e.Configuration()
	if err != nil {
		return nil, err
	}
	gc, ok := c.(*schema.GitHubConnection)
	if !ok {
		return nil, fmt.Errorf("invalid configuration, recieved github webhook for non-github external service: %v", externalServiceID)
	}

	for _, hook := range gc.Webhooks {
		if hook.Secret == "" {
			continue
		}

		if err = gh.ValidateSignature(sig, body, []byte(hook.Secret)); err == nil {
			return e, nil
		}
	}
	return e, nil
}

// findExternalService is the slow path for validating an incoming webhook against a configured
// external service, it iterates over all configured external services and attempts to match
// the signature to the configured secret
// TODO: delete this once old style webhooks are deprecated
func (h *GithubWebhook) findAndValidateExternalService(ctx context.Context, sig string, body []byte) (*repos.ExternalService, error) {
	// 🚨 SECURITY: Try to authenticate the request with any of the stored secrets
	// in GitHub external services config. Since n is usually small here,
	// it's ok for this to be have linear complexity.
	// If there are no secrets or no secret managed to authenticate the request,
	// we return a 401 to the client.
	args := repos.StoreListExternalServicesArgs{Kinds: []string{extsvc.KindGitHub}}
	es, err := h.Repos.ListExternalServices(ctx, args)
	if err != nil {
		return nil, err
	}

	for _, e := range es {
		var c interface{}
		c, err = e.Configuration()
		if err != nil {
			return nil, err
		}
		gc, ok := c.(*schema.GitHubConnection)
		if !ok {
			continue
		}

		for _, hook := range gc.Webhooks {
			if hook.Secret == "" {
				continue
			}

			if err = gh.ValidateSignature(sig, body, []byte(hook.Secret)); err == nil {
				return e, nil
			}
		}
	}
	return nil, fmt.Errorf("couldn't find any external service for webhook")
}
