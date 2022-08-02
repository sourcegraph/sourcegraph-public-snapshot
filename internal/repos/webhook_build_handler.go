package repos

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/sourcegraph/log"
	"github.com/thanhpk/randstr"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/database"
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
}

func newWebhookBuildHandler(store Store) *webhookBuildHandler {
	return &webhookBuildHandler{store: store}
}

func (w *webhookBuildHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	wbj, ok := record.(*webhookworker.Job)
	if !ok {
		return errors.Newf("expected Job, got %T", record)
	}

	switch wbj.ExtSvcKind {
	case "GITHUB":
		return w.handleCaseGitHub(ctx, logger, wbj)
	}

	return nil
}

func (w *webhookBuildHandler) handleCaseGitHub(ctx context.Context, logger log.Logger, wbj *webhookworker.Job) error {
	svcs, err := w.store.ExternalServiceStore().List(ctx, database.ExternalServicesListOptions{})
	if err != nil || len(svcs) != 1 {
		return errors.Wrap(err, "get external service")
	}
	svc := svcs[0]
	baseURL, err := url.Parse("")
	if err != nil {
		return errors.Wrap(err, "parse base URL")
	}

	accounts, err := w.store.UserExternalAccountsStore().List(ctx, database.ExternalAccountsListOptions{})
	if err != nil || len(accounts) < 1 {
		return errors.Wrap(err, "get user accounts")
	}

	_, token, err := github.GetExternalAccountData(&accounts[0].AccountData)
	if err != nil {
		return errors.Wrap(err, "get token")
	}

	cf := httpcli.ExternalClientFactory
	opts := []httpcli.Opt{}
	cli, err := cf.Doer(opts...)
	if err != nil {
		return errors.Wrap(err, "create client")
	}

	client := github.NewV3Client(logger, svc.URN(), baseURL, &auth.OAuthBearerToken{Token: token.AccessToken}, cli)
	id, err := client.FindSyncWebhook(ctx, wbj.RepoName)
	if err != nil {
		return err
	}

	if id != 0 {
		return nil
	}

	secret := randstr.Hex(32)
	if err := addSecretToExtSvc(svc, "someOrg", secret); err != nil {
		return errors.Wrap(err, "add secret to external service")
	}

	_, err = client.CreateSyncWebhook(ctx, wbj.RepoName, globals.ExternalURL().Host, secret)
	if err != nil {
		return errors.Wrap(err, "create webhook")
	}

	return nil
}

func addSecretToExtSvc(svc *types.ExternalService, org, secret string) error {
	var config schema.GitHubConnection
	err := json.Unmarshal([]byte(svc.Config), &config)
	if err != nil {
		return errors.Wrap(err, "unmarshal config")
	}

	config.Webhooks = append(config.Webhooks, &schema.GitHubWebhook{
		Org: org, Secret: secret,
	})

	newConfig, err := json.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "marshal new config")
	}

	svc.Config = string(newConfig)

	return nil
}
