package httpapi

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/throttled/throttled/v2/store/memstore"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
)

func init() {
	txemail.DisableSilently()
}

func newTest(t *testing.T) *httptestutil.Client {
	logger := logtest.Scoped(t)
	enterpriseServices := enterprise.DefaultServices()
	rateLimitStore, _ := memstore.New(1024)
	rateLimiter := graphqlbackend.NewRateLimiteWatcher(logger, rateLimitStore)

	db := database.NewMockDB()

	return httptestutil.NewTest(NewHandler(db,
		jobutil.NewUnimplementedEnterpriseJobs(),
		router.New(mux.NewRouter()),
		nil,
		rateLimiter,
		&Handlers{
			BatchesGitHubWebhook:          enterpriseServices.BatchesGitHubWebhook,
			BatchesGitLabWebhook:          enterpriseServices.BatchesGitLabWebhook,
			GitHubSyncWebhook:             enterpriseServices.ReposGithubWebhook,
			GitLabSyncWebhook:             enterpriseServices.ReposGitLabWebhook,
			BitbucketServerSyncWebhook:    enterpriseServices.ReposBitbucketServerWebhook,
			BitbucketCloudSyncWebhook:     enterpriseServices.ReposBitbucketCloudWebhook,
			BatchesBitbucketServerWebhook: enterpriseServices.BatchesBitbucketServerWebhook,
			BatchesBitbucketCloudWebhook:  enterpriseServices.BatchesBitbucketCloudWebhook,
			BatchesAzureDevOpsWebhook:     enterpriseServices.BatchesAzureDevOpsWebhook,
			SCIMHandler:                   enterpriseServices.SCIMHandler,
			NewCodeIntelUploadHandler:     enterpriseServices.NewCodeIntelUploadHandler,
			NewComputeStreamHandler:       enterpriseServices.NewComputeStreamHandler,
			PermissionsGitHubWebhook:      enterpriseServices.PermissionsGitHubWebhook,
			NewCompletionsStreamHandler:   enterpriseServices.NewCompletionsStreamHandler,
			NewCodeCompletionsHandler:     enterpriseServices.NewCodeCompletionsHandler,
		},
	))
}
