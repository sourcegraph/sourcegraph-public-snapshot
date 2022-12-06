package repos

import (
	"context"
	"encoding/hex"
	"encoding/json"
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
	"github.com/sourcegraph/sourcegraph/internal/types"
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

func (w *webhookBuildHandler) Handle(ctx context.Context, logger log.Logger, job *webhookworker.Job) error {
	switch job.ExtSvcKind {
	case extsvc.KindGitHub:
		return w.handleKindGitHub(ctx, logger, job)
	default:
		return errcode.MakeNonRetryable(errors.Errorf("unable to handle external service kind: %q", job.ExtSvcKind))
	}
}

func (w *webhookBuildHandler) handleKindGitHub(ctx context.Context, logger log.Logger, job *webhookworker.Job) error {
	svc, err := w.store.ExternalServiceStore().GetByID(ctx, job.ExtSvcID)
	if err != nil {
		return errcode.MakeNonRetryable(errors.Wrap(err, "handleKindGitHub: get external service failed"))
	}

	parsed, err := extsvc.ParseEncryptableConfig(ctx, svc.Kind, svc.Config)
	if err != nil {
		return errcode.MakeNonRetryable(errors.Wrap(err, "handleKindGitHub: ParseConfig failed"))
	}

	conn, ok := parsed.(*schema.GitHubConnection)
	if !ok {
		return errcode.MakeNonRetryable(errors.Newf("handleKindGitHub: expected *schema.GitHubConnection, got %T", parsed))
	}

	baseURL, err := url.Parse(conn.Url)
	if err != nil {
		return errcode.MakeNonRetryable(errors.Wrap(err, "handleKindGitHub: parse baseURL failed"))
	}
	client := github.NewV3Client(logger, svc.URN(), baseURL, &auth.OAuthBearerToken{Token: conn.Token}, w.doer)

	// TODO: Not make an API call upon every request. We would need a way to save
	// whether or not we created a hook locally
	webhookPayload, err := client.FindSyncWebhook(ctx, job.RepoName)
	if err != nil && err.Error() != "unable to find webhook" {
		return errors.Wrap(err, "handleKindGitHub: FindSyncWebhook failed")
	}

	// found webhook from GitHub API, don't build a new one
	if webhookPayload != nil {
		if err := addWebhookToExtSvc(svc, conn, job.Org, webhookPayload.Config.Secret); err != nil {
			return errors.Wrap(err, "handleKindGitHub: Webhook found but addWebhookToExtSvc failed")
		}

		logger.Info("webhook found", log.Int("ID", webhookPayload.ID))
		return nil
	}

	secret, err := randomHex(32)
	if err != nil {
		return errcode.MakeNonRetryable(errors.Wrap(err, "handleKindGitHub: secret generation failed"))
	}

	id, err := client.CreateSyncWebhook(ctx, job.RepoName, globals.ExternalURL().Host, secret)
	if err != nil {
		return errors.Wrap(err, "handleKindGitHub: CreateSyncWebhook failed")
	}

	if err := addWebhookToExtSvc(svc, conn, job.Org, secret); err != nil {
		return errors.Wrap(err, "handleKindGitHub: Webhook created but addWebhookToExtSvc failed")
	}

	logger.Info("webhook created", log.Int("ID", id))
	return nil
}

func addWebhookToExtSvc(svc *types.ExternalService, conn *schema.GitHubConnection, org, secret string) error {
	if webhookExistsInConfig(conn.Webhooks, org) {
		return nil
	}

	conn.Webhooks = append(conn.Webhooks, &schema.GitHubWebhook{
		Org: org, Secret: secret,
	})

	serialized, err := json.Marshal(conn)
	if err != nil {
		return err
	}
	svc.Config.Set(string(serialized))
	return nil
}

func webhookExistsInConfig(webhooks []*schema.GitHubWebhook, org string) bool {
	for _, webhook := range webhooks {
		if webhook.Org == org {
			return true
		}
	}
	return false
}

func randomHex(n int) (string, error) {
	r := make([]byte, (n+1)/2)
	_, err := rand.Read(r)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(r)[:n], nil
}
