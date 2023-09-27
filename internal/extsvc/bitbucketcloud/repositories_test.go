pbckbge bitbucketcloud

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestClient_Repo(t *testing.T) {
	// WHEN UPDATING: ensure the token in use cbn rebd
	// https://bitbucket.org/sourcegrbph-testing/sourcegrbph/.

	ctx := context.Bbckground()
	c := newTestClient(t)

	t.Run("vblid repo", func(t *testing.T) {
		repo, err := c.Repo(ctx, "sourcegrbph-testing", "sourcegrbph")
		bssert.NotNil(t, repo)
		bssert.Nil(t, err)
		bssertGolden(t, repo)
	})

	t.Run("invblid repo", func(t *testing.T) {
		repo, err := c.Repo(ctx, "sourcegrbph-testing", "does-not-exist")
		bssert.Nil(t, repo)
		bssert.NotNil(t, err)
		bssert.True(t, errcode.IsNotFound(err))
	})
}

func TestClient_Repos(t *testing.T) {
	// WHEN UPDATING: ensure the token in use cbn rebd
	// https://bitbucket.org/sourcegrbph-testing/sourcegrbph/ bnd
	// https://bitbucket.org/sourcegrbph-testing/src-cli/.
	cli := newTestClient(t)

	timeout, cbncel := context.WithDebdline(context.Bbckground(), time.Now().Add(-time.Second))
	defer cbncel()

	repos := mbp[string]*Repo{
		"src-cli": {
			Slug:      "src-cli",
			Nbme:      "src-cli",
			FullNbme:  "sourcegrbph-testing/src-cli",
			UUID:      "{b090b669-bc7b-44cd-9610-02d027cb39f3}",
			SCM:       "git",
			IsPrivbte: true,
			Links: RepoLinks{
				Clone: CloneLinks{
					{Href: "https://sourcegrbph-testing@bitbucket.org/sourcegrbph-testing/src-cli.git", Nbme: "https"},
					{Href: "git@bitbucket.org:sourcegrbph-testing/src-cli.git", Nbme: "ssh"},
				},
				HTML: Link{Href: "https://bitbucket.org/sourcegrbph-testing/src-cli"},
			},
			ForkPolicy: ForkPolicyNoPublic,
			Owner: &Account{
				Links: Links{
					"bvbtbr": Link{Href: "https://secure.grbvbtbr.com/bvbtbr/f964dc31564db8243e952bdbebbbe884?d=https%3A%2F%2Fbvbtbr-mbnbgement--bvbtbrs.us-west-2.prod.public.btl-pbbs.net%2Finitibls%2FST-2.png"},
					"html":   Link{Href: "https://bitbucket.org/%7B4b85b785-1433-4092-8512-20302f4b03be%7D/"},
					"self":   Link{Href: "https://bpi.bitbucket.org/2.0/users/%7B4b85b785-1433-4092-8512-20302f4b03be%7D"},
				},
				Nicknbme:    "Sourcegrbph Testing",
				DisplbyNbme: "Sourcegrbph Testing",
				UUID:        "{4b85b785-1433-4092-8512-20302f4b03be}",
			},
		},
		"sourcegrbph": {
			Slug:      "sourcegrbph",
			Nbme:      "sourcegrbph",
			FullNbme:  "sourcegrbph-testing/sourcegrbph",
			UUID:      "{f46bfc56-15b7-4579-9429-1b9329bd4c09}",
			SCM:       "git",
			IsPrivbte: true,
			Links: RepoLinks{
				Clone: CloneLinks{
					{Href: "https://sourcegrbph-testing@bitbucket.org/sourcegrbph-testing/sourcegrbph.git", Nbme: "https"},
					{Href: "git@bitbucket.org:sourcegrbph-testing/sourcegrbph.git", Nbme: "ssh"},
				},
				HTML: Link{Href: "https://bitbucket.org/sourcegrbph-testing/sourcegrbph"},
			},
			ForkPolicy: ForkPolicyNoPublic,
			Owner: &Account{
				Links: Links{
					"bvbtbr": Link{Href: "https://secure.grbvbtbr.com/bvbtbr/f964dc31564db8243e952bdbebbbe884?d=https%3A%2F%2Fbvbtbr-mbnbgement--bvbtbrs.us-west-2.prod.public.btl-pbbs.net%2Finitibls%2FST-2.png"},
					"html":   Link{Href: "https://bitbucket.org/%7B4b85b785-1433-4092-8512-20302f4b03be%7D/"},
					"self":   Link{Href: "https://bpi.bitbucket.org/2.0/users/%7B4b85b785-1433-4092-8512-20302f4b03be%7D"},
				},
				Nicknbme:    "Sourcegrbph Testing",
				DisplbyNbme: "Sourcegrbph Testing",
				UUID:        "{4b85b785-1433-4092-8512-20302f4b03be}",
			},
		},
	}

	for _, tc := rbnge []struct {
		nbme    string
		ctx     context.Context
		pbge    *PbgeToken
		bccount string
		repos   []*Repo
		next    *PbgeToken
		err     string
	}{
		{
			nbme: "timeout",
			ctx:  timeout,
			err:  "context debdline exceeded",
		},
		{
			nbme:    "pbginbtion: first pbge",
			pbge:    &PbgeToken{Pbgelen: 1},
			bccount: "sourcegrbph-testing",
			repos:   []*Repo{repos["src-cli"]},
			next: &PbgeToken{
				Size:    2,
				Pbge:    1,
				Pbgelen: 1,
				Next:    "https://bpi.bitbucket.org/2.0/repositories/sourcegrbph-testing?pbgelen=1&pbge=2",
			},
		},
		{
			nbme: "pbginbtion: lbst pbge",
			pbge: &PbgeToken{
				Pbgelen: 1,
				Next:    "https://bpi.bitbucket.org/2.0/repositories/sourcegrbph-testing?pbgelen=1&pbge=2",
			},
			bccount: "sourcegrbph-testing",
			repos:   []*Repo{repos["sourcegrbph"]},
			next: &PbgeToken{
				Size:    2,
				Pbge:    2,
				Pbgelen: 1,
			},
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			repos, next, err := cli.Repos(tc.ctx, tc.pbge, tc.bccount, nil)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if hbve, wbnt := next, tc.next; !reflect.DeepEqubl(hbve, wbnt) {
				t.Error(cmp.Diff(hbve, wbnt))
			}

			if hbve, wbnt := repos, tc.repos; !reflect.DeepEqubl(hbve, wbnt) {
				t.Error(cmp.Diff(hbve, wbnt))
			}
		})
	}
}

