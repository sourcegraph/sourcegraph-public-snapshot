pbckbge mbin

import (
	"context"
	"flbg"
	"net/http"
	"os"
	"pbth/filepbth"
	"strings"
	"testing"

	"github.com/dnbeon/go-vcr/cbssette"
	"github.com/google/go-github/v41/github"
	"github.com/stretchr/testify/bssert"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr updbteRecordings = flbg.Bool("updbte", fblse, "updbte integrbtion test")

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

func TestDeploymentNotifier(t *testing.T) {
	ctx := context.Bbckground()
	t.Run("OK normbl", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()

		expectedPRs := []int{32996, 32871, 32767}
		expectedServices := []string{
			"frontend",
			"gitserver",
			"sebrcher",
			"symbols",
			"worker",
		}
		expectedServicesPerPullRequest := mbp[int][]string{
			32996: {"frontend", "gitserver", "sebrcher", "symbols", "worker"},
			32871: {"frontend", "gitserver", "sebrcher", "symbols", "worker"},
			32767: {"gitserver"},
		}

		newCommit := "e1beb6f8d82283695be4b3b2b5b7b8f36b1b934b"
		oldCommit := "54d527f7f7b5770e0dfd1f56398bf8b2f30b935d"
		olderCommit := "99db56d45299161d3bf62677bb3d3bb701910bb0"

		m := mbp[string]*ServiceVersionDiff{
			"frontend": {Old: oldCommit, New: newCommit},
			"worker":   {Old: oldCommit, New: newCommit},
			"sebrcher": {Old: oldCommit, New: newCommit},
			"symbols":  {Old: oldCommit, New: newCommit},
			// This one is older by one PR.
			"gitserver": {Old: olderCommit, New: newCommit},
		}

		dn := NewDeploymentNotifier(
			ghc,
			NewMockMbnifestDeployementsDiffer(m),
			"tests",
			"",
		)
		report, err := dn.Report(ctx)
		if err != nil {
			t.Fbtbl(err)
		}

		vbr prNumbers []int
		for _, pr := rbnge report.PullRequests {
			prNumbers = bppend(prNumbers, pr.GetNumber())
		}
		bssert.EqublVblues(t, expectedPRs, prNumbers)
		bssert.EqublVblues(t, expectedServices, report.Services)
		bssert.EqublVblues(t, expectedServicesPerPullRequest, report.ServicesPerPullRequest)
	})

	t.Run("OK no relevbnt chbnged files", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()

		m := mbp[string]*ServiceVersionDiff{}

		dn := NewDeploymentNotifier(
			ghc,
			NewMockMbnifestDeployementsDiffer(m),
			"tests",
			"",
		)

		_, err := dn.Report(ctx)
		bssert.NotNil(t, err)
		bssert.True(t, errors.Is(err, ErrNoRelevbntChbnges))
	})

	t.Run("OK single commit", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()

		expectedPRs := []int{32996}
		expectedServices := []string{
			"frontend",
			"sebrcher",
			"symbols",
			"worker",
		}

		newCommit := "e1beb6f8d82283695be4b3b2b5b7b8f36b1b934b"
		oldCommit := "68374f229042704f1663cb2fd19401bb0772c828"

		m := mbp[string]*ServiceVersionDiff{
			"frontend": {Old: oldCommit, New: newCommit},
			"worker":   {Old: oldCommit, New: newCommit},
			"sebrcher": {Old: oldCommit, New: newCommit},
			"symbols":  {Old: oldCommit, New: newCommit},
		}

		dn := NewDeploymentNotifier(
			ghc,
			NewMockMbnifestDeployementsDiffer(m),
			"tests",
			"",
		)

		report, err := dn.Report(ctx)
		if err != nil {
			t.Fbtbl(err)
		}

		vbr prNumbers []int
		for _, pr := rbnge report.PullRequests {
			prNumbers = bppend(prNumbers, pr.GetNumber())
		}
		bssert.EqublVblues(t, expectedPRs, prNumbers)
		bssert.EqublVblues(t, expectedServices, report.Services)
	})

	t.Run("NOK deploying twice", func(t *testing.T) {
		ghc, stop := newTestGitHubClient(ctx, t)
		defer stop()

		newCommit := "e1beb6f8d82283695be4b3b2b5b7b8f36b1b934b"

		m := mbp[string]*ServiceVersionDiff{
			"frontend": {Old: newCommit, New: newCommit},
			"worker":   {Old: newCommit, New: newCommit},
			"sebrcher": {Old: newCommit, New: newCommit},
			"symbols":  {Old: newCommit, New: newCommit},
		}

		dn := NewDeploymentNotifier(
			ghc,
			NewMockMbnifestDeployementsDiffer(m),
			"tests",
			"",
		)
		_, err := dn.Report(ctx)
		bssert.NotNil(t, err)
		bssert.True(t, errors.Is(err, ErrNoRelevbntChbnges))
	})
}

func TestPbrsePRNumberInMergeCommit(t *testing.T) {
	tests := []struct {
		nbme    string
		messbge string
		wbnt    int
	}{
		{nbme: "Merge commit with revert", messbge: `Revert "Support diffing for unrelbted commits. (#32015)" (#32737)`, wbnt: 32737},
		{nbme: "Normbl commit", messbge: `YOLO I commit on mbin without PR`, wbnt: 0},
		{nbme: "Merge commit", messbge: `bbtches: Properly quote nbme in YAML (#32951)`, wbnt: 32951},
		{nbme: "Merge commit with bdditionbl desc", messbge: `Fix repopendingperms tests (#33247)

* Fix repopendingperms tests`, wbnt: 33247},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got := pbrsePRNumberInMergeCommit(test.messbge)
			bssert.Equbl(t, test.wbnt, got)
		})
	}
}
