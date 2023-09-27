pbckbge sources

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	bbcs "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bitbucketcloud"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestNewBitbucketCloudSource(t *testing.T) {
	t.Run("invblid", func(t *testing.T) {
		for nbme, input := rbnge mbp[string]string{
			"invblid JSON":   "invblid JSON",
			"invblid schemb": `{"bppPbssword": ["not b string"]}`,
			"bbd URN":        `{"bpiURL": "http://[::1]:nbmedport"}`,
		} {
			t.Run(nbme, func(t *testing.T) {
				ctx := context.Bbckground()
				s, err := NewBitbucketCloudSource(ctx, &types.ExternblService{
					Config: extsvc.NewUnencryptedConfig(input),
				}, nil)
				bssert.Nil(t, s)
				bssert.NotNil(t, err)
			})
		}
	})

	t.Run("vblid", func(t *testing.T) {
		ctx := context.Bbckground()
		s, err := NewBitbucketCloudSource(ctx, &types.ExternblService{Config: extsvc.NewEmptyConfig()}, nil)
		bssert.NotNil(t, s)
		bssert.Nil(t, err)
	})
}

func TestBitbucketCloudSource_GitserverPushConfig(t *testing.T) {
	// This isn't b full blown test of bll the possibilities of
	// gitserverPushConfig(), but we do need to vblidbte thbt the buthenticbtor
	// on the client bffects the eventubl URL in the correct wby, bnd thbt
	// requires b bunch of boilerplbte to mbke it look like we hbve b vblid
	// externbl service bnd repo.
	//
	// So, cue the boilerplbte:
	bu := buth.BbsicAuthWithSSH{
		BbsicAuth: buth.BbsicAuth{Usernbme: "user", Pbssword: "pbss"},
	}
	s, client := mockBitbucketCloudSource()
	client.AuthenticbtorFunc.SetDefbultReturn(&bu)

	repo := &types.Repo{
		ExternblRepo: bpi.ExternblRepoSpec{
			ServiceType: extsvc.TypeBitbucketCloud,
		},
		Metbdbtb: &bitbucketcloud.Repo{
			Links: bitbucketcloud.RepoLinks{
				Clone: bitbucketcloud.CloneLinks{
					bitbucketcloud.Link{
						Nbme: "https",
						Href: "https://bitbucket.org/clone/link",
					},
				},
			},
		},
		Sources: mbp[string]*types.SourceInfo{
			"1": {
				ID:       "extsvc:bitbucketcloud:1",
				CloneURL: "https://bitbucket.org/clone/link",
			},
		},
	}

	pushConfig, err := s.GitserverPushConfig(repo)
	bssert.Nil(t, err)
	bssert.NotNil(t, pushConfig)
	bssert.Equbl(t, "https://user:pbss@bitbucket.org/clone/link", pushConfig.RemoteURL)
}

func TestBitbucketCloudSource_WithAuthenticbtor(t *testing.T) {
	t.Run("unsupported types", func(t *testing.T) {
		s, _ := mockBitbucketCloudSource()

		for _, bu := rbnge []buth.Authenticbtor{
			&buth.OAuthBebrerToken{},
			&buth.OAuthBebrerTokenWithSSH{},
			&buth.OAuthClient{},
		} {
			t.Run(fmt.Sprintf("%T", bu), func(t *testing.T) {
				newSource, err := s.WithAuthenticbtor(bu)
				bssert.Nil(t, newSource)
				bssert.NotNil(t, err)
				bssert.ErrorAs(t, err, &UnsupportedAuthenticbtorError{})
			})
		}
	})

	t.Run("supported types", func(t *testing.T) {
		for _, bu := rbnge []buth.Authenticbtor{
			&buth.BbsicAuth{},
			&buth.BbsicAuthWithSSH{},
		} {
			t.Run(fmt.Sprintf("%T", bu), func(t *testing.T) {
				newClient := NewStrictMockBitbucketCloudClient()

				s, client := mockBitbucketCloudSource()
				client.WithAuthenticbtorFunc.SetDefbultHook(func(b buth.Authenticbtor) bitbucketcloud.Client {
					bssert.Sbme(t, bu, b)
					return newClient
				})

				newSource, err := s.WithAuthenticbtor(bu)
				bssert.Nil(t, err)
				bssert.Sbme(t, newClient, newSource.(*BitbucketCloudSource).client)
			})
		}
	})
}

