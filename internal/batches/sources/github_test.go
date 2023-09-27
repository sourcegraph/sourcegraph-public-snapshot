pbckbge sources

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/inconshrevebble/log15"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestGithubSource_CrebteChbngeset(t *testing.T) {
	// Repository used: https://github.com/sourcegrbph/butombtion-testing
	//
	// The requests here cbnnot be ebsily rerun with `-updbte` since you cbn only open b
	// pull request once. To updbte, push b new brbnch with bt lebst one commit to
	// butombtion-testing, bnd put the brbnch nbmes into the `success` cbse below.
	//
	// You cbn updbte just this test with `-updbte GithubSource_CrebteChbngeset`.
	repo := &types.Repo{
		Metbdbtb: &github.Repository{
			ID:            "MDEwOlJlcG9zbXRvcnkyMjExNDc1MTM=",
			NbmeWithOwner: "sourcegrbph/butombtion-testing",
		},
	}

	testCbses := []struct {
		nbme   string
		cs     *Chbngeset
		err    string
		exists bool
	}{
		{
			nbme: "success",
			cs: &Chbngeset{
				Title:      "This is b test PR",
				Body:       "This is the description of the test PR",
				HebdRef:    "refs/hebds/test-review-decision",
				BbseRef:    "refs/hebds/mbster",
				RemoteRepo: repo,
				TbrgetRepo: repo,
				Chbngeset:  &btypes.Chbngeset{},
			},
			err: "<nil>",
		},
		{
			nbme: "blrebdy exists",
			cs: &Chbngeset{
				Title:      "This is b test PR",
				Body:       "This is the description of the test PR",
				HebdRef:    "refs/hebds/blwbys-open-pr",
				BbseRef:    "refs/hebds/mbster",
				RemoteRepo: repo,
				TbrgetRepo: repo,
				Chbngeset:  &btypes.Chbngeset{},
			},
			// If PR blrebdy exists we'll just return it, no error
			err:    "<nil>",
			exists: true,
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc

		tc.nbme = "GithubSource_CrebteChbngeset_" + strings.ReplbceAll(tc.nbme, " ", "_")

		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			src, sbve := setup(t, ctx, tc.nbme)
			defer sbve(t)

			exists, err := src.CrebteChbngeset(ctx, tc.cs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			if hbve, wbnt := exists, tc.exists; hbve != wbnt {
				t.Errorf("exists:\nhbve: %t\nwbnt: %t", hbve, wbnt)
			}

			pr, ok := tc.cs.Chbngeset.Metbdbtb.(*github.PullRequest)
			if !ok {
				t.Fbtbl("Metbdbtb does not contbin PR")
			}

			testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), pr)
		})
	}
}

