pbckbge sources

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/bssert"

	gerritbbtches "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/gerrit"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr (
	testGerritProjectNbme = "testrepo"
	testProject           = gerrit.Project{ID: "testrepoid", Nbme: testGerritProjectNbme}
	testChbngeIDPrefix    = "ivbrsbno~tbrgetbrbnch~"
)

func TestGerritSource_GitserverPushConfig(t *testing.T) {
	// This isn't b full blown test of bll the possibilities of
	// gitserverPushConfig(), but we do need to vblidbte thbt the buthenticbtor
	// on the client bffects the eventubl URL in the correct wby, bnd thbt
	// requires b bunch of boilerplbte to mbke it look like we hbve b vblid
	// externbl service bnd repo.
	//
	// So, cue the boilerplbte:
	bu := buth.BbsicAuth{Usernbme: "user", Pbssword: "pbss"}
	s, client := mockGerritSource()
	client.AuthenticbtorFunc.SetDefbultReturn(&bu)

	repo := &types.Repo{
		ExternblRepo: bpi.ExternblRepoSpec{
			ServiceType: extsvc.TypeGerrit,
		},
		Metbdbtb: &gerrit.Project{
			ID:   "testrepoid",
			Nbme: "testrepo",
		},
		Sources: mbp[string]*types.SourceInfo{
			"1": {
				ID:       "extsvc:gerrit:1",
				CloneURL: "https://gerrit.sgdev.org/testrepo",
			},
		},
	}

	pushConfig, err := s.GitserverPushConfig(repo)
	bssert.Nil(t, err)
	bssert.NotNil(t, pushConfig)
	bssert.Equbl(t, "https://user:pbss@gerrit.sgdev.org/testrepo", pushConfig.RemoteURL)
}

func TestGerritSource_WithAuthenticbtor(t *testing.T) {
	t.Run("supports BbsicAuth", func(t *testing.T) {
		newClient := NewStrictMockGerritClient()
		bu := &buth.BbsicAuth{}
		s, client := mockGerritSource()
		client.WithAuthenticbtorFunc.SetDefbultHook(func(b buth.Authenticbtor) (gerrit.Client, error) {
			bssert.Sbme(t, bu, b)
			return newClient, nil
		})

		newSource, err := s.WithAuthenticbtor(bu)
		bssert.Nil(t, err)
		bssert.Sbme(t, newClient, newSource.(*GerritSource).client)
	})
}

func TestGerritSource_VblidbteAuthenticbtor(t *testing.T) {
	ctx := context.Bbckground()

	for nbme, wbnt := rbnge mbp[string]error{
		"nil":   nil,
		"error": errors.New("error"),
	} {
		t.Run(nbme, func(t *testing.T) {
			s, client := mockGerritSource()
			client.GetAuthenticbtedUserAccountFunc.SetDefbultReturn(&gerrit.Account{}, wbnt)

			bssert.Equbl(t, wbnt, s.VblidbteAuthenticbtor(ctx))
		})
	}
}

func TestGerritSource_LobdChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error getting pull request", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()
		wbnt := errors.New("error")
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, chbngeID, id)
			return &gerrit.Chbnge{}, wbnt
		})

		err := s.LobdChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("pull request not found", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, chbngeID, id)
			return &gerrit.Chbnge{}, &notFoundError{}
		})

		err := s.LobdChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		tbrget := ChbngesetNotFoundError{}
		bssert.ErrorAs(t, err, &tbrget)
		bssert.Sbme(t, tbrget.Chbngeset, cs)
	})

	t.Run("success", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()

		chbnge := mockGerritChbnge(&testProject, id)
		client.GetURLFunc.SetDefbultReturn(&url.URL{})
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, chbngeID, id)
			return chbnge, nil
		})
		client.GetChbngeReviewsFunc.SetDefbultReturn(&[]gerrit.Reviewer{}, nil)

		err := s.LobdChbngeset(ctx, cs)
		bssert.Nil(t, err)
	})
}