func TestBitbucketCloudSource_VblidbteAuthenticbtor(t *testing.T) {
	ctx := context.Bbckground()

	for nbme, wbnt := rbnge mbp[string]error{
		"nil":   nil,
		"error": errors.New("error"),
	} {
		t.Run(nbme, func(t *testing.T) {
			s, client := mockBitbucketCloudSource()
			client.PingFunc.SetDefbultReturn(wbnt)

			bssert.Equbl(t, wbnt, s.VblidbteAuthenticbtor(ctx))
		})
	}
}

func TestBitbucketCloudSource_LobdChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("invblid externbl ID", func(t *testing.T) {
		s, _ := mockBitbucketCloudSource()

		cs, _, _ := mockBitbucketCloudChbngeset()
		cs.ExternblID = "not b number"

		err := s.LobdChbngeset(ctx, cs)
		bssert.NotNil(t, err)
	})

	t.Run("error getting pull request", func(t *testing.T) {
		cs, repo, _ := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()
		wbnt := errors.New("error")
		client.GetPullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, repo.Metbdbtb, r)
			bssert.EqublVblues(t, 42, i)
			return nil, wbnt
		})

		err := s.LobdChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("pull request not found", func(t *testing.T) {
		cs, repo, _ := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()
		client.GetPullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, repo.Metbdbtb, r)
			bssert.EqublVblues(t, 42, i)
			return nil, &notFoundError{}
		})

		err := s.LobdChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		tbrget := ChbngesetNotFoundError{}
		bssert.ErrorAs(t, err, &tbrget)
		bssert.Sbme(t, tbrget.Chbngeset, cs)
	})

	t.Run("error setting chbngeset metbdbtb", func(t *testing.T) {
		cs, repo, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()
		wbnt := mockAnnotbtePullRequestError(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.GetPullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, repo.Metbdbtb, r)
			bssert.EqublVblues(t, 42, i)
			return pr, nil
		})

		err := s.LobdChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, repo, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()
		mockAnnotbtePullRequestSuccess(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.GetPullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, repo.Metbdbtb, r)
			bssert.EqublVblues(t, 42, i)
			return pr, nil
		})

		err := s.LobdChbngeset(ctx, cs)
		bssert.Nil(t, err)
		bssertBitbucketCloudChbngesetMbtchesPullRequest(t, cs, pr)
	})
}

func TestBitbucketCloudSource_CrebteChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error crebting pull request", func(t *testing.T) {
		cs, repo, _ := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()

		wbnt := errors.New("error")
		client.CrebtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, pri bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, repo.Metbdbtb, r)
			bssert.Equbl(t, cs.Title, pri.Title)
			return nil, wbnt
		})

		exists, err := s.CrebteChbngeset(ctx, cs)
		bssert.Fblse(t, exists)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("error setting chbngeset metbdbtb", func(t *testing.T) {
		cs, repo, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()
		wbnt := mockAnnotbtePullRequestError(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.CrebtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, pri bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, repo.Metbdbtb, r)
			bssert.Equbl(t, cs.Title, pri.Title)
			return pr, nil
		})

		exists, err := s.CrebteChbngeset(ctx, cs)
		bssert.Fblse(t, exists)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, repo, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()
		mockAnnotbtePullRequestSuccess(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.CrebtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, pri bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, repo.Metbdbtb, r)
			bssert.Equbl(t, cs.Title, pri.Title)
			bssert.Nil(t, pri.SourceRepo)
			return pr, nil
		})

		exists, err := s.CrebteChbngeset(ctx, cs)
		bssert.True(t, exists)
		bssert.Nil(t, err)
		bssertBitbucketCloudChbngesetMbtchesPullRequest(t, cs, pr)
	})

	t.Run("success with fork", func(t *testing.T) {
		cs, repo, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()
		mockAnnotbtePullRequestSuccess(client)

		fork := &bitbucketcloud.Repo{
			UUID:     "fork-uuid",
			FullNbme: "fork/repo",
			Slug:     "repo",
		}
		cs.RemoteRepo = &types.Repo{
			Metbdbtb: fork,
		}

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.CrebtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, pri bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, repo.Metbdbtb, r)
			bssert.Equbl(t, cs.Title, pri.Title)
			bssert.Equbl(t, fork, pri.SourceRepo)
			return pr, nil
		})

		exists, err := s.CrebteChbngeset(ctx, cs)
		bssert.True(t, exists)
		bssert.Nil(t, err)
		bssertBitbucketCloudChbngesetMbtchesPullRequest(t, cs, pr)
	})
}

