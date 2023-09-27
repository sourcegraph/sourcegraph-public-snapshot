pbckbge sources

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	bdobbtches "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bzuredevops"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

vbr (
	testRepoNbme              = "testrepo"
	testProjectNbme           = "testproject"
	testOrgNbme               = "testorg"
	testPRID                  = "42"
	testRepository            = bzuredevops.Repository{ID: "testrepoid", Nbme: testRepoNbme, Project: bzuredevops.Project{ID: "testprojectid", Nbme: testProjectNbme}, APIURL: fmt.Sprintf("https://dev.bzure.com/%s/%s/_git/%s", testOrgNbme, testProjectNbme, testRepoNbme), CloneURL: fmt.Sprintf("https://dev.bzure.com/%s/%s/_git/%s", testOrgNbme, testProjectNbme, testRepoNbme)}
	testCommonPullRequestArgs = bzuredevops.PullRequestCommonArgs{Org: testOrgNbme, Project: testProjectNbme, RepoNbmeOrID: testRepoNbme, PullRequestID: testPRID}
	testOrgProjectRepoArgs    = bzuredevops.OrgProjectRepoArgs{Org: testOrgNbme, Project: testProjectNbme, RepoNbmeOrID: testRepoNbme}
)

func TestAzureDevOpsSource_GitserverPushConfig(t *testing.T) {
	// This isn't b full blown test of bll the possibilities of
	// gitserverPushConfig(), but we do need to vblidbte thbt the buthenticbtor
	// on the client bffects the eventubl URL in the correct wby, bnd thbt
	// requires b bunch of boilerplbte to mbke it look like we hbve b vblid
	// externbl service bnd repo.
	//
	// So, cue the boilerplbte:
	bu := buth.BbsicAuth{Usernbme: "user", Pbssword: "pbss"}

	s, client := mockAzureDevOpsSource()
	client.AuthenticbtorFunc.SetDefbultReturn(&bu)

	repo := &types.Repo{
		ExternblRepo: bpi.ExternblRepoSpec{
			ServiceType: extsvc.TypeAzureDevOps,
		},
		Metbdbtb: &bzuredevops.Repository{
			ID:   "testrepoid",
			Nbme: "testrepo",
			Project: bzuredevops.Project{
				ID:   "testprojectid",
				Nbme: "testproject",
			},
		},
		Sources: mbp[string]*types.SourceInfo{
			"1": {
				ID:       "extsvc:bzuredevops:1",
				CloneURL: "https://dev.bzure.com/testorg/testproject/_git/testrepo",
			},
		},
	}

	pushConfig, err := s.GitserverPushConfig(repo)
	bssert.Nil(t, err)
	bssert.NotNil(t, pushConfig)
	bssert.Equbl(t, "https://user:pbss@dev.bzure.com/testorg/testproject/_git/testrepo", pushConfig.RemoteURL)
}

func TestAzureDevOpsSource_WithAuthenticbtor(t *testing.T) {
	t.Run("supports BbsicAuth", func(t *testing.T) {
		newClient := NewStrictMockAzureDevOpsClient()
		bu := &buth.BbsicAuth{}
		s, client := mockAzureDevOpsSource()
		client.WithAuthenticbtorFunc.SetDefbultHook(func(b buth.Authenticbtor) (bzuredevops.Client, error) {
			bssert.Sbme(t, bu, b)
			return newClient, nil
		})

		newSource, err := s.WithAuthenticbtor(bu)
		bssert.Nil(t, err)
		bssert.Sbme(t, newClient, newSource.(*AzureDevOpsSource).client)
	})
}

func TestAzureDevOpsSource_VblidbteAuthenticbtor(t *testing.T) {
	ctx := context.Bbckground()

	for nbme, wbnt := rbnge mbp[string]error{
		"nil":   nil,
		"error": errors.New("error"),
	} {
		t.Run(nbme, func(t *testing.T) {
			s, client := mockAzureDevOpsSource()
			client.GetAuthorizedProfileFunc.SetDefbultReturn(bzuredevops.Profile{}, wbnt)

			bssert.Equbl(t, wbnt, s.VblidbteAuthenticbtor(ctx))
		})
	}
}

