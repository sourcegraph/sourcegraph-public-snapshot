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
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	webhookworker "github.com/sourcegraph/sourcegraph/internal/repos/webhookworker"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
		return errors.Newf("expected Job, got %T", record)
	}

	switch job.ExtSvcKind {
	case extsvc.KindGitHub:
		return w.handleKindGitHub(ctx, logger, job)
	}

	return nil
}

func (w *webhookBuildHandler) handleKindGitHub(ctx context.Context, logger log.Logger, job *webhookworker.Job) error {
	extsvc, err := w.store.ExternalServiceStore().GetByID(ctx, job.ExtSvcID)
	if err != nil {
		return errors.Wrap(err, "handleKindGitHub: get external service failed")
	}

	token, err := extractTokenFromConfig(extsvc)
	if err != nil {
		return errors.Wrap(err, "handleKindGitHub: extractTokenFromConfig failed")
	}

	baseURL, err := url.Parse("")
	if err != nil {
		return errors.Wrap(err, "handleKindGitHub: parse baseURL failed")
	}

	client := github.NewV3Client(logger, extsvc.URN(), baseURL, &auth.OAuthBearerToken{Token: token}, w.doer)
	id, err := client.FindSyncWebhook(ctx, job.RepoName)
	if err != nil {
		return errors.Wrap(err, "handleKindGitHub: FindSyncWebhook failed")
	}

	// found the webhook
	// don't build a new one
	if id != 0 {
		logger.Info(fmt.Sprintf("Webhook exists with ID: %d", id))
		return nil
	}

	secret := randomHex(32)
	if err := addSecretToExtSvc(extsvc, job.Org, secret); err != nil {
		return errors.Wrap(err, "handleKindGitHub: add secret to external service failed")
	}

	id, err = client.CreateSyncWebhook(ctx, job.RepoName, globals.ExternalURL().Host, secret)
	if err != nil {
		return errors.Wrap(err, "handleKindGitHub: CreateSyncWebhook failed")
	}

	logger.Info(fmt.Sprintf("Created webhook with ID: %d", id))
	return nil
}

func extractTokenFromConfig(extsvc *types.ExternalService) (string, error) {
	var conn schema.GitHubConnection
	if err := json.Unmarshal([]byte(extsvc.Config), &conn); err != nil {
		return "", err
	}

	return conn.Token, nil
}

func randomHex(n int) string {
	r := make([]byte, n/2)
	_, err := rand.Read(r)

	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(r)
}

func addSecretToExtSvc(extsvc *types.ExternalService, org, secret string) error {
	var config schema.GitHubConnection
	err := json.Unmarshal([]byte(extsvc.Config), &config)
	if err != nil {
		return errors.Wrap(err, "addSecretToExtSvc: Unmarshal failed")
	}

	config.Webhooks = append(config.Webhooks, &schema.GitHubWebhook{
		Org: org, Secret: secret,
	})

	newConfig, err := json.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "addSecretToExtSvc: Marshal failed")
	}

	extsvc.Config = string(newConfig)

	return nil
}
