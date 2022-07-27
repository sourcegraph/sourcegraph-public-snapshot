package repos

import (
	"context"
	"net/url"

	"github.com/sourcegraph/log"
	"github.com/tidwall/gjson"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	webhookbuilder "github.com/sourcegraph/sourcegraph/internal/repos/worker"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type webhookBuildHandler struct {
	store Store
}

func (w *webhookBuildHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	wbj, ok := record.(*webhookbuilder.Job)
	if !ok {
		return errors.Newf("expected Job, got %T", record)
	}

	switch wbj.ExtSvcKind {
	case "GITHUB":
		svcs, err := w.store.ExternalServiceStore().List(ctx, database.ExternalServicesListOptions{})
		if err != nil || len(svcs) != 1 {
			return errors.Wrap(err, "get external service")
		}
		svc := svcs[0]

		baseURL, err := url.Parse(gjson.Get(svc.Config, "url").String())
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
		gh := NewGitHubWebhookAPI(client)

		id, err := gh.FindSyncWebhook(ctx, wbj.RepoName)
		if err != nil {
			return err
		}
	}
}
