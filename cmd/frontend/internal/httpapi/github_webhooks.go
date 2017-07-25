package httpapi

import (
	"net/http"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tracking"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var webhookSecretToken = env.Get("SRC_GITHUB_APP_WEBHOOKS_SECRET", "", "GitHub Sourcegraph App webhook secret token.")

func serveReceiveGitHubWebhooks(w http.ResponseWriter, r *http.Request) error {
	payload, err := github.ValidatePayload(r, []byte(webhookSecretToken))
	if err != nil {
		log15.Error("httpapi.serveReceiveGitHubWebhooks: invalid payload or secret", "error", err)
		return err
	}
	eventType := github.WebHookType(r)
	event, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		log15.Error("httpapi.serveReceiveGitHubWebhooks: error parsing payload", "error", err)
		return err
	}

	err = tracking.TrackGitHubWebhook(eventType, event)
	if err != nil {
		log15.Error("httpapi.serveReceiveGitHubWebhooks: error logging webhook", "error", err)
		return err
	}
	return nil
}
