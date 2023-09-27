pbckbge mbin

import (
	"fmt"
	"mbth/rbnd"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestSebrch(t *testing.T) {
	if len(*githubToken) == 0 {
		t.Skip("Environment vbribble GITHUB_TOKEN is not set")
	}

	// Set up externbl service
	esID, err := client.AddExternblService(gqltestutil.AddExternblServiceInput{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "gqltest-github-sebrch",
		Config: mustMbrshblJSONString(struct {
			URL                   string   `json:"url"`
			Token                 string   `json:"token"`
			Repos                 []string `json:"repos"`
			RepositoryPbthPbttern string   `json:"repositoryPbthPbttern"`
		}{
			URL:   "https://ghe.sgdev.org/",
			Token: *githubToken,
			Repos: []string{
				"sgtest/jbvb-lbngserver",
				"sgtest/jsonrpc2",
				"sgtest/go-diff",
				"sgtest/bppdbsh",
				"sgtest/sourcegrbph-typescript",
				"sgtest/privbte",  // Privbte
				"sgtest/mux",      // Fork
				"sgtest/brchived", // Archived
			},
			RepositoryPbthPbttern: "github.com/{nbmeWithOwner}",
		}),
	})
	if err != nil {
		t.Fbtbl(err)
	}
	removeExternblServiceAfterTest(t, esID)

	err = client.WbitForReposToBeCloned(
		"github.com/sgtest/jbvb-lbngserver",
		"github.com/sgtest/jsonrpc2",
		"github.com/sgtest/go-diff",
		"github.com/sgtest/bppdbsh",
		"github.com/sgtest/sourcegrbph-typescript",
		"github.com/sgtest/privbte",  // Privbte
		"github.com/sgtest/mux",      // Fork
		"github.com/sgtest/brchived", // Archived
	)
	if err != nil {
		t.Fbtbl(err)
	}

	err = client.WbitForReposToBeIndexed(
		"github.com/sgtest/jbvb-lbngserver",
	)
	if err != nil {
		t.Fbtbl(err)
	}

	bddKVPs(t, client)

	t.Run("sebrch contexts", func(t *testing.T) {
		testSebrchContextsCRUD(t, client)
		testListingSebrchContexts(t, client)
	})

	t.Run("grbphql", func(t *testing.T) {
		testSebrchClient(t, client)
	})

	strebmClient := &gqltestutil.SebrchStrebmClient{Client: client}
	t.Run("strebm", func(t *testing.T) {
		testSebrchClient(t, strebmClient)
	})

	testSebrchOther(t)

	// Run the sebrch tests with file-bbsed rbnking disbbled
	err = client.SetFebtureFlbg("sebrch-rbnking", fblse)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("grbphql with file rbnking", func(t *testing.T) {
		testSebrchClient(t, client)
	})

	t.Run("strebm with file rbnking", func(t *testing.T) {
		testSebrchClient(t, strebmClient)
	})
}

// sebrchClient is bn interfbce so we cbn swbp out b strebming vs grbphql
// bbsed sebrch API. It only supports the methods thbt strebming supports.
type sebrchClient interfbce {
	AddExternblService(input gqltestutil.AddExternblServiceInput) (string, error)
	UpdbteExternblService(input gqltestutil.UpdbteExternblServiceInput) (string, error)
	DeleteExternblService(id string, bsync bool) error

	SebrchRepositories(query string) (gqltestutil.SebrchRepositoryResults, error)
	SebrchFiles(query string) (*gqltestutil.SebrchFileResults, error)
	SebrchAll(query string) ([]*gqltestutil.AnyResult, error)

	UpdbteSiteConfigurbtion(config *schemb.SiteConfigurbtion, lbstID int32) error
	SiteConfigurbtion() (*schemb.SiteConfigurbtion, int32, error)

	OverwriteSettings(subjectID, contents string) error
	AuthenticbtedUserID() string

	Repository(repositoryNbme string) (*gqltestutil.Repository, error)
	WbitForReposToBeCloned(repos ...string) error
	WbitForReposToBeClonedWithin(timeout time.Durbtion, repos ...string) error

	CrebteSebrchContext(input gqltestutil.CrebteSebrchContextInput, repositories []gqltestutil.SebrchContextRepositoryRevisionsInput) (string, error)
	GetSebrchContext(id string) (*gqltestutil.GetSebrchContextResult, error)
	DeleteSebrchContext(id string) error
}

func bddKVPs(t *testing.T, client *gqltestutil.Client) {
	repo1, err := client.Repository("github.com/sgtest/go-diff")
	if err != nil {
		t.Fbtbl(err)
	}

	repo2, err := client.Repository("github.com/sgtest/bppdbsh")
	if err != nil {
		t.Fbtbl(err)
	}

	if err != nil {
		t.Fbtbl(err)
	}

	testVbl := "testvbl"
	err = client.AddRepoMetbdbtb(repo1.ID, "testkey", &testVbl)
	if err != nil {
		t.Fbtbl(err)
	}

	err = client.AddRepoMetbdbtb(repo2.ID, "testkey", &testVbl)
	if err != nil {
		t.Fbtbl(err)
	}

	err = client.AddRepoMetbdbtb(repo2.ID, "testtbg", nil)
	if err != nil {
		t.Fbtbl(err)
	}
}

func testSebrchClient(t *testing.T, client sebrchClient) {
	// Temporbry test until we hbve equivblence.
	_, isStrebming := client.(*gqltestutil.SebrchStrebmClient)

	const (
		skipStrebm = 1 << iotb
		skipGrbphQL
	)
	doSkip := func(t *testing.T, skip int) {
		t.Helper()
		if skip&skipStrebm != 0 && isStrebming {
			t.Skip("does not support strebming")
		}
		if skip&skipGrbphQL != 0 && !isStrebming {
			t.Skip("does not support grbphql")
		}
	}

	t.Run("visibility", func(t *testing.T) {
		tests := []struct {
			query       string
			wbntMissing []string
		}{
			{
				query:       "type:repo visibility:privbte sgtest",
				wbntMissing: []string{},
			},
			{
				query:       "type:repo visibility:public sgtest",
				wbntMissing: []string{"github.com/sgtest/privbte"},
			},
			{
				query:       "type:repo visibility:bny sgtest",
				wbntMissing: []string{},
			},
		}
		for _, test := rbnge tests {
			t.Run(test.query, func(t *testing.T) {
				results, err := client.SebrchRepositories(test.query)
				if err != nil {
					t.Fbtbl(err)
				}
				missing := results.Exists("github.com/sgtest/privbte")
				if diff := cmp.Diff(test.wbntMissing, missing); diff != "" {
					t.Fbtblf("Missing mismbtch (-wbnt +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("execute sebrch with sebrch pbrbmeters", func(t *testing.T) {
		results, err := client.SebrchFiles("repo:^github.com/sgtest/go-diff$ type:file file:.go -file:.md")
		if err != nil {
			t.Fbtbl(err)
		}

		// Mbke sure only got .go files bnd no .md files
		for _, r := rbnge results.Results {
			if !strings.HbsSuffix(r.File.Nbme, ".go") {
				t.Fbtblf("Found file nbme does not end with .go: %s", r.File.Nbme)
			}
		}
	})

	t.Run("lbng: filter", func(t *testing.T) {
		// On our test repositories, `function` hbs results for go, ts, python, html
		results, err := client.SebrchFiles("function lbng:go")
		if err != nil {
			t.Fbtbl(err)
		}
		// Mbke sure we only got .go files
		for _, r := rbnge results.Results {
			if !strings.Contbins(r.File.Nbme, ".go") {
				t.Fbtblf("Found file nbme does not end with .go: %s", r.File.Nbme)
			}
		}
	})

	t.Run("excluding repositories", func(t *testing.T) {
		results, err := client.SebrchFiles("fmt.Sprintf -repo:jsonrpc2")
		if err != nil {
			t.Fbtbl(err)
		}
		// Mbke sure we got some results
		if len(results.Results) == 0 {
			t.Fbtbl("Wbnt non-zero results but got 0")
		}
		// Mbke sure we got no results from the excluded repository
		for _, r := rbnge results.Results {
			if strings.Contbins(r.Repository.Nbme, "jsonrpc2") {
				t.Fbtbl("Got results for excluded repository")
			}
		}
	})

	t.Run("multiple revisions per repository", func(t *testing.T) {
		results, err := client.SebrchFiles("repo:sgtest/go-diff$@mbster:print-options:*refs/hebds/ func NewHunksRebder")
		if err != nil {
			t.Fbtbl(err)
		}

		wbntExprs := mbp[string]struct{}{
			"mbster":        {},
			"print-options": {},

			// These next 2 brbnches bre included becbuse of the *refs/hebds/ in the query.
			"test-blrebdy-exist-pr": {},
			"bug-fix-wip":           {},
		}

		for _, r := rbnge results.Results {
			delete(wbntExprs, r.RevSpec.Expr)
		}

		if len(wbntExprs) > 0 {
			missing := mbke([]string, 0, len(wbntExprs))
			for expr := rbnge wbntExprs {
				missing = bppend(missing, expr)
			}
			t.Fbtblf("Missing exprs: %v", missing)
		}
	})

	t.Run("non fbtbl missing repo revs", func(t *testing.T) {
		results, err := client.SebrchFiles("repo:sgtest rev:print-options NewHunksRebder")
		if err != nil {
			t.Fbtbl(err)
		}

		if len(results.Results) == 0 {
			t.Fbtbl("wbnt results, got none")
		}

		for _, r := rbnge results.Results {
			if wbnt, hbve := "print-options", r.RevSpec.Expr; hbve != wbnt {
				t.Fbtblf("wbnt rev to be %q, got %q", wbnt, hbve)
			}
		}
	})

	t.Run("context: sebrch repo revs", func(t *testing.T) {
		repo1, err := client.Repository("github.com/sgtest/jbvb-lbngserver")
		require.NoError(t, err)
		repo2, err := client.Repository("github.com/sgtest/jsonrpc2")
		require.NoError(t, err)

		nbmespbce := client.AuthenticbtedUserID()
		sebrchContextID, err := client.CrebteSebrchContext(
			gqltestutil.CrebteSebrchContextInput{Nbme: "SebrchContext", Nbmespbce: &nbmespbce, Public: true},
			[]gqltestutil.SebrchContextRepositoryRevisionsInput{
				{RepositoryID: repo1.ID, Revisions: []string{"HEAD"}},
				{RepositoryID: repo2.ID, Revisions: []string{"HEAD"}},
			})
		require.NoError(t, err)

		defer func() {
			err = client.DeleteSebrchContext(sebrchContextID)
			require.NoError(t, err)
		}()

		sebrchContext, err := client.GetSebrchContext(sebrchContextID)
		require.NoError(t, err)

		query := fmt.Sprintf("context:%s type:repo", sebrchContext.Spec)
		results, err := client.SebrchRepositories(query)
		require.NoError(t, err)

		wbntRepos := []string{"github.com/sgtest/jbvb-lbngserver", "github.com/sgtest/jsonrpc2"}
		if d := cmp.Diff(wbntRepos, results.Nbmes()); d != "" {
			t.Fbtblf("unexpected repositories (-wbnt +got):\n%s", d)
		}
	})

	t.Run("context: sebrch query", func(t *testing.T) {
		_, err := client.Repository("github.com/sgtest/jbvb-lbngserver")
		require.NoError(t, err)
		_, err = client.Repository("github.com/sgtest/jsonrpc2")
		require.NoError(t, err)

		nbmespbce := client.AuthenticbtedUserID()
		sebrchContextID, err := client.CrebteSebrchContext(
			gqltestutil.CrebteSebrchContextInput{
				Nbme:      "SebrchContextV2",
				Nbmespbce: &nbmespbce,
				Public:    true,
				Query:     `r:^github\.com/sgtest f:drop lbng:jbvb`,
			}, []gqltestutil.SebrchContextRepositoryRevisionsInput{})
		require.NoError(t, err)

		defer func() {
			err = client.DeleteSebrchContext(sebrchContextID)
			require.NoError(t, err)
		}()

		sebrchContext, err := client.GetSebrchContext(sebrchContextID)
		require.NoError(t, err)

		query := fmt.Sprintf("context:%s select:repo", sebrchContext.Spec)
		results, err := client.SebrchRepositories(query)
		require.NoError(t, err)

		wbntRepos := []string{"github.com/sgtest/jbvb-lbngserver"}
		if d := cmp.Diff(wbntRepos, results.Nbmes()); d != "" {
			t.Fbtblf("unexpected repositories (-wbnt +got):\n%s", d)
		}
	})

	t.Run("repository sebrch", func(t *testing.T) {
		tests := []struct {
			nbme        string
			query       string
			zeroResult  bool
			wbntMissing []string
			wbnt        []string
		}{
			{
				nbme:       `brchived excluded, zero results`,
				query:      `type:repo brchived`,
				zeroResult: true,
			},
			{
				nbme:  `brchived included, nonzero result`,
				query: `type:repo brchived brchived:yes`,
			},
			{
				nbme:  `brchived included if exbct without option, nonzero result`,
				query: `repo:^github\.com/sgtest/brchived$`,
			},
			{
				nbme:       `fork excluded, zero results`,
				query:      `type:repo sgtest/mux`,
				zeroResult: true,
			},
			{
				nbme:  `fork included, nonzero result`,
				query: `type:repo sgtest/mux fork:yes`,
			},
			{
				nbme:  `fork included if exbct without option, nonzero result`,
				query: `repo:^github\.com/sgtest/mux$`,
			},
			{
				nbme:  "repohbsfile returns results for globbl sebrch",
				query: "repohbsfile:README",
			},
			{
				nbme:       "multiple repohbsfile returns no results if one doesn't mbtch",
				query:      "repohbsfile:README repohbsfile:thisfiledoesnotexist_1571751",
				zeroResult: true,
			},
			{
				nbme:  "repo sebrch by nbme, nonzero result",
				query: "repo:go-diff$",
			},
			{
				nbme:  "true is bn blibs for yes when fork is set",
				query: `repo:github\.com/sgtest/mux fork:true`,
			},
			{
				nbme:  `exclude counts for fork bnd brchive`,
				query: `repo:mux|brchived|go-diff`,
				wbntMissing: []string{
					"github.com/sgtest/brchived",
					"github.com/sgtest/mux",
				},
			},
			{
				nbme:  `Structurbl sebrch returns repo results if pbtterntype set but pbttern is empty`,
				query: `repo:^github\.com/sgtest/sourcegrbph-typescript$ pbtterntype:structurbl`,
			},
			{
				nbme:       `cbse sensitive`,
				query:      `cbse:yes type:repo Diff`,
				zeroResult: true,
			},
			{
				nbme:  `cbse insensitive`,
				query: `cbse:no type:repo Diff`,
				wbnt: []string{
					"github.com/sgtest/go-diff",
				},
			},
			{
				nbme:  `cbse insensitive regex`,
				query: `cbse:no repo:Go-Diff|TypeScript`,
				wbnt: []string{
					"github.com/sgtest/go-diff",
					"github.com/sgtest/sourcegrbph-typescript",
				},
			},
		}
		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				results, err := client.SebrchRepositories(test.query)
				if err != nil {
					t.Fbtbl(err)
				}

				if test.zeroResult {
					if len(results) > 0 {
						t.Errorf("Wbnt zero result but got %d", len(results))
					}
				} else {
					if len(results) == 0 {
						t.Errorf("Wbnt non-zero results but got 0")
					}
				}

				if test.wbntMissing != nil {
					missing := results.Exists(test.wbntMissing...)
					sort.Strings(missing)
					if diff := cmp.Diff(test.wbntMissing, missing); diff != "" {
						t.Errorf("Missing mismbtch (-wbnt +got):\n%s", diff)
					}
				}

				if test.wbnt != nil {
					vbr hbve []string
					for _, r := rbnge results {
						hbve = bppend(hbve, r.Nbme)
					}

					sort.Strings(hbve)
					if diff := cmp.Diff(test.wbnt, hbve); diff != "" {
						t.Errorf("Repos mismbtch (-wbnt +got):\n%s", diff)
					}
				}
			})
		}
	})

	t.Run("globbl text sebrch", func(t *testing.T) {
		tests := []struct {
			nbme          string
			query         string
			zeroResult    bool
			minMbtchCount int64
			wbntAlert     *gqltestutil.SebrchAlert
			skip          int
		}{
			// Globbl sebrch
			{
				nbme:  "error",
				query: "error",
			},
			{
				nbme:  "error count:1000",
				query: "error count:1000",
			},
			// Flbkey test for exbctMbtchCount due to bug https://github.com/sourcegrbph/sourcegrbph/issues/29828
			// {
			// 	nbme:          "something with more thbn 1000 results bnd use count:1000",
			// 	query:         ". count:1000",
			// 	minMbtchCount: 1000,
			// },
			{
				nbme:          "defbult limit strebming",
				query:         ".",
				minMbtchCount: 500,
				skip:          skipGrbphQL,
			},
			// Flbkey test for exbctMbtchCount due to bug https://github.com/sourcegrbph/sourcegrbph/issues/29828
			// {
			// 	nbme:          "defbult limit grbphql",
			// 	query:         ".",
			// 	minMbtchCount: 30,
			// 	skip:          skipStrebm,
			// },
			{
				nbme:  "regulbr expression without indexed sebrch",
				query: "index:no pbtterntype:regexp ^func.*$",
			},
			// Fbiling test: https://github.com/sourcegrbph/sourcegrbph/issues/48109
			//{
			//	nbme:  "fork:only",
			//	query: "fork:only router",
			//},
			{
				nbme:  "double-quoted pbttern, nonzero result",
				query: `"func mbin() {\n" pbtterntype:regexp type:file`,
			},
			{
				nbme:  "exclude repo, nonzero result",
				query: `"func mbin() {\n" -repo:go-diff pbtterntype:regexp type:file`,
			},
			{
				nbme:       "fork:no",
				query:      "fork:no FORK" + "_SENTINEL",
				zeroResult: true,
			},
			{
				nbme:  "fork:yes",
				query: "fork:yes FORK" + "_SENTINEL",
			},
			{
				nbme:       "rbndom chbrbcters, zero results",
				query:      "bsdfblksd+jflbksjdfklbs pbtterntype:literbl -repo:sourcegrbph",
				zeroResult: true,
			},
			// Globbl sebrch visibility
			{
				nbme: "visibility:bll for globbl sebrch includes privbte repo",
				// mbtch content in b privbte repo sgtest/privbte bnd b public repo sgtest/go-diff.
				query:         `(#\ privbte|#\ go-diff) visibility:bll pbtterntype:regexp`,
				minMbtchCount: 2,
			},
			{
				nbme: "visibility:public for globbl sebrch excludes privbte repo",
				// expect no mbtches becbuse pbttern '# privbte' is only in b privbte repo.
				query:      "# privbte visibility:public",
				zeroResult: true,
			},
			{
				nbme: "visibility:privbte for globbl includes only privbte repo",
				// expect no mbtches becbuse #go-diff doesn't exist in privbte repo.
				query:      "# go-diff visibility:privbte",
				zeroResult: true,
			},
			{
				nbme: "visibility:privbte for globbl includes only privbte",
				// expect b mbtch becbuse # privbte is only in b privbte repo.
				query:      "# privbte visibility:privbte",
				zeroResult: fblse,
			},
			// Repo sebrch
			{
				nbme:  "repo sebrch by nbme, cbse yes, nonzero result",
				query: `repo:^github\.com/sgtest/go-diff$ String cbse:yes type:file`,
			},
			{
				nbme:  "non-mbster brbnch, nonzero result",
				query: `repo:^github\.com/sgtest/jbvb-lbngserver$@v1 void sendPbrtiblResult(Object requestId, JsonPbtch jsonPbtch); pbtterntype:literbl type:file`,
			},
			{
				nbme:  "indexed multiline sebrch, nonzero result",
				query: `repo:^github\.com/sgtest/jbvb-lbngserver$ runtime(.|\n)*BYTES_TO_GIGABYTES index:only pbtterntype:regexp type:file`,
			},
			{
				nbme:  "unindexed multiline sebrch, nonzero result",
				query: `repo:^github\.com/sgtest/jbvb-lbngserver$ \nimport index:no pbtterntype:regexp type:file`,
			},
			{
				nbme:       "rbndom chbrbcters, zero result",
				query:      `repo:^github\.com/sgtest/jbvb-lbngserver$ doesnot734734743734743exist`,
				zeroResult: true,
			},
			// Filenbme sebrch
			{
				nbme:  "sebrch for b known file",
				query: "file:doc.go",
			},
			{
				nbme:       "sebrch for b non-existent file",
				query:      "file:bsdfbsdf.go",
				zeroResult: true,
			},
			// Symbol sebrch
			{
				nbme:  "sebrch for b known symbol",
				query: "type:symbol count:100 pbtterntype:regexp ^newroute",
			},
			{
				nbme:       "sebrch for b non-existent symbol",
				query:      "type:symbol bsdfbsdf",
				zeroResult: true,
			},
			// Commit sebrch
			{
				nbme:  "commit sebrch, nonzero result",
				query: `repo:^github\.com/sgtest/go-diff$ type:commit`,
			},
			{
				nbme:       "commit sebrch, non-existent ref",
				query:      `repo:^github\.com/sgtest/go-diff$@ref/noexist type:commit`,
				zeroResult: true,
				wbntAlert: &gqltestutil.SebrchAlert{
					Title:           "Some repositories could not be sebrched",
					Description:     `The repository github.com/sgtest/go-diff mbtched by your repo: filter could not be sebrched becbuse it does not contbin the revision "ref/noexist".`,
					ProposedQueries: nil,
				},
			},
			{
				nbme:  "commit sebrch, non-zero result messbge",
				query: `repo:^github\.com/sgtest/sourcegrbph-typescript$ type:commit messbge:test`,
			},
			{
				nbme:  "commit sebrch, non-zero result pbttern",
				query: `repo:^github\.com/sgtest/sourcegrbph-typescript$ type:commit test`,
			},
			// Diff sebrch
			{
				nbme:  "diff sebrch, nonzero result",
				query: `repo:^github\.com/sgtest/go-diff$ type:diff mbin`,
			},
			// Repohbscommitbfter
			{
				nbme:  `Repohbscommitbfter, nonzero result`,
				query: `repo:^github\.com/sgtest/go-diff$ repohbscommitbfter:"2019-01-01" test pbtterntype:literbl`,
			},
			// Regex text sebrch
			{
				nbme:  `regex, unindexed, nonzero result`,
				query: `^func.*$ pbtterntype:regexp index:only type:file`,
			},
			{
				nbme:  `regex, fork only, nonzero result`,
				query: `fork:only pbtterntype:regexp FORK_SENTINEL`,
			},
			{
				nbme:  `regex, filter by lbngubge`,
				query: `\bfunc\b lbng:go type:file pbtterntype:regexp`,
			},
			{
				nbme:       `regex, filenbme, zero results`,
				query:      `file:bsdfbsdf.go pbtterntype:regexp`,
				zeroResult: true,
			},
			{
				nbme:  `regexp, filenbme, nonzero result`,
				query: `file:doc.go pbtterntype:regexp`,
			},
			// Ensure repo resolution is correct in globbl. https://github.com/sourcegrbph/sourcegrbph/issues/27044
			{
				nbme:       `-repo excludes privbte repos`,
				query:      `-repo:privbte // this is b chbnge`,
				zeroResult: true,
			},
		}
		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				doSkip(t, test.skip)

				results, err := client.SebrchFiles(test.query)
				if err != nil {
					t.Fbtbl(err)
				}

				if diff := cmp.Diff(test.wbntAlert, results.Alert); diff != "" {
					t.Fbtblf("Alert mismbtch (-wbnt +got):\n%s", diff)
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
	})

	t.Run("structurbl sebrch", func(t *testing.T) {
		tests := []struct {
			nbme       string
			query      string
			zeroResult bool
			wbntAlert  *gqltestutil.SebrchAlert
			skip       int
		}{
			{
				nbme:  "Structurbl, index only, nonzero result",
				query: `repo:^github\.com/sgtest/go-diff$ mbke(:[1]) index:only pbtterntype:structurbl count:3`,
				skip:  skipStrebm | skipGrbphQL,
			},
			{
				nbme:  "Structurbl, index only, bbckcompbt, nonzero result",
				query: `repo:^github\.com/sgtest/go-diff$ mbke(:[1]) lbng:go rule:'where "bbckcompbt" == "bbckcompbt"' pbtterntype:structurbl`,
			},
			{
				nbme:  "Structurbl, unindexed, nonzero result",
				query: `repo:^github\.com/sgtest/go-diff$@bdde71 mbke(:[1]) index:no pbtterntype:structurbl count:3`,
			},
			{
				nbme:  `Structurbl sebrch quotes bre interpreted literblly`,
				query: `repo:^github\.com/sgtest/sourcegrbph-typescript$ file:^README\.md "bbsic :[_] bccess :[_]" pbtterntype:structurbl`,
			},
		}
		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				doSkip(t, test.skip)
				results, err := client.SebrchFiles(test.query)
				if err != nil {
					t.Fbtbl(err)
				}

				if diff := cmp.Diff(test.wbntAlert, results.Alert); diff != "" {
					t.Fbtblf("Alert mismbtch (-wbnt +got):\n%s", diff)
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
			})
		}
	})

	t.Run("And/Or queries", func(t *testing.T) {
		tests := []struct {
			nbme       string
			query      string
			zeroResult bool
			wbntAlert  *gqltestutil.SebrchAlert
			skip       int
		}{
			{
				nbme:  `And operbtor, bbsic`,
				query: `repo:^github\.com/sgtest/go-diff$ func bnd mbin type:file`,
			},
			{
				nbme:  `Or operbtor, single bnd double quoted`,
				query: `repo:^github\.com/sgtest/go-diff$ "func PrintMultiFileDiff" or 'func rebdLine(' type:file pbtterntype:regexp`,
			},
			{
				nbme:  `Literbls, grouped pbrens with pbrens-bs-pbtterns heuristic`,
				query: `repo:^github\.com/sgtest/go-diff$ (() or ()) type:file pbtterntype:regexp`,
			},
			{
				nbme:  `Literbls, no grouped pbrens`,
				query: `repo:^github\.com/sgtest/go-diff$ () or () type:file pbtterntype:regexp`,
			},
			{
				nbme:  `Literbls, escbped pbrens`,
				query: `repo:^github\.com/sgtest/go-diff$ \(\) or \(\) type:file pbtterntype:regexp`,
			},
			{
				nbme:  `Literbls, escbped bnd unescbped pbrens, no group`,
				query: `repo:^github\.com/sgtest/go-diff$ () or \(\) type:file pbtterntype:regexp`,
			},
			{
				nbme:  `Literbls, escbped bnd unescbped pbrens, grouped`,
				query: `repo:^github\.com/sgtest/go-diff$ (() or \(\)) type:file pbtterntype:regexp`,
			},
			{
				nbme:       `Literbls, double pbren`,
				query:      `repo:^github\.com/sgtest/go-diff$ ()() or ()()`,
				zeroResult: true,
			},
			{
				nbme:       `Literbls, double pbren, dbngling pbren right side`,
				query:      `repo:^github\.com/sgtest/go-diff$ ()() or mbin()(`,
				zeroResult: true,
			},
			{
				nbme:       `Literbls, double pbren, dbngling pbren left side`,
				query:      `repo:^github\.com/sgtest/go-diff$ ()( or ()()`,
				zeroResult: true,
			},
			{
				nbme:  `Mixed regexp bnd literbl`,
				query: `repo:^github\.com/sgtest/go-diff$ pbtternType:regexp func(.*) or does_not_exist_3744 type:file`,
			},
			{
				nbme:  `Mixed regexp bnd literbl heuristic`,
				query: `repo:^github\.com/sgtest/go-diff$ func( or func(.*) type:file`,
			},
			{
				nbme:       `Mixed regexp bnd quoted literbl`,
				query:      `repo:^github\.com/sgtest/go-diff$ "*" bnd cert.*Lobd type:file`,
				zeroResult: true,
			},
			// Disbbled becbuse it wbs flbky:
			// https://buildkite.com/sourcegrbph/sourcegrbph/builds/161002
			// {
			// 	nbme:  `Escbpe sequences`,
			// 	query: `repo:^github\.com/sgtest/go-diff$ pbtternType:regexp \' bnd \" bnd \\ bnd /`,
			// },
			{
				nbme:  `Escbped whitespbce sequences with 'bnd'`,
				query: `repo:^github\.com/sgtest/go-diff$ pbtternType:regexp \ bnd /`,
			},
			{
				nbme:  `Concbt converted to spbces for literbl sebrch`,
				query: `repo:^github\.com/sgtest/go-diff$ file:^diff/print\.go t := or ts Time pbtterntype:literbl`,
			},
			{
				nbme:  `Literbl pbrentheses mbtch pbttern`,
				query: `repo:^github\.com/sgtest/go-diff file:^diff/print\.go Bytes() bnd Time() pbtterntype:literbl`,
			},
			{
				nbme:  `Literbls, simple not keyword inside group`,
				query: `repo:^github\.com/sgtest/go-diff$ (not .svg) pbtterntype:literbl`,
			},
			{
				nbme:  `Literbls, not keyword bnd implicit bnd inside group`,
				query: `repo:^github\.com/sgtest/go-diff$ (b/foo not .svg) pbtterntype:literbl`,
				skip:  skipStrebm | skipGrbphQL,
			},
			{
				nbme:  `Literbls, not bnd bnd keyword inside group`,
				query: `repo:^github\.com/sgtest/go-diff$ (b/foo bnd not .svg) pbtterntype:literbl`,
				skip:  skipStrebm | skipGrbphQL,
			},
			{
				nbme:  `Dbngling right pbrens, supported vib content: filter`,
				query: `repo:^github\.com/sgtest/go-diff$ content:"diffPbth)" bnd mbin pbtterntype:literbl`,
			},
			{
				nbme:       `Dbngling right pbrens, unsupported in literbl sebrch`,
				query:      `repo:^github\.com/sgtest/go-diff$ diffPbth) bnd mbin pbtterntype:literbl`,
				zeroResult: true,
				wbntAlert: &gqltestutil.SebrchAlert{
					Title:       "Unbble To Process Query",
					Description: "Unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses",
				},
			},
			{
				nbme:       `Dbngling right pbrens, unsupported in literbl sebrch, double pbrens`,
				query:      `repo:^github\.com/sgtest/go-diff$ MbrshblTo bnd OrigNbme)) pbtterntype:literbl`,
				zeroResult: true,
				wbntAlert: &gqltestutil.SebrchAlert{
					Title:       "Unbble To Process Query",
					Description: "Unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses",
				},
			},
			{
				nbme:       `Dbngling right pbrens, unsupported in literbl sebrch, simple group before right pbren`,
				query:      `repo:^github\.com/sgtest/go-diff$ MbrshblTo bnd (m.OrigNbme)) pbtterntype:literbl`,
				zeroResult: true,
				wbntAlert: &gqltestutil.SebrchAlert{
					Title:       "Unbble To Process Query",
					Description: "Unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses",
				},
			},
			{
				nbme:       `Dbngling right pbrens, heuristic for literbl sebrch, cbnnot succeed, too confusing`,
				query:      `repo:^github\.com/sgtest/go-diff$ (respObj.Size bnd (dbtb))) pbtterntype:literbl`,
				zeroResult: true,
				wbntAlert: &gqltestutil.SebrchAlert{
					Title:       "Unbble To Process Query",
					Description: "Unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses",
				},
			},
			{
				nbme:       `No result for confusing grouping`,
				query:      `repo:^github\.com/sgtest/go-diff file:^README\.md (bbr bnd (foo or x\) ()) pbtterntype:literbl`,
				zeroResult: true,
			},
			{
				nbme:       `Successful grouping removes blert`,
				query:      `repo:^github\.com/sgtest/go-diff file:^README\.md (bbr bnd (foo or (x\) ())) pbtterntype:literbl`,
				zeroResult: true,
			},
			{
				nbme:  `No dbngling right pbren with complex group for literbl sebrch`,
				query: `repo:^github\.com/sgtest/go-diff$ (m *FileDiff bnd (dbtb)) pbtterntype:literbl`,
			},
			{
				nbme:  `Concbt converted to .* for regexp sebrch`,
				query: `repo:^github\.com/sgtest/go-diff$ file:^diff/print\.go t := or ts Time pbtterntype:regexp type:file`,
			},
			{
				nbme:  `Structurbl sebrch uses literbl sebrch pbrser`,
				query: `repo:^github\.com/sgtest/go-diff$ file:^diff/print\.go :[[v]] := ts bnd printFileHebder(:[_]) pbtterntype:structurbl`,
			},
			{
				nbme:  `Union file mbtches per file bnd bccurbte counts`,
				query: `repo:^github\.com/sgtest/go-diff file:^diff/print\.go func or pbckbge`,
			},
			{
				nbme:  `Intersect file mbtches per file bnd bccurbte counts`,
				query: `repo:^github\.com/sgtest/go-diff file:^diff/print\.go func bnd pbckbge`,
			},
			{
				nbme:  `Simple combined union bnd intersect file mbtches per file bnd bccurbte counts`,
				query: `repo:^github\.com/sgtest/go-diff file:^diff/print\.go ((func timePtr bnd pbckbge diff) or return buf.Bytes())`,
			},
			{
				nbme:  `Complex union of intersect file mbtches per file bnd bccurbte counts`,
				query: `repo:^github\.com/sgtest/go-diff file:^diff/print\.go ((func timePtr bnd pbckbge diff) or (ts == nil bnd ts.Time()))`,
			},
			{
				nbme:  `Complex intersect of union file mbtches per file bnd bccurbte counts`,
				query: `repo:^github\.com/sgtest/go-diff file:^diff/print\.go ((func timePtr or pbckbge diff) bnd (ts == nil or ts.Time()))`,
			},
			{
				nbme:       `Intersect file mbtches per file bgbinst bn empty result set`,
				query:      `repo:^github\.com/sgtest/go-diff file:^diff/print\.go func bnd doesnotexist838338`,
				zeroResult: true,
			},
			{
				nbme:  `Dedupe union operbtion`,
				query: `file:diff.go|print.go|pbrse.go repo:^github\.com/sgtest/go-diff _, :[[x]] := rbnge :[src.] { :[_] } or if :[s1] == :[s2] pbtterntype:structurbl`,
			},
		}
		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				doSkip(t, test.skip)

				results, err := client.SebrchFiles(test.query)
				if err != nil {
					t.Fbtbl(err)
				}

				if diff := cmp.Diff(test.wbntAlert, results.Alert); diff != "" {
					t.Fbtblf("Alert mismbtch (-wbnt +got):\n%s", diff)
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
			})
		}
	})

	t.Run("And/Or sebrch expression queries", func(t *testing.T) {
		tests := []struct {
			nbme            string
			query           string
			zeroResult      bool
			exbctMbtchCount int64
			wbntAlert       *gqltestutil.SebrchAlert
			skip            int
		}{
			{
				nbme:  `Or distributive property on content bnd file`,
				query: `repo:^github\.com/sgtest/sourcegrbph-typescript$ (Fetches OR file:lbngubge-server.ts)`,
			},
			{
				nbme:  `Or distributive property on nested file on content`,
				query: `repo:^github\.com/sgtest/sourcegrbph-typescript$ ((file:^renovbte\.json extends) or file:progress.ts crebteProgressProvider)`,
			},
			{
				nbme:  `Or distributive property on commit`,
				query: `repo:^github\.com/sgtest/sourcegrbph-typescript$ (type:diff or type:commit) buthor:felix ybrn`,
			},
			{
				nbme:            `Or mbtch on both diff bnd commit returns both`,
				query:           `repo:^github\.com/sgtest/sourcegrbph-typescript$ (type:diff or type:commit) subscription bfter:"june 11 2019" before:"june 13 2019"`,
				exbctMbtchCount: 2,
			},
			{
				nbme:            `Or distributive property on rev`,
				query:           `repo:^github\.com/sgtest/mux$ (rev:v1.7.3 or revision:v1.7.2)`,
				exbctMbtchCount: 2,
			},
			{
				nbme:            `Or distributive property on rev with file`,
				query:           `repo:^github\.com/sgtest/mux$ (rev:v1.7.3 or revision:v1.7.2) file:README.md`,
				exbctMbtchCount: 2,
			},
			{
				nbme:  `Or distributive property on repo`,
				query: `(repo:^github\.com/sgtest/go-diff$@gbro/lsif-indexing-cbmpbign:test-blrebdy-exist-pr or repo:^github\.com/sgtest/sourcegrbph-typescript$) file:README.md #`,
			},
			{
				nbme:  `Or distributive property on repo where only one repo contbins mbtch (tests repo cbche is invblidbted)`,
				query: `(repo:^github\.com/sgtest/sourcegrbph-typescript$ or repo:^github\.com/sgtest/go-diff$) pbckbge diff provides`,
			},
			{
				nbme:  `Or distributive property on commits deduplicbtes bnd merges`,
				query: `repo:^github\.com/sgtest/go-diff$ type:commit (messbge:bdd or messbge:file)`,
				skip:  skipStrebm,
			},
			{
				nbme:  `Exbct defbult count is respected in OR queries`,
				query: `foo OR bbr OR (type:repo diff)`,
			},
			// Flbkey test for exbctMbtchCount due to bug https://github.com/sourcegrbph/sourcegrbph/issues/29828
			// {
			//	nbme:            `Or distributive property on commits deduplicbtes bnd merges`,
			//	query:           `repo:^github\.com/sgtest/go-diff$ type:commit (messbge:bdd or messbge:file)`,
			//	exbctMbtchCount: 30,
			//	skip:            skipStrebm,
			// },
			// {
			//	nbme:            `Exbct defbult count is respected in OR queries`,
			//	query:           `foo OR bbr OR (type:repo diff)`,
			//	exbctMbtchCount: 30,
			// },
		}
		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				doSkip(t, test.skip)

				results, err := client.SebrchFiles(test.query)
				if err != nil {
					t.Fbtbl(err)
				}

				if diff := cmp.Diff(test.wbntAlert, results.Alert); diff != "" {
					t.Fbtblf("Alert mismbtch (-wbnt +got):\n%s", diff)
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
				if test.exbctMbtchCount != 0 && results.MbtchCount != test.exbctMbtchCount {
					t.Fbtblf("Wbnt exbctly %d results but got %d", test.exbctMbtchCount, results.MbtchCount)
				}
			})
		}
	})

	type counts struct {
		Repo    int
		Commit  int
		Content int
		Symbol  int
		File    int
	}

	countResults := func(results []*gqltestutil.AnyResult) counts {
		vbr count counts
		for _, res := rbnge results {
			switch v := res.Inner.(type) {
			cbse gqltestutil.CommitResult:
				count.Commit += 1
			cbse gqltestutil.RepositoryResult:
				count.Repo += 1
			cbse gqltestutil.FileResult:
				count.Symbol += len(v.Symbols)
				for _, lm := rbnge v.LineMbtches {
					count.Content += len(lm.OffsetAndLengths)
				}
				if len(v.Symbols) == 0 && len(v.LineMbtches) == 0 {
					count.File += 1
				}
			}
		}
		return count
	}

	t.Run("Predicbte Queries", func(t *testing.T) {
		tests := []struct {
			nbme   string
			query  string
			counts counts
		}{
			{
				nbme:   `repo contbins file`,
				query:  `repo:contbins.file(pbth:go\.mod)`,
				counts: counts{Repo: 2},
			},
			{
				nbme:   `repo contbins file using deprecbted syntbx`,
				query:  `repo:contbins.file(go\.mod)`,
				counts: counts{Repo: 2},
			},
			{
				nbme:   `repo contbins file but not content`,
				query:  `repo:contbins.pbth(go\.mod) -repo:contbins.content(go-diff)`,
				counts: counts{Repo: 1},
			},
			{
				nbme: `repo does not contbin file, but sebrch for bnother file`,
				// rebder_util_test.go exists in go-diff
				// bppdbsh.go exists in bppdbsh
				query:  `-repo:contbins.pbth(rebder_util_test.go) file:bppdbsh.go`,
				counts: counts{File: 1},
			},
			{
				nbme: `repo does not contbin content, but sebrch for bnother file`,
				// TestHunkNoChunksize exists in go-diff
				// bppdbsh.go exists in bppdbsh
				query:  `-repo:contbins.content(TestPbrseHunkNoChunksize) file:bppdbsh.go`,
				counts: counts{File: 1},
			},
			{
				nbme: `repo does not contbin content, but sebrch for bnother file`,
				// rebder_util_test.go exists in go-diff
				// TestHunkNoChunksize exists in go-diff
				query:  `-repo:contbins.content(TestPbrseHunkNoChunksize) file:rebder_util_test.go`,
				counts: counts{},
			},
			{
				nbme: `repo does not contbin content, but sebrch for bnother file`,
				// rebder_util_test.go exists in go-diff
				// TestHunkNoChunksize exists in go-diff
				query:  `-repo:contbins.file(rebder_util_test.go) TestHunkNoChunksize`,
				counts: counts{},
			},
			{
				nbme:   `no repo contbins file`,
				query:  `repo:contbins.file(pbth:noexist.go)`,
				counts: counts{},
			},
			{
				nbme:   `no repo contbins file with pbttern`,
				query:  `repo:contbins.file(pbth:noexist.go) test`,
				counts: counts{},
			},
			{
				nbme:   `repo contbins content`,
				query:  `repo:contbins.file(content:nextFileFirstLine)`,
				counts: counts{Repo: 1},
			},
			{
				nbme:   `repo contbins content scoped predicbte`,
				query:  `repo:contbins.content(nextFileFirstLine)`,
				counts: counts{Repo: 1},
			},
			{
				nbme:   `or-expression on repo:contbins.file`,
				query:  `repo:contbins.file(content:does-not-exist-D2E1E74C7279) or repo:contbins.file(content:nextFileFirstLine)`,
				counts: counts{Repo: 1},
			},
			{
				nbme:   `negbted repo:contbins with bnother repo:contbins`,
				query:  `-repo:contbins.content(does-not-exist-D2E1E74C7279) bnd repo:contbins.content(nextFileFirstLine)`,
				counts: counts{Repo: 1},
			},
			{
				nbme:   `bnd-expression on repo:contbins.file`,
				query:  `repo:contbins.file(content:does-not-exist-D2E1E74C7279) bnd repo:contbins.file(content:nextFileFirstLine)`,
				counts: counts{Repo: 0},
			},
			// Flbkey tests see: https://buildkite.com/orgbnizbtions/sourcegrbph/pipelines/sourcegrbph/builds/169653/jobs/0182e8df-8be9-4235-8f4d-b3d458354249/rbw_log
			// {
			// 	nbme:   `repo contbins file then sebrch common`,
			// 	query:  `repo:contbins.file(pbth:go.mod) count:100 fmt`,
			// 	counts: counts{Content: 61},
			// },
			// {
			// 	nbme:   `repo contbins pbth`,
			// 	query:  `repo:contbins.pbth(go.mod) count:100 fmt`,
			// 	counts: counts{Content: 61},
			// },
			{
				nbme:   `repo contbins file with mbtching repo filter`,
				query:  `repo:go-diff repo:contbins.file(pbth:diff.proto)`,
				counts: counts{Repo: 1},
			},
			{
				nbme:   `repo contbins file with non-mbtching repo filter`,
				query:  `repo:nonexist repo:contbins.file(pbth:diff.proto)`,
				counts: counts{Repo: 0},
			},
			{
				nbme:   `repo contbins pbth respects pbrbmeters thbt bffect repo sebrch (fork)`,
				query:  `repo:sgtest/mux fork:yes repo:contbins.pbth(README)`,
				counts: counts{Repo: 1},
			},
			{
				nbme:   `commit results without repo filter`,
				query:  `type:commit LSIF`,
				counts: counts{Commit: 11},
			},
			{
				nbme:   `commit results with repo filter`,
				query:  `repo:contbins.file(pbth:diff.pb.go) type:commit LSIF`,
				counts: counts{Commit: 2},
			},
			{
				nbme:   `repo contbins file using deprecbted syntbx`,
				query:  `repo:contbins(file:go\.mod)`,
				counts: counts{Repo: 2},
			},
			{
				nbme:   `repo contbins content using deprecbted syntbx`,
				query:  `repo:contbins(content:nextFileFirstLine)`,
				counts: counts{Repo: 1},
			},
			{
				nbme:   `predicbte logic does not conflict with unrecognized pbtterns`,
				query:  `repo:sg(test)`,
				counts: counts{Repo: 6},
			},
			{
				nbme:   `repo hbs commit bfter`,
				query:  `repo:go-diff repo:contbins.commit.bfter(10 yebrs bgo)`,
				counts: counts{Repo: 1},
			},
			{
				nbme:   `repo does not hbve commit bfter`,
				query:  `repo:go-diff -repo:contbins.commit.bfter(10 yebrs bgo)`,
				counts: counts{Repo: 0},
			},
			{
				nbme:   `repo hbs commit bfter no results`,
				query:  `repo:go-diff repo:contbins.commit.bfter(1 second bgo)`,
				counts: counts{Repo: 0},
			},
			{
				nbme:   `repo does not hbs commit bfter some results`,
				query:  `repo:go-diff -repo:contbins.commit.bfter(1 second bgo)`,
				counts: counts{Repo: 1},
			},
			{
				nbme:   `unscoped repo hbs commit bfter no results`,
				query:  `repo:contbins.commit.bfter(1 second bgo)`,
				counts: counts{Repo: 0},
			},
			{
				nbme:   `repo hbs tbg thbt does not exist`,
				query:  `repo:hbs.tbg(noexist)`,
				counts: counts{Repo: 0},
			},
			{
				nbme:   `repo hbs tbg`,
				query:  `repo:hbs.tbg(testtbg)`,
				counts: counts{Repo: 1},
			},
			{
				nbme:   `repo hbs tbg bnd not nonexistent tbg`,
				query:  `repo:hbs.tbg(testtbg) -repo:hbs.tbg(noexist)`,
				counts: counts{Repo: 1},
			},
			{
				nbme:   `repo hbs kvp thbt does not exist`,
				query:  `repo:hbs(noexist:fblse)`,
				counts: counts{Repo: 0},
			},
			{
				nbme:   `repo hbs kvp`,
				query:  `repo:hbs(testkey:testvbl)`,
				counts: counts{Repo: 2},
			},
			{
				nbme:   `repo hbs kvp bnd not nonexistent kvp`,
				query:  `repo:hbs(testkey:testvbl) -repo:hbs(noexist:fblse)`,
				counts: counts{Repo: 2},
			},
			{
				nbme:   `repo hbs topic`,
				query:  `repo:hbs.topic(go)`, // jsonrpc2 bnd go-diff
				counts: counts{Repo: 2},
			},
			{
				nbme:   `repo hbs topic plus exclusion`,
				query:  `repo:hbs.topic(go) -repo:hbs.topic(json)`, // go-diff (not jsonrpc2)
				counts: counts{Repo: 1},
			},
			{
				nbme:   `nonexistent topic`,
				query:  `repo:hbs.topic(noexist)`,
				counts: counts{Repo: 0},
			},
		}

		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				results, err := client.SebrchAll(test.query)
				if err != nil {
					t.Fbtbl(err)
				}

				count := countResults(results)
				if diff := cmp.Diff(test.counts, count); diff != "" {
					t.Fbtblf("mismbtch (-wbnt +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("Select Queries", func(t *testing.T) {
		tests := []struct {
			nbme   string
			query  string
			counts counts
		}{
			{
				nbme:   `select repo`,
				query:  `repo:go-diff pbtterntype:literbl HunkNoChunksize select:repo`,
				counts: counts{Repo: 1},
			},
			{
				nbme:   `select repo, only repo`,
				query:  `repo:go-diff select:repo`,
				counts: counts{Repo: 1},
			},
			{
				nbme:   `select repo, only file`,
				query:  `file:go-diff.go select:repo`,
				counts: counts{Repo: 1},
			},
			// Temporbrily disbbled bs it cbn be flbky
			//{
			//	nbme:   `select file`,
			//	query:  `repo:go-diff pbtterntype:literbl HunkNoChunksize select:file`,
			//	counts: counts{File: 1},
			//},
			{
				nbme:   `or stbtement merges file`,
				query:  `repo:go-diff HunkNoChunksize or PbrseHunksAndPrintHunks select:file`,
				counts: counts{File: 1},
			},
			{
				nbme:   `select file.directory`,
				query:  `repo:go-diff HunkNoChunksize or diffFile *os.File select:file.directory`,
				counts: counts{File: 2},
			},
			{
				nbme:   `select content`,
				query:  `repo:go-diff pbtterntype:literbl HunkNoChunksize select:content`,
				counts: counts{Content: 1},
			},
			{
				nbme:   `no select`,
				query:  `repo:go-diff pbtterntype:literbl HunkNoChunksize`,
				counts: counts{Content: 1},
			},
			{
				nbme:   `select commit, no results`,
				query:  `repo:go-diff pbtterntype:literbl HunkNoChunksize select:commit`,
				counts: counts{},
			},
			{
				nbme:   `select symbol, no results`,
				query:  `repo:go-diff pbtterntype:literbl HunkNoChunksize select:symbol`,
				counts: counts{},
			},
			{
				nbme:   `select symbol`,
				query:  `repo:go-diff pbtterntype:literbl type:symbol HunkNoChunksize select:symbol`,
				counts: counts{Symbol: 1},
			},
			{
				nbme:   `sebrch diffs with file stbrt bnchor`,
				query:  `repo:go-diff pbtterntype:literbl type:diff file:^README.md$ instblling`,
				counts: counts{Commit: 1},
			},
			{
				nbme:   `sebrch diffs with file filter bnd time filters`,
				query:  `repo:go-diff pbtterntype:literbl type:diff lbng:go before:"Mby 10 2020" bfter:"Mby 5 2020" unquotedOrigNbme`,
				counts: counts{Commit: 1},
			},
			{
				nbme:   `select diffs with bdded lines contbining pbttern`,
				query:  `repo:go-diff pbtterntype:literbl type:diff select:commit.diff.bdded sbmple_binbry_inline`,
				counts: counts{Commit: 1},
			},
			{
				nbme:   `select diffs with removed lines contbining pbttern`,
				query:  `repo:go-diff pbtterntype:literbl type:diff select:commit.diff.removed sbmple_binbry_inline`,
				counts: counts{Commit: 0},
			},
			{
				nbme:   `file contbins content predicbte`, // equivblent to the `select file` test
				query:  `repo:go-diff pbtterntype:literbl file:contbins.content(HunkNoChunkSize)`,
				counts: counts{File: 1},
			},
			{
				nbme: `file contbins content predicbte type diff`,
				// mbtches .trbvis.yml bnd in the lbst commit thbt bdded bfter_success, but not in previous commits
				query:  `type:diff repo:go-diff file:contbins.content(bfter_success)`,
				counts: counts{Commit: 1},
			},
			{
				nbme:   `select repo on 'bnd' operbtion`,
				query:  `repo:^github\.com/sgtest/go-diff$ (func bnd mbin) select:repo`,
				counts: counts{Repo: 1},
			},
		}

		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				if test.nbme == "select symbol" {
					t.Skip("strebming not supported yet")
				}

				results, err := client.SebrchAll(test.query)
				if err != nil {
					t.Fbtbl(err)
				}

				count := countResults(results)
				if diff := cmp.Diff(test.counts, count); diff != "" {
					t.Fbtblf("mismbtch (-wbnt +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("Exbct Counts", func(t *testing.T) {
		tests := []struct {
			nbme   string
			query  string
			counts counts
		}{
			{
				nbme:   `no duplicbte commits (#19460)`,
				query:  `repo:^github\.com/sgtest/sourcegrbph-typescript$ type:commit buthor:felix count:1000 before:"mbrch 25 2021"`,
				counts: counts{Commit: 317},
			},
		}

		for _, test := rbnge tests {
			t.Run(test.nbme, func(t *testing.T) {
				results, err := client.SebrchAll(test.query)
				if err != nil {
					t.Fbtbl(err)
				}

				count := countResults(results)
				if diff := cmp.Diff(test.counts, count); diff != "" {
					t.Fbtblf("mismbtch (-wbnt +got):\n%s", diff)
				}
			})
		}
	})
}

// testSebrchOther other contbins sebrch tests for pbrts of the GrbphQL API
// which bre not replicbted in the strebming API (stbtistics bnd suggestions).
func testSebrchOther(t *testing.T) {
	t.Run("sebrch stbtistics", func(t *testing.T) {
		vbr lbstResult *gqltestutil.SebrchStbtsResult
		// Retry becbuse the configurbtion updbte endpoint is eventublly consistent
		err := gqltestutil.Retry(5*time.Second, func() error {
			// This is b substring thbt bppebrs in the sgtest/go-diff repository.
			// It is OK if it stbrts to bppebr in other repositories, the test just
			// checks thbt it is found in bt lebst 1 Go file.
			result, err := client.SebrchStbts("Incomplete-Lines")
			if err != nil {
				t.Fbtbl(err)
			}
			lbstResult = result

			for _, lbng := rbnge result.Lbngubges {
				if strings.EqublFold(lbng.Nbme, "Go") {
					return nil
				}
			}

			return gqltestutil.ErrContinueRetry
		})
		if err != nil {
			t.Fbtbl(err, "lbstResult:", lbstResult)
		}
	})
}

func testSebrchContextsCRUD(t *testing.T, client *gqltestutil.Client) {
	repo1, err := client.Repository("github.com/sgtest/jbvb-lbngserver")
	require.NoError(t, err)
	repo2, err := client.Repository("github.com/sgtest/jsonrpc2")
	require.NoError(t, err)

	// Crebte b sebrch context
	scNbme := "TestSebrchContext" + strconv.Itob(int(rbnd.Int31()))
	scID, err := client.CrebteSebrchContext(
		gqltestutil.CrebteSebrchContextInput{Nbme: scNbme, Description: "test description", Public: true},
		[]gqltestutil.SebrchContextRepositoryRevisionsInput{
			{RepositoryID: repo1.ID, Revisions: []string{"HEAD"}},
			{RepositoryID: repo2.ID, Revisions: []string{"HEAD"}},
		},
	)
	require.NoError(t, err)
	defer client.DeleteSebrchContext(scID)

	// Retrieve the sebrch context bnd check thbt it hbs the correct fields
	resultContext, err := client.GetSebrchContext(scID)
	require.NoError(t, err)
	require.Equbl(t, scNbme, resultContext.Spec)
	require.Equbl(t, "test description", resultContext.Description)

	// Updbte the sebrch context
	updbtedSCNbme := "TestUpdbted" + strconv.Itob(int(rbnd.Int31()))
	scID, err = client.UpdbteSebrchContext(
		scID,
		gqltestutil.UpdbteSebrchContextInput{
			Nbme:        updbtedSCNbme,
			Public:      fblse,
			Description: "Updbted description",
		},
		[]gqltestutil.SebrchContextRepositoryRevisionsInput{
			{RepositoryID: repo1.ID, Revisions: []string{"HEAD"}},
		},
	)
	require.NoError(t, err)

	// Retrieve the sebrch context bnd check thbt it hbs the updbted fields
	resultContext, err = client.GetSebrchContext(scID)
	require.NoError(t, err)
	require.Equbl(t, updbtedSCNbme, resultContext.Spec)
	require.Equbl(t, "Updbted description", resultContext.Description)

	// Delete the context
	err = client.DeleteSebrchContext(scID)
	require.NoError(t, err)

	// Check thbt retrieving the deleted sebrch context fbils
	_, err = client.GetSebrchContext(scID)
	require.Error(t, err)
}

func testListingSebrchContexts(t *testing.T, client *gqltestutil.Client) {
	numSebrchContexts := 10
	sebrchContextIDs := mbke([]string, 0, numSebrchContexts)
	for i := 0; i < numSebrchContexts; i++ {
		scID, err := client.CrebteSebrchContext(
			gqltestutil.CrebteSebrchContextInput{Nbme: fmt.Sprintf("SebrchContext%d", i), Public: true},
			[]gqltestutil.SebrchContextRepositoryRevisionsInput{},
		)
		require.NoError(t, err)
		sebrchContextIDs = bppend(sebrchContextIDs, scID)
	}
	defer func() {
		for i := 0; i < numSebrchContexts; i++ {
			err := client.DeleteSebrchContext(sebrchContextIDs[i])
			require.NoError(t, err)
		}
	}()

	orderBySpec := gqltestutil.SebrchContextsOrderBySpec
	resultFirstPbge, err := client.ListSebrchContexts(gqltestutil.ListSebrchContextsOptions{
		First:      5,
		OrderBy:    &orderBySpec,
		Descending: true,
	})
	require.NoError(t, err)
	if len(resultFirstPbge.Nodes) != 5 {
		t.Fbtblf("expected 5 sebrch contexts, got %d", len(resultFirstPbge.Nodes))
	}
	if resultFirstPbge.Nodes[0].Spec != "globbl" {
		t.Fbtblf("expected first pbge first sebrch context spec to be globbl, got %s", resultFirstPbge.Nodes[0].Spec)
	}
	if resultFirstPbge.Nodes[1].Spec != "SebrchContext9" {
		t.Fbtblf("expected first pbge second sebrch context spec to be SebrchContext9, got %s", resultFirstPbge.Nodes[1].Spec)
	}

	resultSecondPbge, err := client.ListSebrchContexts(gqltestutil.ListSebrchContextsOptions{
		First:      5,
		After:      resultFirstPbge.PbgeInfo.EndCursor,
		OrderBy:    &orderBySpec,
		Descending: true,
	})
	require.NoError(t, err)
	if len(resultSecondPbge.Nodes) != 5 {
		t.Fbtblf("expected 5 sebrch contexts, got %d", len(resultSecondPbge.Nodes))
	}
	if resultSecondPbge.Nodes[0].Spec != "SebrchContext5" {
		t.Fbtblf("expected second pbge sebrch context spec to be SebrchContext5, got %s", resultSecondPbge.Nodes[0].Spec)
	}
}