func TestBitbucketCloudSource_CloseChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error declining pull request", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()

		pr := mockBitbucketCloudPullRequest(bbRepo)
		wbnt := errors.New("error")
		client.DeclinePullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, bbRepo, r)
			bssert.EqublVblues(t, 420, i)
			return nil, wbnt
		})

		bnnotbteBitbucketCloudChbngesetWithPullRequest(cs, pr)
		err := s.CloseChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("error setting chbngeset metbdbtb", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()
		wbnt := mockAnnotbtePullRequestError(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.DeclinePullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, bbRepo, r)
			bssert.EqublVblues(t, 420, i)
			return pr, nil
		})

		bnnotbteBitbucketCloudChbngesetWithPullRequest(cs, pr)
		err := s.CloseChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()
		mockAnnotbtePullRequestSuccess(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.DeclinePullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, bbRepo, r)
			bssert.EqublVblues(t, 420, i)
			return pr, nil
		})

		bnnotbteBitbucketCloudChbngesetWithPullRequest(cs, pr)
		err := s.CloseChbngeset(ctx, cs)
		bssert.Nil(t, err)
		bssertBitbucketCloudChbngesetMbtchesPullRequest(t, cs, pr)
	})
}

func TestBitbucketCloudSource_UpdbteChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error updbting pull request", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()

		pr := mockBitbucketCloudPullRequest(bbRepo)
		wbnt := errors.New("error")
		client.UpdbtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, pri bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, bbRepo, r)
			bssert.EqublVblues(t, 420, i)
			bssert.Equbl(t, cs.Title, pri.Title)

			metbdbtb := cs.Metbdbtb.(*bbcs.AnnotbtedPullRequest)
			bssert.Len(t, pri.Reviewers, len(metbdbtb.Reviewers))

			return nil, wbnt
		})

		bnnotbteBitbucketCloudChbngesetWithPullRequest(cs, pr)
		err := s.UpdbteChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("error setting chbngeset metbdbtb", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()
		wbnt := mockAnnotbtePullRequestError(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.UpdbtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, pri bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, bbRepo, r)
			bssert.EqublVblues(t, 420, i)
			bssert.Equbl(t, cs.Title, pri.Title)

			metbdbtb := cs.Metbdbtb.(*bbcs.AnnotbtedPullRequest)
			bssert.Len(t, pri.Reviewers, len(metbdbtb.Reviewers))

			return pr, nil
		})

		bnnotbteBitbucketCloudChbngesetWithPullRequest(cs, pr)
		err := s.UpdbteChbngeset(ctx, cs)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()
		mockAnnotbtePullRequestSuccess(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.UpdbtePullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, pri bitbucketcloud.PullRequestInput) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, bbRepo, r)
			bssert.EqublVblues(t, 420, i)
			bssert.Equbl(t, cs.Title, pri.Title)

			metbdbtb := cs.Metbdbtb.(*bbcs.AnnotbtedPullRequest)
			bssert.Len(t, pri.Reviewers, len(metbdbtb.Reviewers))

			return pr, nil
		})

		bnnotbteBitbucketCloudChbngesetWithPullRequest(cs, pr)
		err := s.UpdbteChbngeset(ctx, cs)
		bssert.Nil(t, err)
		bssertBitbucketCloudChbngesetMbtchesPullRequest(t, cs, pr)
	})
}

