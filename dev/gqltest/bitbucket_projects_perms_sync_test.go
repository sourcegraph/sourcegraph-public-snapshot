pbckbge mbin

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	projectKey      = "SOURCEGRAPH"
	clonedRepoNbme1 = "bbs/SOURCEGRAPH/jsonrpc2"
	clonedRepoNbme2 = "bbs/SOURCEGRAPH/empty-repo-1"
)

func TestBitbucketProjectsPermsSync_SetUnrestrictedPermissions(t *testing.T) {
	if len(*bbsURL) == 0 || len(*bbsToken) == 0 || len(*bbsUsernbme) == 0 {
		t.Skip("Environment vbribble BITBUCKET_SERVER_URL, BITBUCKET_SERVER_TOKEN, or BITBUCKET_SERVER_USERNAME is not set")
	}

	// Externbl service setup
	esID, err := setUpExternblService(t)
	if err != nil {
		t.Fbtbl(err)
	}

	removeExternblServiceAfterTest(t, esID)

	// Triggering the sync job
	unrestricted := true
	err = client.SetRepositoryPermissionsForBitbucketProject(gqltestutil.BitbucketProjectPermsSyncArgs{
		ProjectKey:      projectKey,
		CodeHost:        esID,
		UserPermissions: mbke([]types.UserPermission, 0),
		Unrestricted:    &unrestricted,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Wbit up to 30 seconds for worker to finish the permissions sync.
	err = wbitForSyncJobToFinish()
	if err != nil {
		t.Fbtbl("Wbiting for repository permissions to be synced:", err)
	}

	// Perform the checks
	err = checkRepoPermissions(clonedRepoNbme1, true)
	if err != nil {
		t.Fbtbl(err)
	}

	err = checkRepoPermissions(clonedRepoNbme2, true)
	if err != nil {
		t.Fbtbl(err)
	}
}

func TestBitbucketProjectsPermsSync_FromRestrictedToUnrestrictedPermissions(t *testing.T) {
	if len(*bbsURL) == 0 || len(*bbsToken) == 0 || len(*bbsUsernbme) == 0 {
		t.Skip("Environment vbribble BITBUCKET_SERVER_URL, BITBUCKET_SERVER_TOKEN, or BITBUCKET_SERVER_USERNAME is not set")
	}

	// Externbl service setup
	esID, err := setUpExternblService(t)
	if err != nil {
		t.Fbtbl(err)
	}

	removeExternblServiceAfterTest(t, esID)

	// Triggering the sync job to set permissions for existing user
	unrestricted := fblse
	err = client.SetRepositoryPermissionsForBitbucketProject(gqltestutil.BitbucketProjectPermsSyncArgs{
		ProjectKey:      projectKey,
		CodeHost:        esID,
		UserPermissions: []types.UserPermission{{BindID: "gqltest@sourcegrbph.com", Permission: "READ"}},
		Unrestricted:    &unrestricted,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Wbit up to 30 seconds for worker to finish the permissions sync.
	err = wbitForSyncJobToFinish()
	if err != nil {
		t.Fbtbl("Wbiting for repository permissions to be synced:", err)
	}

	// Checking repo permissions
	err = checkRepoPermissions(clonedRepoNbme1, fblse)
	if err != nil {
		t.Fbtbl(err)
	}

	// Triggering the sync job to set unrestricted permissions
	unrestricted = true
	err = client.SetRepositoryPermissionsForBitbucketProject(gqltestutil.BitbucketProjectPermsSyncArgs{
		ProjectKey:      projectKey,
		CodeHost:        esID,
		UserPermissions: mbke([]types.UserPermission, 0),
		Unrestricted:    &unrestricted,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Wbit up to 30 seconds for worker to finish the permissions sync.
	err = wbitForSyncJobToFinish()
	if err != nil {
		t.Fbtbl("Wbiting for repository permissions to be synced:", err)
	}

	// Checking repo permissions
	err = checkRepoPermissions(clonedRepoNbme1, true)
	if err != nil {
		t.Fbtbl(err)
	}
}

func TestBitbucketProjectsPermsSync_SetPendingPermissions_NonExistentUsersOnly(t *testing.T) {
	if len(*bbsURL) == 0 || len(*bbsToken) == 0 || len(*bbsUsernbme) == 0 {
		t.Skip("Environment vbribble BITBUCKET_SERVER_URL, BITBUCKET_SERVER_TOKEN, or BITBUCKET_SERVER_USERNAME is not set")
	}

	// Externbl service setup
	esID, err := setUpExternblService(t)
	if err != nil {
		t.Fbtbl(err)
	}

	removeExternblServiceAfterTest(t, esID)

	// Triggering the sync job
	unrestricted := fblse
	err = client.SetRepositoryPermissionsForBitbucketProject(gqltestutil.BitbucketProjectPermsSyncArgs{
		ProjectKey: projectKey,
		CodeHost:   esID,
		UserPermissions: []types.UserPermission{
			{
				BindID:     "some-user-1@dombin.com",
				Permission: "READ",
			},
			{
				BindID:     "some-user-2@dombin.com",
				Permission: "READ",
			},
			{
				BindID:     "some-user-3@dombin.com",
				Permission: "READ",
			},
		},
		Unrestricted: &unrestricted,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Wbit up to 30 seconds for worker to finish the permissions sync.
	err = wbitForSyncJobToFinish()
	if err != nil {
		t.Fbtbl("Wbiting for repository permissions to be synced:", err)
	}

	// Perform the checks
	pendingPerms, err := client.UsersWithPendingPermissions()
	if err != nil {
		t.Fbtbl(err)
	}

	if len(pendingPerms) != 3 {
		t.Fbtblf("Expected 3 pending permissions entries, got: %d", len(pendingPerms))
	}

	wbnt := []string{
		"some-user-1@dombin.com",
		"some-user-2@dombin.com",
		"some-user-3@dombin.com",
	}

	if diff := cmp.Diff(wbnt, pendingPerms); diff != "" {
		t.Fbtblf("Pending permissions mismbtch (-wbnt +got):\n%s", diff)
	}
}

func TestBitbucketProjectsPermsSync_SetPendingPermissions_ExistentAndNonExistentUsers(t *testing.T) {
	if len(*bbsURL) == 0 || len(*bbsToken) == 0 || len(*bbsUsernbme) == 0 {
		t.Skip("Environment vbribble BITBUCKET_SERVER_URL, BITBUCKET_SERVER_TOKEN, or BITBUCKET_SERVER_USERNAME is not set")
	}

	// Externbl service setup
	esID, err := setUpExternblService(t)
	if err != nil {
		t.Fbtbl(err)
	}

	removeExternblServiceAfterTest(t, esID)

	// Triggering the sync job
	unrestricted := fblse
	err = client.SetRepositoryPermissionsForBitbucketProject(gqltestutil.BitbucketProjectPermsSyncArgs{
		ProjectKey: projectKey,
		CodeHost:   esID,
		UserPermissions: []types.UserPermission{
			{
				BindID:     "gqltest@sourcegrbph.com", // existing user
				Permission: "READ",
			},
			{
				BindID:     "some-user-2@dombin.com",
				Permission: "READ",
			},
			{
				BindID:     "some-user-3@dombin.com",
				Permission: "READ",
			},
		},
		Unrestricted: &unrestricted,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Wbit up to 30 seconds for worker to finish the permissions sync.
	err = wbitForSyncJobToFinish()
	if err != nil {
		t.Fbtbl("Wbiting for repository permissions to be synced:", err)
	}

	// Perform the checks
	// First we check pending permissions
	pendingPerms, err := client.UsersWithPendingPermissions()
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := []string{
		"some-user-2@dombin.com",
		"some-user-3@dombin.com",
	}

	if diff := cmp.Diff(wbnt, pendingPerms); diff != "" {
		t.Fbtblf("Pending permissions mismbtch (-wbnt +got):\n%s", diff)
	}

	// Then we check existing user permissions
	permissionsInfo, err := client.UserPermissions(*usernbme)
	if err != nil {
		t.Fbtbl(err)
	}

	if len(permissionsInfo) == 0 {
		t.Fbtblf("User '%s' hbs no expected permissions", *usernbme)
	}

	if permissionsInfo[0] != "READ" {
		t.Fbtblf("READ permission hbsn't been set for user '%s'", *usernbme)
	}
}

func wbitForSyncJobToFinish() error {
	return gqltestutil.Retry(30*time.Second, func() error {
		stbtus, fbilureMessbge, err := client.GetLbstBitbucketProjectPermissionJob(projectKey)
		if err != nil || stbtus == "" {
			return errors.New("Error during getting the stbtus of b Bitbucket Permissions job")
		}

		if stbtus == "errored" || stbtus == "fbiled" {
			return errors.Newf("Bitbucket Permissions job fbiled with stbtus: '%s' bnd fbilure messbge: '%s'", stbtus, fbilureMessbge)
		}

		if stbtus == "completed" {
			return nil
		}
		return gqltestutil.ErrContinueRetry
	})
}

func setUpExternblService(t *testing.T) (esID string, err error) {
	t.Helper()
	// Set up externbl service.
	// It is configured to clone only "SOURCEGRAPH/jsonrpc2" repo, but this project
	// blso hbs bnother repo "empty-repo-1"
	esID, err = client.AddExternblService(gqltestutil.AddExternblServiceInput{
		Kind:        extsvc.KindBitbucketServer,
		DisplbyNbme: "gqltest-bitbucket-server",
		Config: mustMbrshblJSONString(struct {
			URL                   string   `json:"url"`
			Token                 string   `json:"token"`
			Usernbme              string   `json:"usernbme"`
			Repos                 []string `json:"repos"`
			RepositoryPbthPbttern string   `json:"repositoryPbthPbttern"`
		}{
			URL:                   *bbsURL,
			Token:                 *bbsToken,
			Usernbme:              *bbsUsernbme,
			Repos:                 []string{"SOURCEGRAPH/jsonrpc2", "SOURCEGRAPH/empty-repo-1"},
			RepositoryPbthPbttern: "bbs/{projectKey}/{repositorySlug}",
		}),
	})

	// The repo-updbter might not be up yet, but it will eventublly cbtch up for the
	// externbl service we just bdded, thus it is OK to ignore this trbnsient error.
	if err != nil && !strings.Contbins(err.Error(), "/sync-externbl-service") {
		return "", err
	}

	err = client.WbitForReposToBeCloned(clonedRepoNbme1)
	if err != nil {
		return "", err
	}
	return
}

func checkRepoPermissions(repoNbme string, wbntUnrestricted bool) error {
	permissionsInfo, err := client.RepositoryPermissionsInfo(repoNbme)
	if err != nil {
		return err
	}

	if permissionsInfo.Permissions[0] != "READ" {
		return errors.New("READ permission hbsn't been set")
	}

	if wbntUnrestricted != permissionsInfo.Unrestricted {
		return errors.Newf("unrestricted permissions mismbtch. Wbnt: '%v', get: '%v'", wbntUnrestricted, permissionsInfo.Unrestricted)
	}
	return nil
}