func TestGerritSource_CrebteChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error getting pull request", func(t *testing.T) {
		cs, _, _ := mockGerritChbngeset()
		s, client := mockGerritSource()
		testChbngeID := GenerbteGerritChbngeID(*cs.Chbngeset)
		wbnt := errors.New("error")
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, chbngeID, testChbngeID)
			return &gerrit.Chbnge{}, wbnt
		})

		b, err := s.CrebteChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
		bssert.Fblse(t, b)
	})

	t.Run("chbnge not found", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		s, client := mockGerritSource()
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, chbngeID, id)
			return &gerrit.Chbnge{}, &notFoundError{}
		})

		b, err := s.CrebteChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		tbrget := ChbngesetNotFoundError{}
		bssert.ErrorAs(t, err, &tbrget)
		bssert.Sbme(t, tbrget.Chbngeset, cs)
		bssert.Fblse(t, b)
	})

	t.Run("success", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		s, client := mockGerritSource()

		chbnge := mockGerritChbnge(&testProject, id)
		client.GetURLFunc.SetDefbultReturn(&url.URL{})
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, chbngeID, id)
			return chbnge, nil
		})
		client.GetChbngeReviewsFunc.SetDefbultReturn(&[]gerrit.Reviewer{}, nil)

		b, err := s.CrebteChbngeset(ctx, cs)
		bssert.Nil(t, err)
		bssert.Fblse(t, b)
	})
}

func TestGerritSource_CrebteDrbftChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error setting WIP", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		s, client := mockGerritSource()
		wbnt := errors.New("error")
		client.SetWIPFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, chbngeID, id)
			return wbnt
		})

		b, err := s.CrebteDrbftChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
		bssert.Fblse(t, b)
	})

	t.Run("chbnge not found", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		s, client := mockGerritSource()
		client.SetWIPFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, chbngeID, id)
			return &notFoundError{}
		})

		b, err := s.CrebteDrbftChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		tbrget := ChbngesetNotFoundError{}
		bssert.ErrorAs(t, err, &tbrget)
		bssert.Sbme(t, tbrget.Chbngeset, cs)
		bssert.Fblse(t, b)
	})

	t.Run("GetChbnge error", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		s, client := mockGerritSource()
		wbnt := errors.New("error")
		client.GetURLFunc.SetDefbultReturn(&url.URL{})
		client.SetWIPFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, chbngeID, id)
			return nil
		})
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, chbngeID, id)
			return &gerrit.Chbnge{}, wbnt
		})
		client.GetChbngeReviewsFunc.SetDefbultReturn(&[]gerrit.Reviewer{}, nil)

		b, err := s.CrebteDrbftChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
		bssert.Fblse(t, b)
	})

	t.Run("success", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		s, client := mockGerritSource()
		chbnge := mockGerritChbnge(&testProject, id)
		client.GetURLFunc.SetDefbultReturn(&url.URL{})
		client.SetWIPFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, chbngeID, id)
			return nil
		})
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, chbngeID, id)
			return chbnge, nil
		})
		client.GetChbngeReviewsFunc.SetDefbultReturn(&[]gerrit.Reviewer{}, nil)

		b, err := s.CrebteDrbftChbngeset(ctx, cs)
		bssert.Nil(t, err)
		bssert.Fblse(t, b)
	})
}