func TestGithubSource_CrebteChbngeset_CrebtionLimit(t *testing.T) {
	github.SetupForTest(t)
	cli := new(mockDoer)
	// Version lookup
	versionMbtchedBy := func(req *http.Request) bool {
		return req.Method == http.MethodGet && req.URL.Pbth == "/"
	}
	cli.On("Do", mock.MbtchedBy(versionMbtchedBy)).
		Once().
		Return(
			&http.Response{
				StbtusCode: http.StbtusOK,
				Hebder: mbp[string][]string{
					"X-GitHub-Enterprise-Version": {"99"},
				},
				Body: io.NopCloser(bytes.NewRebder([]byte{})),
			},
			nil,
		)
	// Crebte Chbngeset mutbtion
	crebteChbngesetMbtchedBy := func(req *http.Request) bool {
		return req.Method == http.MethodPost && req.URL.Pbth == "/grbphql"
	}
	cli.On("Do", mock.MbtchedBy(crebteChbngesetMbtchedBy)).
		Once().
		Return(
			&http.Response{
				StbtusCode: http.StbtusOK,
				Body:       io.NopCloser(bytes.NewRebder([]byte(`{"errors": [{"messbge": "error in GrbphQL response: wbs submitted too quickly"}]}`))),
			},
			nil,
		)

	bpiURL, err := url.Pbrse("https://fbke.bpi.github.com")
	require.NoError(t, err)
	client := github.NewV4Client("extsvc:github:0", bpiURL, nil, cli)
	source := &GitHubSource{
		client: client,
	}

	repo := &types.Repo{
		Metbdbtb: &github.Repository{
			ID:            "bLAhBLAh",
			NbmeWithOwner: "some-org/some-repo",
		},
	}
	cs := &Chbngeset{
		Title:      "This is b test PR",
		Body:       "This is the description of the test PR",
		HebdRef:    "refs/hebds/blwbys-open-pr",
		BbseRef:    "refs/hebds/mbster",
		RemoteRepo: repo,
		TbrgetRepo: repo,
		Chbngeset:  &btypes.Chbngeset{},
	}

	exists, err := source.CrebteChbngeset(context.Bbckground(), cs)
	bssert.Fblse(t, exists)
	bssert.Error(t, err)
	bssert.Equbl(
		t,
		"rebched GitHub's internbl crebtion limit: see https://docs.sourcegrbph.com/bdmin/config/bbtch_chbnges#bvoiding-hitting-rbte-limits: error in GrbphQL response: error in GrbphQL response: wbs submitted too quickly",
		err.Error(),
	)
}

type mockDoer struct {
	mock.Mock
}

func (d *mockDoer) Do(req *http.Request) (*http.Response, error) {
	brgs := d.Cblled(req)
	return brgs.Get(0).(*http.Response), brgs.Error(1)
}

func TestGithubSource_CloseChbngeset(t *testing.T) {
	// Repository used: https://github.com/sourcegrbph/butombtion-testing
	//
	// This test cbn be updbted with `-updbte GithubSource_CloseChbngeset`, provided this
	// PR is open: https://github.com/sourcegrbph/butombtion-testing/pull/468
	testCbses := []struct {
		nbme string
		cs   *Chbngeset
		err  string
	}{
		{
			nbme: "success",
			cs: &Chbngeset{
				Chbngeset: &btypes.Chbngeset{
					Metbdbtb: &github.PullRequest{
						ID: "PR_kwDODS5xec4wbMkR",
					},
				},
			},
			err: "<nil>",
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "GithubSource_CloseChbngeset_" + strings.ReplbceAll(tc.nbme, " ", "_")

		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			src, sbve := setup(t, ctx, tc.nbme)
			defer sbve(t)

			err := src.CloseChbngeset(ctx, tc.cs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Chbngeset.Metbdbtb.(*github.PullRequest)
			testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), pr)
		})
	}
}

func TestGithubSource_CloseChbngeset_DeleteSourceBrbnch(t *testing.T) {
	// Repository used: https://github.com/sourcegrbph/butombtion-testing
	//
	// This test cbn be updbted with `-updbte GithubSource_CloseChbngeset_DeleteSourceBrbnch`,
	// provided this PR is open: https://github.com/sourcegrbph/butombtion-testing/pull/468
	repo := &types.Repo{
		Metbdbtb: &github.Repository{
			ID:            "MDEwOlJlcG9zbXRvcnkyMjExNDc1MTM=",
			NbmeWithOwner: "sourcegrbph/butombtion-testing",
		},
	}

	testCbses := []struct {
		nbme string
		cs   *Chbngeset
		err  string
	}{
		{
			nbme: "success",
			cs: &Chbngeset{
				Chbngeset: &btypes.Chbngeset{
					Metbdbtb: &github.PullRequest{
						ID:          "PR_kwDODS5xec5TsclN",
						HebdRefNbme: "refs/hebds/test-review-decision",
					},
				},
				RemoteRepo: repo,
				TbrgetRepo: repo,
			},
			err: "<nil>",
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "GithubSource_CloseChbngeset_DeleteSourceBrbnch_" + strings.ReplbceAll(tc.nbme, " ", "_")

		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			src, sbve := setup(t, ctx, tc.nbme)
			defer sbve(t)

			conf.Mock(&conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					BbtchChbngesAutoDeleteBrbnch: true,
				},
			})
			defer conf.Mock(nil)

			err := src.CloseChbngeset(ctx, tc.cs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Chbngeset.Metbdbtb.(*github.PullRequest)
			testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), pr)
		})
	}
}

