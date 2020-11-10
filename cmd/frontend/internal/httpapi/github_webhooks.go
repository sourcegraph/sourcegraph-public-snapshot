package httpapi

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	gh "github.com/google/go-github/v28/github"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

type GithubWebhook struct {
	handlers map[string][]WebhookHandler
	Repos    repos.Store
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
	for _, handler := range h.handlers[eventType] {
		err := handler(ctx, extSvc, e)
		if err != nil {
			return err
		}
	}
	return nil
}

// Register associates a given event type with the specified handler(s).
// Handlers are organized into a stack and executed sequentially, so the order in
// which they are provided is significant.
func (h *GithubWebhook) Register(eventType string, handler ...WebhookHandler) {
	// Next we need to see if the service recognizes the event.
	handlers, ok := h.handlers[eventType]
	if !ok {
		handlers = make([]WebhookHandler, 0)
	}

	// Next we can append the incoming handlers to the event.
	handlers = append(handlers, handler...)

	if h.handlers == nil {
		h.handlers = make(map[string][]WebhookHandler)
	}
	// Finally we can ensure that the service map is up-to-date.
	h.handlers[eventType] = handlers
}

type WebhookHandler func(ctx context.Context, extSvc *repos.ExternalService, event interface{}) error

func (h *GithubWebhook) getExternalService(r *http.Request, body []byte) (*repos.ExternalService, error) {
	// 🚨 SECURITY: Try to authenticate the request with any of the stored secrets
	// in GitHub external services config. Since n is usually small here,
	// it's ok for this to be have linear complexity.
	// If there are no secrets or no secret managed to authenticate the request,
	// we return a 401 to the client.
	args := repos.StoreListExternalServicesArgs{Kinds: []string{extsvc.KindGitHub}}
	es, err := h.Repos.ListExternalServices(r.Context(), args)
	if err != nil {
		return nil, err
	}

	var (
		sig               = r.Header.Get("X-Hub-Signature")
		rawID             = r.FormValue(extsvc.IDParam)
		extSvc            *repos.ExternalService
		externalServiceID int64
	)

	// If a webhook was setup before we introduced the externalServiceID as part of the URL,
	// the webhook requests may not contain the external service ID, so we need to fall back.
	if rawID != "" {
		externalServiceID, err = strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	for _, e := range es {
		if externalServiceID != 0 && e.ID != externalServiceID {
			continue
		}
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
				extSvc = e
				break
			}
		}
	}
	if extSvc == nil {
		return nil, fmt.Errorf("could not find any external service for webhook")
	}
	return extSvc, nil
}