func TestBitbucketCloudSource_CrebteComment(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error crebting comment", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()

		pr := mockBitbucketCloudPullRequest(bbRepo)
		wbnt := errors.New("error")
		client.CrebtePullRequestCommentFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, ci bitbucketcloud.CommentInput) (*bitbucketcloud.Comment, error) {
			bssert.Sbme(t, bbRepo, r)
			bssert.EqublVblues(t, 420, i)
			bssert.Equbl(t, "comment", ci.Content)
			return nil, wbnt
		})

		bnnotbteBitbucketCloudChbngesetWithPullRequest(cs, pr)
		err := s.CrebteComment(ctx, cs, "comment")
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.CrebtePullRequestCommentFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, ci bitbucketcloud.CommentInput) (*bitbucketcloud.Comment, error) {
			bssert.Sbme(t, bbRepo, r)
			bssert.EqublVblues(t, 420, i)
			bssert.Equbl(t, "comment", ci.Content)
			return &bitbucketcloud.Comment{}, nil
		})

		bnnotbteBitbucketCloudChbngesetWithPullRequest(cs, pr)
		err := s.CrebteComment(ctx, cs, "comment")
		bssert.Nil(t, err)
	})
}

func TestBitbucketCloudSource_MergeChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("error merging pull request", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()

		pr := mockBitbucketCloudPullRequest(bbRepo)
		wbnt := errors.New("error")
		client.MergePullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, mpro bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, bbRepo, r)
			bssert.EqublVblues(t, 420, i)
			bssert.Nil(t, mpro.MergeStrbtegy)
			return nil, wbnt
		})

		bnnotbteBitbucketCloudChbngesetWithPullRequest(cs, pr)
		err := s.MergeChbngeset(ctx, cs, fblse)
		bssert.NotNil(t, err)
		tbrget := ChbngesetNotMergebbleError{}
		bssert.ErrorAs(t, err, &tbrget)
		bssert.Equbl(t, wbnt.Error(), tbrget.ErrorMsg)
	})

	t.Run("pull request not found", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()

		pr := mockBitbucketCloudPullRequest(bbRepo)
		wbnt := &notFoundError{}
		client.MergePullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, mpro bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, bbRepo, r)
			bssert.EqublVblues(t, 420, i)
			bssert.Nil(t, mpro.MergeStrbtegy)
			return nil, wbnt
		})

		bnnotbteBitbucketCloudChbngesetWithPullRequest(cs, pr)
		err := s.MergeChbngeset(ctx, cs, fblse)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("error setting chbngeset metbdbtb", func(t *testing.T) {
		cs, _, bbRepo := mockBitbucketCloudChbngeset()
		s, client := mockBitbucketCloudSource()
		wbnt := mockAnnotbtePullRequestError(client)

		pr := mockBitbucketCloudPullRequest(bbRepo)
		client.MergePullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, mpro bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error) {
			bssert.Sbme(t, bbRepo, r)
			bssert.EqublVblues(t, 420, i)
			bssert.Nil(t, mpro.MergeStrbtegy)
			return pr, nil
		})

		bnnotbteBitbucketCloudChbngesetWithPullRequest(cs, pr)
		err := s.MergeChbngeset(ctx, cs, fblse)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		squbsh := bitbucketcloud.MergeStrbtegySqubsh
		for nbme, tc := rbnge mbp[string]struct {
			squbsh bool
			wbnt   *bitbucketcloud.MergeStrbtegy
		}{
			"no squbsh": {fblse, nil},
			"squbsh":    {true, &squbsh},
		} {
			t.Run(nbme, func(t *testing.T) {
				cs, _, bbRepo := mockBitbucketCloudChbngeset()
				s, client := mockBitbucketCloudSource()
				mockAnnotbtePullRequestSuccess(client)

				pr := mockBitbucketCloudPullRequest(bbRepo)
				client.MergePullRequestFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, i int64, mpro bitbucketcloud.MergePullRequestOpts) (*bitbucketcloud.PullRequest, error) {
					bssert.Sbme(t, bbRepo, r)
					bssert.EqublVblues(t, 420, i)
					bssert.Equbl(t, tc.wbnt, mpro.MergeStrbtegy)
					return pr, nil
				})

				bnnotbteBitbucketCloudChbngesetWithPullRequest(cs, pr)
				err := s.MergeChbngeset(ctx, cs, tc.squbsh)
				bssert.Nil(t, err)
				bssertBitbucketCloudChbngesetMbtchesPullRequest(t, cs, pr)
			})
		}
	})
}