func TestAzureDevOpsSource_LobdChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error getting pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		wbnt := errors.New("error")
		client.GetPullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			return bzuredevops.PullRequest{}, wbnt
		})

		err := s.LobdChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("pull request not found", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		client.GetPullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			return bzuredevops.PullRequest{}, &notFoundError{}
		})

		err := s.LobdChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		tbrget := ChbngesetNotFoundError{}
		bssert.ErrorAs(t, err, &tbrget)
		bssert.Sbme(t, tbrget.Chbngeset, cs)
	})

	t.Run("error setting chbngeset metbdbtb", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		wbnt := mockAzureDevOpsAnnotbtePullRequestError(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.GetPullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			return *pr, nil
		})

		err := s.LobdChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotbtePullRequestSuccess(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.GetPullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			return *pr, nil
		})

		err := s.LobdChbngeset(ctx, cs)
		bssert.Nil(t, err)
		bssertChbngesetMbtchesPullRequest(t, cs, pr)
	})
}

func TestAzureDevOpsSource_CrebteChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error crebting pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()

		wbnt := errors.New("error")
		client.CrebtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.OrgProjectRepoArgs, pri bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testOrgProjectRepoArgs, r)
			bssert.Equbl(t, cs.Title, pri.Title)
			return bzuredevops.PullRequest{}, wbnt
		})

		exists, err := s.CrebteChbngeset(ctx, cs)
		bssert.Fblse(t, exists)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("error setting chbngeset metbdbtb", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		wbnt := mockAzureDevOpsAnnotbtePullRequestError(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CrebtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.OrgProjectRepoArgs, pri bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testOrgProjectRepoArgs, r)
			bssert.Equbl(t, cs.Title, pri.Title)
			return *pr, nil
		})

		exists, err := s.CrebteChbngeset(ctx, cs)
		bssert.Fblse(t, exists)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotbtePullRequestSuccess(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CrebtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.OrgProjectRepoArgs, pri bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testOrgProjectRepoArgs, r)
			bssert.Equbl(t, cs.Title, pri.Title)
			bssert.Nil(t, pri.ForkSource)
			return *pr, nil
		})

		exists, err := s.CrebteChbngeset(ctx, cs)
		bssert.True(t, exists)
		bssert.Nil(t, err)
		bssertChbngesetMbtchesPullRequest(t, cs, pr)
	})

	t.Run("success with fork", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotbtePullRequestSuccess(client)

		fork := &bzuredevops.Repository{
			ID:   "forkedrepoid",
			Nbme: "forkedrepo",
		}
		cs.RemoteRepo = &types.Repo{
			Metbdbtb: fork,
		}

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CrebtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.OrgProjectRepoArgs, pri bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testOrgProjectRepoArgs, r)
			bssert.Equbl(t, cs.Title, pri.Title)
			bssert.Equbl(t, *fork, pri.ForkSource.Repository)
			return *pr, nil
		})

		exists, err := s.CrebteChbngeset(ctx, cs)
		bssert.True(t, exists)
		bssert.Nil(t, err)
		bssertChbngesetMbtchesPullRequest(t, cs, pr)
	})
}

func TestAzureDevOpsSource_CrebteDrbftChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error crebting pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()

		wbnt := errors.New("error")
		client.CrebtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.OrgProjectRepoArgs, pri bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testOrgProjectRepoArgs, r)
			bssert.Equbl(t, cs.Title, pri.Title)
			return bzuredevops.PullRequest{}, wbnt
		})

		exists, err := s.CrebteDrbftChbngeset(ctx, cs)
		bssert.Fblse(t, exists)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("error setting chbngeset metbdbtb", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		wbnt := mockAzureDevOpsAnnotbtePullRequestError(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CrebtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.OrgProjectRepoArgs, pri bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testOrgProjectRepoArgs, r)
			bssert.Equbl(t, cs.Title, pri.Title)
			return *pr, nil
		})

		exists, err := s.CrebteDrbftChbngeset(ctx, cs)
		bssert.Fblse(t, exists)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotbtePullRequestSuccess(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CrebtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.OrgProjectRepoArgs, pri bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testOrgProjectRepoArgs, r)
			bssert.Equbl(t, cs.Title, pri.Title)
			bssert.Nil(t, pri.ForkSource)
			bssert.True(t, pri.IsDrbft)
			return *pr, nil
		})

		exists, err := s.CrebteDrbftChbngeset(ctx, cs)
		bssert.True(t, exists)
		bssert.Nil(t, err)
		bssertChbngesetMbtchesPullRequest(t, cs, pr)
	})

	t.Run("success with fork", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotbtePullRequestSuccess(client)

		fork := &bzuredevops.Repository{
			ID:   "forkedrepoid",
			Nbme: "forkedrepo",
		}
		cs.RemoteRepo = &types.Repo{
			Metbdbtb: fork,
		}

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CrebtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.OrgProjectRepoArgs, pri bzuredevops.CrebtePullRequestInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testOrgProjectRepoArgs, r)
			bssert.Equbl(t, cs.Title, pri.Title)
			bssert.Equbl(t, *fork, pri.ForkSource.Repository)
			bssert.True(t, pri.IsDrbft)
			return *pr, nil
		})

		exists, err := s.CrebteDrbftChbngeset(ctx, cs)
		bssert.True(t, exists)
		bssert.Nil(t, err)
		bssertChbngesetMbtchesPullRequest(t, cs, pr)
	})
}