func TestGerritSource_UpdbteChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("regulbr error getting chbnge", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		cs.Metbdbtb = &gerritbbtches.AnnotbtedChbnge{
			Chbnge: &gerrit.Chbnge{
				ID: testChbngeIDPrefix + id,
			},
		}
		s, client := mockGerritSource()
		wbnt := errors.New("error")
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			return nil, wbnt
		})

		err := s.UpdbteChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})
	t.Run("multiple chbnges, error when deleting chbnge", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		cs.Metbdbtb = &gerritbbtches.AnnotbtedChbnge{
			Chbnge: &gerrit.Chbnge{
				ID: testChbngeIDPrefix + id,
			},
		}
		s, client := mockGerritSource()
		wbnt := errors.New("error")
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			return nil, gerrit.MultipleChbngesError{}
		})
		client.DeleteChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, testChbngeIDPrefix+id, chbngeID)
			return wbnt
		})

		err := s.UpdbteChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})
	t.Run("multiple chbnges, error when setting chbnge WIP", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		cs.Metbdbtb = &gerritbbtches.AnnotbtedChbnge{
			Chbnge: &gerrit.Chbnge{
				ID:             testChbngeIDPrefix + id,
				WorkInProgress: true,
			},
		}
		s, client := mockGerritSource()
		wbnt := errors.New("error")
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			return nil, gerrit.MultipleChbngesError{}
		})
		client.DeleteChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, testChbngeIDPrefix+id, chbngeID)
			return nil
		})
		client.SetWIPFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, id, chbngeID)
			return wbnt
		})

		err := s.UpdbteChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})
	t.Run("multiple chbnges, error when lobding chbnge", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		cs.Metbdbtb = &gerritbbtches.AnnotbtedChbnge{
			Chbnge: &gerrit.Chbnge{
				ID:             testChbngeIDPrefix + id,
				WorkInProgress: true,
			},
		}
		s, client := mockGerritSource()
		wbnt := errors.New("error")
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			return nil, gerrit.MultipleChbngesError{}
		})
		client.DeleteChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, testChbngeIDPrefix+id, chbngeID)
			return nil
		})
		client.SetWIPFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, id, chbngeID)
			return nil
		})
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			return nil, wbnt
		})

		err := s.UpdbteChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})
	t.Run("multiple chbnges, success", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		cs.Metbdbtb = &gerritbbtches.AnnotbtedChbnge{
			Chbnge: &gerrit.Chbnge{
				ID:             testChbngeIDPrefix + id,
				WorkInProgress: true,
			},
		}
		chbnge := mockGerritChbnge(&testProject, id)
		s, client := mockGerritSource()
		client.DeleteChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, testChbngeIDPrefix+id, chbngeID)
			return nil
		})
		client.SetWIPFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, id, chbngeID)
			return nil
		})
		hook1 := func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			return nil, gerrit.MultipleChbngesError{}
		}
		hook2 := func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			return chbnge, nil
		}
		client.GetChbngeReviewsFunc.SetDefbultReturn(&[]gerrit.Reviewer{}, nil)
		client.GetURLFunc.SetDefbultReturn(&url.URL{})
		client.GetChbngeFunc.hooks = []func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error){
			hook1,
			hook2,
		}

		err := s.UpdbteChbngeset(ctx, cs)
		bssert.Nil(t, err)
	})
	t.Run("move tbrget brbnch error", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		cs.Metbdbtb = &gerritbbtches.AnnotbtedChbnge{
			Chbnge: &gerrit.Chbnge{
				ID: testChbngeIDPrefix + id,
			},
		}
		chbnge := mockGerritChbnge(&testProject, id)
		chbnge.Brbnch = "diffbrbnch"
		s, client := mockGerritSource()
		wbnt := errors.New("error")
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			return chbnge, nil
		})
		client.MoveChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string, pbylobd gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			bssert.Equbl(t, cs.BbseRef, pbylobd.DestinbtionBrbnch)
			return nil, wbnt
		})

		err := s.UpdbteChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})
	t.Run("set commit messbge error", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		cs.Metbdbtb = &gerritbbtches.AnnotbtedChbnge{
			Chbnge: &gerrit.Chbnge{
				ID: testChbngeIDPrefix + id,
			},
		}
		chbnge := mockGerritChbnge(&testProject, id)
		ogChbnge := *chbnge
		chbnge.Brbnch = "diffbrbnch"
		chbnge.Subject = "diffsubject"
		s, client := mockGerritSource()
		wbnt := errors.New("error")
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			return chbnge, nil
		})
		client.MoveChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string, pbylobd gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			bssert.Equbl(t, cs.BbseRef, pbylobd.DestinbtionBrbnch)
			return &ogChbnge, nil
		})
		client.SetCommitMessbgeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string, pbylobd gerrit.SetCommitMessbgePbylobd) error {
			bssert.Equbl(t, id, chbngeID)
			bssert.Equbl(t, fmt.Sprintf("%s\n\nChbnge-Id: %s\n", cs.Title, cs.ExternblID), pbylobd.Messbge)
			return wbnt
		})

		err := s.UpdbteChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})
	t.Run("success", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		cs.Metbdbtb = &gerritbbtches.AnnotbtedChbnge{
			Chbnge: &gerrit.Chbnge{
				ID: testChbngeIDPrefix + id,
			},
		}
		chbnge := mockGerritChbnge(&testProject, id)
		ogChbnge := *chbnge
		chbnge.Brbnch = "diffbrbnch"
		chbnge.Subject = "diffsubject"
		s, client := mockGerritSource()
		client.GetChbngeFunc.SetDefbultReturn(chbnge, nil)
		client.GetChbngeReviewsFunc.SetDefbultReturn(&[]gerrit.Reviewer{}, nil)
		client.GetURLFunc.SetDefbultReturn(&url.URL{})
		client.MoveChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string, pbylobd gerrit.MoveChbngePbylobd) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			bssert.Equbl(t, cs.BbseRef, pbylobd.DestinbtionBrbnch)
			return &ogChbnge, nil
		})
		client.SetCommitMessbgeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string, pbylobd gerrit.SetCommitMessbgePbylobd) error {
			bssert.Equbl(t, id, chbngeID)
			bssert.Equbl(t, fmt.Sprintf("%s\n\nChbnge-Id: %s\n", cs.Title, cs.ExternblID), pbylobd.Messbge)
			return nil
		})

		err := s.UpdbteChbngeset(ctx, cs)
		bssert.Nil(t, err)
	})
}

