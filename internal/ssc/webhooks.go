package ssc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2/clientcredentials"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/schema"
)

type WebhookEvent int

// WebhookEventUserUpgradedToCodyPro is sent when a Sourcegraph user has upgraded to
// Cody Pro (for the free trial offer, from 12/2023 to 2/2024).
const WebhookEventUserUpgradedToCodyPro WebhookEvent = 1

// EmitWebhook sends a webhook to the Self-Service Cody backend.
//
// Any transport errors will be swallowed and ignored, no attempt will be made to persist
// or retry delivery. Since this is best-effort, webhooks should be posted on Goroutines
// as to not add any latency for the original request.
func EmitWebhook(logger log.Logger, event WebhookEvent, payload map[string]string) {
	if !envvar.SourcegraphDotComMode() {
		return
	}
	logger.Debug("sending SSC webhook", log.Int("event", int(event)))

	// We intentionally do not use the HTTP request context from whatever triggered the webhook to be sent.
	// Doing so would cause instability, such as if the original HTTP request were canceled before the
	// webhook deliver is complete.
	ctx := context.Background()
	logger = trace.Logger(ctx, logger)

	// Get all SAMS identity providers. Dotcom may have both a -dev and -prod instance configured.
	var samsProviders []*schema.OpenIDConnectAuthProvider
	for _, p := range conf.Get().AuthProviders {
		if p.Openidconnect != nil && strings.HasPrefix(p.Openidconnect.ClientID, "sams_") {
			samsProviders = append(samsProviders, p.Openidconnect)
		}
	}
	if len(samsProviders) == 0 {
		logger.Error("unable to send webhook because no SAMS IdPs registered")
		return
	}

	for _, samsProvider := range samsProviders {
		sendWebhookToSAMS(ctx, logger, samsProvider, event, payload)
	}
}

func sendWebhookToSAMS(
	ctx context.Context, logger log.Logger, sams *schema.OpenIDConnectAuthProvider,
	event WebhookEvent, payload map[string]string) {
	// Create a generic OAuth client using the supplied SAMS credentials. Only the "dotcom" SAMS client
	// will have the ability to create tokens with the "client.dotcom" scope. This how the recipient of
	// this token will be able to authorize the request.
	clientCreds := &clientcredentials.Config{
		ClientID:     sams.ClientID,
		ClientSecret: sams.ClientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth/token", sams.Issuer),
		Scopes:       []string{"client.dotcom"},
	}
	token, err := clientCreds.Token(ctx)
	if err != nil {
		logger.Error("error issuing SAMS access token", log.Error(err))
		return
	}

	// Stuff the event into the payload itself to simplify the message protocol.
	payload["event"] = fmt.Sprint(event)
	bodyJSON, err := json.Marshal(payload)
	if err != nil {
		logger.Error("marshalling webhook body", log.Error(err))
	}

	// Make the request.
	url := fmt.Sprintf("%s/cody/api/rest/svc/webhook", sams.Issuer)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		logger.Error("creating HTTP request", log.Error(err))
		return
	}
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	resp, err := httpcli.ExternalClient.Do(req)
	if err != nil {
		logger.Error("sending webhook request", log.Error(err))
		return
	}
	defer func() { _ = resp.Body.Close() }()

	// Note that we ignore the response body, assuming it to be empty.
	if resp.StatusCode > 299 { // wiggle room for OK, NoContent, Created, etc.
		logger.Error("Unsuccessful request response", log.Int("status", resp.StatusCode))
	}
}