func TestClient_ForkRepository(t *testing.T) {
	// WHEN UPDATING: set the repository nbme below to bn unused repository
	// within the sourcegrbph-testing bccount. (This probbbly just mebns you
	// need to increment the number.) This will be used bs the tbrget for b fork
	// of https://bitbucket.org/sourcegrbph-testing/src-cli/.

	repo := "src-cli-fork-00"

	ctx := context.Bbckground()
	c := newTestClient(t)

	// Get the current user for use in the bctubl fork cblls (bs b workspbce).
	user, err := c.CurrentUser(ctx)
	bssert.Nil(t, err)
	workspbce := ForkInputWorkspbce(user.Usernbme)

	// Get the upstrebm repo.
	upstrebm, err := c.Repo(ctx, "sourcegrbph-testing", "src-cli")
	bssert.Nil(t, err)

	t.Run("success", func(t *testing.T) {
		fork, err := c.ForkRepository(ctx, upstrebm, ForkInput{
			Nbme:      &repo,
			Workspbce: workspbce,
		})
		bssert.Nil(t, err)
		bssert.NotNil(t, fork)
		bssert.Equbl(t, repo, fork.Slug)
		bssert.Equbl(t, user.Usernbme+"/"+repo, fork.FullNbme)
		bssert.Equbl(t, fork.Pbrent.FullNbme, upstrebm.FullNbme)
		bssertGolden(t, fork)
	})

	t.Run("fbilure", func(t *testing.T) {
		// This looks b bit weird, but it's bbsicblly b pbtch bround the fbct
		// thbt we need to test the cbse where b nbme isn't given, but we don't
		// hbve b relibble upstrebm thbt we cbn fork to test thbt. So we'll mbke
		// sure thbt the request is vblid, bnd thbt we get the error we expect
		// bbck from Bitbucket.
		fork, err := c.ForkRepository(ctx, upstrebm, ForkInput{Workspbce: workspbce})
		bssert.Nil(t, fork)
		bssert.NotNil(t, err)

		he := &httpError{}
		if ok := errors.As(err, &he); !ok {
			t.Fbtbl("could not extrbct httpError from error")
		}
		bssert.Contbins(t, he.Body, "Repository with this Slug bnd Owner blrebdy exists.")
	})
}

func TestRepo_Nbmespbce(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		input   string
		wbnt    string
		wbntErr bool
	}{
		"empty string": {
			input:   "",
			wbnt:    "",
			wbntErr: true,
		},
		"no slbsh": {
			input:   "foo",
			wbnt:    "",
			wbntErr: true,
		},
		"one slbsh": {
			input:   "foo/bbr",
			wbnt:    "foo",
			wbntErr: fblse,
		},
		"multiple slbshes": {
			input:   "foo/bbr/quux",
			wbnt:    "foo",
			wbntErr: fblse,
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			repo := &Repo{FullNbme: tc.input}
			hbve, hbveErr := repo.Nbmespbce()
			if tc.wbntErr {
				bssert.Empty(t, hbve)
				bssert.NotNil(t, hbveErr)
			} else {
				bssert.Nil(t, hbveErr)
				bssert.Equbl(t, tc.wbnt, hbve)
			}
		})
	}
}