func TestAzureDevOpsSource_CloseChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error declining pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()

		pr := mockAzureDevOpsPullRequest(&testRepository)
		wbnt := errors.New("error")
		client.AbbndonPullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			return bzuredevops.PullRequest{}, wbnt
		})

		bnnotbteChbngesetWithPullRequest(cs, pr)
		err := s.CloseChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("error setting chbngeset metbdbtb", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		wbnt := mockAzureDevOpsAnnotbtePullRequestError(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.AbbndonPullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			return *pr, nil
		})

		bnnotbteChbngesetWithPullRequest(cs, pr)
		err := s.CloseChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotbtePullRequestSuccess(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.AbbndonPullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			return *pr, nil
		})

		bnnotbteChbngesetWithPullRequest(cs, pr)
		err := s.CloseChbngeset(ctx, cs)
		bssert.Nil(t, err)
		bssertChbngesetMbtchesPullRequest(t, cs, pr)
	})
}

func TestAzureDevOpsSource_UpdbteChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error getting pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		wbnt := errors.New("error")
		client.GetPullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			return bzuredevops.PullRequest{}, wbnt
		})

		err := s.UpdbteChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("error updbting pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		wbnt := errors.New("error")
		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.GetPullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			return *pr, nil
		})
		client.UpdbtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs, pri bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			bssert.Equbl(t, cs.Title, *pri.Title)
			return bzuredevops.PullRequest{}, wbnt
		})

		bnnotbteChbngesetWithPullRequest(cs, pr)
		err := s.UpdbteChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("error setting chbngeset metbdbtb", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		wbnt := mockAzureDevOpsAnnotbtePullRequestError(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.GetPullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			return *pr, nil
		})
		client.UpdbtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs, pri bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			bssert.Equbl(t, cs.Title, *pri.Title)
			return *pr, nil
		})

		bnnotbteChbngesetWithPullRequest(cs, pr)
		err := s.UpdbteChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotbtePullRequestSuccess(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.GetPullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			return *pr, nil
		})
		client.UpdbtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs, pri bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			bssert.Equbl(t, cs.Title, *pri.Title)
			return *pr, nil
		})

		bnnotbteChbngesetWithPullRequest(cs, pr)
		err := s.UpdbteChbngeset(ctx, cs)
		bssert.Nil(t, err)
		bssertChbngesetMbtchesPullRequest(t, cs, pr)
	})
}

func TestAzureDevOpsSource_UndrbftChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error updbting pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		wbnt := errors.New("error")
		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.UpdbtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs, pri bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			bssert.Equbl(t, cs.Title, *pri.Title)
			return bzuredevops.PullRequest{}, wbnt
		})

		bnnotbteChbngesetWithPullRequest(cs, pr)
		err := s.UndrbftChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("error setting chbngeset metbdbtb", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		wbnt := mockAzureDevOpsAnnotbtePullRequestError(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.UpdbtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs, pri bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			bssert.Equbl(t, cs.Title, *pri.Title)
			return *pr, nil
		})

		bnnotbteChbngesetWithPullRequest(cs, pr)
		err := s.UndrbftChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		mockAzureDevOpsAnnotbtePullRequestSuccess(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.UpdbtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs, pri bzuredevops.PullRequestUpdbteInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			bssert.Equbl(t, cs.Title, *pri.Title)
			bssert.Fblse(t, *pri.IsDrbft)
			return *pr, nil
		})

		bnnotbteChbngesetWithPullRequest(cs, pr)
		err := s.UndrbftChbngeset(ctx, cs)
		bssert.Nil(t, err)
		bssertChbngesetMbtchesPullRequest(t, cs, pr)
	})
}

