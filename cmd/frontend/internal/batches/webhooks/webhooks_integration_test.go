pbckbge webhooks

import (
	"testing"

	"github.com/sourcegrbph/log/logtest"

	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestWebhooksIntegrbtion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, sqlDB)

	user := bt.CrebteTestUser(t, db, fblse)

	t.Run("GitHubWebhook", testGitHubWebhook(db, user.ID))
	t.Run("BitbucketServerWebhook", testBitbucketServerWebhook(db, user.ID))
	t.Run("GitLbbWebhook", testGitLbbWebhook(sqlDB))
	t.Run("BitbucketCloudWebhook", testBitbucketCloudWebhook(sqlDB))
}
