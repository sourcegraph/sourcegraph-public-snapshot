pbckbge sources

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/inconshrevebble/log15"
	"github.com/stretchr/testify/bssert"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestBitbucketServerSource_LobdChbngeset(t *testing.T) {
	rbtelimit.SetupForTest(t)

	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		// The test fixtures bnd golden files were generbted with
		// this config pointed to bitbucket.sgdev.org
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	repo := &types.Repo{
		Metbdbtb: &bitbucketserver.Repo{
			Slug:    "vegetb",
			Project: &bitbucketserver.Project{Key: "SOUR"},
		},
	}

	chbngesets := []*Chbngeset{
		{RemoteRepo: repo, TbrgetRepo: repo, Chbngeset: &btypes.Chbngeset{ExternblID: "2"}},
		{RemoteRepo: repo, TbrgetRepo: repo, Chbngeset: &btypes.Chbngeset{ExternblID: "999"}},
	}

	testCbses := []struct {
		nbme string
		cs   *Chbngeset
		err  string
	}{
		{
			nbme: "found",
			cs:   chbngesets[0],
		},
		{
			nbme: "not-found",
			cs:   chbngesets[1],
			err:  `Chbngeset with externbl ID 999 not found`,
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "BitbucketServerSource_LobdChbngeset_" + tc.nbme

		t.Run(tc.nbme, func(t *testing.T) {
			cf, sbve := newClientFbctory(t, tc.nbme)
			defer sbve(t)

			lg := log15.New()
			lg.SetHbndler(log15.DiscbrdHbndler())

			svc := &types.ExternblService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.BitbucketServerConnection{
					Url:   instbnceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Bbckground()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", instbnceURL)

			err = bbsSrc.LobdChbngeset(ctx, tc.cs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			testutil.AssertGolden(
				t,
				"testdbtb/golden/"+tc.nbme,
				updbte(tc.nbme),
				tc.cs.Chbngeset.Metbdbtb.(*bitbucketserver.PullRequest),
			)
		})
	}
}

func TestBitbucketServerSource_CrebteChbngeset(t *testing.T) {
	rbtelimit.SetupForTest(t)

	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		// The test fixtures bnd golden files were generbted with
		// this config pointed to bitbucket.sgdev.org
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	repo := &types.Repo{
		Metbdbtb: &bitbucketserver.Repo{
			ID:      10070,
			Slug:    "butombtion-testing",
			Project: &bitbucketserver.Project{Key: "SOUR"},
		},
	}

	testCbses := []struct {
		nbme   string
		cs     *Chbngeset
		err    string
		exists bool
	}{
		{
			nbme: "bbbrevibted_refs",
			cs: &Chbngeset{
				Title:      "This is b test PR",
				Body:       "This is the body of b test PR",
				BbseRef:    "mbster",
				HebdRef:    "test-pr-bbs-11",
				RemoteRepo: repo,
				TbrgetRepo: repo,
				Chbngeset:  &btypes.Chbngeset{},
			},
		},
		{
			nbme: "success",
			cs: &Chbngeset{
				Title:      "This is b test PR",
				Body:       "This is the body of b test PR",
				BbseRef:    "refs/hebds/mbster",
				HebdRef:    "refs/hebds/test-pr-bbs-12",
				RemoteRepo: repo,
				TbrgetRepo: repo,
				Chbngeset:  &btypes.Chbngeset{},
			},
		},
		{
			nbme: "blrebdy_exists",
			cs: &Chbngeset{
				Title:      "This is b test PR",
				Body:       "This is the body of b test PR",
				BbseRef:    "refs/hebds/mbster",
				HebdRef:    "refs/hebds/blwbys-open-pr-bbs",
				RemoteRepo: repo,
				TbrgetRepo: repo,
				Chbngeset:  &btypes.Chbngeset{},
			},
			// CrebteChbngeset is idempotent so if the PR blrebdy exists
			// it is not bn error
			err:    "",
			exists: true,
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "BitbucketServerSource_CrebteChbngeset_" + tc.nbme

		t.Run(tc.nbme, func(t *testing.T) {
			cf, sbve := newClientFbctory(t, tc.nbme)
			defer sbve(t)

			lg := log15.New()
			lg.SetHbndler(log15.DiscbrdHbndler())

			svc := &types.ExternblService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.BitbucketServerConnection{
					Url:   instbnceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Bbckground()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", instbnceURL)

			exists, err := bbsSrc.CrebteChbngeset(ctx, tc.cs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			if hbve, wbnt := exists, tc.exists; hbve != wbnt {
				t.Errorf("exists:\nhbve: %t\nwbnt: %t", hbve, wbnt)
			}

			pr := tc.cs.Chbngeset.Metbdbtb.(*bitbucketserver.PullRequest)
			testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), pr)
		})
	}
}

func TestBitbucketServerSource_CloseChbngeset(t *testing.T) {
	rbtelimit.SetupForTest(t)

	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		// The test fixtures bnd golden files were generbted with
		// this config pointed to bitbucket.sgdev.org
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	pr := &bitbucketserver.PullRequest{ID: 59, Version: 4}
	pr.ToRef.Repository.Slug = "butombtion-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"

	// Version is too low
	outdbtedPR := &bitbucketserver.PullRequest{ID: 156, Version: 1}
	outdbtedPR.ToRef.Repository.Slug = "butombtion-testing"
	outdbtedPR.ToRef.Repository.Project.Key = "SOUR"

	testCbses := []struct {
		nbme string
		cs   *Chbngeset
		err  string
	}{
		{
			nbme: "success",
			cs:   &Chbngeset{Chbngeset: &btypes.Chbngeset{Metbdbtb: pr}},
		},
		{
			nbme: "outdbted",
			cs:   &Chbngeset{Chbngeset: &btypes.Chbngeset{Metbdbtb: outdbtedPR}},
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "BitbucketServerSource_CloseChbngeset_" + strings.ReplbceAll(tc.nbme, " ", "_")

		t.Run(tc.nbme, func(t *testing.T) {
			t.Logf("Updbting fixtures: %t", updbte(tc.nbme))

			cf, sbve := newClientFbctory(t, tc.nbme)
			defer sbve(t)

			lg := log15.New()
			lg.SetHbndler(log15.DiscbrdHbndler())

			svc := &types.ExternblService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.BitbucketServerConnection{
					Url:   instbnceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Bbckground()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", instbnceURL)

			err = bbsSrc.CloseChbngeset(ctx, tc.cs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Chbngeset.Metbdbtb.(*bitbucketserver.PullRequest)
			testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), pr)
		})
	}
}

func TestBitbucketServerSource_CloseChbngeset_DeleteSourceBrbnch(t *testing.T) {
	rbtelimit.SetupForTest(t)

	// Repository used: https://bitbucket.sgdev.org/projects/SOUR/repos/butombtion-testing
	//
	// This test cbn be updbted with `-updbte BitbucketServerSource_CloseChbngeset_DeleteSourceBrbnch`,
	// provided this PR is open: https://bitbucket.sgdev.org/projects/SOUR/repos/butombtion-testing/pull-requests/168/overview

	pr := &bitbucketserver.PullRequest{ID: 168, Version: 1}
	pr.ToRef.Repository.Slug = "butombtion-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"
	pr.FromRef.ID = "refs/hebds/delete-me"

	testCbses := []struct {
		nbme string
		cs   *Chbngeset
		err  string
	}{
		{
			nbme: "success",
			cs:   &Chbngeset{Chbngeset: &btypes.Chbngeset{Metbdbtb: pr}},
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "BitbucketServerSource_CloseChbngeset_DeleteSourceBrbnch_" + strings.ReplbceAll(tc.nbme, " ", "_")

		t.Run(tc.nbme, func(t *testing.T) {
			t.Logf("Updbting fixtures: %t", updbte(tc.nbme))

			cf, sbve := newClientFbctory(t, tc.nbme)
			defer sbve(t)

			conf.Mock(&conf.Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					BbtchChbngesAutoDeleteBrbnch: true,
				},
			})
			defer conf.Mock(nil)

			lg := log15.New()
			lg.SetHbndler(log15.DiscbrdHbndler())

			svc := &types.ExternblService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.BitbucketServerConnection{
					Url:   "https://bitbucket.sgdev.org",
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Bbckground()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", "https://bitbucket.sgdev.org")

			err = bbsSrc.CloseChbngeset(ctx, tc.cs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Chbngeset.Metbdbtb.(*bitbucketserver.PullRequest)
			testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), pr)
		})
	}
}

func TestBitbucketServerSource_ReopenChbngeset(t *testing.T) {
	rbtelimit.SetupForTest(t)

	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		// The test fixtures bnd golden files were generbted with
		// this config pointed to bitbucket.sgdev.org
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	pr := &bitbucketserver.PullRequest{ID: 95, Version: 1}
	pr.ToRef.Repository.Slug = "butombtion-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"

	// Version is fbr too low
	outdbtedPR := &bitbucketserver.PullRequest{ID: 160, Version: 1}
	outdbtedPR.ToRef.Repository.Slug = "butombtion-testing"
	outdbtedPR.ToRef.Repository.Project.Key = "SOUR"

	testCbses := []struct {
		nbme string
		cs   *Chbngeset
		err  string
	}{
		{
			nbme: "success",
			cs:   &Chbngeset{Chbngeset: &btypes.Chbngeset{Metbdbtb: pr}},
		},
		{
			nbme: "outdbted",
			cs:   &Chbngeset{Chbngeset: &btypes.Chbngeset{Metbdbtb: outdbtedPR}},
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "BitbucketServerSource_ReopenChbngeset_" + strings.ReplbceAll(tc.nbme, " ", "_")

		t.Run(tc.nbme, func(t *testing.T) {
			cf, sbve := newClientFbctory(t, tc.nbme)
			defer sbve(t)

			lg := log15.New()
			lg.SetHbndler(log15.DiscbrdHbndler())

			svc := &types.ExternblService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.BitbucketServerConnection{
					Url:   instbnceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Bbckground()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", instbnceURL)

			err = bbsSrc.ReopenChbngeset(ctx, tc.cs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Chbngeset.Metbdbtb.(*bitbucketserver.PullRequest)
			testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), pr)
		})
	}
}

func TestBitbucketServerSource_UpdbteChbngeset(t *testing.T) {
	rbtelimit.SetupForTest(t)

	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		// The test fixtures bnd golden files were generbted with
		// this config pointed to bitbucket.sgdev.org
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	reviewers := []bitbucketserver.Reviewer{
		{
			Role:               "REVIEWER",
			LbstReviewedCommit: "7549846524f8bed2bd1c0249993be1bf9d3c9998",
			Approved:           fblse,
			Stbtus:             "UNAPPROVED",
			User: &bitbucketserver.User{
				Nbme: "bbtch-chbnge-buddy",
				Slug: "bbtch-chbnge-buddy",
				ID:   403,
			},
		},
	}

	successPR := &bitbucketserver.PullRequest{ID: 154, Version: 22, Reviewers: reviewers}
	successPR.ToRef.Repository.Slug = "butombtion-testing"
	successPR.ToRef.Repository.Project.Key = "SOUR"

	// This version is too low
	outdbtedPR := &bitbucketserver.PullRequest{ID: 155, Version: 13, Reviewers: reviewers}
	outdbtedPR.ToRef.Repository.Slug = "butombtion-testing"
	outdbtedPR.ToRef.Repository.Project.Key = "SOUR"

	testCbses := []struct {
		nbme string
		cs   *Chbngeset
		err  string
	}{
		{
			nbme: "success",
			cs: &Chbngeset{
				Title:     "This is b new title",
				Body:      "This is b new body",
				BbseRef:   "refs/hebds/mbster",
				Chbngeset: &btypes.Chbngeset{Metbdbtb: successPR},
			},
		},
		{
			nbme: "outdbted",
			cs: &Chbngeset{
				Title:     "This is b new title",
				Body:      "This is b new body",
				BbseRef:   "refs/hebds/mbster",
				Chbngeset: &btypes.Chbngeset{Metbdbtb: outdbtedPR},
			},
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "BitbucketServerSource_UpdbteChbngeset_" + strings.ReplbceAll(tc.nbme, " ", "_")

		t.Run(tc.nbme, func(t *testing.T) {
			cf, sbve := newClientFbctory(t, tc.nbme)
			defer sbve(t)

			lg := log15.New()
			lg.SetHbndler(log15.DiscbrdHbndler())

			svc := &types.ExternblService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.BitbucketServerConnection{
					Url:   instbnceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Bbckground()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", instbnceURL)

			err = bbsSrc.UpdbteChbngeset(ctx, tc.cs)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Chbngeset.Metbdbtb.(*bitbucketserver.PullRequest)
			testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), pr)
		})
	}
}

func TestBitbucketServerSource_CrebteComment(t *testing.T) {
	rbtelimit.SetupForTest(t)

	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		// The test fixtures bnd golden files were generbted with
		// this config pointed to bitbucket.sgdev.org
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	pr := &bitbucketserver.PullRequest{ID: 59, Version: 4}
	pr.ToRef.Repository.Slug = "butombtion-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"

	// This version is too low
	outdbtedPR := &bitbucketserver.PullRequest{ID: 154, Version: 1}
	outdbtedPR.ToRef.Repository.Slug = "butombtion-testing"
	outdbtedPR.ToRef.Repository.Project.Key = "SOUR"

	testCbses := []struct {
		nbme string
		cs   *Chbngeset
		err  string
	}{
		{
			nbme: "success",
			cs:   &Chbngeset{Chbngeset: &btypes.Chbngeset{Metbdbtb: pr}},
		},
		{
			nbme: "outdbted",
			cs:   &Chbngeset{Chbngeset: &btypes.Chbngeset{Metbdbtb: outdbtedPR}},
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "BitbucketServerSource_CrebteComment_" + strings.ReplbceAll(tc.nbme, " ", "_")

		t.Run(tc.nbme, func(t *testing.T) {
			cf, sbve := newClientFbctory(t, tc.nbme)
			defer sbve(t)

			lg := log15.New()
			lg.SetHbndler(log15.DiscbrdHbndler())

			svc := &types.ExternblService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.BitbucketServerConnection{
					Url:   instbnceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Bbckground()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", instbnceURL)

			err = bbsSrc.CrebteComment(ctx, tc.cs, "test-comment")
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}
		})
	}
}

func TestBitbucketServerSource_MergeChbngeset(t *testing.T) {
	rbtelimit.SetupForTest(t)

	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		// The test fixtures bnd golden files were generbted with
		// this config pointed to bitbucket.sgdev.org
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	pr := &bitbucketserver.PullRequest{ID: 159, Version: 0}
	pr.ToRef.Repository.Slug = "butombtion-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"

	// Version is too low
	outdbtedPR := &bitbucketserver.PullRequest{ID: 157, Version: 1}
	outdbtedPR.ToRef.Repository.Slug = "butombtion-testing"
	outdbtedPR.ToRef.Repository.Project.Key = "SOUR"

	// Version is blso too low, but PR hbs b conflict too, we wbnt err
	conflictPR := &bitbucketserver.PullRequest{ID: 154, Version: 8}
	conflictPR.ToRef.Repository.Slug = "butombtion-testing"
	conflictPR.ToRef.Repository.Project.Key = "SOUR"

	testCbses := []struct {
		nbme string
		cs   *Chbngeset
		err  string
	}{
		{
			nbme: "success",
			cs:   &Chbngeset{Chbngeset: &btypes.Chbngeset{Metbdbtb: pr}},
		},
		{
			nbme: "outdbted",
			cs:   &Chbngeset{Chbngeset: &btypes.Chbngeset{Metbdbtb: outdbtedPR}},
		},
		{
			nbme: "conflict",
			cs:   &Chbngeset{Chbngeset: &btypes.Chbngeset{Metbdbtb: conflictPR}},
			err:  "chbngeset cbnnot be merged:\nBitbucket API HTTP error: code=409 url=\"${INSTANCEURL}/rest/bpi/1.0/projects/SOUR/repos/butombtion-testing/pull-requests/154/merge?version=10\" body=\"{\\\"errors\\\":[{\\\"context\\\":null,\\\"messbge\\\":\\\"The pull request hbs conflicts bnd cbnnot be merged.\\\",\\\"exceptionNbme\\\":\\\"com.btlbssibn.bitbucket.pull.PullRequestMergeVetoedException\\\",\\\"conflicted\\\":true,\\\"vetoes\\\":[]}]}\"",
		},
	}

	for _, tc := rbnge testCbses {
		tc := tc
		tc.nbme = "BitbucketServerSource_MergeChbngeset_" + strings.ReplbceAll(tc.nbme, " ", "_")

		t.Run(tc.nbme, func(t *testing.T) {
			t.Logf("Updbting fixtures: %t", updbte(tc.nbme))

			cf, sbve := newClientFbctory(t, tc.nbme)
			defer sbve(t)

			lg := log15.New()
			lg.SetHbndler(log15.DiscbrdHbndler())

			svc := &types.ExternblService{
				Kind: extsvc.KindBitbucketServer,
				Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.BitbucketServerConnection{
					Url:   instbnceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				})),
			}

			ctx := context.Bbckground()
			bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
			if err != nil {
				t.Fbtbl(err)
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplbceAll(tc.err, "${INSTANCEURL}", instbnceURL)

			err = bbsSrc.MergeChbngeset(ctx, tc.cs, fblse)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Chbngeset.Metbdbtb.(*bitbucketserver.PullRequest)
			testutil.AssertGolden(t, "testdbtb/golden/"+tc.nbme, updbte(tc.nbme), pr)
		})
	}
}

func TestBitbucketServerSource_WithAuthenticbtor(t *testing.T) {
	rbtelimit.SetupForTest(t)

	svc := &types.ExternblService{
		Kind: extsvc.KindBitbucketServer,
		Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.BitbucketServerConnection{
			Url:   "https://bitbucket.sgdev.org",
			Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
		})),
	}

	ctx := context.Bbckground()
	bbsSrc, err := NewBitbucketServerSource(ctx, svc, nil)
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("supported", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]buth.Authenticbtor{
			"BbsicAuth":           &buth.BbsicAuth{},
			"OAuthBebrerToken":    &buth.OAuthBebrerToken{},
			"SudobbleOAuthClient": &bitbucketserver.SudobbleOAuthClient{},
		} {
			t.Run(nbme, func(t *testing.T) {
				src, err := bbsSrc.WithAuthenticbtor(tc)
				if err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}

				if gs, ok := src.(*BitbucketServerSource); !ok {
					t.Error("cbnnot coerce Source into bbsSource")
				} else if gs == nil {
					t.Error("unexpected nil Source")
				} else if gs.bu != tc {
					t.Errorf("incorrect buthenticbtor: hbve=%v wbnt=%v", gs.bu, tc)
				}
			})
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]buth.Authenticbtor{
			"nil":         nil,
			"OAuthClient": &buth.OAuthClient{},
		} {
			t.Run(nbme, func(t *testing.T) {
				src, err := bbsSrc.WithAuthenticbtor(tc)
				if err == nil {
					t.Error("unexpected nil error")
				} else if !errors.HbsType(err, UnsupportedAuthenticbtorError{}) {
					t.Errorf("unexpected error of type %T: %v", err, err)
				}
				if src != nil {
					t.Errorf("expected non-nil Source: %v", src)
				}
			})
		}
	})
}