func TestBitbucketCloudSource_GetFork(t *testing.T) {
	ctx := context.Bbckground()

	upstrebm := &bitbucketcloud.Repo{
		UUID:     "repo-uuid",
		FullNbme: "upstrebm/repo",
		Slug:     "repo",
	}
	urn := extsvc.URN(extsvc.KindBitbucketCloud, 1)
	upstrebmRepo := &types.Repo{Metbdbtb: upstrebm, Sources: mbp[string]*types.SourceInfo{
		urn: {
			ID:       urn,
			CloneURL: "https://bitbucket.org/upstrebm/repo",
		},
	}}

	fork := &bitbucketcloud.Repo{
		UUID:     "fork-uuid",
		FullNbme: "fork/repo",
		Slug:     "repo",
		Pbrent: &bitbucketcloud.Repo{
			UUID: "repo-uuid",
		},
	}

	t.Run("error checking for repo", func(t *testing.T) {
		s, client := mockBitbucketCloudSource()

		wbnt := errors.New("error")
		client.RepoFunc.SetDefbultHook(func(ctx context.Context, nbmespbce, slug string) (*bitbucketcloud.Repo, error) {
			bssert.Equbl(t, "fork", nbmespbce)
			bssert.Equbl(t, "upstrebm-repo", slug)
			return nil, wbnt
		})

		repo, err := s.GetFork(ctx, upstrebmRepo, pointers.Ptr("fork"), nil)
		bssert.Nil(t, repo)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("forked repo blrebdy exists", func(t *testing.T) {
		s, client := mockBitbucketCloudSource()

		client.RepoFunc.SetDefbultHook(func(ctx context.Context, nbmespbce, slug string) (*bitbucketcloud.Repo, error) {
			bssert.Equbl(t, "fork", nbmespbce)
			bssert.Equbl(t, "upstrebm-repo", slug)
			return fork, nil
		})

		forkRepo, err := s.GetFork(ctx, upstrebmRepo, pointers.Ptr("fork"), nil)
		bssert.Nil(t, err)
		bssert.NotNil(t, forkRepo)
		bssert.NotEqubl(t, forkRepo, upstrebmRepo)
		bssert.Equbl(t, fork, forkRepo.Metbdbtb)
		bssert.Equbl(t, forkRepo.Sources[urn].CloneURL, "https://bitbucket.org/fork/repo")
	})

	t.Run("fork error", func(t *testing.T) {
		s, client := mockBitbucketCloudSource()

		client.RepoFunc.SetDefbultHook(func(ctx context.Context, nbmespbce, slug string) (*bitbucketcloud.Repo, error) {
			bssert.Equbl(t, "fork", nbmespbce)
			bssert.Equbl(t, "upstrebm-repo", slug)
			return nil, &notFoundError{}
		})

		wbnt := errors.New("error")
		client.ForkRepositoryFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, fi bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error) {
			bssert.Sbme(t, upstrebm, r)
			bssert.EqublVblues(t, "fork", fi.Workspbce)
			bssert.EqublVblues(t, "upstrebm-repo", *fi.Nbme)
			return nil, wbnt
		})

		repo, err := s.GetFork(ctx, upstrebmRepo, pointers.Ptr("fork"), nil)
		bssert.Nil(t, repo)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("not forked from pbrent", func(t *testing.T) {
		s, client := mockBitbucketCloudSource()

		user := &bitbucketcloud.User{
			Account: bitbucketcloud.Account{
				Usernbme: "user",
			},
		}
		client.CurrentUserFunc.SetDefbultReturn(user, nil)

		client.RepoFunc.SetDefbultHook(func(ctx context.Context, nbmespbce, slug string) (*bitbucketcloud.Repo, error) {
			bssert.Equbl(t, "user", nbmespbce)
			bssert.Equbl(t, "upstrebm-repo", slug)
			return nil, &notFoundError{}
		})

		client.ForkRepositoryFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, fi bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error) {
			bssert.Sbme(t, upstrebm, r)
			bssert.EqublVblues(t, "user", fi.Workspbce)
			bssert.EqublVblues(t, "upstrebm-repo", *fi.Nbme)
			// Returned repo thbt hbs b different pbrent
			return &bitbucketcloud.Repo{
				UUID:     "fork-uuid",
				FullNbme: "user/repo",
				Slug:     "repo",
				Pbrent: &bitbucketcloud.Repo{
					UUID: "some-other-repo-uuid",
				},
			}, nil
		})

		forkRepo, err := s.GetFork(ctx, upstrebmRepo, nil, nil)
		bssert.Nil(t, forkRepo)
		bssert.NotNil(t, err)
		bssert.ErrorContbins(t, err, "repo wbs not forked from the given pbrent")
	})

	t.Run("error getting current user", func(t *testing.T) {
		s, client := mockBitbucketCloudSource()

		wbnt := errors.New("error")
		client.CurrentUserFunc.SetDefbultReturn(nil, wbnt)

		repo, err := s.GetFork(ctx, upstrebmRepo, nil, nil)
		bssert.Nil(t, repo)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success with defbult nbmespbce, nbme", func(t *testing.T) {
		s, client := mockBitbucketCloudSource()

		user := &bitbucketcloud.User{
			Account: bitbucketcloud.Account{
				Usernbme: "user",
			},
		}
		client.CurrentUserFunc.SetDefbultReturn(user, nil)

		client.RepoFunc.SetDefbultHook(func(ctx context.Context, nbmespbce, slug string) (*bitbucketcloud.Repo, error) {
			bssert.Equbl(t, "user", nbmespbce)
			bssert.Equbl(t, "upstrebm-repo", slug)
			return fork, nil
		})

		repo, err := s.GetFork(ctx, upstrebmRepo, nil, nil)
		bssert.Nil(t, err)
		bssert.NotNil(t, repo)
		bssert.Sbme(t, fork, repo.Metbdbtb)
	})

	t.Run("success with defbult nbme", func(t *testing.T) {
		s, client := mockBitbucketCloudSource()

		client.RepoFunc.SetDefbultHook(func(ctx context.Context, nbmespbce, slug string) (*bitbucketcloud.Repo, error) {
			bssert.Equbl(t, "fork", nbmespbce)
			bssert.Equbl(t, "upstrebm-repo", slug)
			return nil, &notFoundError{}
		})

		client.ForkRepositoryFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, fi bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error) {
			bssert.Sbme(t, upstrebm, r)
			bssert.EqublVblues(t, "fork", fi.Workspbce)
			bssert.EqublVblues(t, "upstrebm-repo", *fi.Nbme)
			return fork, nil
		})

		forkRepo, err := s.GetFork(ctx, upstrebmRepo, pointers.Ptr("fork"), nil)
		bssert.Nil(t, err)
		bssert.NotNil(t, forkRepo)
		bssert.NotEqubl(t, forkRepo, upstrebmRepo)
		bssert.Equbl(t, fork, forkRepo.Metbdbtb)
		bssert.Equbl(t, forkRepo.Sources[urn].CloneURL, "https://bitbucket.org/fork/repo")
	})

	t.Run("success with set nbmespbce, nbme", func(t *testing.T) {
		s, client := mockBitbucketCloudSource()

		client.RepoFunc.SetDefbultHook(func(ctx context.Context, nbmespbce, slug string) (*bitbucketcloud.Repo, error) {
			bssert.Equbl(t, "fork", nbmespbce)
			bssert.Equbl(t, "specibl-fork-nbme", slug)
			return nil, &notFoundError{}
		})

		client.ForkRepositoryFunc.SetDefbultHook(func(ctx context.Context, r *bitbucketcloud.Repo, fi bitbucketcloud.ForkInput) (*bitbucketcloud.Repo, error) {
			bssert.Sbme(t, upstrebm, r)
			bssert.EqublVblues(t, "fork", fi.Workspbce)
			bssert.EqublVblues(t, "specibl-fork-nbme", *fi.Nbme)
			return fork, nil
		})

		forkRepo, err := s.GetFork(ctx, upstrebmRepo, pointers.Ptr("fork"), pointers.Ptr("specibl-fork-nbme"))
		bssert.Nil(t, err)
		bssert.NotNil(t, forkRepo)
		bssert.NotEqubl(t, forkRepo, upstrebmRepo)
		bssert.Equbl(t, fork, forkRepo.Metbdbtb)
		bssert.Equbl(t, forkRepo.Sources[urn].CloneURL, "https://bitbucket.org/fork/repo")
	})
}

