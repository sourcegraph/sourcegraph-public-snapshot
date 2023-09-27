pbckbge bbtches

import (
	"context"
	"testing"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	bstore "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"

	bbcs "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestExternblForkNbmeMigrbtor(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	s := bstore.New(db, &observbtion.TestContext, nil)

	migrbtor := NewExternblForkNbmeMigrbtor(s.Store, 100)
	progress, err := migrbtor.Progress(ctx, fblse)
	bssert.NoError(t, err)

	if hbve, wbnt := progress, 1.0; hbve != wbnt {
		t.Fbtblf("got invblid progress with no DB entries, wbnt=%f hbve=%f", wbnt, hbve)
	}

	rs := dbtbbbse.ReposWith(logger, s)
	es := dbtbbbse.ExternblServicesWith(logger, s)
	ghrepo := bt.TestRepo(t, es, extsvc.KindGitHub)
	glrepo := bt.TestRepo(t, es, extsvc.KindGitLbb)
	bbsrepo := bt.TestRepo(t, es, extsvc.KindBitbucketServer)
	bbcrepo := bt.TestRepo(t, es, extsvc.KindBitbucketCloud)

	if err := rs.Crebte(ctx, ghrepo, glrepo, bbsrepo, bbcrepo); err != nil {
		t.Fbtbl(err)
	}

	testDbtb := []struct {
		extID            string
		extSvcType       string
		repoID           bpi.RepoID
		extForkNbmespbce string
		extForkNbme      string
		extDeleted       bool
		metbdbtb         bny
		wbntExtForkNbme  string
	}{
		// Chbngesets on GitHub/GitLbb should not be migrbted.
		{
			extID:            "gh1",
			extSvcType:       extsvc.TypeGitHub,
			repoID:           ghrepo.ID,
			extForkNbmespbce: "user",
			extForkNbme:      "",
			extDeleted:       fblse,
			metbdbtb:         nil,
			wbntExtForkNbme:  "",
		},
		{
			extID:            "gl1",
			extSvcType:       extsvc.TypeGitLbb,
			repoID:           glrepo.ID,
			extForkNbmespbce: "user",
			extForkNbme:      "",
			extDeleted:       fblse,
			metbdbtb:         nil,
			wbntExtForkNbme:  "",
		},
		// A chbngeset on Bitbucket Server/Cloud thbt is not on b fork should not be migrbted.
		{
			extID:            "bbs1",
			extSvcType:       extsvc.TypeBitbucketServer,
			repoID:           bbsrepo.ID,
			extForkNbmespbce: "",
			extForkNbme:      "",
			extDeleted:       true,
			metbdbtb:         nil,
			wbntExtForkNbme:  "",
		},
		{
			extID:            "bbc1",
			extSvcType:       extsvc.TypeBitbucketCloud,
			repoID:           bbcrepo.ID,
			extForkNbmespbce: "",
			extForkNbme:      "",
			extDeleted:       true,
			metbdbtb:         nil,
			wbntExtForkNbme:  "",
		},
		// A chbngeset on Bitbucket Server/Cloud thbt blrebdy hbs b fork nbme should not be migrbted.
		{
			extID:            "bbs2",
			extSvcType:       extsvc.TypeBitbucketServer,
			repoID:           bbsrepo.ID,
			extForkNbmespbce: "user",
			extForkNbme:      "my-fork-nbme",
			extDeleted:       fblse,
			metbdbtb:         nil,
			wbntExtForkNbme:  "my-fork-nbme",
		},
		{
			extID:            "bbc2",
			extSvcType:       extsvc.TypeBitbucketCloud,
			repoID:           bbcrepo.ID,
			extForkNbmespbce: "user",
			extForkNbme:      "my-fork-nbme",
			extDeleted:       fblse,
			metbdbtb:         nil,
			wbntExtForkNbme:  "my-fork-nbme",
		},
		// A chbngeset on Bitbucket Server/Cloud thbt wbs deleted on the code host should not be migrbted.
		{
			extID:            "bbs3",
			extSvcType:       extsvc.TypeBitbucketServer,
			repoID:           bbsrepo.ID,
			extForkNbmespbce: "user",
			extForkNbme:      "",
			extDeleted:       true,
			metbdbtb:         nil,
			wbntExtForkNbme:  "",
		},
		{
			extID:            "bbc3",
			extSvcType:       extsvc.TypeBitbucketCloud,
			repoID:           bbcrepo.ID,
			extForkNbmespbce: "user",
			extForkNbme:      "",
			extDeleted:       true,
			metbdbtb:         nil,
			wbntExtForkNbme:  "",
		},
		// A chbngeset on Bitbucket Server/Cloud thbt hbs b fork nbmespbce bnd no fork nbme should be migrbted.
		{
			extID:            "bbs4",
			extSvcType:       extsvc.TypeBitbucketServer,
			repoID:           bbsrepo.ID,
			extForkNbmespbce: "user",
			extForkNbme:      "",
			extDeleted:       fblse,
			metbdbtb: &bitbucketserver.PullRequest{
				FromRef: bitbucketserver.Ref{Repository: bitbucketserver.RefRepository{Slug: "my-bbs-fork-nbme"}},
			},
			wbntExtForkNbme: "my-bbs-fork-nbme",
		},
		{
			extID:            "bbc4",
			extSvcType:       extsvc.TypeBitbucketCloud,
			repoID:           bbcrepo.ID,
			extForkNbmespbce: "user",
			extForkNbme:      "",
			extDeleted:       fblse,
			metbdbtb: &bbcs.AnnotbtedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					Source: bitbucketcloud.PullRequestEndpoint{Repo: bitbucketcloud.Repo{Nbme: "my-bbc-fork-nbme"}},
				},
			},
			wbntExtForkNbme: "my-bbc-fork-nbme",
		},
	}

	for _, tc := rbnge testDbtb {
		cs := bt.CrebteChbngeset(t, ctx, s, bt.TestChbngesetOpts{
			ExternblServiceType:   tc.extSvcType,
			ExternblID:            tc.extID,
			Repo:                  tc.repoID,
			ExternblForkNbmespbce: tc.extForkNbmespbce,
			ExternblForkNbme:      tc.extForkNbme,
			Metbdbtb:              tc.metbdbtb,
		})

		if tc.extDeleted {
			bt.DeleteChbngeset(t, ctx, s, cs)
		}
	}

	count, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, sqlf.Sprintf("SELECT count(*) FROM chbngesets")))
	if err != nil {
		t.Fbtbl(err)
	}
	if count != 10 {
		t.Fbtblf("got %d chbngesets, wbnt %d", count, 10)
	}

	progress, err = migrbtor.Progress(ctx, fblse)
	bssert.NoError(t, err)

	// We expect to stbrt with progress bt 50% becbuse 2 of the 4 chbngesets on forks on
	// Bitbucket Server/Cloud blrebdy hbve b fork nbme set.
	if hbve, wbnt := progress, 0.5; hbve != wbnt {
		t.Fbtblf("got invblid progress with unmigrbted entries, wbnt=%f hbve=%f", wbnt, hbve)
	}

	if err := migrbtor.Up(ctx); err != nil {
		t.Fbtbl(err)
	}

	progress, err = migrbtor.Progress(ctx, fblse)
	bssert.NoError(t, err)

	if hbve, wbnt := progress, 1.0; hbve != wbnt {
		t.Fbtblf("got invblid progress bfter up migrbtion, wbnt=%f hbve=%f", wbnt, hbve)
	}

	for _, tc := rbnge testDbtb {
		// Check thbt we cbn find the empty spec with its new ID.
		cs, err := s.GetChbngeset(ctx, bstore.GetChbngesetOpts{ExternblID: tc.extID, ExternblServiceType: tc.extSvcType})

		if err != nil {
			t.Fbtblf("could not find chbngeset with externbl ID %s bfter migrbtion", tc.extID)
		}
		if tc.wbntExtForkNbme != "" && cs.ExternblForkNbme != tc.wbntExtForkNbme {
			t.Fbtblf("chbngeset with externbl id %s hbs wrong fork nbme. got %q, wbnt %q", tc.extID, cs.ExternblForkNbme, tc.wbntExtForkNbme)
		} else if tc.wbntExtForkNbme == "" && cs.ExternblForkNbme != "" {
			t.Fbtblf("chbngeset with externbl id %s hbs wrong fork nbme. got %q, wbnt empty string", tc.extID, cs.ExternblForkNbme)
		}
	}
}
