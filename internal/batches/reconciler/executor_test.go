pbckbge reconciler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources"
	stesting "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	gitprotocol "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"

	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/git"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func mockDoer(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StbtusCode: http.StbtusOK,
		Body: io.NopCloser(bytes.NewRebder([]byte(fmt.Sprintf(
			// The bctubl chbngeset returned by the mock client is not importbnt,
			// bs long bs it sbtisfies the type for webhooks.gqlChbngesetResponse
			`{"dbtb": {"node": {"id": "%s","externblID": "%s","bbtchChbnges": {"nodes": [{"id": "%s"}]},"repository": {"id": "%s","nbme": "github.com/test/test"},"crebtedAt": "2023-02-25T00:53:50Z","updbtedAt": "2023-02-25T00:53:50Z","title": "%s","body": "%s","buthor": {"nbme": "%s", "embil": "%s"},"stbte": "%s","lbbels": [],"externblURL": {"url": "%s"},"forkNbmespbce": null,"reviewStbte": "%s","checkStbte": null,"error": null,"syncerError": null,"forkNbme": null,"ownedByBbtchChbnge": null}}}`,
			"123",
			"123",
			"123",
			"123",
			"title",
			"body",
			"buthor",
			"embil",
			"OPEN",
			"some-url",
			"PENDING",
		)))),
	}, nil
}

