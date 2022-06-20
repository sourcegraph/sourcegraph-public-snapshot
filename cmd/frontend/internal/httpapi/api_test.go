package httpapi

import (
	"github.com/gorilla/mux"
	"github.com/throttled/throttled/v2/store/memstore"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
)

func init() {
	txemail.DisableSilently()
}

func newTest() *httptestutil.Client {
	enterpriseServices := enterprise.DefaultServices()
	rateLimitStore, _ := memstore.New(1024)
	rateLimiter := graphqlbackend.NewRateLimiteWatcher(rateLimitStore)

	return httptestutil.NewTest(NewHandler(database.NewMockDB(),
		router.New(mux.NewRouter()),
		nil,
		rateLimiter,
		&Handlers{
			GitHubWebhook:             enterpriseServices.GitHubWebhook,
			GitLabWebhook:             enterpriseServices.GitLabWebhook,
			BitbucketServerWebhook:    enterpriseServices.BitbucketServerWebhook,
			BitbucketCloudWebhook:     enterpriseServices.BitbucketCloudWebhook,
			NewCodeIntelUploadHandler: enterpriseServices.NewCodeIntelUploadHandler,
			NewComputeStreamHandler:   enterpriseServices.NewComputeStreamHandler,
		},
	))
}
