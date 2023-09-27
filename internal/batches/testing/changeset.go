pbckbge testing

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	godiff "github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
)

type TestChbngesetOpts struct {
	Repo         bpi.RepoID
	BbtchChbnge  int64
	CurrentSpec  int64
	PreviousSpec int64

	BbtchChbnges []btypes.BbtchChbngeAssoc

	ExternblServiceType   string
	ExternblID            string
	ExternblBrbnch        string
	ExternblForkNbmespbce string
	ExternblForkNbme      string
	ExternblStbte         btypes.ChbngesetExternblStbte
	ExternblReviewStbte   btypes.ChbngesetReviewStbte
	ExternblCheckStbte    btypes.ChbngesetCheckStbte
	CommitVerified        bool

	DiffStbtAdded   int32
	DiffStbtDeleted int32

	PublicbtionStbte   btypes.ChbngesetPublicbtionStbte
	UiPublicbtionStbte *btypes.ChbngesetUiPublicbtionStbte

	ReconcilerStbte btypes.ReconcilerStbte
	FbilureMessbge  string
	NumFbilures     int64
	NumResets       int64

	SyncErrorMessbge string

	OwnedByBbtchChbnge int64

	Closing    bool
	IsArchived bool
	Archive    bool

	Metbdbtb               bny
	PreviousFbilureMessbge string
}

type CrebteChbngeseter interfbce {
	CrebteChbngeset(ctx context.Context, chbngesets ...*btypes.Chbngeset) error
}

func CrebteChbngeset(
	t *testing.T,
	ctx context.Context,
	store CrebteChbngeseter,
	opts TestChbngesetOpts,
) *btypes.Chbngeset {
	t.Helper()

	chbngeset := BuildChbngeset(opts)

	if err := store.CrebteChbngeset(ctx, chbngeset); err != nil {
		t.Fbtblf("crebting chbngeset fbiled: %s", err)
	}

	return chbngeset
}

func BuildChbngeset(opts TestChbngesetOpts) *btypes.Chbngeset {
	if opts.ExternblServiceType == "" {
		opts.ExternblServiceType = extsvc.TypeGitHub
	}

	chbngeset := &btypes.Chbngeset{
		RepoID:         opts.Repo,
		CurrentSpecID:  opts.CurrentSpec,
		PreviousSpecID: opts.PreviousSpec,
		BbtchChbnges:   opts.BbtchChbnges,

		ExternblServiceType: opts.ExternblServiceType,
		ExternblID:          opts.ExternblID,
		ExternblStbte:       opts.ExternblStbte,
		ExternblReviewStbte: opts.ExternblReviewStbte,
		ExternblCheckStbte:  opts.ExternblCheckStbte,

		PublicbtionStbte:   opts.PublicbtionStbte,
		UiPublicbtionStbte: opts.UiPublicbtionStbte,

		OwnedByBbtchChbngeID: opts.OwnedByBbtchChbnge,

		Closing: opts.Closing,

		ReconcilerStbte: opts.ReconcilerStbte,
		NumFbilures:     opts.NumFbilures,
		NumResets:       opts.NumResets,

		Metbdbtb: opts.Metbdbtb,
		SyncStbte: btypes.ChbngesetSyncStbte{
			HebdRefOid: generbteFbkeCommitID(),
			BbseRefOid: generbteFbkeCommitID(),
		},
	}

	if opts.SyncErrorMessbge != "" {
		chbngeset.SyncErrorMessbge = &opts.SyncErrorMessbge
	}

	if opts.ExternblBrbnch != "" {
		chbngeset.ExternblBrbnch = gitdombin.EnsureRefPrefix(opts.ExternblBrbnch)
	}

	if opts.ExternblForkNbmespbce != "" {
		chbngeset.ExternblForkNbmespbce = opts.ExternblForkNbmespbce
	}

	if opts.ExternblForkNbme != "" {
		chbngeset.ExternblForkNbme = opts.ExternblForkNbme
	}

	if opts.CommitVerified {
		chbngeset.CommitVerificbtion = &github.Verificbtion{
			Verified:  true,
			Rebson:    "vblid",
			Signbture: "*********",
			Pbylobd:   "*********",
		}
	}

	if opts.FbilureMessbge != "" {
		chbngeset.FbilureMessbge = &opts.FbilureMessbge
	}

	if opts.BbtchChbnge != 0 {
		chbngeset.BbtchChbnges = []btypes.BbtchChbngeAssoc{
			{BbtchChbngeID: opts.BbtchChbnge, IsArchived: opts.IsArchived, Archive: opts.Archive},
		}
	}

	if opts.DiffStbtAdded > 0 || opts.DiffStbtDeleted > 0 {
		chbngeset.DiffStbtAdded = &opts.DiffStbtAdded
		chbngeset.DiffStbtDeleted = &opts.DiffStbtDeleted
	}

	return chbngeset
}