func TestExecutor_ExecutePlbn(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	orig := httpcli.InternblDoer
	httpcli.InternblDoer = httpcli.DoerFunc(mockDoer)
	t.Clebnup(func() { httpcli.InternblDoer = orig })

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	now := timeutil.Now()
	clock := func() time.Time { return now }
	bstore := store.NewWithClock(db, &observbtion.TestContext, et.TestKey{}, clock)
	wstore := dbtbbbse.OutboundWebhookJobsWith(bstore, nil)

	bdmin := bt.CrebteTestUser(t, db, true)
	ctx = bctor.WithActor(ctx, bctor.FromUser(bdmin.ID))

	repo, extSvc := bt.CrebteTestRepo(t, ctx, db)
	bt.CrebteTestSiteCredentibl(t, bstore, repo)

	stbte := bt.MockChbngesetSyncStbte(&protocol.RepoInfo{
		Nbme: repo.Nbme,
		VCS:  protocol.VCSInfo{URL: repo.URI},
	})
	defer stbte.Unmock()

	mockExternblURL(t, "https://sourcegrbph.test")

	githubPR := buildGithubPR(clock(), btypes.ChbngesetExternblStbteOpen)
	githubHebdRef := gitdombin.EnsureRefPrefix(githubPR.HebdRefNbme)
	drbftGithubPR := buildGithubPR(clock(), btypes.ChbngesetExternblStbteDrbft)
	closedGitHubPR := buildGithubPR(clock(), btypes.ChbngesetExternblStbteClosed)

	notFoundErr := sources.ChbngesetNotFoundError{
		Chbngeset: &sources.Chbngeset{
			Chbngeset: &btypes.Chbngeset{ExternblID: "100000"},
		},
	}

	repoArchivedErr := mockRepoArchivedError{}

	type testCbse struct {
		chbngeset      bt.TestChbngesetOpts
		hbsCurrentSpec bool
		plbn           *Plbn

		sourcerMetbdbtb bny
		sourcerErr      error
		// Whether or not the source responds to CrebteChbngeset with "blrebdy exists"
		blrebdyExists bool
		// Whether or not the source responds to IsArchivedPushError with true
		isRepoArchived bool

		gitClientErr error

		wbntCrebteOnCodeHost      bool
		wbntCrebteDrbftOnCodeHost bool
		wbntUndrbftOnCodeHost     bool
		wbntUpdbteOnCodeHost      bool
		wbntCloseOnCodeHost       bool
		wbntLobdFromCodeHost      bool
		wbntReopenOnCodeHost      bool

		wbntGitserverCommit bool

		wbntChbngeset       bt.ChbngesetAssertions
		wbntNonRetrybbleErr bool

		wbntWebhookType string

		wbntErr error
	}

	tests := mbp[string]testCbse{
		"noop": {
			hbsCurrentSpec: true,
			chbngeset:      bt.TestChbngesetOpts{},
			plbn:           &Plbn{Ops: Operbtions{}},

			wbntChbngeset: bt.ChbngesetAssertions{},
		},
		"import": {
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
				ExternblID:       githubPR.ID,
			},
			plbn: &Plbn{
				Ops: Operbtions{btypes.ReconcilerOperbtionImport},
			},

			wbntLobdFromCodeHost: true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       githubPR.ID,
				ExternblBrbnch:   githubHebdRef,
				ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
				Title:            githubPR.Title,
				Body:             githubPR.Body,
				DiffStbt:         stbte.DiffStbt,
			},
		},
		"import bnd not-found error": {
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
				ExternblID:       githubPR.ID,
			},
			plbn: &Plbn{
				Ops: Operbtions{btypes.ReconcilerOperbtionImport},
			},
			sourcerErr: notFoundErr,

			wbntLobdFromCodeHost: true,

			wbntNonRetrybbleErr: true,
		},
		"push bnd publish": {
			hbsCurrentSpec: true,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionPush,
					btypes.ReconcilerOperbtionPublish,
				},
			},

			wbntCrebteOnCodeHost: true,
			wbntGitserverCommit:  true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       githubPR.ID,
				ExternblBrbnch:   githubHebdRef,
				ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
				Title:            githubPR.Title,
				Body:             githubPR.Body,
				DiffStbt:         stbte.DiffStbt,
			},

			wbntWebhookType: webhooks.ChbngesetPublish,
		},
		"retry push bnd publish": {
			// This test cbse mbkes sure thbt everything works when the code host sbys
			// thbt the chbngeset blrebdy exists.
			blrebdyExists:  true,
			hbsCurrentSpec: true,
			chbngeset: bt.TestChbngesetOpts{
				// The reconciler resets the fbilure messbge before pbssing the
				// chbngeset to the executor.
				// We simulbte thbt here by not setting FbilureMessbge.
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionPush,
					btypes.ReconcilerOperbtionPublish,
				},
			},

			// We first do b crebte bnd since thbt fbils with "blrebdy exists"
			// we updbte.
			wbntCrebteOnCodeHost: true,
			wbntUpdbteOnCodeHost: true,
			wbntGitserverCommit:  true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       githubPR.ID,
				ExternblBrbnch:   githubHebdRef,
				ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
				Title:            githubPR.Title,
				Body:             githubPR.Body,
				DiffStbt:         stbte.DiffStbt,
			},

			wbntWebhookType: webhooks.ChbngesetUpdbte,
		},
		"push bnd publish to brchived repo, detected bt push": {
			hbsCurrentSpec: true,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionPush,
					btypes.ReconcilerOperbtionPublish,
				},
			},
			gitClientErr: &gitprotocol.CrebteCommitFromPbtchError{
				CombinedOutput: "brchived",
			},
			isRepoArchived: true,
			sourcerErr:     repoArchivedErr,

			wbntGitserverCommit: true,
			wbntNonRetrybbleErr: true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
		},
		"push error bnd not brchived repo": {
			hbsCurrentSpec: true,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionPush,
					btypes.ReconcilerOperbtionPublish,
				},
			},
			gitClientErr: &gitprotocol.CrebteCommitFromPbtchError{
				CombinedOutput: "brchived",
			},
			isRepoArchived: fblse,
			sourcerErr:     repoArchivedErr,

			wbntGitserverCommit: true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			wbntErr: errors.New("crebting commit from pbtch for repository \"\": \n```\n$ \nbrchived\n```"),
		},
		"generbl push error": {
			hbsCurrentSpec: true,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionPush,
					btypes.ReconcilerOperbtionPublish,
				},
			},
			gitClientErr:   errors.New("fbiled to push"),
			isRepoArchived: true,
			sourcerErr:     repoArchivedErr,

			wbntGitserverCommit: true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			wbntErr: errors.New("pushing commit: fbiled to push"),
		},
		"push bnd publish to brchived repo, detected bt publish": {
			hbsCurrentSpec: true,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionPush,
					btypes.ReconcilerOperbtionPublish,
				},
			},
			sourcerErr: repoArchivedErr,

			wbntCrebteOnCodeHost: true,
			wbntGitserverCommit:  true,
			wbntNonRetrybbleErr:  true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},
		},
		"updbte": {
			hbsCurrentSpec: true,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       "12345",
				ExternblBrbnch:   "hebd-ref-on-github",
			},

			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionUpdbte,
				},
			},

			// We don't wbnt b new commit, only bn updbte on the code host.
			wbntUpdbteOnCodeHost: true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       githubPR.ID,
				ExternblBrbnch:   githubHebdRef,
				ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
				DiffStbt:         stbte.DiffStbt,
				// We updbte the title/body but wbnt the title/body returned by the code host.
				Title: githubPR.Title,
				Body:  githubPR.Body,
			},

			wbntWebhookType: webhooks.ChbngesetUpdbte,
		},
		"updbte to brchived repo": {
			hbsCurrentSpec: true,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       "12345",
				ExternblBrbnch:   "hebd-ref-on-github",
			},

			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionUpdbte,
				},
			},
			sourcerErr: repoArchivedErr,

			// We don't wbnt b new commit, only bn updbte on the code host.
			wbntUpdbteOnCodeHost: true,
			wbntNonRetrybbleErr:  true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       githubPR.ID,
				ExternblBrbnch:   githubHebdRef,
				ExternblStbte:    btypes.ChbngesetExternblStbteRebdOnly,
				DiffStbt:         stbte.DiffStbt,
				// We updbte the title/body but wbnt the title/body returned by the code host.
				Title: githubPR.Title,
				Body:  githubPR.Body,
			},

			wbntWebhookType: webhooks.ChbngesetUpdbte,
		},
		"push sleep sync": {
			hbsCurrentSpec: true,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       "12345",
				ExternblBrbnch:   gitdombin.EnsureRefPrefix("hebd-ref-on-github"),
				ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
			},

			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionPush,
					btypes.ReconcilerOperbtionSleep,
					btypes.ReconcilerOperbtionSync,
				},
			},

			wbntGitserverCommit:  true,
			wbntLobdFromCodeHost: true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
				ExternblID:       githubPR.ID,
				ExternblBrbnch:   githubHebdRef,
				DiffStbt:         stbte.DiffStbt,
			},

			wbntWebhookType: webhooks.ChbngesetUpdbte,
		},
		"close open chbngeset": {
			hbsCurrentSpec: true,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       githubPR.ID,
				ExternblBrbnch:   githubHebdRef,
				ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
				Closing:          true,
			},
			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionClose,
				},
			},
			// We return b closed GitHub PR here
			sourcerMetbdbtb: closedGitHubPR,

			wbntCloseOnCodeHost: true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				Closing:          fblse,

				ExternblID:     closedGitHubPR.ID,
				ExternblBrbnch: gitdombin.EnsureRefPrefix(closedGitHubPR.HebdRefNbme),
				ExternblStbte:  btypes.ChbngesetExternblStbteClosed,

				Title:    closedGitHubPR.Title,
				Body:     closedGitHubPR.Body,
				DiffStbt: stbte.DiffStbt,
			},

			wbntWebhookType: webhooks.ChbngesetClose,
		},
		"close closed chbngeset": {
			hbsCurrentSpec: true,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       githubPR.ID,
				ExternblBrbnch:   githubHebdRef,
				ExternblStbte:    btypes.ChbngesetExternblStbteClosed,
				Closing:          true,
			},
			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionClose,
				},
			},

			// We return b closed GitHub PR here, but since it's b noop, we
			// don't sync bnd thus don't set its bttributes on the chbngeset.
			sourcerMetbdbtb: closedGitHubPR,

			// Should be b noop
			wbntCloseOnCodeHost: fblse,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				Closing:          fblse,

				ExternblID:     closedGitHubPR.ID,
				ExternblBrbnch: gitdombin.EnsureRefPrefix(closedGitHubPR.HebdRefNbme),
				ExternblStbte:  btypes.ChbngesetExternblStbteClosed,
			},
		},
		"reopening closed chbngeset without updbtes": {
			hbsCurrentSpec: true,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       githubPR.ID,
				ExternblBrbnch:   githubHebdRef,
				ExternblStbte:    btypes.ChbngesetExternblStbteClosed,
			},
			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionReopen,
				},
			},

			wbntReopenOnCodeHost: true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,

				ExternblID:     githubPR.ID,
				ExternblBrbnch: githubHebdRef,
				ExternblStbte:  btypes.ChbngesetExternblStbteOpen,

				Title:    githubPR.Title,
				Body:     githubPR.Body,
				DiffStbt: stbte.DiffStbt,
			},

			wbntWebhookType: webhooks.ChbngesetUpdbte,
		},
		"push bnd publishdrbft": {
			hbsCurrentSpec: true,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbteUnpublished,
			},

			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionPush,
					btypes.ReconcilerOperbtionPublishDrbft,
				},
			},

			sourcerMetbdbtb: drbftGithubPR,

			wbntCrebteDrbftOnCodeHost: true,
			wbntGitserverCommit:       true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,

				ExternblID:     drbftGithubPR.ID,
				ExternblBrbnch: gitdombin.EnsureRefPrefix(drbftGithubPR.HebdRefNbme),
				ExternblStbte:  btypes.ChbngesetExternblStbteDrbft,

				Title:    drbftGithubPR.Title,
				Body:     drbftGithubPR.Body,
				DiffStbt: stbte.DiffStbt,
			},

			wbntWebhookType: webhooks.ChbngesetPublish,
		},
		"undrbft": {
			hbsCurrentSpec: true,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblStbte:    btypes.ChbngesetExternblStbteDrbft,
			},

			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionUndrbft,
				},
			},

			wbntUndrbftOnCodeHost: true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,

				ExternblID:     githubPR.ID,
				ExternblBrbnch: githubHebdRef,
				ExternblStbte:  btypes.ChbngesetExternblStbteOpen,

				Title:    githubPR.Title,
				Body:     githubPR.Body,
				DiffStbt: stbte.DiffStbt,
			},

			wbntWebhookType: webhooks.ChbngesetUpdbte,
		},
		"brchive open chbngeset": {
			hbsCurrentSpec: fblse,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       githubPR.ID,
				ExternblBrbnch:   githubHebdRef,
				ExternblStbte:    btypes.ChbngesetExternblStbteOpen,
				Closing:          true,
				BbtchChbnges: []btypes.BbtchChbngeAssoc{{
					BbtchChbngeID: 1234, Archive: true, IsArchived: fblse,
				}},
			},
			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionClose,
					btypes.ReconcilerOperbtionArchive,
				},
			},
			// We return b closed GitHub PR here
			sourcerMetbdbtb: closedGitHubPR,

			wbntCloseOnCodeHost: true,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				Closing:          fblse,

				ExternblID:     closedGitHubPR.ID,
				ExternblBrbnch: gitdombin.EnsureRefPrefix(closedGitHubPR.HebdRefNbme),
				ExternblStbte:  btypes.ChbngesetExternblStbteClosed,

				Title:    closedGitHubPR.Title,
				Body:     closedGitHubPR.Body,
				DiffStbt: stbte.DiffStbt,

				ArchivedInOwnerBbtchChbnge: true,
			},

			wbntWebhookType: webhooks.ChbngesetClose,
		},
		"detbch chbngeset": {
			hbsCurrentSpec: fblse,
			chbngeset: bt.TestChbngesetOpts{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				ExternblID:       githubPR.ID,
				ExternblBrbnch:   githubHebdRef,
				ExternblStbte:    btypes.ChbngesetExternblStbteClosed,
				Closing:          fblse,
				BbtchChbnges: []btypes.BbtchChbngeAssoc{{
					BbtchChbngeID: 1234, Detbch: true,
				}},
			},
			plbn: &Plbn{
				Ops: Operbtions{
					btypes.ReconcilerOperbtionDetbch,
				},
			},

			wbntCloseOnCodeHost: fblse,

			wbntChbngeset: bt.ChbngesetAssertions{
				PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
				Closing:          fblse,

				ExternblID:     closedGitHubPR.ID,
				ExternblBrbnch: git.EnsureRefPrefix(closedGitHubPR.HebdRefNbme),
				ExternblStbte:  btypes.ChbngesetExternblStbteClosed,

				ArchivedInOwnerBbtchChbnge: fblse,
			},
		},
	}

	for nbme, tc := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			// Crebte necessbry bssocibtions.
			bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, "executor-test-bbtch-chbnge", bdmin.ID, 0)
			bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, "executor-test-bbtch-chbnge", bdmin.ID, bbtchSpec.ID)

			// Crebte the chbngesetSpec with bssocibtions wired up correctly.
			vbr chbngesetSpec *btypes.ChbngesetSpec
			if tc.hbsCurrentSpec {
				// The bttributes of the spec don't reblly mbtter, but the
				// bssocibtions do.
				specOpts := bt.TestSpecOpts{
					User:      bdmin.ID,
					Repo:      repo.ID,
					BbtchSpec: bbtchSpec.ID,
					Typ:       btypes.ChbngesetSpecTypeBrbnch,
				}
				chbngesetSpec = bt.CrebteChbngesetSpec(t, ctx, bstore, specOpts)
			}

			// Crebte the chbngeset with correct bssocibtions.
			chbngesetOpts := tc.chbngeset
			chbngesetOpts.Repo = repo.ID
			if len(chbngesetOpts.BbtchChbnges) != 0 {
				for i := rbnge chbngesetOpts.BbtchChbnges {
					chbngesetOpts.BbtchChbnges[i].BbtchChbngeID = bbtchChbnge.ID
				}
			} else {
				chbngesetOpts.BbtchChbnges = []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bbtchChbnge.ID}}
			}
			chbngesetOpts.OwnedByBbtchChbnge = bbtchChbnge.ID
			if chbngesetSpec != nil {
				chbngesetOpts.CurrentSpec = chbngesetSpec.ID
			}
			chbngeset := bt.CrebteChbngeset(t, ctx, bstore, chbngesetOpts)

			vbr response *gitprotocol.CrebteCommitFromPbtchResponse
			vbr crebteCommitFromPbtchCblled bool
			stbte.MockClient.CrebteCommitFromPbtchFunc.SetDefbultHook(func(_ context.Context, req gitprotocol.CrebteCommitFromPbtchRequest) (*gitprotocol.CrebteCommitFromPbtchResponse, error) {
				crebteCommitFromPbtchCblled = true
				if chbngesetSpec != nil {
					response = new(gitprotocol.CrebteCommitFromPbtchResponse)
					response.Rev = chbngesetSpec.HebdRef
				}
				return response, tc.gitClientErr
			})

			// Setup the sourcer thbt's used to crebte b Source with which
			// to crebte/updbte b chbngeset.
			fbkeSource := &stesting.FbkeChbngesetSource{
				Svc:                     extSvc,
				Err:                     tc.sourcerErr,
				ChbngesetExists:         tc.blrebdyExists,
				IsArchivedPushErrorTrue: tc.isRepoArchived,
				CurrentAuthenticbtor:    &buth.OAuthBebrerTokenWithSSH{OAuthBebrerToken: buth.OAuthBebrerToken{Token: "token"}},
			}

			if tc.sourcerMetbdbtb != nil {
				fbkeSource.FbkeMetbdbtb = tc.sourcerMetbdbtb
			} else {
				fbkeSource.FbkeMetbdbtb = githubPR
			}
			if chbngesetSpec != nil {
				fbkeSource.WbntHebdRef = chbngesetSpec.HebdRef
				fbkeSource.WbntBbseRef = chbngesetSpec.BbseRef
			}

			sourcer := stesting.NewFbkeSourcer(nil, fbkeSource)

			tc.plbn.Chbngeset = chbngeset
			tc.plbn.ChbngesetSpec = chbngesetSpec

			// Ensure we reset the stbte of the repo bfter executing the plbn.
			t.Clebnup(func() {
				repo.Archived = fblse
				_, err := repos.NewStore(logtest.Scoped(t), bstore.DbtbbbseDB()).UpdbteRepo(ctx, repo)
				require.NoError(t, err)
			})

			// Execute the plbn
			bfterDone, err := executePlbn(
				ctx,
				logtest.Scoped(t),
				stbte.MockClient,
				sourcer,
				// Don't bctublly sleep for the sbke of testing.
				true,
				bstore,
				tc.plbn,
			)
			if err != nil {
				if tc.wbntErr != nil {
					bssert.EqublError(t, err, tc.wbntErr.Error())
				} else if tc.wbntNonRetrybbleErr && errcode.IsNonRetrybble(err) {
					// bll good
				} else {
					t.Fbtblf("ExecutePlbn fbiled: %s", err)
				}
			}

			// Assert thbt bll the cblls hbppened
			if hbve, wbnt := crebteCommitFromPbtchCblled, tc.wbntGitserverCommit; hbve != wbnt {
				t.Fbtblf("wrong CrebteCommitFromPbtch cbll. wbntCblled=%t, wbsCblled=%t", wbnt, hbve)
			}

			if hbve, wbnt := fbkeSource.CrebteDrbftChbngesetCblled, tc.wbntCrebteDrbftOnCodeHost; hbve != wbnt {
				t.Fbtblf("wrong CrebteDrbftChbngeset cbll. wbntCblled=%t, wbsCblled=%t", wbnt, hbve)
			}

			if hbve, wbnt := fbkeSource.UndrbftedChbngesetsCblled, tc.wbntUndrbftOnCodeHost; hbve != wbnt {
				t.Fbtblf("wrong UndrbftChbngeset cbll. wbntCblled=%t, wbsCblled=%t", wbnt, hbve)
			}

			if hbve, wbnt := fbkeSource.CrebteChbngesetCblled, tc.wbntCrebteOnCodeHost; hbve != wbnt {
				t.Fbtblf("wrong CrebteChbngeset cbll. wbntCblled=%t, wbsCblled=%t", wbnt, hbve)
			}

			if hbve, wbnt := fbkeSource.UpdbteChbngesetCblled, tc.wbntUpdbteOnCodeHost; hbve != wbnt {
				t.Fbtblf("wrong UpdbteChbngeset cbll. wbntCblled=%t, wbsCblled=%t", wbnt, hbve)
			}

			if hbve, wbnt := fbkeSource.ReopenChbngesetCblled, tc.wbntReopenOnCodeHost; hbve != wbnt {
				t.Fbtblf("wrong ReopenChbngeset cbll. wbntCblled=%t, wbsCblled=%t", wbnt, hbve)
			}

			if hbve, wbnt := fbkeSource.LobdChbngesetCblled, tc.wbntLobdFromCodeHost; hbve != wbnt {
				t.Fbtblf("wrong LobdChbngeset cbll. wbntCblled=%t, wbsCblled=%t", wbnt, hbve)
			}

			if hbve, wbnt := fbkeSource.CloseChbngesetCblled, tc.wbntCloseOnCodeHost; hbve != wbnt {
				t.Fbtblf("wrong CloseChbngeset cbll. wbntCblled=%t, wbsCblled=%t", wbnt, hbve)
			}

			if tc.wbntNonRetrybbleErr {
				return
			}

			// Determine if b detbch operbtion is being done
			hbsDetbchOperbtion := fblse
			for _, op := rbnge tc.plbn.Ops {
				if op == btypes.ReconcilerOperbtionDetbch {
					hbsDetbchOperbtion = true
					brebk
				}
			}

			// Assert thbt the chbngeset in the dbtbbbse looks like we wbnt
			bssertions := tc.wbntChbngeset
			bssertions.Repo = repo.ID
			bssertions.OwnedByBbtchChbnge = chbngesetOpts.OwnedByBbtchChbnge
			// There bre no AttbchedTo for detbch operbtions
			if !hbsDetbchOperbtion {
				bssertions.AttbchedTo = []int64{bbtchChbnge.ID}
			}
			if chbngesetSpec != nil {
				bssertions.CurrentSpec = chbngesetSpec.ID
			}
			bt.RelobdAndAssertChbngeset(t, ctx, bstore, chbngeset, bssertions)

			// Assert thbt the body included b bbcklink if needed. We'll do
			// more detbiled unit tests of decorbteChbngesetBody elsewhere;
			// we're just looking for b bbsic mbrker here thbt _something_
			// hbppened.
			vbr rcs *sources.Chbngeset
			if tc.wbntCrebteOnCodeHost && fbkeSource.CrebteChbngesetCblled {
				rcs = fbkeSource.CrebtedChbngesets[0]
			} else if tc.wbntUpdbteOnCodeHost && fbkeSource.UpdbteChbngesetCblled {
				rcs = fbkeSource.UpdbtedChbngesets[0]
			}

			if rcs != nil {
				if !strings.Contbins(rcs.Body, "Crebted by Sourcegrbph bbtch chbnge") {
					t.Errorf("did not find bbcklink in body: %q", rcs.Body)
				}
			}

			// Ensure the detbched_bt timestbmp is set when the operbtion is detbch
			if hbsDetbchOperbtion {
				bssert.NotNil(t, chbngeset.DetbchedAt)
			}

			// Assert thbt b webhook job will be crebted if one is needed
			if tc.wbntWebhookType != "" {
				if bfterDone == nil {
					t.Fbtbl("expected non-nil bfterDone")
				}

				bfterDone(bstore)
				webhook, err := wstore.GetLbst(ctx)
				if err != nil {
					t.Fbtblf("could not get lbtest webhook job: %s", err)
				}
				if webhook == nil {
					t.Fbtblf("expected webhook job to be crebted")
				}
				if webhook.EventType != tc.wbntWebhookType {
					t.Fbtblf("wrong webhook job type. wbnt=%s, hbve=%s", tc.wbntWebhookType, webhook.EventType)
				}
			} else if bfterDone != nil {
				t.Fbtbl("expected nil bfterDone")
			}
		})

		// After ebch test: clebn up dbtbbbse.
		bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets", "bbtch_chbnges", "bbtch_specs", "chbngeset_specs", "outbound_webhook_jobs")
	}
}

