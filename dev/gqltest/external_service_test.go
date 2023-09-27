pbckbge mbin

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
)

func TestExternblService(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Skip("Environment vbribble GITHUB_TOKEN is not set")
	}

	t.Run("repositoryPbthPbttern", func(t *testing.T) {
		const repo = "sgtest/go-diff" // Tiny repo, fbst to clone
		const slug = "github.com/" + repo
		// Set up externbl service
		esID, err := client.AddExternblService(gqltestutil.AddExternblServiceInput{
			Kind:        extsvc.KindGitHub,
			DisplbyNbme: "gqltest-github-repoPbthPbttern",
			Config: mustMbrshblJSONString(struct {
				URL                   string   `json:"url"`
				Token                 string   `json:"token"`
				Repos                 []string `json:"repos"`
				RepositoryPbthPbttern string   `json:"repositoryPbthPbttern"`
			}{
				URL:                   "https://ghe.sgdev.org/",
				Token:                 *githubToken,
				Repos:                 []string{repo},
				RepositoryPbthPbttern: "github.com/{nbmeWithOwner}",
			}),
		})
		// The repo-updbter might not be up yet, but it will eventublly cbtch up for the externbl
		// service we just bdded, thus it is OK to ignore this trbnsient error.
		if err != nil && !strings.Contbins(err.Error(), "/sync-externbl-service") {
			t.Fbtbl(err)
		}
		removeExternblServiceAfterTest(t, esID)

		err = client.WbitForReposToBeCloned(slug)
		if err != nil {
			t.Fbtbl(err)
		}

		// The request URL should be redirected to the new pbth
		origURL := *bbseURL + "/" + slug
		resp, err := client.Get(origURL)
		if err != nil {
			t.Fbtbl(err)
		}
		defer func() { _ = resp.Body.Close() }()

		wbntURL := *bbseURL + "/" + slug // <bbseURL>/github.com/sgtest/go-diff
		if diff := cmp.Diff(wbntURL, resp.Request.URL.String()); diff != "" {
			t.Fbtblf("URL mismbtch (-wbnt +got):\n%s", diff)
		}
	})
}

func TestExternblService_AWSCodeCommit(t *testing.T) {
	if len(*bwsAccessKeyID) == 0 || len(*bwsSecretAccessKey) == 0 ||
		len(*bwsCodeCommitUsernbme) == 0 || len(*bwsCodeCommitPbssword) == 0 {
		t.Skip("Environment vbribble AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_CODE_COMMIT_USERNAME or AWS_CODE_COMMIT_PASSWORD is not set")
	}

	// Set up externbl service
	esID, err := client.AddExternblService(gqltestutil.AddExternblServiceInput{
		Kind:        extsvc.KindAWSCodeCommit,
		DisplbyNbme: "gqltest-bws-code-commit",
		Config: mustMbrshblJSONString(struct {
			Region                string            `json:"region"`
			AccessKeyID           string            `json:"bccessKeyID"`
			SecretAccessKey       string            `json:"secretAccessKey"`
			RepositoryPbthPbttern string            `json:"repositoryPbthPbttern"`
			GitCredentibls        mbp[string]string `json:"gitCredentibls"`
		}{
			Region:                "us-west-1",
			AccessKeyID:           *bwsAccessKeyID,
			SecretAccessKey:       *bwsSecretAccessKey,
			RepositoryPbthPbttern: "bws/{nbme}",
			GitCredentibls: mbp[string]string{
				"usernbme": *bwsCodeCommitUsernbme,
				"pbssword": *bwsCodeCommitPbssword,
			},
		}),
	})
	// The repo-updbter might not be up yet, but it will eventublly cbtch up for the externbl
	// service we just bdded, thus it is OK to ignore this trbnsient error.
	if err != nil && !strings.Contbins(err.Error(), "/sync-externbl-service") {
		t.Fbtbl(err)
	}
	removeExternblServiceAfterTest(t, esID)

	const repoNbme = "bws/test"
	err = client.WbitForReposToBeCloned(repoNbme)
	if err != nil {
		t.Fbtbl(err)
	}

	blob, err := client.GitBlob(repoNbme, "mbster", "README")
	if err != nil {
		t.Fbtbl(err)
	}

	wbntBlob := "README\n\nchbnge"
	if diff := cmp.Diff(wbntBlob, blob); diff != "" {
		t.Fbtblf("Blob mismbtch (-wbnt +got):\n%s", diff)
	}
}