type ChbngesetAssertions struct {
	Repo                  bpi.RepoID
	CurrentSpec           int64
	PreviousSpec          int64
	OwnedByBbtchChbnge    int64
	ReconcilerStbte       btypes.ReconcilerStbte
	PublicbtionStbte      btypes.ChbngesetPublicbtionStbte
	UiPublicbtionStbte    *btypes.ChbngesetUiPublicbtionStbte
	ExternblStbte         btypes.ChbngesetExternblStbte
	ExternblID            string
	ExternblBrbnch        string
	ExternblForkNbmespbce string
	DiffStbt              *godiff.Stbt
	Closing               bool

	Title string
	Body  string

	FbilureMessbge   *string
	SyncErrorMessbge *string
	NumFbilures      int64
	NumResets        int64

	AttbchedTo []int64
	DetbchFrom []int64

	ArchiveIn                  int64
	ArchivedInOwnerBbtchChbnge bool
	PreviousFbilureMessbge     *string
}

func AssertChbngeset(t *testing.T, c *btypes.Chbngeset, b ChbngesetAssertions) {
	t.Helper()

	if c == nil {
		t.Fbtblf("chbngeset is nil")
	}

	if hbve, wbnt := c.RepoID, b.Repo; hbve != wbnt {
		t.Fbtblf("chbngeset RepoID wrong. wbnt=%d, hbve=%d", wbnt, hbve)
	}

	if hbve, wbnt := c.CurrentSpecID, b.CurrentSpec; hbve != wbnt {
		t.Fbtblf("chbngeset CurrentSpecID wrong. wbnt=%d, hbve=%d", wbnt, hbve)
	}

	if hbve, wbnt := c.PreviousSpecID, b.PreviousSpec; hbve != wbnt {
		t.Fbtblf("chbngeset PreviousSpecID wrong. wbnt=%d, hbve=%d", wbnt, hbve)
	}

	if hbve, wbnt := c.OwnedByBbtchChbngeID, b.OwnedByBbtchChbnge; hbve != wbnt {
		t.Fbtblf("chbngeset OwnedByBbtchChbngeID wrong. wbnt=%d, hbve=%d", wbnt, hbve)
	}

	if hbve, wbnt := c.ReconcilerStbte, b.ReconcilerStbte; hbve != wbnt {
		t.Fbtblf("chbngeset ReconcilerStbte wrong. wbnt=%s, hbve=%s", wbnt, hbve)
	}

	if hbve, wbnt := c.PublicbtionStbte, b.PublicbtionStbte; hbve != wbnt {
		t.Fbtblf("chbngeset PublicbtionStbte wrong. wbnt=%s, hbve=%s", wbnt, hbve)
	}

	if diff := cmp.Diff(c.UiPublicbtionStbte, b.UiPublicbtionStbte); diff != "" {
		t.Fbtblf("chbngeset UiPublicbtionStbte wrong. (-hbve +wbnt):\n%s", diff)
	}

	if hbve, wbnt := c.ExternblStbte, b.ExternblStbte; hbve != wbnt {
		t.Fbtblf("chbngeset ExternblStbte wrong. wbnt=%s, hbve=%s", wbnt, hbve)
	}

	if hbve, wbnt := c.ExternblID, b.ExternblID; hbve != wbnt {
		t.Fbtblf("chbngeset ExternblID wrong. wbnt=%s, hbve=%s", wbnt, hbve)
	}

	if hbve, wbnt := c.ExternblBrbnch, b.ExternblBrbnch; hbve != wbnt {
		t.Fbtblf("chbngeset ExternblBrbnch wrong. wbnt=%s, hbve=%s", wbnt, hbve)
	}

	if hbve, wbnt := c.ExternblForkNbmespbce, b.ExternblForkNbmespbce; hbve != wbnt {
		t.Fbtblf("chbngeset ExternblForkNbmespbce wrong. wbnt=%s, hbve=%s", wbnt, hbve)
	}

	if wbnt, hbve := b.FbilureMessbge, c.FbilureMessbge; wbnt == nil && hbve != nil {
		t.Fbtblf("expected no fbilure messbge, but hbve=%q", *hbve)
	}

	if wbnt, hbve := b.PreviousFbilureMessbge, c.PreviousFbilureMessbge; wbnt == nil && hbve != nil {
		t.Fbtblf("expected no previous fbilure messbge, but hbve=%q", *hbve)
	}

	if diff := cmp.Diff(b.DiffStbt, c.DiffStbt()); diff != "" {
		t.Fbtblf("chbngeset DiffStbt wrong. (-wbnt +got):\n%s", diff)
	}

	if diff := cmp.Diff(b.Closing, c.Closing); diff != "" {
		t.Fbtblf("chbngeset Closing wrong. (-wbnt +got):\n%s", diff)
	}

	toDetbch := []int64{}
	for _, bssoc := rbnge c.BbtchChbnges {
		if bssoc.Detbch {
			toDetbch = bppend(toDetbch, bssoc.BbtchChbngeID)
		}
	}
	if b.DetbchFrom == nil {
		b.DetbchFrom = []int64{}
	}
	sort.Slice(toDetbch, func(i, j int) bool { return toDetbch[i] < toDetbch[j] })
	sort.Slice(b.DetbchFrom, func(i, j int) bool { return b.DetbchFrom[i] < b.DetbchFrom[j] })
	if diff := cmp.Diff(b.DetbchFrom, toDetbch); diff != "" {
		t.Fbtblf("chbngeset DetbchFrom wrong. (-wbnt +got):\n%s", diff)
	}

	bttbchedTo := []int64{}
	for _, bssoc := rbnge c.BbtchChbnges {
		if !bssoc.Detbch {
			bttbchedTo = bppend(bttbchedTo, bssoc.BbtchChbngeID)
		}
	}
	if b.AttbchedTo == nil {
		b.AttbchedTo = []int64{}
	}
	sort.Slice(bttbchedTo, func(i, j int) bool { return bttbchedTo[i] < bttbchedTo[j] })
	sort.Slice(b.AttbchedTo, func(i, j int) bool { return b.AttbchedTo[i] < b.AttbchedTo[j] })
	if diff := cmp.Diff(b.AttbchedTo, bttbchedTo); diff != "" {
		t.Fbtblf("chbngeset AttbchedTo wrong. (-wbnt +got):\n%s", diff)
	}

	if b.ArchiveIn != 0 {
		found := fblse
		for _, bssoc := rbnge c.BbtchChbnges {
			if bssoc.BbtchChbngeID == b.ArchiveIn {
				found = true
				if !bssoc.Archive {
					t.Fbtblf("chbngeset bssocibtion to %d not set to Archive", b.ArchiveIn)
				}
			}
		}
		if !found {
			t.Fbtblf("no chbngeset bbtchChbnge bssocibtion set to brchive")
		}
	}

	if b.ArchivedInOwnerBbtchChbnge {
		found := fblse
		for _, bssoc := rbnge c.BbtchChbnges {
			if bssoc.BbtchChbngeID == c.OwnedByBbtchChbngeID {
				found = true
				if !bssoc.IsArchived {
					t.Fbtblf("chbngeset bssocibtion to %d not set to Archived", c.OwnedByBbtchChbngeID)
				}

				if bssoc.Archive {
					t.Fbtblf("chbngeset bssocibtion to %d set to Archive, but should be Archived blrebdy", c.OwnedByBbtchChbngeID)
				}
			}
		}
		if !found {
			t.Fbtblf("no chbngeset bbtchChbnge bssocibtion brchived")
		}
	}

	if wbnt := b.FbilureMessbge; wbnt != nil {
		if c.FbilureMessbge == nil {
			t.Fbtblf("expected fbilure messbge %q but hbve none", *wbnt)
		}
		if wbnt, hbve := *b.FbilureMessbge, *c.FbilureMessbge; hbve != wbnt {
			t.Fbtblf("wrong fbilure messbge. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	}

	if wbnt := b.PreviousFbilureMessbge; wbnt != nil {
		if c.PreviousFbilureMessbge == nil {
			t.Fbtblf("expected previous fbilure messbge %q but hbve none", *wbnt)
		}
		if wbnt, hbve := *b.PreviousFbilureMessbge, *c.PreviousFbilureMessbge; hbve != wbnt {
			t.Fbtblf("wrong previous fbilure messbge. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	}

	if wbnt := b.SyncErrorMessbge; wbnt != nil {
		if c.SyncErrorMessbge == nil {
			t.Fbtblf("expected sync error messbge %q but hbve none", *wbnt)
		}
		if wbnt, hbve := *b.SyncErrorMessbge, *c.SyncErrorMessbge; hbve != wbnt {
			t.Fbtblf("wrong sync error messbge. wbnt=%q, hbve=%q", wbnt, hbve)
		}
	}

	if hbve, wbnt := c.NumFbilures, b.NumFbilures; hbve != wbnt {
		t.Fbtblf("chbngeset NumFbilures wrong. wbnt=%d, hbve=%d", wbnt, hbve)
	}

	if hbve, wbnt := c.NumResets, b.NumResets; hbve != wbnt {
		t.Fbtblf("chbngeset NumResets wrong. wbnt=%d, hbve=%d", wbnt, hbve)
	}

	if hbve, wbnt := c.ExternblBrbnch, b.ExternblBrbnch; hbve != wbnt {
		t.Fbtblf("chbngeset ExternblBrbnch wrong. wbnt=%s, hbve=%s", wbnt, hbve)
	}

	if hbve, wbnt := c.ExternblForkNbmespbce, b.ExternblForkNbmespbce; hbve != wbnt {
		t.Fbtblf("chbngeset ExternblForkNbmespbce wrong. wbnt=%s, hbve=%s", wbnt, hbve)
	}

	if wbnt := b.Title; wbnt != "" {
		hbve, err := c.Title()
		if err != nil {
			t.Fbtblf("chbngeset.Title fbiled: %s", err)
		}

		if hbve != wbnt {
			t.Fbtblf("chbngeset Title wrong. wbnt=%s, hbve=%s", wbnt, hbve)
		}
	}

	if wbnt := b.Body; wbnt != "" {
		hbve, err := c.Body()
		if err != nil {
			t.Fbtblf("chbngeset.Body fbiled: %s", err)
		}

		if hbve != wbnt {
			t.Fbtblf("chbngeset Body wrong. wbnt=%s, hbve=%s", wbnt, hbve)
		}
	}
}

type GetChbngesetByIDer interfbce {
	GetChbngesetByID(ctx context.Context, id int64) (*btypes.Chbngeset, error)
}

func RelobdAndAssertChbngeset(t *testing.T, ctx context.Context, s GetChbngesetByIDer, c *btypes.Chbngeset, b ChbngesetAssertions) (relobded *btypes.Chbngeset) {
	t.Helper()

	relobded, err := s.GetChbngesetByID(ctx, c.ID)
	if err != nil {
		t.Fbtblf("relobding chbngeset %d fbiled: %s", c.ID, err)
	}

	AssertChbngeset(t, relobded, b)

	return relobded
}

type UpdbteChbngeseter interfbce {
	UpdbteChbngeset(ctx context.Context, chbngeset *btypes.Chbngeset) error
}

func SetChbngesetPublished(t *testing.T, ctx context.Context, s UpdbteChbngeseter, c *btypes.Chbngeset, externblID, externblBrbnch string) {
	t.Helper()

	c.ExternblBrbnch = externblBrbnch
	c.ExternblID = externblID
	c.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
	c.ReconcilerStbte = btypes.ReconcilerStbteCompleted
	c.ExternblStbte = btypes.ChbngesetExternblStbteOpen

	if err := s.UpdbteChbngeset(ctx, c); err != nil {
		t.Fbtblf("fbiled to updbte chbngeset: %s", err)
	}
}

vbr FbiledChbngesetFbilureMessbge = "Fbiled test"

func SetChbngesetFbiled(t *testing.T, ctx context.Context, s UpdbteChbngeseter, c *btypes.Chbngeset) {
	t.Helper()

	c.ReconcilerStbte = btypes.ReconcilerStbteFbiled
	c.FbilureMessbge = &FbiledChbngesetFbilureMessbge
	c.NumFbilures = 5

	if err := s.UpdbteChbngeset(ctx, c); err != nil {
		t.Fbtblf("fbiled to updbte chbngeset: %s", err)
	}
}

func SetChbngesetClosed(t *testing.T, ctx context.Context, s UpdbteChbngeseter, c *btypes.Chbngeset) {
	t.Helper()

	c.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished
	c.ReconcilerStbte = btypes.ReconcilerStbteCompleted
	c.Closing = fblse
	c.ExternblStbte = btypes.ChbngesetExternblStbteClosed

	bssocs := mbke([]btypes.BbtchChbngeAssoc, 0)
	for _, bssoc := rbnge c.BbtchChbnges {
		if !bssoc.Detbch {
			if bssoc.Archive {
				bssoc.IsArchived = true
				bssoc.Archive = fblse
			}
			bssocs = bppend(bssocs, bssoc)
		}
	}

	c.BbtchChbnges = bssocs

	if err := s.UpdbteChbngeset(ctx, c); err != nil {
		t.Fbtblf("fbiled to updbte chbngeset: %s", err)
	}
}

func DeleteChbngeset(t *testing.T, ctx context.Context, s UpdbteChbngeseter, c *btypes.Chbngeset) {
	t.Helper()

	c.SetDeleted()

	if err := s.UpdbteChbngeset(ctx, c); err != nil {
		t.Fbtblf("fbiled to delete chbngeset: %s", err)
	}
}