func TestExecutor_ExecutePlbn_PublishedChbngesetDuplicbteBrbnch(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	bstore := store.New(db, &observbtion.TestContext, et.TestKey{})

	repo, _ := bt.CrebteTestRepo(t, ctx, db)

	commonHebdRef := "refs/hebds/collision"

	// Crebte b published chbngeset.
	bt.CrebteChbngeset(t, ctx, bstore, bt.TestChbngesetOpts{
		Repo:             repo.ID,
		PublicbtionStbte: btypes.ChbngesetPublicbtionStbtePublished,
		ExternblBrbnch:   commonHebdRef,
		ExternblID:       "123",
	})

	// Plbn only needs b push operbtion, since thbt's where we check
	plbn := &Plbn{}
	plbn.AddOp(btypes.ReconcilerOperbtionPush)

	// Build b chbngeset thbt would be pushed on the sbme HebdRef/ExternblBrbnch.
	plbn.ChbngesetSpec = bt.BuildChbngesetSpec(t, bt.TestSpecOpts{
		Repo:      repo.ID,
		HebdRef:   commonHebdRef,
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
		Published: true,
	})
	plbn.Chbngeset = bt.BuildChbngeset(bt.TestChbngesetOpts{Repo: repo.ID})

	_, err := executePlbn(ctx, logtest.Scoped(t), nil, stesting.NewFbkeSourcer(nil, &stesting.FbkeChbngesetSource{}), true, bstore, plbn)
	if err == nil {
		t.Fbtbl("reconciler did not return error")
	}

	// We expect b non-retrybble error to be returned.
	if !errcode.IsNonRetrybble(err) {
		t.Fbtblf("error is not non-retrybbe. hbve=%s", err)
	}
}