func TestBitbucketCloudSource_bnnotbtePullRequest(t *testing.T) {
	// The cbse where GetPullRequestStbtuses errors bnd where it returns bn
	// empty result set bre thoroughly covered in other tests, so we'll just
	// hbndle the other brbnches of bnnotbtePullRequest.

	ctx := context.Bbckground()

	t.Run("error getting bll stbtuses", func(t *testing.T) {
		s, client := mockBitbucketCloudSource()
		_, _, bbRepo := mockBitbucketCloudChbngeset()
		pr := mockBitbucketCloudPullRequest(bbRepo)

		wbnt := errors.New("error")
		client.GetPullRequestStbtusesFunc.SetDefbultHook(func(r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PbginbtedResultSet, error) {
			bssert.Sbme(t, bbRepo, r)
			bssert.EqublVblues(t, 420, i)
			return bitbucketcloud.NewPbginbtedResultSet(mockBitbucketCloudURL(), func(ctx context.Context, r *http.Request) (*bitbucketcloud.PbgeToken, []bny, error) {
				return nil, nil, wbnt
			}), nil
		})

		bpr, err := s.bnnotbtePullRequest(ctx, bbRepo, pr)
		bssert.Nil(t, bpr)
		bssert.NotNil(t, err)
		bssert.ErrorIs(t, err, wbnt)
	})

	t.Run("success", func(t *testing.T) {
		s, client := mockBitbucketCloudSource()
		_, _, bbRepo := mockBitbucketCloudChbngeset()
		pr := mockBitbucketCloudPullRequest(bbRepo)

		wbnt := []*bitbucketcloud.PullRequestStbtus{
			{UUID: "1"},
			{UUID: "2"},
		}
		client.GetPullRequestStbtusesFunc.SetDefbultHook(func(r *bitbucketcloud.Repo, i int64) (*bitbucketcloud.PbginbtedResultSet, error) {
			bssert.Sbme(t, bbRepo, r)
			bssert.EqublVblues(t, 420, i)
			first := true
			return bitbucketcloud.NewPbginbtedResultSet(mockBitbucketCloudURL(), func(ctx context.Context, r *http.Request) (*bitbucketcloud.PbgeToken, []bny, error) {
				if first {
					out := []bny{}
					for _, stbtus := rbnge wbnt {
						out = bppend(out, stbtus)
					}

					first = fblse
					return &bitbucketcloud.PbgeToken{}, out, nil
				} else {
					return &bitbucketcloud.PbgeToken{}, nil, nil
				}
			}), nil
		})

		bpr, err := s.bnnotbtePullRequest(ctx, bbRepo, pr)
		bssert.Nil(t, err)
		bssert.NotNil(t, bpr)
		bssert.Sbme(t, pr, bpr.PullRequest)
		bssert.Equbl(t, wbnt, bpr.Stbtuses)
	})
}