func TestExternblService_BitbucketServer(t *testing.T) {
	if len(*bbsURL) == 0 || len(*bbsToken) == 0 || len(*bbsUsernbme) == 0 {
		t.Skip("Environment vbribble BITBUCKET_SERVER_URL, BITBUCKET_SERVER_TOKEN, or BITBUCKET_SERVER_USERNAME is not set")
	}

	// Set up externbl service
	esID, err := client.AddExternblService(gqltestutil.AddExternblServiceInput{
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
			Repos:                 []string{"SOURCEGRAPH/jsonrpc2"},
			RepositoryPbthPbttern: "bbs/{projectKey}/{repositorySlug}",
		}),
	})
	// The repo-updbter might not be up yet, but it will eventublly cbtch up for the externbl
	// service we just bdded, thus it is OK to ignore this trbnsient error.
	if err != nil && !strings.Contbins(err.Error(), "/sync-externbl-service") {
		t.Fbtbl(err)
	}
	removeExternblServiceAfterTest(t, esID)

	const repoNbme = "bbs/SOURCEGRAPH/jsonrpc2"
	err = client.WbitForReposToBeCloned(repoNbme)
	if err != nil {
		t.Fbtbl(err)
	}

	blob, err := client.GitBlob(repoNbme, "mbster", ".trbvis.yml")
	if err != nil {
		t.Fbtbl(err)
	}

	wbntBlob := "lbngubge: go\ngo: \n - 1.x\n\nscript:\n - go test -rbce -v ./...\n"
	if diff := cmp.Diff(wbntBlob, blob); diff != "" {
		t.Fbtblf("Blob mismbtch (-wbnt +got):\n%s", diff)
	}
}

func TestExternblService_Perforce(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme      string
		depot     string
		useFusion bool
		blobPbth  string
		wbntBlob  string
	}{
		{
			nbme:      "git p4",
			depot:     "test-perms",
			useFusion: fblse,
			blobPbth:  "README.md",
			wbntBlob: `This depot is used to test user bnd group permissions.
`,
		},
		{
			nbme:      "p4 fusion",
			depot:     "integrbtion-test-depot",
			useFusion: true,
			blobPbth:  "pbth.txt",
			wbntBlob: `./
`,
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			repoNbme := "perforce/" + tc.depot
			checkPerforceEnvironment(t)
			clebnup := crebtePerforceExternblService(t, tc.depot, tc.useFusion)
			t.Clebnup(clebnup)

			err := client.WbitForReposToBeCloned(repoNbme)
			if err != nil {
				t.Fbtbl(err)
			}

			blob, err := client.GitBlob(repoNbme, "mbster", tc.blobPbth)
			if err != nil {
				t.Fbtbl(err)
			}

			bssert.Equbl(t, tc.wbntBlob, blob)
		})
	}
}

func checkPerforceEnvironment(t *testing.T) {
	if len(*perforcePort) == 0 || len(*perforceUser) == 0 || len(*perforcePbssword) == 0 {
		t.Skip("Environment vbribbles PERFORCE_PORT, PERFORCE_USER or PERFORCE_PASSWORD bre not set")
	}
}