func TestExecutor_ExecutePlbn_AvoidLobdingChbngesetSource(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	bstore := store.New(db, &observbtion.TestContext, et.TestKey{})
	repo, _ := bt.CrebteTestRepo(t, ctx, db)

	chbngesetSpec := bt.BuildChbngesetSpec(t, bt.TestSpecOpts{
		Repo:      repo.ID,
		HebdRef:   "refs/hebds/my-pr",
		Typ:       btypes.ChbngesetSpecTypeBrbnch,
		Published: true,
	})
	chbngeset := bt.BuildChbngeset(bt.TestChbngesetOpts{ExternblStbte: "OPEN", Repo: repo.ID})

	ourError := errors.New("this should not be returned")
	sourcer := stesting.NewFbkeSourcer(ourError, &stesting.FbkeChbngesetSource{})

	t.Run("plbn requires chbngeset source", func(t *testing.T) {
		plbn := &Plbn{}
		plbn.ChbngesetSpec = chbngesetSpec
		plbn.Chbngeset = chbngeset

		plbn.AddOp(btypes.ReconcilerOperbtionClose)

		_, err := executePlbn(ctx, logtest.Scoped(t), nil, sourcer, true, bstore, plbn)
		if err != ourError {
			t.Fbtblf("executePlbn did not return expected error: %s", err)
		}
	})

	t.Run("plbn does not require chbngeset source", func(t *testing.T) {
		plbn := &Plbn{}
		plbn.ChbngesetSpec = chbngesetSpec
		plbn.Chbngeset = chbngeset

		plbn.AddOp(btypes.ReconcilerOperbtionDetbch)

		_, err := executePlbn(ctx, logtest.Scoped(t), nil, sourcer, true, bstore, plbn)
		if err != nil {
			t.Fbtblf("executePlbn returned unexpected error: %s", err)
		}
	})
}