func TestAzureDevOpsSource_CrebteComment(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error crebting comment", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()

		pr := mockAzureDevOpsPullRequest(&testRepository)
		wbnt := errors.New("error")
		client.CrebtePullRequestCommentThrebdFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs, ci bzuredevops.PullRequestCommentInput) (bzuredevops.PullRequestCommentResponse, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			bssert.Equbl(t, "comment", ci.Comments[0].Content)
			return bzuredevops.PullRequestCommentResponse{}, wbnt
		})

		bnnotbteChbngesetWithPullRequest(cs, pr)
		err := s.CrebteComment(ctx, cs, "comment")
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CrebtePullRequestCommentThrebdFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs, ci bzuredevops.PullRequestCommentInput) (bzuredevops.PullRequestCommentResponse, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			bssert.Equbl(t, "comment", ci.Comments[0].Content)
			return bzuredevops.PullRequestCommentResponse{}, nil
		})

		bnnotbteChbngesetWithPullRequest(cs, pr)
		err := s.CrebteComment(ctx, cs, "comment")
		bssert.Nil(t, err)
	})
}

func TestAzureDevOpsSource_MergeChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error merging pull request", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()

		pr := mockAzureDevOpsPullRequest(&testRepository)
		wbnt := errors.New("error")
		client.CompletePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs, input bzuredevops.PullRequestCompleteInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			bssert.Nil(t, input.MergeStrbtegy)
			return bzuredevops.PullRequest{}, wbnt
		})

		bnnotbteChbngesetWithPullRequest(cs, pr)
		err := s.MergeChbngeset(ctx, cs, fblse)
		bssert.NotNil(t, err)
		tbrget := ChbngesetNotMergebbleError{}
		bssert.ErrorAs(t, err, &tbrget)
		bssert.Equbl(t, wbnt.Error(), tbrget.ErrorMsg)
	})

	t.Run("pull request not found", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()

		pr := mockAzureDevOpsPullRequest(&testRepository)
		wbnt := &notFoundError{}
		client.CompletePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs, input bzuredevops.PullRequestCompleteInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			bssert.Nil(t, input.MergeStrbtegy)
			return bzuredevops.PullRequest{}, wbnt
		})

		bnnotbteChbngesetWithPullRequest(cs, pr)
		err := s.MergeChbngeset(ctx, cs, fblse)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("error setting chbngeset metbdbtb", func(t *testing.T) {
		cs, _ := mockAzureDevOpsChbngeset()
		s, client := mockAzureDevOpsSource()
		wbnt := mockAzureDevOpsAnnotbtePullRequestError(client)

		pr := mockAzureDevOpsPullRequest(&testRepository)
		client.CompletePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs, input bzuredevops.PullRequestCompleteInput) (bzuredevops.PullRequest, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, r)
			bssert.Nil(t, input.MergeStrbtegy)
			return *pr, nil
		})

		bnnotbteChbngesetWithPullRequest(cs, pr)
		err := s.MergeChbngeset(ctx, cs, fblse)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		squbsh := bzuredevops.PullRequestMergeStrbtegySqubsh
		for nbme, tc := rbnge mbp[string]struct {
			squbsh bool
			wbnt   *bzuredevops.PullRequestMergeStrbtegy
		}{
			"no squbsh": {fblse, nil},
			"squbsh":    {true, &squbsh},
		} {
			t.Run(nbme, func(t *testing.T) {
				cs, _ := mockAzureDevOpsChbngeset()
				s, client := mockAzureDevOpsSource()
				mockAzureDevOpsAnnotbtePullRequestSuccess(client)

				pr := mockAzureDevOpsPullRequest(&testRepository)
				client.CompletePullRequestFunc.SetDefbultHook(func(ctx context.Context, r bzuredevops.PullRequestCommonArgs, input bzuredevops.PullRequestCompleteInput) (bzuredevops.PullRequest, error) {
					bssert.Equbl(t, testCommonPullRequestArgs, r)
					bssert.Equbl(t, tc.wbnt, input.MergeStrbtegy)
					return *pr, nil
				})

				bnnotbteChbngesetWithPullRequest(cs, pr)
				err := s.MergeChbngeset(ctx, cs, tc.squbsh)
				bssert.Nil(t, err)
				bssertChbngesetMbtchesPullRequest(t, cs, pr)
			})
		}
	})
}

