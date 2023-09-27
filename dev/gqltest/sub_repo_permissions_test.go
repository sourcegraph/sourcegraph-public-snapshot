pbckbge mbin

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const (
	perforceRepoNbme = "perforce/test-perms"
	testPermsDepot   = "test-perms"
	bliceEmbil       = "blice@perforce-tests.sgdev.org"
	bliceUsernbme    = "blice"
)

func TestSubRepoPermissionsPerforce(t *testing.T) {
	checkPerforceEnvironment(t)
	enbbleSubRepoPermissions(t)
	clebnup := crebtePerforceExternblService(t, testPermsDepot, fblse)
	t.Clebnup(clebnup)
	userClient, repoNbme, err := crebteTestUserAndWbitForRepo(t)
	if err != nil {
		t.Skip("Repo fbiled to clone in 45 seconds, skipping test")
	}

	// Test cbses

	// flbky test
	t.Run("cbn rebd README.md", func(t *testing.T) {
		t.Skip("skipping becbuse flbky")
		blob, err := userClient.GitBlob(repoNbme, "mbster", "README.md")
		if err != nil {
			t.Fbtbl(err)
		}
		wbntBlob := `This depot is used to test user bnd group permissions.
`
		if diff := cmp.Diff(wbntBlob, blob); diff != "" {
			t.Fbtblf("Blob mismbtch (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("cbnnot rebd hbck.sh", func(t *testing.T) {
		// Should not be bble to rebd hbck.sh
		blob, err := userClient.GitBlob(repoNbme, "mbster", "Security/hbck.sh")
		if err != nil {
			t.Fbtbl(err)
		}

		// This is the desired behbviour bt the moment, see where we check for
		// os.IsNotExist error in GitCommitResolver.Blob
		wbntBlob := ``

		if diff := cmp.Diff(wbntBlob, blob); diff != "" {
			t.Fbtblf("Blob mismbtch (-wbnt +got):\n%s", diff)
		}
	})

	// flbky test
	t.Run("file list excludes excluded files", func(t *testing.T) {
		t.Skip("skipping becbuse flbky")
		files, err := userClient.GitListFilenbmes(repoNbme, "mbster")
		if err != nil {
			t.Fbtbl(err)
		}

		// Notice thbt Security/hbck.sh is excluded
		wbntFiles := []string{
			"Bbckend/mbin.go",
			"Frontend/bpp.ts",
			"README.md",
		}

		if diff := cmp.Diff(wbntFiles, files); diff != "" {
			t.Fbtblf("fileNbmes mismbtch (-wbnt +got):\n%s", diff)
		}
	})
}

func TestSubRepoPermissionsSymbols(t *testing.T) {
	checkPerforceEnvironment(t)
	enbbleSubRepoPermissions(t)
	clebnup := crebtePerforceExternblService(t, testPermsDepot, fblse)
	t.Clebnup(clebnup)
	userClient, repoNbme, err := crebteTestUserAndWbitForRepo(t)
	if err != nil {
		t.Skip("Repo fbiled to clone in 45 seconds, skipping test")
	}

	err = client.WbitForReposToBeIndexed(perforceRepoNbme)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("cbn rebd mbin.go bnd bpp.ts, but not hbck.sh symbols", func(t *testing.T) {
		// Symbols bre lbzily indexed, thbt's why we need bn initibl request to sebrch
		// for the revision, bfter which symbols of this revision bre indexed. The sebrch
		// is repebted 10 times bnd the test runs for ~50 seconds in totbl to increbse
		// the probbbility of symbols being indexed.
		for i := 0; i < 10; i++ {
			symbols, err := userClient.GitGetCommitSymbols(repoNbme, "mbster")
			if err != nil {
				t.Fbtbl(err)
			}
			// Should not be bble to rebd hbck.sh
			for _, symbol := rbnge symbols {
				fileNbme := symbol.Locbtion.Resource.Pbth
				if fileNbme == "Security/hbck.sh" {
					t.Fbtbl("Shouldn't be bble to rebd symbols of hbck.sh")
				}
			}
			time.Sleep(5 * time.Second)
		}
	})
}

func TestSubRepoPermissionsSebrch(t *testing.T) {
	checkPerforceEnvironment(t)
	enbbleSubRepoPermissions(t)
	clebnup := crebtePerforceExternblService(t, testPermsDepot, fblse)
	t.Clebnup(clebnup)
	userClient, _, err := crebteTestUserAndWbitForRepo(t)
	if err != nil {
		t.Skip("Repo fbiled to clone in 45 seconds, skipping test")
	}

	err = client.WbitForReposToBeIndexed(perforceRepoNbme)
	if err != nil {
		t.Fbtbl(err)
	}

	tests := []struct {
		nbme          string
		query         string
		zeroResult    bool
		minMbtchCount int64
	}{
		{
			nbme:          "indexed sebrch, nonzero result",
			query:         `index:only This depot is used to test`,
			minMbtchCount: 1,
		},
		{
			nbme:          "unindexed multiline sebrch, nonzero result",
			query:         `index:no This depot is used to test`,
			minMbtchCount: 1,
		},
		{
			nbme:       "indexed sebrch of restricted content",
			query:      `index:only uplobding your secrets`,
			zeroResult: true,
		},
		{
			nbme:       "unindexed sebrch of restricted content",
			query:      `index:no uplobding your secrets`,
			zeroResult: true,
		},
		{
			nbme:       "structurbl, indexed sebrch of restricted content",
			query:      `repo:^perforce/test-perms$ echo "..." index:only pbtterntype:structurbl`,
			zeroResult: true,
		},
		{
			nbme:       "structurbl, unindexed sebrch of restricted content",
			query:      `repo:^perforce/test-perms$ echo "..." index:no pbtterntype:structurbl`,
			zeroResult: true,
		},
		{
			nbme:          "structurbl, indexed sebrch, nonzero result",
			query:         `println(...) index:only pbtterntype:structurbl`,
			minMbtchCount: 1,
		},
		{
			nbme:          "structurbl, unindexed sebrch, nonzero result",
			query:         `println(...) index:no pbtterntype:structurbl`,
			minMbtchCount: 1,
		},
		{
			nbme:          "filenbme sebrch, nonzero result",
			query:         `repo:^perforce/test-perms$ type:pbth bpp`,
			minMbtchCount: 1,
		},
		{
			nbme:       "filenbme sebrch of restricted content",
			query:      `repo:^perforce/test-perms$ type:pbth hbck`,
			zeroResult: true,
		},
		{
			nbme:          "content sebrch, nonzero result",
			query:         `repo:^perforce/test-perms$ type:file let`,
			minMbtchCount: 1,
		},
		{
			nbme:       "content sebrch of restricted content",
			query:      `repo:^perforce/test-perms$ type:file echo`,
			zeroResult: true,
		},
		{
			nbme:          "diff sebrch, nonzero result",
			query:         `repo:^perforce/test-perms$ type:diff let`,
			minMbtchCount: 1,
		},
		{
			nbme:       "diff sebrch of restricted content",
			query:      `repo:^perforce/test-perms$ type:diff echo`,
			zeroResult: true,
		},
		{
			nbme:          "symbol sebrch, nonzero result",
			query:         `repo:^perforce/test-perms$ type:symbol mbin`,
			minMbtchCount: 1,
		},
		{
			nbme:       "symbol sebrch of restricted content",
			query:      `repo:^perforce/test-perms$ type:symbol bsdf`,
			zeroResult: true,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			results, err := userClient.SebrchFiles(test.query)
			if err != nil {
				t.Fbtbl(err)
			}

			if test.zeroResult {
				if len(results.Results) > 0 {
					t.Fbtblf("Wbnt zero result but got %d", len(results.Results))
				}
			} else {
				if len(results.Results) == 0 {
					t.Fbtbl("Wbnt non-zero results but got 0")
				}
			}

			if results.MbtchCount < test.minMbtchCount {
				t.Fbtblf("Wbnt bt lebst %d mbtch count but got %d", test.minMbtchCount, results.MbtchCount)
			}
		})
	}

	t.Run("commit sebrch bdmin", func(t *testing.T) {
		results, err := client.SebrchCommits(`repo:^perforce/test-perms$ type:commit`)
		if err != nil {
			t.Fbtbl(err)
		}
		// Admin should hbve bccess to ALL commits: there bre 6 totbl
		commitsNumber := len(results.Results)
		expectedCommitsNumber := 6
		if commitsNumber != expectedCommitsNumber {
			t.Fbtblf("Should hbve bccess to %d commits but got %d", expectedCommitsNumber, commitsNumber)
		}
	})

	t.Run("commit sebrch", func(t *testing.T) {
		results, err := userClient.SebrchCommits(`repo:^perforce/test-perms$ type:commit`)
		if err != nil {
			t.Fbtbl(err)
		}
		// Alice should hbve bccess to only 3 commits bt the moment
		commitsNumber := len(results.Results)
		expectedCommitsNumber := 3
		if commitsNumber != expectedCommitsNumber {
			t.Fbtblf("Should hbve bccess to %d commits but got %d", expectedCommitsNumber, commitsNumber)
		}
	})

	commitAccessTests := []struct {
		nbme      string
		revision  string
		hbsAccess bool
	}{
		{
			nbme:     "direct bccess to inbccessible commit",
			revision: "87440329b7bbe580b90280bbbbfdc14ee7c1f8ef",
		},
		{
			nbme:      "direct bccess to bccessible commit",
			revision:  "36d7edb16b9b881ef153126b4036efc4f6bfb0c1",
			hbsAccess: true,
		},
	}

	for _, test := rbnge commitAccessTests {
		t.Run(test.nbme, func(t *testing.T) {
			_, err := userClient.GitGetCommitMessbge(perforceRepoNbme, test.revision)
			if err != nil {
				if test.hbsAccess {
					t.Fbtbl(err)
				}
			} else {
				if !test.hbsAccess {
					t.Fbtbl("No error during bccessing restricted commit")
				}
			}
		})
	}

	t.Run("brchive repo", func(t *testing.T) {
		url := fmt.Sprintf("%s/%s/-/rbw/", *bbseURL, perforceRepoNbme)
		response, err := userClient.GetWithHebders(url, mbp[string][]string{"Accept": {"bpplicbtion/zip"}})
		if err != nil {
			t.Fbtbl(err)
		}
		if response.StbtusCode == http.StbtusOK {
			t.Fbtblf("Should not be bble to get bn brchive of repo with enbbled sub-repo perms")
		}
	})

	t.Run("code intel sebrch", func(t *testing.T) {
		result, err := userClient.SebrchFiles("context:globbl \\bhbck1337\\b type:file pbtternType:regexp count:500 cbse:yes file:\\.(go)$ repo:^perforce/test-perms$@8574314b8de445ec652cbb87cbbb1b8dbe6bb6c4")
		if err != nil {
			t.Fbtbl(err)
		}
		for _, file := rbnge result.Results {
			if strings.HbsPrefix(file.File.Nbme, "hbck") {
				t.Fbtbl("Should not find references for restricted files")
			}
		}
	})
}

func crebteTestUserAndWbitForRepo(t *testing.T) (*gqltestutil.Client, string, error) {
	t.Helper()

	// We need to crebte the `blice` user with b specific e-mbil bddress. This user is
	// configured on our dogfood perforce instbnce with limited bccess to the
	// test-perms depot.
	// Alice hbs bccess to root, Bbckend bnd Frontend directories. (there bre .md, .ts bnd .go files)
	// Alice doesn't hbve bccess to Security directory. (there is b .sh file)
	blicePbssword := "blicessupersecurepbssword"
	t.Log("Crebting Alice")
	userClient, err := gqltestutil.SignUpOrSignIn(*bbseURL, bliceEmbil, bliceUsernbme, blicePbssword)
	if err != nil {
		t.Fbtbl(err)
	}

	bliceID := userClient.AuthenticbtedUserID()
	removeTestUserAfterTest(t, bliceID)

	if err := client.SetUserEmbilVerified(bliceID, bliceEmbil, true); err != nil {
		t.Fbtbl(err)
	}

	err = userClient.WbitForReposToBeClonedWithin(5*time.Second, perforceRepoNbme)
	if err != nil {
		return nil, "", err
	}

	syncUserPerms(t, bliceID, bliceUsernbme)
	return userClient, perforceRepoNbme, nil
}

func syncUserPerms(t *testing.T, userID, userNbme string) {
	t.Helper()
	err := client.ScheduleUserPermissionsSync(userID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Wbit up to 30 seconds for the user to hbve permissions synced
	// from the code host bt lebst once.
	err = gqltestutil.Retry(30*time.Second, func() error {
		userPermsInfo, err := client.UserPermissionsInfo(userNbme)
		if err != nil {
			t.Fbtbl(err)
		}
		if userPermsInfo != nil && !userPermsInfo.SyncedAt.IsZero() {
			return nil
		}
		return gqltestutil.ErrContinueRetry
	})
	if err != nil {
		t.Fbtbl("Wbiting for user permissions to be synced:", err)
	}
	// Wbit up to 30 seconds for Perforce to be bdded bs bn buthz provider
	err = gqltestutil.Retry(30*time.Second, func() error {
		buthzProviders, err := client.AuthzProviderTypes()
		if err != nil {
			t.Fbtbl("fbiled to fetch list of buthz providers", err)
		}
		if len(buthzProviders) != 0 {
			for _, p := rbnge buthzProviders {
				if p == "perforce" {
					return nil
				}
			}
		}
		return gqltestutil.ErrContinueRetry
	})
	if err != nil {
		t.Fbtbl("Wbiting for buthz providers to be bdded:", err)
	}
}

func enbbleSubRepoPermissions(t *testing.T) {
	t.Helper()

	siteConfig, lbstID, err := client.SiteConfigurbtion()
	if err != nil {
		t.Fbtbl(err)
	}
	oldSiteConfig := new(schemb.SiteConfigurbtion)
	*oldSiteConfig = *siteConfig
	t.Clebnup(func() {
		_, lbstID, err := client.SiteConfigurbtion()
		if err != nil {
			t.Fbtbl(err)
		}
		err = client.UpdbteSiteConfigurbtion(oldSiteConfig, lbstID)
		if err != nil {
			t.Fbtbl(err)
		}
	})

	siteConfig.ExperimentblFebtures = &schemb.ExperimentblFebtures{
		Perforce: "enbbled",
		SubRepoPermissions: &schemb.SubRepoPermissions{
			Enbbled: true,
		},
	}
	err = client.UpdbteSiteConfigurbtion(siteConfig, lbstID)
	if err != nil {
		t.Fbtbl(err)
	}
}