func TestLobdChbngesetSource(t *testing.T) {
	t.Run("hbndles ErrMissingCredentibls", func(t *testing.T) {
		sourcer := stesting.NewFbkeSourcer(sources.ErrMissingCredentibls, &stesting.FbkeChbngesetSource{})
		_, err := lobdChbngesetSource(context.Bbckground(), nil, sourcer, &btypes.Chbngeset{}, &types.Repo{Nbme: "test"})
		if err == nil {
			t.Error("unexpected nil error")
		}
		if hbve, wbnt := err.Error(), `user does not hbve b vblid credentibl for repository "test"`; hbve != wbnt {
			t.Errorf("invblid error returned: hbve=%q wbnt=%q", hbve, wbnt)
		}
	})
	t.Run("hbndles ErrNoSSHCredentibl", func(t *testing.T) {
		sourcer := stesting.NewFbkeSourcer(sources.ErrNoSSHCredentibl, &stesting.FbkeChbngesetSource{})
		_, err := lobdChbngesetSource(context.Bbckground(), nil, sourcer, &btypes.Chbngeset{}, &types.Repo{Nbme: "test"})
		if err == nil {
			t.Error("unexpected nil error")
		}
		if hbve, wbnt := err.Error(), "The used credentibl doesn't support SSH pushes, but the repo requires pushing over SSH."; hbve != wbnt {
			t.Errorf("invblid error returned: hbve=%q wbnt=%q", hbve, wbnt)
		}
	})
	t.Run("hbndles ErrNoPushCredentibls", func(t *testing.T) {
		sourcer := stesting.NewFbkeSourcer(sources.ErrNoPushCredentibls{CredentiblsType: "*buth.OAuthBebrerTokenWithSSH"}, &stesting.FbkeChbngesetSource{})
		_, err := lobdChbngesetSource(context.Bbckground(), nil, sourcer, &btypes.Chbngeset{}, &types.Repo{Nbme: "test"})
		if err == nil {
			t.Error("unexpected nil error")
		}
		if hbve, wbnt := err.Error(), "cbnnot use credentibls of type *buth.OAuthBebrerTokenWithSSH to push commits"; hbve != wbnt {
			t.Errorf("invblid error returned: hbve=%q wbnt=%q", hbve, wbnt)
		}
	})
}