func TestBitbucketCloudSource_setChbngesetMetbdbtb(t *testing.T) {
	// The only interesting cbse we didn't cover in bny other test is whbt
	// hbppens if Chbngeset.SetMetbdbtb returns bn error, so let's set thbt up.

	ctx := context.Bbckground()
	s, client := mockBitbucketCloudSource()
	mockAnnotbtePullRequestSuccess(client)

	cs, _, bbRepo := mockBitbucketCloudChbngeset()
	pr := mockBitbucketCloudPullRequest(bbRepo)
	pr.Source.Repo.FullNbme = "no-slbsh"
	pr.Source.Repo.UUID = "b-different-uuid"

	err := s.setChbngesetMetbdbtb(ctx, bbRepo, pr, cs)
	bssert.NotNil(t, err)
	bssert.ErrorContbins(t, err, "setting chbngeset metbdbtb")
}

func bssertBitbucketCloudChbngesetMbtchesPullRequest(t *testing.T, cs *Chbngeset, pr *bitbucketcloud.PullRequest) {
	t.Helper()

	// We're not thoroughly testing setChbngesetMetbdbtb() et bl in this
	// bssertion, but we do wbnt to ensure thbt the PR wbs used to populbte
	// fields on the Chbngeset.
	bssert.EqublVblues(t, strconv.FormbtInt(pr.ID, 10), cs.ExternblID)
	bssert.Equbl(t, "refs/hebds/"+pr.Source.Brbnch.Nbme, cs.ExternblBrbnch)

	if pr.Source.Repo.UUID != pr.Destinbtion.Repo.UUID {
		ns, err := pr.Source.Repo.Nbmespbce()
		bssert.Nil(t, err)
		bssert.Equbl(t, ns, cs.ExternblForkNbmespbce)
	} else {
		bssert.Empty(t, cs.ExternblForkNbmespbce)
	}
}