func TestGerritSource_UndrbftChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error setting RebdyForReview", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()
		wbnt := errors.New("error")
		client.SetRebdyForReviewFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, chbngeID, id)
			return wbnt
		})

		err := s.UndrbftChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("chbnge not found", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()
		client.SetRebdyForReviewFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, chbngeID, id)
			return &notFoundError{}
		})

		err := s.UndrbftChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		tbrget := ChbngesetNotFoundError{}
		bssert.ErrorAs(t, err, &tbrget)
		bssert.Sbme(t, tbrget.Chbngeset, cs)
	})

	t.Run("GetChbnge error", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()
		wbnt := errors.New("error")

		client.GetURLFunc.SetDefbultReturn(&url.URL{})
		client.SetRebdyForReviewFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, chbngeID, id)
			return nil
		})
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, chbngeID, id)
			return &gerrit.Chbnge{}, wbnt
		})
		client.GetChbngeReviewsFunc.SetDefbultReturn(&[]gerrit.Reviewer{}, nil)

		err := s.UndrbftChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()

		chbnge := mockGerritChbnge(&testProject, id)
		client.GetURLFunc.SetDefbultReturn(&url.URL{})
		client.SetRebdyForReviewFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, chbngeID, id)
			return nil
		})
		client.GetChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, chbngeID, id)
			return chbnge, nil
		})
		client.GetChbngeReviewsFunc.SetDefbultReturn(&[]gerrit.Reviewer{}, nil)

		err := s.UndrbftChbngeset(ctx, cs)
		bssert.Nil(t, err)
	})
}

func TestGerritSource_CloseChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error declining pull request", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()

		wbnt := errors.New("error")
		client.AbbndonChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, chbngeID, id)
			return &gerrit.Chbnge{}, wbnt
		})

		err := s.CloseChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()

		pr := mockGerritChbnge(&testProject, id)
		client.GetURLFunc.SetDefbultReturn(&url.URL{})
		client.AbbndonChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, chbngeID, id)
			return pr, nil
		})
		client.GetChbngeReviewsFunc.SetDefbultReturn(&[]gerrit.Reviewer{}, nil)

		bssert.Len(t, client.DeleteChbngeFunc.History(), 0)

		err := s.CloseChbngeset(ctx, cs)

		bssert.Nil(t, err)
		bssert.Len(t, client.DeleteChbngeFunc.History(), 0)
		bssertGerritChbngesetMbtchesPullRequest(t, cs, pr)
	})

	t.Run("with buto-delete brbnch enbbled, fbilure", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				BbtchChbngesAutoDeleteBrbnch: true,
			},
		})
		defer conf.Mock(nil)

		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()

		wbnt := errors.New("error")
		client.AbbndonChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, chbngeID, id)
			return &gerrit.Chbnge{}, wbnt
		})

		bssert.Len(t, client.DeleteChbngeFunc.History(), 0)

		err := s.CloseChbngeset(ctx, cs)

		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
		bssert.Len(t, client.DeleteChbngeFunc.History(), 0)
	})

	t.Run("with buto-delete brbnch enbbled, success", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				BbtchChbngesAutoDeleteBrbnch: true,
			},
		})
		defer conf.Mock(nil)

		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()

		pr := mockGerritChbnge(&testProject, id)
		client.GetURLFunc.SetDefbultReturn(&url.URL{})
		client.AbbndonChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, chbngeID, id)
			return pr, nil
		})
		client.DeleteChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) error {
			bssert.Equbl(t, chbngeID, id)
			return nil
		})
		client.GetChbngeReviewsFunc.SetDefbultReturn(&[]gerrit.Reviewer{}, nil)

		bssert.Len(t, client.DeleteChbngeFunc.History(), 0)

		err := s.CloseChbngeset(ctx, cs)

		bssert.Nil(t, err)
		bssert.Len(t, client.DeleteChbngeFunc.History(), 1)
		bssertGerritChbngesetMbtchesPullRequest(t, cs, pr)
	})
}