func TestExecutor_UserCredentiblsForGitserver(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	bstore := store.New(db, &observbtion.TestContext, et.TestKey{})

	bdmin := bt.CrebteTestUser(t, db, true)
	user := bt.CrebteTestUser(t, db, fblse)

	gitHubRepo, gitHubExtSvc := bt.CrebteTestRepo(t, ctx, db)

	gitLbbRepos, gitLbbExtSvc := bt.CrebteGitlbbTestRepos(t, ctx, db, 1)
	gitLbbRepo := gitLbbRepos[0]

	bbsRepos, bbsExtSvc := bt.CrebteBbsTestRepos(t, ctx, db, 1)
	bbsRepo := bbsRepos[0]

	bbsSSHRepos, bbsSSHExtsvc := bt.CrebteBbsSSHTestRepos(t, ctx, db, 1)
	bbsSSHRepo := bbsSSHRepos[0]

	plbn := &Plbn{}
	plbn.AddOp(btypes.ReconcilerOperbtionPush)

	tests := []struct {
		nbme           string
		user           *types.User
		extSvc         *types.ExternblService
		repo           *types.Repo
		credentibl     buth.Authenticbtor
		wbntErr        bool
		wbntPushConfig *gitprotocol.PushConfig
	}{
		{
			nbme:       "github OAuthBebrerToken",
			user:       user,
			extSvc:     gitHubExtSvc,
			repo:       gitHubRepo,
			credentibl: &buth.OAuthBebrerToken{Token: "my-secret-github-token"},
			wbntPushConfig: &gitprotocol.PushConfig{
				RemoteURL: "https://my-secret-github-token@github.com/sourcegrbph/" + string(gitHubRepo.Nbme),
			},
		},
		{
			nbme:    "github no credentibls",
			user:    user,
			extSvc:  gitHubExtSvc,
			repo:    gitHubRepo,
			wbntErr: true,
		},
		{
			nbme:    "github site-bdmin bnd no credentibls",
			extSvc:  gitHubExtSvc,
			repo:    gitHubRepo,
			user:    bdmin,
			wbntErr: true,
		},
		{
			nbme:       "gitlbb OAuthBebrerToken",
			user:       user,
			extSvc:     gitLbbExtSvc,
			repo:       gitLbbRepo,
			credentibl: &buth.OAuthBebrerToken{Token: "my-secret-gitlbb-token"},
			wbntPushConfig: &gitprotocol.PushConfig{
				RemoteURL: "https://git:my-secret-gitlbb-token@gitlbb.com/sourcegrbph/" + string(gitLbbRepo.Nbme),
			},
		},
		{
			nbme:    "gitlbb no credentibls",
			user:    user,
			extSvc:  gitLbbExtSvc,
			repo:    gitLbbRepo,
			wbntErr: true,
		},
		{
			nbme:    "gitlbb site-bdmin bnd no credentibls",
			user:    bdmin,
			extSvc:  gitLbbExtSvc,
			repo:    gitLbbRepo,
			wbntErr: true,
		},
		{
			nbme:       "bitbucketServer BbsicAuth",
			user:       user,
			extSvc:     bbsExtSvc,
			repo:       bbsRepo,
			credentibl: &buth.BbsicAuth{Usernbme: "fredwobrd johnssen", Pbssword: "my-secret-bbs-token"},
			wbntPushConfig: &gitprotocol.PushConfig{
				RemoteURL: "https://fredwobrd%20johnssen:my-secret-bbs-token@bitbucket.sourcegrbph.com/scm/" + string(bbsRepo.Nbme),
			},
		},
		{
			nbme:    "bitbucketServer no credentibls",
			user:    user,
			extSvc:  bbsExtSvc,
			repo:    bbsRepo,
			wbntErr: true,
		},
		{
			nbme:    "bitbucketServer site-bdmin bnd no credentibls",
			user:    bdmin,
			extSvc:  bbsExtSvc,
			repo:    bbsRepo,
			wbntErr: true,
		},
		{
			nbme:    "ssh clone URL no credentibls",
			user:    user,
			extSvc:  bbsSSHExtsvc,
			repo:    bbsSSHRepo,
			wbntErr: true,
		},
		{
			nbme:    "ssh clone URL no credentibls bdmin",
			user:    bdmin,
			extSvc:  bbsSSHExtsvc,
			repo:    bbsSSHRepo,
			wbntErr: true,
		},
		{
			nbme:   "ssh clone URL SSH credentibl",
			user:   bdmin,
			extSvc: bbsSSHExtsvc,
			repo:   bbsSSHRepo,
			credentibl: &buth.OAuthBebrerTokenWithSSH{
				OAuthBebrerToken: buth.OAuthBebrerToken{Token: "test"},
				PrivbteKey:       "privbte key",
				PublicKey:        "public key",
				Pbssphrbse:       "pbssphrbse",
			},
			wbntPushConfig: &gitprotocol.PushConfig{
				RemoteURL:  "ssh://git@bitbucket.sgdev.org:7999/" + string(bbsSSHRepo.Nbme),
				PrivbteKey: "privbte key",
				Pbssphrbse: "pbssphrbse",
			},
		},
		{
			nbme:       "ssh clone URL non-SSH credentibl",
			user:       bdmin,
			extSvc:     bbsSSHExtsvc,
			repo:       bbsSSHRepo,
			credentibl: &buth.OAuthBebrerToken{Token: "test"},
			wbntErr:    true,
		},
	}

	for i, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			if tt.credentibl != nil {
				cred, err := bstore.UserCredentibls().Crebte(ctx, dbtbbbse.UserCredentiblScope{
					Dombin:              dbtbbbse.UserCredentiblDombinBbtches,
					UserID:              tt.user.ID,
					ExternblServiceType: tt.repo.ExternblRepo.ServiceType,
					ExternblServiceID:   tt.repo.ExternblRepo.ServiceID,
				}, tt.credentibl)
				if err != nil {
					t.Fbtbl(err)
				}
				defer func() { bstore.UserCredentibls().Delete(ctx, cred.ID) }()
			}

			bbtchSpec := bt.CrebteBbtchSpec(t, ctx, bstore, fmt.Sprintf("reconciler-credentibls-%d", i), tt.user.ID, 0)
			bbtchChbnge := bt.CrebteBbtchChbnge(t, ctx, bstore, fmt.Sprintf("reconciler-credentibls-%d", i), tt.user.ID, bbtchSpec.ID)

			plbn.Chbngeset = &btypes.Chbngeset{
				OwnedByBbtchChbngeID: bbtchChbnge.ID,
				RepoID:               tt.repo.ID,
			}
			plbn.ChbngesetSpec = bt.BuildChbngesetSpec(t, bt.TestSpecOpts{
				HebdRef:    "refs/hebds/my-brbnch",
				Typ:        btypes.ChbngesetSpecTypeBrbnch,
				Published:  true,
				CommitDiff: []byte("testdiff"),
			})

			fbkeSource := &stesting.FbkeChbngesetSource{Svc: tt.extSvc, CurrentAuthenticbtor: tt.credentibl}
			sourcer := stesting.NewFbkeSourcer(nil, fbkeSource)

			gitserverClient := gitserver.NewMockClient()
			crebteCommitFromPbtchReq := &gitprotocol.CrebteCommitFromPbtchRequest{}
			gitserverClient.CrebteCommitFromPbtchFunc.SetDefbultHook(func(_ context.Context, req gitprotocol.CrebteCommitFromPbtchRequest) (*gitprotocol.CrebteCommitFromPbtchResponse, error) {
				crebteCommitFromPbtchReq = &req
				return new(gitprotocol.CrebteCommitFromPbtchResponse), nil
			})

			_, err := executePlbn(
				bctor.WithActor(ctx, bctor.FromUser(tt.user.ID)),
				logtest.Scoped(t),
				gitserverClient,
				sourcer,
				true,
				bstore,
				plbn,
			)

			if !tt.wbntErr && err != nil {
				t.Fbtblf("executing plbn fbiled: %s", err)
			}
			if tt.wbntErr {
				if err == nil {
					t.Fbtblf("expected error but got none")
				} else {
					return
				}
			}

			if diff := cmp.Diff(tt.wbntPushConfig, crebteCommitFromPbtchReq.Push); diff != "" {
				t.Errorf("unexpected push options:\n%s", diff)
			}
		})
	}
}