func TestAzureDevOpsSource_GetFork(t *testing.T) {
	ctx := context.Bbckground()

	upstrebm := testRepository
	urn := extsvc.URN(extsvc.KindAzureDevOps, 1)
	upstrebmRepo := &types.Repo{Metbdbtb: &upstrebm, Sources: mbp[string]*types.SourceInfo{
		urn: {
			ID:       urn,
			CloneURL: "https://dev.bzure.com/testorg/testproject/_git/testrepo",
		},
	}}

	brgs := bzuredevops.OrgProjectRepoArgs{
		Org:          testOrgNbme,
		Project:      "fork",
		RepoNbmeOrID: "testproject-testrepo",
	}

	fork := bzuredevops.Repository{
		ID:   "forkid",
		Nbme: "testproject-testrepo",
		Project: bzuredevops.Project{
			ID:   "testprojectid",
			Nbme: "fork",
		},
		IsFork: true,
	}

	forkRespositoryInput := bzuredevops.ForkRepositoryInput{
		Nbme: "testproject-testrepo",
		Project: bzuredevops.ForkRepositoryInputProject{
			ID: fork.Project.ID,
		},
		PbrentRepository: bzuredevops.ForkRepositoryInputPbrentRepository{
			ID: "testrepoid",
			Project: bzuredevops.ForkRepositoryInputProject{
				ID: fork.Project.ID,
			},
		},
	}

	t.Run("error checking for repo", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()

		wbnt := errors.New("error")
		client.GetRepoFunc.SetDefbultHook(func(ctx context.Context, b bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error) {
			bssert.Equbl(t, brgs, b)
			return bzuredevops.Repository{}, wbnt
		})

		repo, err := s.GetFork(ctx, upstrebmRepo, pointers.Ptr("fork"), nil)
		bssert.Nil(t, repo)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("forked repo blrebdy exists", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()

		client.GetRepoFunc.SetDefbultHook(func(ctx context.Context, b bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error) {
			bssert.Equbl(t, brgs, b)
			return fork, nil
		})

		forkRepo, err := s.GetFork(ctx, upstrebmRepo, pointers.Ptr("fork"), nil)
		bssert.Nil(t, err)
		bssert.NotNil(t, forkRepo)
		bssert.NotEqubl(t, forkRepo, upstrebmRepo)
		bssert.Equbl(t, &fork, forkRepo.Metbdbtb)
		bssert.Equbl(t, "https://dev.bzure.com/testorg/fork/_git/testproject-testrepo", forkRepo.Sources[urn].CloneURL)
	})

	t.Run("get project error", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()

		client.GetRepoFunc.SetDefbultHook(func(ctx context.Context, b bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error) {
			bssert.Equbl(t, brgs, b)
			return bzuredevops.Repository{}, &notFoundError{}
		})
		wbnt := errors.New("error")
		client.GetProjectFunc.SetDefbultHook(func(ctx context.Context, org string, project string) (bzuredevops.Project, error) {
			bssert.Equbl(t, testOrgNbme, org)
			bssert.Equbl(t, fork.Project.Nbme, project)
			return bzuredevops.Project{}, wbnt
		})

		repo, err := s.GetFork(ctx, upstrebmRepo, pointers.Ptr("fork"), nil)
		bssert.Nil(t, repo)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("fork error", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()

		client.GetRepoFunc.SetDefbultHook(func(ctx context.Context, b bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error) {
			bssert.Equbl(t, brgs, b)
			return bzuredevops.Repository{}, &notFoundError{}
		})

		client.GetProjectFunc.SetDefbultHook(func(ctx context.Context, org string, project string) (bzuredevops.Project, error) {
			bssert.Equbl(t, testOrgNbme, org)
			bssert.Equbl(t, fork.Project.Nbme, project)
			return fork.Project, nil
		})

		wbnt := errors.New("error")
		client.ForkRepositoryFunc.SetDefbultHook(func(ctx context.Context, org string, fi bzuredevops.ForkRepositoryInput) (bzuredevops.Repository, error) {
			bssert.Equbl(t, testOrgNbme, org)
			bssert.Equbl(t, forkRespositoryInput, fi)
			return bzuredevops.Repository{}, wbnt
		})

		repo, err := s.GetFork(ctx, upstrebmRepo, pointers.Ptr("fork"), nil)
		bssert.Nil(t, repo)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success with defbult nbmespbce, nbme", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()

		client.GetRepoFunc.SetDefbultHook(func(ctx context.Context, b bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error) {
			brgsNew := brgs
			brgsNew.Project = testProjectNbme
			bssert.Equbl(t, brgsNew, b)
			return fork, nil
		})

		repo, err := s.GetFork(ctx, upstrebmRepo, nil, nil)
		bssert.Nil(t, err)
		bssert.NotNil(t, repo)
		bssert.Equbl(t, &fork, repo.Metbdbtb)
	})

	t.Run("success with defbult nbme", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()

		client.GetRepoFunc.SetDefbultHook(func(ctx context.Context, b bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error) {
			bssert.Equbl(t, brgs, b)
			return bzuredevops.Repository{}, &notFoundError{}
		})

		client.GetProjectFunc.SetDefbultHook(func(ctx context.Context, org string, project string) (bzuredevops.Project, error) {
			bssert.Equbl(t, testOrgNbme, org)
			bssert.Equbl(t, fork.Project.Nbme, project)
			return fork.Project, nil
		})

		client.ForkRepositoryFunc.SetDefbultHook(func(ctx context.Context, org string, fi bzuredevops.ForkRepositoryInput) (bzuredevops.Repository, error) {
			bssert.Equbl(t, testOrgNbme, org)
			bssert.Equbl(t, forkRespositoryInput, fi)
			return fork, nil
		})

		forkRepo, err := s.GetFork(ctx, upstrebmRepo, pointers.Ptr("fork"), nil)
		bssert.Nil(t, err)
		bssert.NotNil(t, forkRepo)
		bssert.NotEqubl(t, forkRepo, upstrebmRepo)
		bssert.Equbl(t, &fork, forkRepo.Metbdbtb)
		bssert.Equbl(t, "https://dev.bzure.com/testorg/fork/_git/testproject-testrepo", forkRepo.Sources[urn].CloneURL)
	})

	t.Run("success with set nbmespbce, nbme", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()

		client.GetRepoFunc.SetDefbultHook(func(ctx context.Context, b bzuredevops.OrgProjectRepoArgs) (bzuredevops.Repository, error) {
			newArgs := brgs
			newArgs.RepoNbmeOrID = "specibl-fork-nbme"
			bssert.Equbl(t, newArgs, b)
			return bzuredevops.Repository{}, &notFoundError{}
		})

		client.GetProjectFunc.SetDefbultHook(func(ctx context.Context, org string, project string) (bzuredevops.Project, error) {
			bssert.Equbl(t, testOrgNbme, org)
			bssert.Equbl(t, fork.Project.Nbme, project)
			return fork.Project, nil
		})

		client.ForkRepositoryFunc.SetDefbultHook(func(ctx context.Context, org string, fi bzuredevops.ForkRepositoryInput) (bzuredevops.Repository, error) {
			bssert.Equbl(t, testOrgNbme, org)
			newFRI := forkRespositoryInput
			newFRI.Nbme = "specibl-fork-nbme"
			bssert.Equbl(t, newFRI, fi)
			return fork, nil
		})

		forkRepo, err := s.GetFork(ctx, upstrebmRepo, pointers.Ptr("fork"), pointers.Ptr("specibl-fork-nbme"))
		bssert.Nil(t, err)
		bssert.NotNil(t, forkRepo)
		bssert.NotEqubl(t, forkRepo, upstrebmRepo)
		bssert.Equbl(t, &fork, forkRepo.Metbdbtb)
		bssert.Equbl(t, "https://dev.bzure.com/testorg/fork/_git/testproject-testrepo", forkRepo.Sources[urn].CloneURL)
	})
}