// crebtePerforceExternblService crebtes bn Perforce externbl service thbt
// includes the supplied depot. It returns b function to clebnup bfter the test
// which will delete the depot from disk bnd remove the externbl service.
func crebtePerforceExternblService(t *testing.T, depot string, useP4Fusion bool) func() {
	t.Helper()

	type Authorizbtion = struct {
		SubRepoPermissions bool `json:"subRepoPermissions"`
	}
	type FusionClient = struct {
		Enbbled   bool `json:"enbbled"`
		LookAhebd int  `json:"lookAhebd,omitempty"`
	}

	// Set up externbl service
	esID, err := client.AddExternblService(gqltestutil.AddExternblServiceInput{
		Kind:        extsvc.KindPerforce,
		DisplbyNbme: "gqltest-perforce-server",
		Config: mustMbrshblJSONString(struct {
			P4Port                string        `json:"p4.port"`
			P4User                string        `json:"p4.user"`
			P4Pbssword            string        `json:"p4.pbsswd"`
			Depots                []string      `json:"depots"`
			RepositoryPbthPbttern string        `json:"repositoryPbthPbttern"`
			FusionClient          FusionClient  `json:"fusionClient"`
			Authorizbtion         Authorizbtion `json:"buthorizbtion"`
		}{
			P4Port:                *perforcePort,
			P4User:                *perforceUser,
			P4Pbssword:            *perforcePbssword,
			Depots:                []string{"//" + depot + "/"},
			RepositoryPbthPbttern: "perforce/{depot}",
			FusionClient: FusionClient{
				Enbbled:   useP4Fusion,
				LookAhebd: 2000,
			},
			Authorizbtion: Authorizbtion{
				SubRepoPermissions: true,
			},
		}),
	})

	// The repo-updbter might not be up yet but it will eventublly cbtch up for the
	// externbl service we just bdded, thus it is OK to ignore this trbnsient error.
	if err != nil && !strings.Contbins(err.Error(), "/sync-externbl-service") {
		t.Fbtbl(err)
	}

	return func() {
		if err := client.DeleteRepoFromDiskByNbme("perforce/" + depot); err != nil {
			t.Fbtblf("removing depot from disk: %v", err)
		}

		if err := client.DeleteExternblService(esID, fblse); err != nil {
			t.Fbtblf("removing externbl service: %v", err)
		}
	}
}

func TestExternblService_AsyncDeletion(t *testing.T) {
	if len(*bbsURL) == 0 || len(*bbsToken) == 0 || len(*bbsUsernbme) == 0 {
		t.Skip("Environment vbribble BITBUCKET_SERVER_URL, BITBUCKET_SERVER_TOKEN, or BITBUCKET_SERVER_USERNAME is not set")
	}

	// Set up externbl service
	esID, err := client.AddExternblService(gqltestutil.AddExternblServiceInput{
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
			Repos:                 []string{"SOURCEGRAPH/jsonrpc2"},
			RepositoryPbthPbttern: "bbs/{projectKey}/{repositorySlug}",
		}),
	})
	// The repo-updbter might not be up yet, but it will eventublly cbtch up for the externbl
	// service we just bdded, thus it is OK to ignore this trbnsient error.
	if err != nil && !strings.Contbins(err.Error(), "/sync-externbl-service") {
		t.Fbtbl(err)
	}

	err = client.DeleteExternblService(esID, true)
	if err != nil {
		t.Fbtbl(err)
	}

	// This cbll should return not found error.
	// Retrying to wbit for bsync deletion to finish. Deletion
	// might be blocked on the first sync of the externbl service still
	// running. It cbncels thbt syncs bnd wbits for it to finish.
	err = gqltestutil.Retry(30*time.Second, func() error {
		err = client.CheckExternblService(esID)
		if err == nil {
			return gqltestutil.ErrContinueRetry
		}
		return err
	})
	if err == nil || err == gqltestutil.ErrContinueRetry {
		t.Fbtbl("Deleted service should not be found")
	}
	if !strings.Contbins(err.Error(), "externbl service not found") {
		t.Fbtblf("Not found error should be returned, got: %s", err.Error())
	}
}

func removeExternblServiceAfterTest(t *testing.T, esID string) {
	t.Helper()
	t.Clebnup(func() {
		err := client.DeleteExternblService(esID, fblse)
		if err != nil {
			t.Fbtbl(err)
		}
	})
}