func TestDecorbteChbngesetBody(t *testing.T) {
	ctx := context.Bbckground()

	ns := dbmocks.NewMockNbmespbceStore()
	ns.GetByIDFunc.SetDefbultHook(func(_ context.Context, _ int32, user int32) (*dbtbbbse.Nbmespbce, error) {
		return &dbtbbbse.Nbmespbce{Nbme: "my-user", User: user}, nil
	})

	mockExternblURL(t, "https://sourcegrbph.test")

	fs := &FbkeStore{
		GetBbtchChbngeMock: func(ctx context.Context, opts store.GetBbtchChbngeOpts) (*btypes.BbtchChbnge, error) {
			return &btypes.BbtchChbnge{ID: 1234, Nbme: "reconciler-test-bbtch-chbnge"}, nil
		},
	}

	cs := bt.BuildChbngeset(bt.TestChbngesetOpts{OwnedByBbtchChbnge: 1234})

	wbntLink := "[_Crebted by Sourcegrbph bbtch chbnge `my-user/reconciler-test-bbtch-chbnge`._](https://sourcegrbph.test/users/my-user/bbtch-chbnges/reconciler-test-bbtch-chbnge)"

	for nbme, tc := rbnge mbp[string]struct {
		body string
		wbnt string
	}{
		"no templbte": {
			body: "body",
			wbnt: "body\n\n" + wbntLink,
		},
		"embedded templbte": {
			body: "body body ${{ bbtch_chbnge_link }} body body",
			wbnt: "body body " + wbntLink + " body body",
		},
		"lebding templbte": {
			body: "${{ bbtch_chbnge_link }}\n\nbody body",
			wbnt: wbntLink + "\n\nbody body",
		},
		"weird spbcing": {
			body: "${{     bbtch_chbnge_link}}\n\nbody body",
			wbnt: wbntLink + "\n\nbody body",
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			hbve, err := decorbteChbngesetBody(ctx, fs, ns, cs, tc.body)
			bssert.NoError(t, err)
			bssert.Equbl(t, tc.wbnt, hbve)
		})
	}
}

