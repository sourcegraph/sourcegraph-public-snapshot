package repos

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos/webhookworker"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type webhookBuildHandler struct {
	store Store
	doer  httpcli.Doer
}

func newWebhookBuildHandler(store Store, doer httpcli.Doer) *webhookBuildHandler {
	return &webhookBuildHandler{store: store, doer: doer}
}

func (w *webhookBuildHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	job, ok := record.(*webhookworker.Job)
	if !ok {
		return errcode.MakeNonRetryable((errors.Newf("expected Job, got %T", record)))
	}

	switch job.ExtSvcKind {
	case extsvc.KindGitHub:
		return w.handleKindGitHub(ctx, logger, job)
	default:
		return errcode.MakeNonRetryable(errors.New("Unable to determine external service kind"))
	}
}

func (w *webhookBuildHandler) handleKindGitHub(ctx context.Context, logger log.Logger, job *webhookworker.Job) error {
	svc, err := w.store.ExternalServiceStore().GetByID(ctx, job.ExtSvcID)
	if err != nil {
		return errcode.MakeNonRetryable(errors.Wrap(err, "handleKindGitHub: get external service failed"))
	}

	parsed, err := extsvc.ParseConfig(svc.Kind, svc.Config)
	if err != nil {
		return errcode.MakeNonRetryable(errors.Wrap(err, "handleKindGitHub: ParseConfig failed"))
	}

	conn, ok := parsed.(*schema.GitHubConnection)
	if !ok {
		return errcode.MakeNonRetryable(errors.Newf("handleKindGitHub: expected *schema.GitHubConnection, got %T", parsed))
	}

	baseURL, err := url.Parse("")
	if err != nil {
		return errcode.MakeNonRetryable(errors.Wrap(err, "handleKindGitHub: parse baseURL failed"))
	}

	client := github.NewV3Client(logger, svc.URN(), baseURL, &auth.OAuthBearerToken{Token: conn.Token}, w.doer)
	id, err := client.FindSyncWebhook(ctx, job.RepoName) // TODO: Don't make API calls every time
	if err != nil {
		return errors.Wrap(err, "handleKindGitHub: FindSyncWebhook failed")
	}

	// found the webhook
	// don't build a new one
	if id != 0 {
		logger.Info(fmt.Sprintf("Webhook exists with ID: %d", id))
		return nil
	}

	secret, err := randomHex(32)
	if err != nil {
		return errcode.MakeNonRetryable(errors.Wrap(err, "handleKindGitHub: secret generation failed"))
	}

	found := findOrgInWebhooks(conn.Webhooks, job.Org)
	if !found {
		conn.Webhooks = append(conn.Webhooks, &schema.GitHubWebhook{
			Org: job.Org, Secret: secret,
		})

		newConfig, err := json.Marshal(conn)
		if err != nil {
			return errcode.MakeNonRetryable(errors.Wrap(err, "handleKindGitHub: Marshal failed"))
		}

		svc.Config = string(newConfig)
	}

	id, err = client.CreateSyncWebhook(ctx, job.RepoName, globals.ExternalURL().Host, secret) // TODO: Add to DB
	if err != nil {
		return errors.Wrap(err, "handleKindGitHub: CreateSyncWebhook failed")
	}

	logger.Info(fmt.Sprintf("Created webhook with ID: %d", id))
	return nil
}

func findOrgInWebhooks(webhooks []*schema.GitHubWebhook, org string) bool {
	for _, webhook := range webhooks {
		if webhook.Org == org {
			return true
		}
	}
	return false
}

func randomHex(n int) (string, error) {
	r := make([]byte, n/2)
	_, err := rand.Read(r)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(r), nil
}