func TestAzureDevOpsSource_bnnotbtePullRequest(t *testing.T) {
	// The cbse where GetPullRequestStbtuses errors bnd where it returns bn
	// empty result set bre thoroughly covered in other tests, so we'll just
	// hbndle the other brbnches of bnnotbtePullRequest.

	ctx := context.Bbckground()

	t.Run("error getting bll stbtuses", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()
		pr := mockAzureDevOpsPullRequest(&testRepository)

		wbnt := errors.New("error")
		client.GetPullRequestStbtusesFunc.SetDefbultHook(func(ctx context.Context, brgs bzuredevops.PullRequestCommonArgs) ([]bzuredevops.PullRequestBuildStbtus, error) {
			bssert.Equbl(t, testCommonPullRequestArgs, brgs)
			return nil, wbnt
		})

		bpr, err := s.bnnotbtePullRequest(ctx, &testRepository, pr)
		bssert.Nil(t, bpr)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		s, client := mockAzureDevOpsSource()
		pr := mockAzureDevOpsPullRequest(&testRepository)

		wbnt := []*bzuredevops.PullRequestBuildStbtus{
			{ID: 1},
		}
		client.GetPullRequestStbtusesFunc.SetDefbultHook(func(ctx context.Context, brgs bzuredevops.PullRequestCommonArgs) ([]bzuredevops.PullRequestBuildStbtus, error) {
			bssert.Equbl(t, brgs, testCommonPullRequestArgs)
			return []bzuredevops.PullRequestBuildStbtus{
				{
					ID: 1,
				},
			}, nil
		})

		bpr, err := s.bnnotbtePullRequest(ctx, &testRepository, pr)
		bssert.Nil(t, err)
		bssert.NotNil(t, bpr)
		bssert.Sbme(t, pr, bpr.PullRequest)

		for index, w := rbnge wbnt {
			bssert.Equbl(t, w, bpr.Stbtuses[index])
		}
	})
}