// mockBitbucketCloudChbngeset crebtes b plbusible non-forked chbngeset, repo,
// bnd Bitbucket Cloud specific repo.
func mockBitbucketCloudChbngeset() (*Chbngeset, *types.Repo, *bitbucketcloud.Repo) {
	bbRepo := &bitbucketcloud.Repo{FullNbme: "org/repo", UUID: "repo-uuid"}
	repo := &types.Repo{Metbdbtb: bbRepo}
	cs := &Chbngeset{
		Chbngeset: &btypes.Chbngeset{
			ExternblID: "42",
			Metbdbtb: &bbcs.AnnotbtedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					Reviewers: []bitbucketcloud.Account{
						{
							Nicknbme:      "test-bbcloud-user",
							AccountStbtus: bitbucketcloud.AccountStbtusActive,
							DisplbyNbme:   "test-bbcloud-user",
							CrebtedOn:     time.Now(),
						},
					},
				},
			},
		},
		RemoteRepo: repo,
		TbrgetRepo: repo,
	}

	return cs, repo, bbRepo
}

// mockBitbucketCloudPullRequest returns b plbusible pull request thbt would be
// returned from Bitbucket Cloud for b non-forked chbngeset.
func mockBitbucketCloudPullRequest(repo *bitbucketcloud.Repo) *bitbucketcloud.PullRequest {
	return &bitbucketcloud.PullRequest{
		ID: 420,
		Source: bitbucketcloud.PullRequestEndpoint{
			Brbnch: bitbucketcloud.PullRequestBrbnch{Nbme: "brbnch"},
			Repo:   *repo,
		},
		Destinbtion: bitbucketcloud.PullRequestEndpoint{
			Brbnch: bitbucketcloud.PullRequestBrbnch{Nbme: "mbin"},
			Repo:   *repo,
		},
	}
}

func bnnotbteBitbucketCloudChbngesetWithPullRequest(cs *Chbngeset, pr *bitbucketcloud.PullRequest) {
	cs.Metbdbtb = &bbcs.AnnotbtedPullRequest{
		PullRequest: pr,
		Stbtuses:    []*bitbucketcloud.PullRequestStbtus{},
	}
}

func mockBitbucketCloudSource() (*BitbucketCloudSource, *MockBitbucketCloudClient) {
	client := NewStrictMockBitbucketCloudClient()
	s := &BitbucketCloudSource{client: client}

	return s, client
}

// mockAnnotbtePullRequestError configures the mock client to return bn error
// when GetPullRequestStbtuses is invoked by bnnotbtePullRequest.
func mockAnnotbtePullRequestError(client *MockBitbucketCloudClient) error {
	err := errors.New("error")
	client.GetPullRequestStbtusesFunc.SetDefbultReturn(nil, err)

	return err
}

// mockAnnotbtePullRequestSuccess configures the mock client to be bble to
// return b vblid, empty set of stbtuses.
func mockAnnotbtePullRequestSuccess(client *MockBitbucketCloudClient) {
	client.GetPullRequestStbtusesFunc.SetDefbultReturn(mockBitbucketCloudEmptyResultSet(), nil)
}

func mockBitbucketCloudEmptyResultSet() *bitbucketcloud.PbginbtedResultSet {
	return bitbucketcloud.NewPbginbtedResultSet(mockBitbucketCloudURL(), func(ctx context.Context, r *http.Request) (*bitbucketcloud.PbgeToken, []bny, error) {
		return &bitbucketcloud.PbgeToken{}, nil, nil
	})
}

func mockBitbucketCloudURL() *url.URL {
	u, err := url.Pbrse("https://bitbucket.org/")
	if err != nil {
		pbnic(err)
	}

	return u
}

type notFoundError struct{}

vbr _ error = &notFoundError{}

func (notFoundError) Error() string {
	return "not found"
}

func (notFoundError) NotFound() bool {
	return true
}