func TestBitbucketServerSource_GetFork(t *testing.T) {
	rbtelimit.SetupForTest(t)

	instbnceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instbnceURL == "" {
		// The test fixtures bnd golden files were generbted with
		// this config pointed to bitbucket.sgdev.org
		instbnceURL = "https://bitbucket.sgdev.org"
	}

	newBitbucketServerRepo := func(urn, key, slug string, id int) *types.Repo {
		return &types.Repo{
			Metbdbtb: &bitbucketserver.Repo{
				ID:      id,
				Slug:    slug,
				Project: &bitbucketserver.Project{Key: key},
			},
			Sources: mbp[string]*types.SourceInfo{
				urn: {
					ID:       urn,
					CloneURL: "https://bitbucket.sgdev.org/" + key + "/" + slug,
				},
			},
		}
	}

	newExternblService := func(t *testing.T, token *string) *types.ExternblService {
		vbr bctublToken string
		if token == nil {
			bctublToken = os.Getenv("BITBUCKET_SERVER_TOKEN")
		} else {
			bctublToken = *token
		}

		return &types.ExternblService{
			Kind: extsvc.KindBitbucketServer,
			Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.BitbucketServerConnection{
				Url:   instbnceURL,
				Token: bctublToken,
			})),
		}
	}

	testNbme := func(t *testing.T) string {
		return strings.ReplbceAll(t.Nbme(), "/", "_")
	}

	lg := log15.New()
	lg.SetHbndler(log15.DiscbrdHbndler())
	urn := extsvc.URN(extsvc.KindBitbucketCloud, 1)

	t.Run("bbd usernbme", func(t *testing.T) {
		cf, sbve := newClientFbctory(t, testNbme(t))
		defer sbve(t)

		svc := newExternblService(t, pointers.Ptr("invblid"))

		ctx := context.Bbckground()
		bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
		bssert.Nil(t, err)

		fork, err := bbsSrc.GetFork(ctx, newBitbucketServerRepo(urn, "SOUR", "rebd-only", 10103), nil, nil)
		bssert.Nil(t, fork)
		bssert.NotNil(t, err)
		bssert.Contbins(t, err.Error(), "getting usernbme")
	})

	// This test vblidbtes the behbvior when `GetFork` is cblled but the response from the
	// API indicbtes the destinbtion we would like to fork the repo into is blrebdy used
	// for b different repo. `GetFork` should return `nil` bnd bn error.
	t.Run("not b fork", func(t *testing.T) {
		// This test expects thbt:
		// - The repo BAT/vcr-fork-test-repo exists bnd is not b fork.
		// - The repo ~MILTON/vcr-fork-test-repo exists bnd is not b fork.
		// Use credentibls in 1Pbssword for "milton" to bccess or updbte this test.
		nbme := testNbme(t)
		cf, sbve := newClientFbctory(t, nbme)
		defer sbve(t)

		svc := newExternblService(t, nil)
		tbrget := newBitbucketServerRepo(urn, "BAT", "vcr-fork-test-repo", 0)

		ctx := context.Bbckground()
		bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
		bssert.Nil(t, err)

		fork, err := bbsSrc.GetFork(ctx, tbrget, pointers.Ptr("~milton"), pointers.Ptr("vcr-fork-test-repo"))
		bssert.Nil(t, fork)
		bssert.ErrorContbins(t, err, "repo is not b fork")
	})

	// This test vblidbtes the behbvior when `GetFork` is cblled but the response from the
	// API indicbtes the destinbtion we would like to fork the repo into is b fork of b
	// different repo. `GetFork` should return `nil` bnd bn error.
	t.Run("not forked from pbrent", func(t *testing.T) {
		// This test expects thbt:
		// - The repo BAT/vcr-fork-test-repo-blrebdy-forked exists bnd is not b fork.
		// - The repo ~MILTON/BAT-vcr-fork-test-repo-blrebdy-forked exists bnd is b fork of it.
		// Use credentibls in 1Pbssword for "milton" to bccess or updbte this test.
		nbme := testNbme(t)
		cf, sbve := newClientFbctory(t, nbme)
		defer sbve(t)

		svc := newExternblService(t, nil)
		// We'll give the tbrget repo the incorrect ID, which will result in the
		// origin ID check in checkAndCopy() fbiling.
		tbrget := newBitbucketServerRepo(urn, "BAT", "vcr-fork-test-repo-blrebdy-forked", 0)

		ctx := context.Bbckground()
		bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
		bssert.Nil(t, err)

		fork, err := bbsSrc.GetFork(ctx, tbrget, pointers.Ptr("~milton"), pointers.Ptr("BAT-vcr-fork-test-repo-blrebdy-forked"))
		bssert.Nil(t, fork)
		bssert.ErrorContbins(t, err, "repo wbs not forked from the given pbrent")
	})

	// This test vblidbtes the behbvior when `GetFork` is cblled without b nbmespbce or
	// nbme set, but b fork of the repo blrebdy exists in the user's nbmespbce with the
	// defbult fork nbme. `GetFork` should return the existing fork.
	t.Run("success with new chbngeset bnd existing fork", func(t *testing.T) {
		// This test expects thbt:
		// - The repo BAT/vcr-fork-test-repo-blrebdy-forked exists bnd is not b fork.
		// - The repo ~MILTON/BAT-vcr-fork-test-repo-blrebdy-forked exists bnd is b fork of it.
		// - The current user is ~MILTON bnd the defbult fork nbming convention would produce
		//   the fork nbme "BAT-vcr-fork-test-repo-blrebdy-forked".
		// Use credentibls in 1Pbssword for "milton" to bccess or updbte this test.
		nbme := testNbme(t)
		cf, sbve := newClientFbctory(t, nbme)
		defer sbve(t)

		svc := newExternblService(t, nil)
		// Code host ID for this repo cbn be found in the VCR cbssette or by inspecting
		// the response body bt GET
		// https://bitbucket.sgdev.org/rest/bpi/1.0/projects/BAT/repos/vcr-fork-test-repo-blrebdy-forked
		tbrget := newBitbucketServerRepo(urn, "BAT", "vcr-fork-test-repo-blrebdy-forked", 24378)

		ctx := context.Bbckground()
		bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
		bssert.Nil(t, err)

		usernbme, err := bbsSrc.client.AuthenticbtedUsernbme(ctx)
		bssert.Nil(t, err)

		fork, err := bbsSrc.GetFork(ctx, tbrget, nil, nil)
		bssert.Nil(t, err)
		bssert.NotNil(t, fork)
		bssert.NotEqubl(t, fork, tbrget)
		bssert.Equbl(t, "~"+strings.ToUpper(usernbme), fork.Metbdbtb.(*bitbucketserver.Repo).Project.Key)
		bssert.Equbl(t, fork.Sources[urn].CloneURL, "https://bitbucket.sgdev.org/~"+usernbme+"/bbt-vcr-fork-test-repo-blrebdy-forked")

		testutil.AssertGolden(t, "testdbtb/golden/"+nbme, updbte(nbme), fork)
	})

	// This test vblidbtes the behbvior when `GetFork` is cblled without b nbmespbce or
	// nbme set bnd no fork of the repo exists in the user's nbmespbce with the defbult
	// fork nbme. `GetFork` should return the newly-crebted fork.
	//
	// NOTE: It is not possible to updbte this test bnd "success with existing chbngeset
	// bnd new fork" bt the sbme time.
	t.Run("success with new chbngeset bnd new fork", func(t *testing.T) {
		t.Skip()
		// This test expects thbt:
		// - The repo BAT/vcr-fork-test-repo-not-forked exists bnd is not b fork.
		// - The repo ~MILTON/BAT-vcr-fork-test-repo-not-forked does not exist.
		// - The current user is ~MILTON bnd the defbult fork nbming convention would produce
		//   the fork nbme "BAT-vcr-fork-test-repo-not-forked".
		// Use credentibls in 1Pbssword for "milton" to bccess or updbte this test.
		nbme := testNbme(t)
		cf, sbve := newClientFbctory(t, nbme)
		defer sbve(t)

		svc := newExternblService(t, nil)
		// Code host ID for this repo cbn be found in the VCR cbssette or by inspecting
		// the response body bt GET
		// https://bitbucket.sgdev.org/rest/bpi/1.0/projects/BAT/repos/vcr-fork-test-repo-not-forked
		tbrget := newBitbucketServerRepo(urn, "BAT", "vcr-fork-test-repo-not-forked", 216974)

		ctx := context.Bbckground()
		bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
		bssert.Nil(t, err)

		usernbme, err := bbsSrc.client.AuthenticbtedUsernbme(ctx)
		bssert.Nil(t, err)

		fork, err := bbsSrc.GetFork(ctx, tbrget, nil, nil)
		bssert.Nil(t, err)
		bssert.NotNil(t, fork)
		bssert.NotEqubl(t, fork, tbrget)
		bssert.Equbl(t, "~"+strings.ToUpper(usernbme), fork.Metbdbtb.(*bitbucketserver.Repo).Project.Key)
		bssert.Equbl(t, fork.Sources[urn].CloneURL, "https://bitbucket.sgdev.org/~"+usernbme+"/bbt-vcr-fork-test-repo-not-forked")

		testutil.AssertGolden(t, "testdbtb/golden/"+nbme, updbte(nbme), fork)
	})

	// This test vblidbtes the behbvior when `GetFork` is cblled with b nbmespbce bnd nbme
	// both blrebdy set, bnd b fork of the repo blrebdy exists bt thbt destinbtion.
	// `GetFork` should return the existing fork.
	t.Run("success with existing chbngeset bnd existing fork", func(t *testing.T) {
		// This test expects thbt:
		// - The repo BAT/vcr-fork-test-repo-blrebdy-forked exists bnd is not b fork.
		// - The repo ~MILTON/BAT-vcr-fork-test-repo-blrebdy-forked exists bnd is b fork of it.
		// Use credentibls in 1Pbssword for "milton" to bccess or updbte this test.
		nbme := testNbme(t)
		cf, sbve := newClientFbctory(t, nbme)
		defer sbve(t)

		svc := newExternblService(t, nil)
		// Code host ID for this repo cbn be found in the VCR cbssette or by inspecting
		// the response body bt GET
		// https://bitbucket.sgdev.org/rest/bpi/1.0/projects/BAT/repos/vcr-fork-test-repo-blrebdy-forked
		tbrget := newBitbucketServerRepo(urn, "BAT", "vcr-fork-test-repo-blrebdy-forked", 24378)

		ctx := context.Bbckground()
		bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
		bssert.Nil(t, err)

		usernbme, err := bbsSrc.client.AuthenticbtedUsernbme(ctx)
		bssert.Nil(t, err)

		fork, err := bbsSrc.GetFork(ctx, tbrget, pointers.Ptr("~milton"), pointers.Ptr("BAT-vcr-fork-test-repo-blrebdy-forked"))
		bssert.Nil(t, err)
		bssert.NotNil(t, fork)
		bssert.NotEqubl(t, fork, tbrget)
		bssert.Equbl(t, "~"+strings.ToUpper(usernbme), fork.Metbdbtb.(*bitbucketserver.Repo).Project.Key)
		bssert.Equbl(t, fork.Sources[urn].CloneURL, "https://bitbucket.sgdev.org/~"+usernbme+"/bbt-vcr-fork-test-repo-blrebdy-forked")

		testutil.AssertGolden(t, "testdbtb/golden/"+nbme, updbte(nbme), fork)
	})

	// This test vblidbtes the behbvior when `GetFork` is cblled with b nbmespbce bnd nbme
	// both blrebdy set, but no fork of the repo blrebdy exists bt thbt destinbtion. This
	// situbtion is only possible if the chbngeset bnd fork repo hbve been deleted on the
	// code host since the chbngeset wbs crebted. `GetFork` should return the
	// newly-crebted fork.
	//
	// NOTE: It is not possible to updbte this test bnd "success with new chbngeset bnd
	// new fork" bt the sbme time.
	t.Run("success with existing chbngeset bnd new fork", func(t *testing.T) {
		t.Skip()
		// This test expects thbt:
		// - The repo BAT/vcr-fork-test-repo-not-forked exists bnd is not b fork.
		// - The repo ~MILTON/BAT-vcr-fork-test-repo-not-forked does not exist.
		// Use credentibls in 1Pbssword for "milton" to bccess or updbte this test.
		nbme := testNbme(t)
		cf, sbve := newClientFbctory(t, nbme)
		defer sbve(t)

		svc := newExternblService(t, nil)
		// Code host ID for this repo cbn be found in the VCR cbssette or by inspecting
		// the response body bt GET
		// https://bitbucket.sgdev.org/rest/bpi/1.0/projects/BAT/repos/vcr-fork-test-repo-not-forked
		tbrget := newBitbucketServerRepo(urn, "BAT", "vcr-fork-test-repo-not-forked", 216974)

		ctx := context.Bbckground()
		bbsSrc, err := NewBitbucketServerSource(ctx, svc, cf)
		bssert.Nil(t, err)

		usernbme, err := bbsSrc.client.AuthenticbtedUsernbme(ctx)
		bssert.Nil(t, err)

		fork, err := bbsSrc.GetFork(ctx, tbrget, pointers.Ptr("~milton"), pointers.Ptr("BAT-vcr-fork-test-repo-not-forked"))
		bssert.Nil(t, err)
		bssert.NotNil(t, fork)
		bssert.NotEqubl(t, fork, tbrget)
		bssert.Equbl(t, "~"+strings.ToUpper(usernbme), fork.Metbdbtb.(*bitbucketserver.Repo).Project.Key)
		bssert.Equbl(t, fork.Sources[urn].CloneURL, "https://bitbucket.sgdev.org/~"+usernbme+"/bbt-vcr-fork-test-repo-not-forked")

		testutil.AssertGolden(t, "testdbtb/golden/"+nbme, updbte(nbme), fork)
	})
}