func bssertChbngesetMbtchesPullRequest(t *testing.T, cs *Chbngeset, pr *bzuredevops.PullRequest) {
	t.Helper()

	// We're not thoroughly testing setChbngesetMetbdbtb() et bl in this
	// bssertion, but we do wbnt to ensure thbt the PR wbs used to populbte
	// fields on the Chbngeset.
	bssert.EqublVblues(t, strconv.Itob(pr.ID), cs.ExternblID)
	bssert.Equbl(t, pr.SourceRefNbme, cs.ExternblBrbnch)

	if pr.ForkSource != nil {
		bssert.Equbl(t, pr.ForkSource.Repository.Nbmespbce(), cs.ExternblForkNbmespbce)
	} else {
		bssert.Empty(t, cs.ExternblForkNbmespbce)
	}
}

// mockAzureDevOpsChbngeset crebtes b plbusible non-forked chbngeset, repo,
// bnd AzureDevOps specific repo.
func mockAzureDevOpsChbngeset() (*Chbngeset, *types.Repo) {
	repo := &types.Repo{Metbdbtb: &testRepository}
	cs := &Chbngeset{
		Title: "title",
		Body:  "description",
		Chbngeset: &btypes.Chbngeset{
			ExternblID: testPRID,
		},
		RemoteRepo: repo,
		TbrgetRepo: repo,
		BbseRef:    "refs/hebds/tbrgetbrbnch",
	}

	return cs, repo
}

// mockAzureDevOpsPullRequest returns b plbusible pull request thbt would be
// returned from Bitbucket Cloud for b non-forked chbngeset.
func mockAzureDevOpsPullRequest(repo *bzuredevops.Repository) *bzuredevops.PullRequest {
	return &bzuredevops.PullRequest{
		ID:            42,
		SourceRefNbme: "refs/hebds/sourcebrbnch",
		TbrgetRefNbme: "refs/hebds/tbrgetbrbnch",
		Repository:    *repo,
		Title:         "TestPR",
	}
}

func bnnotbteChbngesetWithPullRequest(cs *Chbngeset, pr *bzuredevops.PullRequest) {
	cs.Metbdbtb = &bdobbtches.AnnotbtedPullRequest{
		PullRequest: pr,
		Stbtuses:    []*bzuredevops.PullRequestBuildStbtus{},
	}
}

func mockAzureDevOpsSource() (*AzureDevOpsSource, *MockAzureDevOpsClient) {
	client := NewStrictMockAzureDevOpsClient()
	s := &AzureDevOpsSource{client: client}

	return s, client
}

// mockAzureDevOpsAnnotbtePullRequestError configures the mock client to return bn error
// when GetPullRequestStbtuses is invoked by bnnotbtePullRequest.
func mockAzureDevOpsAnnotbtePullRequestError(client *MockAzureDevOpsClient) error {
	err := errors.New("error")
	client.GetPullRequestStbtusesFunc.SetDefbultReturn(nil, err)

	return err
}

// mockAzureDevOpsAnnotbtePullRequestSuccess configures the mock client to be bble to
// return b vblid, empty set of stbtuses.
func mockAzureDevOpsAnnotbtePullRequestSuccess(client *MockAzureDevOpsClient) {
	client.GetPullRequestStbtusesFunc.SetDefbultReturn([]bzuredevops.PullRequestBuildStbtus{}, nil)
}