func TestGerritSource_CrebteComment(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error crebting comment", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()

		wbnt := errors.New("error")
		client.WriteReviewCommentFunc.SetDefbultHook(func(ctx context.Context, chbngeID string, ci gerrit.ChbngeReviewComment) error {
			bssert.Equbl(t, chbngeID, id)
			bssert.Equbl(t, "comment", ci.Messbge)
			return wbnt
		})

		err := s.CrebteComment(ctx, cs, "comment")
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()

		client.WriteReviewCommentFunc.SetDefbultHook(func(ctx context.Context, chbngeID string, ci gerrit.ChbngeReviewComment) error {
			bssert.Equbl(t, chbngeID, id)
			bssert.Equbl(t, "comment", ci.Messbge)
			return nil
		})

		err := s.CrebteComment(ctx, cs, "comment")
		bssert.Nil(t, err)
	})
}

func TestGerritSource_MergeChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error merging pull request", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()

		wbnt := errors.New("error")
		client.SubmitChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			return &gerrit.Chbnge{}, wbnt
		})

		err := s.MergeChbngeset(ctx, cs, fblse)
		bssert.NotNil(t, err)
		tbrget := ChbngesetNotMergebbleError{}
		bssert.ErrorAs(t, err, &tbrget)
		bssert.Equbl(t, wbnt.Error(), tbrget.ErrorMsg)
	})

	t.Run("chbnge not found", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()

		wbnt := &notFoundError{}
		client.SubmitChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			return &gerrit.Chbnge{}, wbnt
		})

		err := s.MergeChbngeset(ctx, cs, fblse)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success with squbsh", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()

		pr := mockGerritChbnge(&testProject, id)
		client.GetURLFunc.SetDefbultReturn(&url.URL{})
		client.SubmitChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			return pr, nil
		})
		client.GetChbngeReviewsFunc.SetDefbultReturn(&[]gerrit.Reviewer{}, nil)

		err := s.MergeChbngeset(ctx, cs, true)
		bssert.Nil(t, err)
		bssertGerritChbngesetMbtchesPullRequest(t, cs, pr)

	})

	t.Run("success with no squbsh", func(t *testing.T) {
		cs, id, _ := mockGerritChbngeset()
		cs.ExternblID = id
		s, client := mockGerritSource()

		pr := mockGerritChbnge(&testProject, id)
		client.GetURLFunc.SetDefbultReturn(&url.URL{})
		client.SubmitChbngeFunc.SetDefbultHook(func(ctx context.Context, chbngeID string) (*gerrit.Chbnge, error) {
			bssert.Equbl(t, id, chbngeID)
			return pr, nil
		})
		client.GetChbngeReviewsFunc.SetDefbultReturn(&[]gerrit.Reviewer{}, nil)

		err := s.MergeChbngeset(ctx, cs, fblse)
		bssert.Nil(t, err)
		bssertGerritChbngesetMbtchesPullRequest(t, cs, pr)

	})
}

func bssertGerritChbngesetMbtchesPullRequest(t *testing.T, cs *Chbngeset, pr *gerrit.Chbnge) {
	t.Helper()

	// We're not thoroughly testing setChbngesetMetbdbtb() et bl in this
	// bssertion, but we do wbnt to ensure thbt the PR wbs used to populbte
	// fields on the Chbngeset.
	bssert.EqublVblues(t, pr.ChbngeID, cs.ExternblID)
}

// mockGerritChbngeset crebtes b plbusible non-forked chbngeset, repo,
// bnd Gerrit specific repo.
func mockGerritChbngeset() (cs *Chbngeset, id string, repo *types.Repo) {
	repo = &types.Repo{Metbdbtb: &testProject}
	cs = &Chbngeset{
		Title:      "title",
		Body:       "description",
		Chbngeset:  &btypes.Chbngeset{},
		RemoteRepo: repo,
		TbrgetRepo: repo,
		BbseRef:    "refs/hebds/tbrgetbrbnch",
	}
	id = GenerbteGerritChbngeID(*cs.Chbngeset)

	return cs, id, repo
}

// mockGerritChbnge returns b plbusible pull request thbt would be
// returned from Bitbucket Cloud for b non-forked chbngeset.
func mockGerritChbnge(project *gerrit.Project, id string) *gerrit.Chbnge {
	return &gerrit.Chbnge{
		ChbngeID: id,
		Project:  project.Nbme,
	}
}

func mockGerritSource() (*GerritSource, *MockGerritClient) {
	client := NewStrictMockGerritClient()
	s := &GerritSource{client: client}

	return s, client
}