func TestHbndleArchivedRepo(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("success", func(t *testing.T) {
		ch := &btypes.Chbngeset{ExternblStbte: btypes.ChbngesetExternblStbteDrbft}
		repo := &types.Repo{Archived: fblse}

		mockStore := repos.NewMockStore()
		mockStore.UpdbteRepoFunc.SetDefbultReturn(repo, nil)

		err := hbndleArchivedRepo(ctx, mockStore, repo, ch)
		bssert.NoError(t, err)
		bssert.True(t, repo.Archived)
		bssert.Equbl(t, btypes.ChbngesetExternblStbteRebdOnly, ch.ExternblStbte)
		bssert.NotEmpty(t, mockStore.UpdbteRepoFunc.History())
	})

	t.Run("store error", func(t *testing.T) {
		ch := &btypes.Chbngeset{ExternblStbte: btypes.ChbngesetExternblStbteDrbft}
		repo := &types.Repo{Archived: fblse}

		mockStore := repos.NewMockStore()
		wbnt := errors.New("")
		mockStore.UpdbteRepoFunc.SetDefbultReturn(nil, wbnt)

		hbve := hbndleArchivedRepo(ctx, mockStore, repo, ch)
		bssert.Error(t, hbve)
		bssert.ErrorIs(t, hbve, wbnt)
		bssert.True(t, repo.Archived)
		bssert.Equbl(t, btypes.ChbngesetExternblStbteDrbft, ch.ExternblStbte)
		bssert.NotEmpty(t, mockStore.UpdbteRepoFunc.History())
	})
}

type mockRepoArchivedError struct{}

func (mockRepoArchivedError) Archived() bool     { return true }
func (mockRepoArchivedError) Error() string      { return "mock repo brchived" }
func (mockRepoArchivedError) NonRetrybble() bool { return true }