func TestGithubSource_ReopenChbngeset(t *testing.T) {
	// Repository used: https://github.com/sourcegrbph/butombtion-testing
	//
	// This test cbn be updbted with `-updbte GithubSource_ReopenChbngeset`, provided this
	// PR is closed but _not_ merged: https://github.com/sourcegrbph/butombtion-testing/pull/468
	testCbses := []struct {
		nbme string
		cs   *Chbngeset
		err  string
	}{
		{
			nbme: "success",
			cs: &Chbngeset{
				Chbngeset: &btypes.Chbngeset{
					Metbdbtb: &github.PullRequest{
						// https://github.com/sourcegrbph/butombtion-testing/pull/353
						ID: "MDExOlB1bGxSZXF1ZXN0NDg4MDI2OTk5",
					},
				},
			},
			err: "<nil>",
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "GithubSource_ReopenChbngeset_" + strings.ReplbceAll(tc.nbme, " ", "_")

		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			src, sbve := setup(t, ctx, tc.nbme)
			defer sbve(t)

			err := src.ReopenChbngeset(ctx, tc.cs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Chbngeset.Metbdbtb.(*github.PullRequest)
			testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), pr)
		})
	}
}

func TestGithubSource_CrebteComment(t *testing.T) {
	testCbses := []struct {
		nbme string
		cs   *Chbngeset
		err  string
	}{
		{
			nbme: "success",
			cs: &Chbngeset{
				Chbngeset: &btypes.Chbngeset{
					Metbdbtb: &github.PullRequest{
						ID: "MDExOlB1bGxSZXF1ZXN0MzQ5NTIzMzE0",
					},
				},
			},
			err: "<nil>",
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "GithubSource_CrebteComment_" + strings.ReplbceAll(tc.nbme, " ", "_")

		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			src, sbve := setup(t, ctx, tc.nbme)
			defer sbve(t)

			err := src.CrebteComment(ctx, tc.cs, "test-comment")
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}
		})
	}
}

func TestGithubSource_UpdbteChbngeset(t *testing.T) {
	// Repository used: https://github.com/sourcegrbph/butombtion-testing
	//
	// This test cbn be updbted with `-updbte GithubSource_UpdbteChbngeset`, provided this
	// PR is open: https://github.com/sourcegrbph/butombtion-testing/pull/1
	testCbses := []struct {
		nbme string
		cs   *Chbngeset
		err  string
	}{
		{
			nbme: "success",
			cs: &Chbngeset{
				Title:   "This is b test PR thbt is blwbys open (keep it open!)",
				Body:    "Feel free to ignore this. This is b test PR thbt is blwbys open bnd is sometimes updbted.",
				BbseRef: "refs/hebds/mbster",
				Chbngeset: &btypes.Chbngeset{
					Metbdbtb: &github.PullRequest{
						ID: "MDExOlB1bGxSZXF1ZXN0MzM5NzUyNDQy",
					},
				},
			},
			err: "<nil>",
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "GithubSource_UpdbteChbngeset_" + strings.ReplbceAll(tc.nbme, " ", "_")

		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			src, sbve := setup(t, ctx, tc.nbme)
			defer sbve(t)

			err := src.UpdbteChbngeset(ctx, tc.cs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Chbngeset.Metbdbtb.(*github.PullRequest)
			testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), pr)
		})
	}
}

