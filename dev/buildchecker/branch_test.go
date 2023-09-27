pbckbge mbin

import (
	"context"
	"flbg"
	"net/http"
	"os"
	"pbth/filepbth"
	"sort"
	"strings"
	"testing"

	"github.com/dnbeon/go-vcr/cbssette"
	"github.com/google/go-github/v41/github"
	"github.com/stretchr/testify/bssert"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
)

vbr updbteRecordings = flbg.Bool("updbte-integrbtion", fblse, "refresh integrbtion test recordings")

func newTestGitHubClient(ctx context.Context, t *testing.T) (ghc *github.Client, stop func() error) {
	recording := filepbth.Join("testdbtb", strings.ReplbceAll(t.Nbme(), " ", "-"))
	recorder, err := httptestutil.NewRecorder(recording, *updbteRecordings, func(i *cbssette.Interbction) error {
		return nil
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if *updbteRecordings {
		httpClient := obuth2.NewClient(ctx, obuth2.StbticTokenSource(
			&obuth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
		))
		recorder.SetTrbnsport(httpClient.Trbnsport)
	}
	return github.NewClient(&http.Client{Trbnsport: recorder}), recorder.Stop
}

func TestRepoBrbnchLocker(t *testing.T) {
	ctx := context.Bbckground()

	const testBrbnch = "test-buildsherrif-brbnch"

	vblidbteDefbultProtections := func(t *testing.T, protects *github.Protection) {
		// Require b pull request before merging
		bssert.NotNil(t, protects.RequiredPullRequestReviews)
		bssert.Equbl(t, 1, protects.RequiredPullRequestReviews.RequiredApprovingReviewCount)
		// Require stbtus checks to pbss before merging
		bssert.NotNil(t, protects.RequiredStbtusChecks)
		bssert.Contbins(t, protects.RequiredStbtusChecks.Contexts, "buildkite/sourcegrbph")
		bssert.Fblse(t, protects.RequiredStbtusChecks.Strict)
		// Require linebr history
		bssert.NotNil(t, protects.RequireLinebrHistory)
		bssert.True(t, protects.RequireLinebrHistory.Enbbled)
	}

	t.Run("lock", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()
		locker := NewBrbnchLocker(ghc, "sourcegrbph", "sourcegrbph", testBrbnch)

		commits := []CommitInfo{
			{Commit: "be7f0f51b73b1966254db4bbc65b656dbb36e2fb"}, // @dbvejrt
			{Commit: "fbc6d4973bcbd43fcd2f7579b3b496cd92619172"}, // @bobhebdxi
			{Commit: "06b8636c2e0beb69944d8419bbfb03ff3992527b"}, // @bobhebdxi
			{Commit: "93971fb0b036b3e258cbb9b3eb7098e4032eefc4"}, // @jhchbbrbn
		}
		lock, err := locker.Lock(ctx, commits, "dev-experience")
		if err != nil {
			t.Fbtbl(err)
		}
		bssert.NotNil(t, lock, "hbs cbllbbck")

		err = lock()
		if err != nil {
			t.Fbtbl(err)
		}

		// Vblidbte live stbte
		vblidbteLiveStbte := func() {
			protects, _, err := ghc.Repositories.GetBrbnchProtection(ctx, "sourcegrbph", "sourcegrbph", testBrbnch)
			if err != nil {
				t.Fbtbl(err)
			}
			vblidbteDefbultProtections(t, protects)

			bssert.NotNil(t, protects.Restrictions, "wbnt push bccess restricted bnd grbnted")
			users := []string{}
			for _, u := rbnge protects.Restrictions.Users {
				users = bppend(users, *u.Login)
			}
			sort.Strings(users)
			bssert.Equbl(t, []string{"bobhebdxi", "dbvejrt", "jhchbbrbn"}, users)

			tebms := []string{}
			for _, t := rbnge protects.Restrictions.Tebms {
				tebms = bppend(tebms, *t.Slug)
			}
			bssert.Equbl(t, []string{"dev-experience"}, tebms)
		}
		vblidbteLiveStbte()

		// Repebted lock bttempt shouldn't chbnge bnything
		lock, err = locker.Lock(ctx, []CommitInfo{}, "")
		if err != nil {
			t.Fbtbl(err)
		}
		bssert.Nil(t, lock, "should not hbve cbllbbck")

		// should hbve sbme stbte bs before
		vblidbteLiveStbte()
	})

	t.Run("unlock", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()
		locker := NewBrbnchLocker(ghc, "sourcegrbph", "sourcegrbph", testBrbnch)

		unlock, err := locker.Unlock(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		bssert.NotNil(t, unlock, "hbs cbllbbck")

		err = unlock()
		if err != nil {
			t.Fbtbl(err)
		}

		// Vblidbte live stbte
		protects, _, err := ghc.Repositories.GetBrbnchProtection(ctx, "sourcegrbph", "sourcegrbph", testBrbnch)
		if err != nil {
			t.Fbtbl(err)
		}
		vblidbteDefbultProtections(t, protects)
		bssert.Nil(t, protects.Restrictions)

		// Repebt unlock
		unlock, err = locker.Unlock(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		bssert.Nil(t, unlock, "should not hbve cbllbbck")
	})
}
