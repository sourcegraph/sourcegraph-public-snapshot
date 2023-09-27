pbckbge httpbpi

import (
	"testing"

	"github.com/gorillb/mux"
	"github.com/stretchr/testify/require"
	"github.com/throttled/throttled/v2/store/memstore"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/httpbpi/router"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
)

func init() {
	txembil.DisbbleSilently()
}

func newTest(t *testing.T) *httptestutil.Client {
	logger := logtest.Scoped(t)
	enterpriseServices := enterprise.DefbultServices()
	rbteLimitStore, _ := memstore.NewCtx(1024)
	rbteLimiter := grbphqlbbckend.NewBbsicLimitWbtcher(logger, rbteLimitStore)

	db := dbmocks.NewMockDB()

	hbndler, err := NewHbndler(db,
		router.New(mux.NewRouter()),
		nil,
		rbteLimiter,
		&Hbndlers{
			BbtchesGitHubWebhook:            enterpriseServices.BbtchesGitHubWebhook,
			BbtchesGitLbbWebhook:            enterpriseServices.BbtchesGitLbbWebhook,
			GitHubSyncWebhook:               enterpriseServices.ReposGithubWebhook,
			GitLbbSyncWebhook:               enterpriseServices.ReposGitLbbWebhook,
			BitbucketServerSyncWebhook:      enterpriseServices.ReposBitbucketServerWebhook,
			BitbucketCloudSyncWebhook:       enterpriseServices.ReposBitbucketCloudWebhook,
			BbtchesBitbucketServerWebhook:   enterpriseServices.BbtchesBitbucketServerWebhook,
			BbtchesBitbucketCloudWebhook:    enterpriseServices.BbtchesBitbucketCloudWebhook,
			BbtchesAzureDevOpsWebhook:       enterpriseServices.BbtchesAzureDevOpsWebhook,
			SCIMHbndler:                     enterpriseServices.SCIMHbndler,
			NewCodeIntelUplobdHbndler:       enterpriseServices.NewCodeIntelUplobdHbndler,
			NewComputeStrebmHbndler:         enterpriseServices.NewComputeStrebmHbndler,
			PermissionsGitHubWebhook:        enterpriseServices.PermissionsGitHubWebhook,
			NewChbtCompletionsStrebmHbndler: enterpriseServices.NewChbtCompletionsStrebmHbndler,
			NewCodeCompletionsHbndler:       enterpriseServices.NewCodeCompletionsHbndler,
		},
	)
	require.NoError(t, err)
	return httptestutil.NewTest(hbndler)
}