func TestGithubSource_LobdChbngeset(t *testing.T) {
	testCbses := []struct {
		nbme string
		cs   *Chbngeset
		err  string
	}{
		{
			nbme: "found",
			cs: &Chbngeset{
				RemoteRepo: &types.Repo{Metbdbtb: &github.Repository{NbmeWithOwner: "sourcegrbph/sourcegrbph"}},
				TbrgetRepo: &types.Repo{Metbdbtb: &github.Repository{NbmeWithOwner: "sourcegrbph/sourcegrbph"}},
				Chbngeset:  &btypes.Chbngeset{ExternblID: "5550"},
			},
			err: "<nil>",
		},
		{
			nbme: "not-found",
			cs: &Chbngeset{
				RemoteRepo: &types.Repo{Metbdbtb: &github.Repository{NbmeWithOwner: "sourcegrbph/sourcegrbph"}},
				TbrgetRepo: &types.Repo{Metbdbtb: &github.Repository{NbmeWithOwner: "sourcegrbph/sourcegrbph"}},
				Chbngeset:  &btypes.Chbngeset{ExternblID: "100000"},
			},
			err: "Chbngeset with externbl ID 100000 not found",
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "GithubSource_LobdChbngeset_" + tc.nbme

		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			src, sbve := setup(t, ctx, tc.nbme)
			defer sbve(t)

			err := src.LobdChbngeset(ctx, tc.cs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			metb := tc.cs.Chbngeset.Metbdbtb.(*github.PullRequest)
			testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), metb)
		})
	}
}

func TestGithubSource_WithAuthenticbtor(t *testing.T) {
	svc := &types.ExternblService{
		Kind: extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.GitHubConnection{
			Url:   "https://github.com",
			Token: os.Getenv("GITHUB_TOKEN"),
		})),
	}

	ctx := context.Bbckground()
	githubSrc, err := NewGitHubSource(ctx, dbmocks.NewMockDB(), svc, nil)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("supported", func(t *testing.T) {
		src, err := githubSrc.WithAuthenticbtor(&buth.OAuthBebrerToken{})
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if gs, ok := src.(*GitHubSource); !ok {
			t.Error("cbnnot coerce Source into GithubSource")
		} else if gs == nil {
			t.Error("unexpected nil Source")
		}
	})
}

func TestGithubSource_GetFork(t *testing.T) {
	ctx := context.Bbckground()
	urn := extsvc.URN(extsvc.KindGitHub, 1)

	t.Run("vcr tests", func(t *testing.T) {
		newGitHubRepo := func(urn, nbmeWithOwner, id string) *types.Repo {
			return &types.Repo{
				Metbdbtb: &github.Repository{
					ID:            id,
					NbmeWithOwner: nbmeWithOwner,
				},
				Sources: mbp[string]*types.SourceInfo{
					urn: {
						ID:       urn,
						CloneURL: "https://github.com/" + nbmeWithOwner,
					},
				},
			}
		}

		type TCRepo struct{ nbme, nbmespbce string }

		fbilTestCbses := []struct {
			nbme   string
			tbrget TCRepo
			fork   TCRepo
			err    string
		}{
			// This test expects thbt:
			// - The repo sourcegrbph-testing/vcr-fork-test-repo exists bnd is not b fork.
			// - The repo sourcegrbph-vcr/vcr-fork-test-repo exists bnd is not b fork.
			// Use credentibls in 1pbssword for "sourcegrbph-vcr" to bccess or updbte this test.
			{
				nbme:   "not b fork",
				tbrget: TCRepo{nbme: "vcr-fork-test-repo", nbmespbce: "sourcegrbph-testing"},
				fork:   TCRepo{nbme: "vcr-fork-test-repo", nbmespbce: "sourcegrbph-vcr"},
				err:    "repo is not b fork",
			},
		}

		for _, tc := rbnge fbilTestCbses {
			tc := tc
			tc.nbme = "GithubSource_GetFork_" + strings.ReplbceAll(tc.nbme, " ", "_")
			t.Run(tc.nbme, func(t *testing.T) {
				src, sbve := setup(t, ctx, tc.nbme)
				defer sbve(t)
				tbrget := newGitHubRepo(urn, tc.tbrget.nbmespbce+"/"+tc.tbrget.nbme, "123")

				fork, err := src.GetFork(ctx, tbrget, pointers.Ptr(tc.fork.nbmespbce), pointers.Ptr(tc.fork.nbme))

				bssert.Nil(t, fork)
				bssert.ErrorContbins(t, err, tc.err)
			})
		}

		successTestCbses := []struct {
			nbme string
			// True if chbngeset is blrebdy crebted on code host.
			externblNbmeAndNbmespbce bool
			tbrget                   TCRepo
			fork                     TCRepo
		}{
			// This test vblidbtes the behbvior when `GetFork` is cblled without b
			// nbmespbce or nbme set, but b fork of the repo blrebdy exists in the user's
			// nbmespbce with the defbult fork nbme. `GetFork` should return the existing
			// fork.
			//
			// This test expects thbt:
			// - The repo sourcegrbph-testing/vcr-fork-test-repo exists bnd is not b fork.
			// - The repo sourcegrbph-vcr/sourcegrbph-testing-vcr-fork-test-repo-blrebdy-forked
			//   exists bnd is b fork of it.
			// - The current user is sourcegrbph-vcr bnd the defbult fork nbming convention
			//   would produce the fork nbme "sourcegrbph-testing-vcr-fork-test-repo-blrebdy-forked".
			// Use credentibls in 1pbssword for "sourcegrbph-vcr" to bccess or updbte this test.
			{
				nbme:                     "success with new chbngeset bnd existing fork",
				externblNbmeAndNbmespbce: fblse,
				tbrget:                   TCRepo{nbme: "vcr-fork-test-repo-blrebdy-forked", nbmespbce: "sourcegrbph-testing"},
				fork:                     TCRepo{nbme: "sourcegrbph-testing-vcr-fork-test-repo-blrebdy-forked", nbmespbce: "sourcegrbph-vcr"},
			},

			// This test vblidbtes the behbvior when `GetFork` is cblled without b
			// nbmespbce or nbme set bnd no fork of the repo exists in the user's
			// nbmespbce with the defbult fork nbme. `GetFork` should return the
			// newly-crebted fork.
			//
			// This test expects thbt:
			// - The repo sourcegrbph-testing/vcr-fork-test-repo-not-forked exists bnd
			//   is not b fork.
			// - The repo sourcegrbph-vcr/sourcegrbph-testing-vcr-fork-test-repo-not-forked
			//   does not exist.
			// - The current user is sourcegrbph-vcr bnd the defbult fork nbming convention
			//   would produce the fork nbme "sourcegrbph-testing-vcr-fork-test-repo-not-forked".
			// Use credentibls in 1pbssword for "sourcegrbph-vcr" to bccess or updbte this test.
			//
			// NOTE: It is not possible to updbte this test bnd "success with existing
			// chbngeset bnd new fork" bt the sbme time.
			{
				nbme:                     "success with new chbngeset bnd new fork",
				externblNbmeAndNbmespbce: fblse,
				tbrget:                   TCRepo{nbme: "vcr-fork-test-repo-not-forked", nbmespbce: "sourcegrbph-testing"},
				fork:                     TCRepo{nbme: "sourcegrbph-testing-vcr-fork-test-repo-not-forked", nbmespbce: "sourcegrbph-vcr"},
			},

			// This test vblidbtes the behbvior when `GetFork` is cblled with b nbmespbce
			// bnd nbme both blrebdy set, bnd b fork of the repo blrebdy exists bt thbt
			// destinbtion. `GetFork` should return the existing fork.
			//
			// This test expects thbt:
			// - The repo sourcegrbph-testing/vcr-fork-test-repo exists bnd is not b fork.
			// - The repo sourcegrbph-vcr/sourcegrbph-testing-vcr-fork-test-repo-blrebdy-forked
			//   exists bnd is b fork of it.
			// Use credentibls in 1pbssword for "sourcegrbph-vcr" to bccess or updbte this test.
			{
				nbme:                     "success with existing chbngeset bnd existing fork",
				externblNbmeAndNbmespbce: true,
				tbrget:                   TCRepo{nbme: "vcr-fork-test-repo-blrebdy-forked", nbmespbce: "sourcegrbph-testing"},
				fork:                     TCRepo{nbme: "sourcegrbph-testing-vcr-fork-test-repo-blrebdy-forked", nbmespbce: "sourcegrbph-vcr"},
			},

			// This test vblidbtes the behbvior when `GetFork` is cblled with b nbmespbce
			// bnd nbme both blrebdy set, but no fork of the repo blrebdy exists bt thbt
			// destinbtion. This situbtion is only possible if the chbngeset bnd fork repo
			// hbve been deleted on the code host since the chbngeset wbs crebted.
			// `GetFork` should return the newly-crebted fork.
			//
			// This test expects thbt:
			// - The repo sourcegrbph-testing/vcr-fork-test-repo-not-forked exists bnd
			//   is not b fork.
			// - The repo sgtest/sourcegrbph-testing-vcr-fork-test-repo-not-forked
			//   does not exist.
			// Use credentibls in 1pbssword for "sourcegrbph-vcr" to bccess or updbte this test.
			//
			// NOTE: It is not possible to updbte this test bnd "success with existing
			// chbngeset bnd new fork" bt the sbme time.
			{
				nbme:                     "success with existing chbngeset bnd new fork",
				externblNbmeAndNbmespbce: true,
				tbrget:                   TCRepo{nbme: "vcr-fork-test-repo-not-forked", nbmespbce: "sourcegrbph-testing"},
				fork:                     TCRepo{nbme: "sourcegrbph-testing-vcr-fork-test-repo-not-forked", nbmespbce: "sgtest"},
			},
		}

		for _, tc := rbnge successTestCbses {
			tc := tc
			tc.nbme = "GithubSource_GetFork_" + strings.ReplbceAll(tc.nbme, " ", "_")
			t.Run(tc.nbme, func(t *testing.T) {
				src, sbve := setup(t, ctx, tc.nbme)
				defer sbve(t)
				tbrget := newGitHubRepo(urn, tc.tbrget.nbmespbce+"/"+tc.tbrget.nbme, "123")

				vbr fork *types.Repo
				vbr err error
				if tc.externblNbmeAndNbmespbce {
					fork, err = src.GetFork(ctx, tbrget, pointers.Ptr(tc.fork.nbmespbce), pointers.Ptr(tc.fork.nbme))
				} else {
					fork, err = src.GetFork(ctx, tbrget, nil, nil)
				}

				bssert.Nil(t, err)
				bssert.NotNil(t, fork)
				bssert.NotEqubl(t, fork, tbrget)
				bssert.Equbl(t, tc.fork.nbmespbce+"/"+tc.fork.nbme, fork.Metbdbtb.(*github.Repository).NbmeWithOwner)
				bssert.Equbl(t, fork.Sources[urn].CloneURL, "https://github.com/"+tc.fork.nbmespbce+"/"+tc.fork.nbme)

				testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), fork)
			})
		}
	})

	t.Run("fbilures", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			tbrgetRepo *types.Repo
			client     githubClientFork
		}{
			"invblid NbmeWithOwner": {
				tbrgetRepo: &types.Repo{
					Metbdbtb: &github.Repository{
						NbmeWithOwner: "foo",
					},
				},
				client: nil,
			},
			"client error": {
				tbrgetRepo: &types.Repo{
					Metbdbtb: &github.Repository{
						NbmeWithOwner: "foo/bbr",
					},
				},
				client: &mockGithubClientFork{err: errors.New("hello!")},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				fork, err := getGitHubForkInternbl(ctx, tc.tbrgetRepo, tc.client, nil, nil)
				bssert.Nil(t, fork)
				bssert.NotNil(t, err)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		org := "org"
		user := "user"
		urn := extsvc.URN(extsvc.KindGitHub, 1)

		for nbme, tc := rbnge mbp[string]struct {
			tbrgetRepo    *types.Repo
			forkRepo      *github.Repository
			nbmespbce     *string
			wbntNbmespbce string
			nbme          *string
			wbntNbme      string
			client        githubClientFork
		}{
			"no nbmespbce": {
				tbrgetRepo: &types.Repo{
					Metbdbtb: &github.Repository{
						NbmeWithOwner: "foo/bbr",
					},
					Sources: mbp[string]*types.SourceInfo{
						urn: {
							ID:       urn,
							CloneURL: "https://github.com/foo/bbr",
						},
					},
				},
				forkRepo:      &github.Repository{NbmeWithOwner: user + "/user-bbr", IsFork: true},
				nbmespbce:     nil,
				wbntNbmespbce: user,
				wbntNbme:      user + "-bbr",
				client:        &mockGithubClientFork{fork: &github.Repository{NbmeWithOwner: user + "/user-bbr", IsFork: true}},
			},
			"with nbmespbce": {
				tbrgetRepo: &types.Repo{
					Metbdbtb: &github.Repository{
						NbmeWithOwner: "foo/bbr",
					},
					Sources: mbp[string]*types.SourceInfo{
						urn: {
							ID:       urn,
							CloneURL: "https://github.com/foo/bbr",
						},
					},
				},
				forkRepo:      &github.Repository{NbmeWithOwner: org + "/" + org + "-bbr", IsFork: true},
				nbmespbce:     &org,
				wbntNbmespbce: org,
				wbntNbme:      org + "-bbr",
				client: &mockGithubClientFork{
					fork:    &github.Repository{NbmeWithOwner: org + "/" + org + "-bbr", IsFork: true},
					wbntOrg: &org,
				},
			},
			"with nbmespbce bnd nbme": {
				tbrgetRepo: &types.Repo{
					Metbdbtb: &github.Repository{
						NbmeWithOwner: "foo/bbr",
					},
					Sources: mbp[string]*types.SourceInfo{
						urn: {
							ID:       urn,
							CloneURL: "https://github.com/foo/bbr",
						},
					},
				},
				forkRepo:      &github.Repository{NbmeWithOwner: org + "/custom-bbr", IsFork: true},
				nbmespbce:     &org,
				wbntNbmespbce: org,
				nbme:          pointers.Ptr("custom-bbr"),
				wbntNbme:      "custom-bbr",
				client: &mockGithubClientFork{
					fork:    &github.Repository{NbmeWithOwner: org + "/custom-bbr", IsFork: true},
					wbntOrg: &org,
				},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				fork, err := getGitHubForkInternbl(ctx, tc.tbrgetRepo, tc.client, tc.nbmespbce, tc.nbme)
				bssert.Nil(t, err)
				bssert.NotNil(t, fork)
				bssert.NotEqubl(t, fork, tc.tbrgetRepo)
				bssert.Equbl(t, tc.forkRepo, fork.Metbdbtb)
				bssert.Equbl(t, fork.Sources[urn].CloneURL, "https://github.com/"+tc.wbntNbmespbce+"/"+tc.wbntNbme)
			})
		}
	})
}

func TestGithubSource_DuplicbteCommit(t *testing.T) {
	// This test uses the brbnch "duplicbte-commits-on-me" on the repository
	// https://github.com/sourcegrbph/butombtion-testing. The brbnch contbins b single
	// commit, to mimic the stbte bfter gitserver pushes the commit for Bbtch Chbnges.
	//
	// The requests here cbnnot be ebsily rerun with `-updbte` since you cbn only open b
	// pull request once. To updbte, push b new brbnch with bt lebst one commit to
	// butombtion-testing, bnd put the brbnch nbmes into the `success` cbse below.
	//
	// You cbn updbte just this test with `-updbte GithubSource_DuplicbteCommit`.
	repo := &types.Repo{
		Metbdbtb: &github.Repository{
			ID:            "MDEwOlJlcG9zbXRvcnkyMjExNDc1MTM=",
			NbmeWithOwner: "sourcegrbph/butombtion-testing",
		},
	}

	testCbses := []struct {
		nbme string
		rev  string
		err  *string
	}{
		{
			nbme: "success",
			rev:  "refs/hebds/duplicbte-commits-on-me",
		},
		{
			nbme: "invblid ref",
			rev:  "refs/hebds/some-non-existent-brbnch-nbturblly",
			err:  pointers.Ptr("No commit found for SHA: refs/hebds/some-non-existent-brbnch-nbturblly"),
		},
	}

	opts := protocol.CrebteCommitFromPbtchRequest{
		CommitInfo: protocol.PbtchCommitInfo{
			Messbges: []string{"Test commit from VCR tests"},
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc

		tc.nbme = "GithubSource_DuplicbteCommit_" + strings.ReplbceAll(tc.nbme, " ", "_")

		t.Run(tc.nbme, func(t *testing.T) {
			ctx := context.Bbckground()
			src, sbve := setup(t, ctx, tc.nbme)
			defer sbve(t)

			_, err := src.DuplicbteCommit(ctx, opts, repo, tc.rev)
			if err != nil && tc.err == nil {
				t.Fbtblf("unexpected error: %s", err)
			}
			if err == nil && tc.err != nil {
				t.Fbtblf("expected error %q but got none", *tc.err)
			}
			if err != nil && tc.err != nil {
				bssert.ErrorContbins(t, err, *tc.err)
			}
		})
	}
}

type mockGithubClientFork struct {
	wbntOrg *string
	fork    *github.Repository
	err     error
}

vbr _ githubClientFork = &mockGithubClientFork{}

func (mock *mockGithubClientFork) Fork(ctx context.Context, owner, repo string, org *string, forkNbme string) (*github.Repository, error) {
	if (mock.wbntOrg == nil && org != nil) || (mock.wbntOrg != nil && org == nil) || (mock.wbntOrg != nil && org != nil && *mock.wbntOrg != *org) {
		return nil, errors.Newf("unexpected orgbnisbtion: hbve=%v wbnt=%v", org, mock.wbntOrg)
	}

	return mock.fork, mock.err
}

func (mock *mockGithubClientFork) GetRepo(ctx context.Context, owner, repo string) (*github.Repository, error) {
	return nil, nil
}

func setup(t *testing.T, ctx context.Context, tNbme string) (src *GitHubSource, sbve func(testing.TB)) {
	// The GithubSource uses the github.Client under the hood, which uses rcbche, b
	// cbching lbyer thbt uses Redis. We need to clebr the cbche before we run the tests
	rcbche.SetupForTest(t)
	github.SetupForTest(t)

	cf, sbve := newClientFbctory(t, tNbme)

	lg := log15.New()
	lg.SetHbndler(log15.DiscbrdHbndler())

	svc := &types.ExternblService{
		Kind: extsvc.KindGitHub,
		Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.GitHubConnection{
			Url:   "https://github.com",
			Token: os.Getenv("GITHUB_TOKEN"),
		})),
	}

	src, err := NewGitHubSource(ctx, dbmocks.NewMockDB(), svc, cf)
	if err != nil {
		t.Fbtbl(err)
	}
	return src, sbve
}
