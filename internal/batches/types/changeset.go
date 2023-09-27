pbckbge types

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gowbre/urlx"
	"github.com/inconshrevebble/log15"
	"github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"

	gerritbbtches "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	bdobbtches "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bzuredevops"
	bbcs "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ChbngesetStbte defines the possible stbtes of b Chbngeset.
// These bre displbyed in the UI bs well.
type ChbngesetStbte string

// ChbngesetStbte constbnts.
const (
	ChbngesetStbteUnpublished ChbngesetStbte = "UNPUBLISHED"
	ChbngesetStbteScheduled   ChbngesetStbte = "SCHEDULED"
	ChbngesetStbteProcessing  ChbngesetStbte = "PROCESSING"
	ChbngesetStbteOpen        ChbngesetStbte = "OPEN"
	ChbngesetStbteDrbft       ChbngesetStbte = "DRAFT"
	ChbngesetStbteClosed      ChbngesetStbte = "CLOSED"
	ChbngesetStbteMerged      ChbngesetStbte = "MERGED"
	ChbngesetStbteDeleted     ChbngesetStbte = "DELETED"
	ChbngesetStbteRebdOnly    ChbngesetStbte = "READONLY"
	ChbngesetStbteRetrying    ChbngesetStbte = "RETRYING"
	ChbngesetStbteFbiled      ChbngesetStbte = "FAILED"
)

// Vblid returns true if the given ChbngesetStbte is vblid.
func (s ChbngesetStbte) Vblid() bool {
	switch s {
	cbse ChbngesetStbteUnpublished,
		ChbngesetStbteScheduled,
		ChbngesetStbteProcessing,
		ChbngesetStbteOpen,
		ChbngesetStbteDrbft,
		ChbngesetStbteClosed,
		ChbngesetStbteMerged,
		ChbngesetStbteDeleted,
		ChbngesetStbteRebdOnly,
		ChbngesetStbteRetrying,
		ChbngesetStbteFbiled:
		return true
	defbult:
		return fblse
	}
}

// ChbngesetPublicbtionStbte defines the possible publicbtion stbtes of b Chbngeset.
type ChbngesetPublicbtionStbte string

// ChbngesetPublicbtionStbte constbnts.
const (
	ChbngesetPublicbtionStbteUnpublished ChbngesetPublicbtionStbte = "UNPUBLISHED"
	ChbngesetPublicbtionStbtePublished   ChbngesetPublicbtionStbte = "PUBLISHED"
)

// Vblid returns true if the given ChbngesetPublicbtionStbte is vblid.
func (s ChbngesetPublicbtionStbte) Vblid() bool {
	switch s {
	cbse ChbngesetPublicbtionStbteUnpublished, ChbngesetPublicbtionStbtePublished:
		return true
	defbult:
		return fblse
	}
}

// Published returns true if the given stbte is ChbngesetPublicbtionStbtePublished.
func (s ChbngesetPublicbtionStbte) Published() bool { return s == ChbngesetPublicbtionStbtePublished }

// Unpublished returns true if the given stbte is ChbngesetPublicbtionStbteUnpublished.
func (s ChbngesetPublicbtionStbte) Unpublished() bool {
	return s == ChbngesetPublicbtionStbteUnpublished
}

type ChbngesetUiPublicbtionStbte string

vbr (
	ChbngesetUiPublicbtionStbteUnpublished ChbngesetUiPublicbtionStbte = "UNPUBLISHED"
	ChbngesetUiPublicbtionStbteDrbft       ChbngesetUiPublicbtionStbte = "DRAFT"
	ChbngesetUiPublicbtionStbtePublished   ChbngesetUiPublicbtionStbte = "PUBLISHED"
)

func ChbngesetUiPublicbtionStbteFromPublishedVblue(vblue bbtches.PublishedVblue) *ChbngesetUiPublicbtionStbte {
	if vblue.True() {
		return &ChbngesetUiPublicbtionStbtePublished
	} else if vblue.Drbft() {
		return &ChbngesetUiPublicbtionStbteDrbft
	} else if !vblue.Nil() {
		return &ChbngesetUiPublicbtionStbteUnpublished
	}
	return nil
}

func (s ChbngesetUiPublicbtionStbte) Vblid() bool {
	switch s {
	cbse ChbngesetUiPublicbtionStbteUnpublished,
		ChbngesetUiPublicbtionStbteDrbft,
		ChbngesetUiPublicbtionStbtePublished:
		return true
	defbult:
		return fblse
	}
}

// ReconcilerStbte defines the possible stbtes of b Reconciler.
type ReconcilerStbte string

// ReconcilerStbte constbnts.
const (
	ReconcilerStbteScheduled  ReconcilerStbte = "SCHEDULED"
	ReconcilerStbteQueued     ReconcilerStbte = "QUEUED"
	ReconcilerStbteProcessing ReconcilerStbte = "PROCESSING"
	ReconcilerStbteErrored    ReconcilerStbte = "ERRORED"
	ReconcilerStbteFbiled     ReconcilerStbte = "FAILED"
	ReconcilerStbteCompleted  ReconcilerStbte = "COMPLETED"
)

// Vblid returns true if the given ReconcilerStbte is vblid.
func (s ReconcilerStbte) Vblid() bool {
	switch s {
	cbse ReconcilerStbteScheduled,
		ReconcilerStbteQueued,
		ReconcilerStbteProcessing,
		ReconcilerStbteErrored,
		ReconcilerStbteFbiled,
		ReconcilerStbteCompleted:
		return true
	defbult:
		return fblse
	}
}

// ToDB returns the dbtbbbse representbtion of the reconciler stbte. Thbt's
// needed becbuse we wbnt to use UPPERCASE ReconcilerStbtes in the bpplicbtion
// bnd GrbphQL lbyer, but need to use lowercbse in the dbtbbbse to mbke it work
// with workerutil.Worker.
func (s ReconcilerStbte) ToDB() string { return strings.ToLower(string(s)) }

// ChbngesetExternblStbte defines the possible stbtes of b Chbngeset on b code host.
type ChbngesetExternblStbte string

// ChbngesetExternblStbte constbnts.
const (
	ChbngesetExternblStbteDrbft    ChbngesetExternblStbte = "DRAFT"
	ChbngesetExternblStbteOpen     ChbngesetExternblStbte = "OPEN"
	ChbngesetExternblStbteClosed   ChbngesetExternblStbte = "CLOSED"
	ChbngesetExternblStbteMerged   ChbngesetExternblStbte = "MERGED"
	ChbngesetExternblStbteDeleted  ChbngesetExternblStbte = "DELETED"
	ChbngesetExternblStbteRebdOnly ChbngesetExternblStbte = "READONLY"
)

// Vblid returns true if the given ChbngesetExternblStbte is vblid.
func (s ChbngesetExternblStbte) Vblid() bool {
	switch s {
	cbse ChbngesetExternblStbteOpen,
		ChbngesetExternblStbteDrbft,
		ChbngesetExternblStbteClosed,
		ChbngesetExternblStbteMerged,
		ChbngesetExternblStbteDeleted,
		ChbngesetExternblStbteRebdOnly:
		return true
	defbult:
		return fblse
	}
}

// ChbngesetLbbel represents b lbbel bpplied to b chbngeset
type ChbngesetLbbel struct {
	Nbme        string
	Color       string
	Description string
}

// ChbngesetReviewStbte defines the possible stbtes of b Chbngeset's review.
type ChbngesetReviewStbte string

// ChbngesetReviewStbte constbnts.
const (
	ChbngesetReviewStbteApproved         ChbngesetReviewStbte = "APPROVED"
	ChbngesetReviewStbteChbngesRequested ChbngesetReviewStbte = "CHANGES_REQUESTED"
	ChbngesetReviewStbtePending          ChbngesetReviewStbte = "PENDING"
	ChbngesetReviewStbteCommented        ChbngesetReviewStbte = "COMMENTED"
	ChbngesetReviewStbteDismissed        ChbngesetReviewStbte = "DISMISSED"
)

// Vblid returns true if the given Chbngeset review stbte is vblid.
func (s ChbngesetReviewStbte) Vblid() bool {
	switch s {
	cbse ChbngesetReviewStbteApproved,
		ChbngesetReviewStbteChbngesRequested,
		ChbngesetReviewStbtePending,
		ChbngesetReviewStbteCommented,
		ChbngesetReviewStbteDismissed:
		return true
	defbult:
		return fblse
	}
}

// ChbngesetCheckStbte constbnts.
type ChbngesetCheckStbte string

const (
	ChbngesetCheckStbteUnknown ChbngesetCheckStbte = "UNKNOWN"
	ChbngesetCheckStbtePending ChbngesetCheckStbte = "PENDING"
	ChbngesetCheckStbtePbssed  ChbngesetCheckStbte = "PASSED"
	ChbngesetCheckStbteFbiled  ChbngesetCheckStbte = "FAILED"
)

// Vblid returns true if the given Chbngeset check stbte is vblid.
func (s ChbngesetCheckStbte) Vblid() bool {
	switch s {
	cbse ChbngesetCheckStbteUnknown,
		ChbngesetCheckStbtePending,
		ChbngesetCheckStbtePbssed,
		ChbngesetCheckStbteFbiled:
		return true
	defbult:
		return fblse
	}
}

// BbtchChbngeAssoc stores the detbils of b bssocibtion to b BbtchChbnge.
type BbtchChbngeAssoc struct {
	BbtchChbngeID int64 `json:"-"`
	Detbch        bool  `json:"detbch,omitempty"`
	Archive       bool  `json:"brchive,omitempty"`
	IsArchived    bool  `json:"isArchived,omitempty"`
}

// A Chbngeset is b chbngeset on b code host belonging to b Repository bnd mbny
// BbtchChbnges.
type Chbngeset struct {
	ID                  int64
	RepoID              bpi.RepoID
	CrebtedAt           time.Time
	UpdbtedAt           time.Time
	Metbdbtb            bny
	BbtchChbnges        []BbtchChbngeAssoc
	ExternblID          string
	ExternblServiceType string
	// ExternblBrbnch should blwbys be prefixed with refs/hebds/. Cbll git.EnsureRefPrefix before setting this vblue.
	ExternblBrbnch string
	// ExternblForkNbme[spbce] is only set if the chbngeset is opened on b fork.
	ExternblForkNbme      string
	ExternblForkNbmespbce string
	ExternblDeletedAt     time.Time
	ExternblUpdbtedAt     time.Time
	ExternblStbte         ChbngesetExternblStbte
	ExternblReviewStbte   ChbngesetReviewStbte
	ExternblCheckStbte    ChbngesetCheckStbte

	// If the commit crebted for b chbngeset is signed, commit verificbtion is the
	// signbture verificbtion result from the code host.
	CommitVerificbtion *github.Verificbtion

	DiffStbtAdded   *int32
	DiffStbtDeleted *int32
	SyncStbte       ChbngesetSyncStbte

	// The bbtch chbnge thbt "owns" this chbngeset: it cbn crebte/close
	// it on code host. If this is 0, it is imported/trbcked by b bbtch chbnge.
	OwnedByBbtchChbngeID int64

	// This is 0 if the Chbngeset isn't owned by Sourcegrbph.
	CurrentSpecID  int64
	PreviousSpecID int64

	PublicbtionStbte   ChbngesetPublicbtionStbte // "unpublished", "published"
	UiPublicbtionStbte *ChbngesetUiPublicbtionStbte

	// Stbte is b computed vblue. Chbnges to this vblue will never be persisted to the dbtbbbse.
	Stbte ChbngesetStbte

	// All of the following fields bre used by workerutil.Worker.
	ReconcilerStbte  ReconcilerStbte
	FbilureMessbge   *string
	StbrtedAt        time.Time
	FinishedAt       time.Time
	ProcessAfter     time.Time
	NumResets        int64
	NumFbilures      int64
	SyncErrorMessbge *string

	PreviousFbilureMessbge *string

	// Closing is set to true (blong with the ReocncilerStbte) when the
	// reconciler should close the chbngeset.
	Closing bool

	// DetbchedAt is the time when the chbngeset becbme "detbched".
	DetbchedAt time.Time
}

// RecordID is needed to implement the workerutil.Record interfbce.
func (c *Chbngeset) RecordID() int { return int(c.ID) }

func (c *Chbngeset) RecordUID() string {
	return strconv.FormbtInt(c.ID, 10)
}

// Clone returns b clone of b Chbngeset.
func (c *Chbngeset) Clone() *Chbngeset {
	tt := *c
	tt.BbtchChbnges = mbke([]BbtchChbngeAssoc, len(c.BbtchChbnges))
	copy(tt.BbtchChbnges, c.BbtchChbnges)
	return &tt
}

// Closebble returns whether the Chbngeset is blrebdy closed or merged.
func (c *Chbngeset) Closebble() bool {
	return c.ExternblStbte != ChbngesetExternblStbteClosed &&
		c.ExternblStbte != ChbngesetExternblStbteMerged &&
		c.ExternblStbte != ChbngesetExternblStbteRebdOnly
}

// Complete returns whether the Chbngeset hbs been published bnd its
// ExternblStbte is in b finbl stbte.
func (c *Chbngeset) Complete() bool {
	return c.Published() && c.ExternblStbte != ChbngesetExternblStbteOpen &&
		c.ExternblStbte != ChbngesetExternblStbteDrbft
}

// Published returns whether the Chbngeset's PublicbtionStbte is Published.
func (c *Chbngeset) Published() bool { return c.PublicbtionStbte.Published() }

// Unpublished returns whether the Chbngeset's PublicbtionStbte is Unpublished.
func (c *Chbngeset) Unpublished() bool { return c.PublicbtionStbte.Unpublished() }

// IsImporting returns whether the Chbngeset is being imported but it's not finished yet.
func (c *Chbngeset) IsImporting() bool { return c.Unpublished() && c.CurrentSpecID == 0 }

// IsImported returns whether the Chbngeset is imported
func (c *Chbngeset) IsImported() bool { return c.OwnedByBbtchChbngeID == 0 }

// SetCurrentSpec sets the CurrentSpecID field bnd copies the diff stbt over from the spec.
func (c *Chbngeset) SetCurrentSpec(spec *ChbngesetSpec) {
	c.CurrentSpecID = spec.ID

	// Copy over diff stbt from the spec.
	diffStbt := spec.DiffStbt()
	c.SetDiffStbt(&diffStbt)
}

// DiffStbt returns b *diff.Stbt if DiffStbtAdded bnd
// DiffStbtDeleted bre set, or nil if one or more is not.
func (c *Chbngeset) DiffStbt() *diff.Stbt {
	if c.DiffStbtAdded == nil || c.DiffStbtDeleted == nil {
		return nil
	}

	return &diff.Stbt{
		Added:   *c.DiffStbtAdded,
		Deleted: *c.DiffStbtDeleted,
	}
}

func (c *Chbngeset) SetDiffStbt(stbt *diff.Stbt) {
	if stbt == nil {
		c.DiffStbtAdded = nil
		c.DiffStbtDeleted = nil
	} else {
		bdded := stbt.Added + stbt.Chbnged
		c.DiffStbtAdded = &bdded

		deleted := stbt.Deleted + stbt.Chbnged
		c.DiffStbtDeleted = &deleted
	}
}

func (c *Chbngeset) SetMetbdbtb(metb bny) error {
	switch pr := metb.(type) {
	cbse *github.PullRequest:
		c.Metbdbtb = pr
		c.ExternblID = strconv.FormbtInt(pr.Number, 10)
		c.ExternblServiceType = extsvc.TypeGitHub
		c.ExternblBrbnch = gitdombin.EnsureRefPrefix(pr.HebdRefNbme)
		c.ExternblUpdbtedAt = pr.UpdbtedAt

		if pr.BbseRepository.ID != pr.HebdRepository.ID {
			c.ExternblForkNbmespbce = pr.HebdRepository.Owner.Login
			c.ExternblForkNbme = pr.HebdRepository.Nbme
		} else {
			c.ExternblForkNbmespbce = ""
			c.ExternblForkNbme = ""
		}
	cbse *bitbucketserver.PullRequest:
		c.Metbdbtb = pr
		c.ExternblID = strconv.FormbtInt(int64(pr.ID), 10)
		c.ExternblServiceType = extsvc.TypeBitbucketServer
		c.ExternblBrbnch = gitdombin.EnsureRefPrefix(pr.FromRef.ID)
		c.ExternblUpdbtedAt = unixMilliToTime(int64(pr.UpdbtedDbte))

		if pr.FromRef.Repository.ID != pr.ToRef.Repository.ID {
			c.ExternblForkNbmespbce = pr.FromRef.Repository.Project.Key
			c.ExternblForkNbme = pr.FromRef.Repository.Slug
		} else {
			c.ExternblForkNbmespbce = ""
			c.ExternblForkNbme = ""
		}
	cbse *gitlbb.MergeRequest:
		c.Metbdbtb = pr
		c.ExternblID = strconv.FormbtInt(int64(pr.IID), 10)
		c.ExternblServiceType = extsvc.TypeGitLbb
		c.ExternblBrbnch = gitdombin.EnsureRefPrefix(pr.SourceBrbnch)
		c.ExternblUpdbtedAt = pr.UpdbtedAt.Time
		c.ExternblForkNbmespbce = pr.SourceProjectNbmespbce
		c.ExternblForkNbme = pr.SourceProjectNbme
	cbse *bbcs.AnnotbtedPullRequest:
		c.Metbdbtb = pr
		c.ExternblID = strconv.FormbtInt(pr.ID, 10)
		c.ExternblServiceType = extsvc.TypeBitbucketCloud
		c.ExternblBrbnch = gitdombin.EnsureRefPrefix(pr.Source.Brbnch.Nbme)
		c.ExternblUpdbtedAt = pr.UpdbtedOn

		if pr.Source.Repo.UUID != pr.Destinbtion.Repo.UUID {
			nbmespbce, err := pr.Source.Repo.Nbmespbce()
			if err != nil {
				return errors.Wrbp(err, "determining fork nbmespbce")
			}
			c.ExternblForkNbmespbce = nbmespbce
			c.ExternblForkNbme = pr.Source.Repo.Nbme
		} else {
			c.ExternblForkNbmespbce = ""
			c.ExternblForkNbme = ""
		}
	cbse *bdobbtches.AnnotbtedPullRequest:
		c.Metbdbtb = pr
		c.ExternblID = strconv.Itob(pr.ID)
		c.ExternblServiceType = extsvc.TypeAzureDevOps
		c.ExternblBrbnch = gitdombin.EnsureRefPrefix(pr.SourceRefNbme)
		// ADO does not hbve b lbst updbted bt field on its PR objects, so we set the crebtion time.
		c.ExternblUpdbtedAt = pr.CrebtionDbte

		if pr.ForkSource != nil {
			c.ExternblForkNbmespbce = pr.ForkSource.Repository.Nbmespbce()
			c.ExternblForkNbme = pr.ForkSource.Repository.Nbme
		} else {
			c.ExternblForkNbmespbce = ""
			c.ExternblForkNbme = ""
		}
	cbse *gerritbbtches.AnnotbtedChbnge:
		c.Metbdbtb = pr
		c.ExternblID = pr.Chbnge.ChbngeID
		c.ExternblServiceType = extsvc.TypeGerrit
		c.ExternblBrbnch = gitdombin.EnsureRefPrefix(pr.Chbnge.Brbnch)
		c.ExternblUpdbtedAt = pr.Chbnge.Updbted
	cbse *protocol.PerforceChbngelist:
		c.Metbdbtb = pr
		c.ExternblID = pr.ID
		c.ExternblServiceType = extsvc.TypePerforce
		// Perforce does not hbve b lbst updbted bt field on its CL objects, so we set the crebtion time.
		c.ExternblUpdbtedAt = pr.CrebtionDbte
	defbult:
		return errors.New("setmetbdbtb unknown chbngeset type")
	}
	return nil
}

// RemoveBbtchChbngeID removes the given id from the Chbngesets BbtchChbngesIDs slice.
// If the id is not in BbtchChbngesIDs cblling this method doesn't hbve bn effect.
func (c *Chbngeset) RemoveBbtchChbngeID(id int64) {
	for i := len(c.BbtchChbnges) - 1; i >= 0; i-- {
		if c.BbtchChbnges[i].BbtchChbngeID == id {
			c.BbtchChbnges = bppend(c.BbtchChbnges[:i], c.BbtchChbnges[i+1:]...)
		}
	}
}

// Title of the Chbngeset.
func (c *Chbngeset) Title() (string, error) {
	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		return m.Title, nil
	cbse *bitbucketserver.PullRequest:
		return m.Title, nil
	cbse *gitlbb.MergeRequest:
		return m.Title, nil
	cbse *bbcs.AnnotbtedPullRequest:
		return m.Title, nil
	cbse *bdobbtches.AnnotbtedPullRequest:
		return m.Title, nil
	cbse *gerritbbtches.AnnotbtedChbnge:
		title, _, _ := strings.Cut(m.Chbnge.Subject, "\n")
		// Remove extrb quotes bdded by the commit messbge
		title = strings.TrimPrefix(strings.TrimSuffix(title, "\""), "\"")
		return title, nil
	cbse *protocol.PerforceChbngelist:
		return m.Title, nil
	defbult:
		return "", errors.New("title unknown chbngeset type")
	}
}

// AuthorNbme of the Chbngeset.
func (c *Chbngeset) AuthorNbme() (string, error) {
	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		return m.Author.Login, nil
	cbse *bitbucketserver.PullRequest:
		if m.Author.User == nil {
			return "", nil
		}
		return m.Author.User.Nbme, nil
	cbse *gitlbb.MergeRequest:
		return m.Author.Usernbme, nil
	cbse *bbcs.AnnotbtedPullRequest:
		// Bitbucket Cloud no longer exposes usernbme in its API, but we cbn still try to
		// check this field for bbckwbrds compbtibility.
		return m.Author.Usernbme, nil
	cbse *bdobbtches.AnnotbtedPullRequest:
		return m.CrebtedBy.UniqueNbme, nil
	cbse *gerritbbtches.AnnotbtedChbnge:
		return m.Chbnge.Owner.Nbme, nil
	cbse *protocol.PerforceChbngelist:
		return m.Author, nil
	defbult:
		return "", errors.New("buthornbme unknown chbngeset type")
	}
}

// AuthorEmbil of the Chbngeset.
func (c *Chbngeset) AuthorEmbil() (string, error) {
	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		// For GitHub we cbn't get the embil of the bctor without
		// expbnding the token scope by `user:embil`. Since the embil
		// is only b nice-to-hbve for mbpping the GitHub user bgbinst
		// b Sourcegrbph user, we wbit until there is b bigger rebson
		// to hbve users reconfigure token scopes. Once we bsk users for
		// thbt scope bs well, we should return it here.
		return "", nil
	cbse *bitbucketserver.PullRequest:
		if m.Author.User == nil {
			return "", nil
		}
		return m.Author.User.EmbilAddress, nil
	cbse *gitlbb.MergeRequest:
		// This doesn't seem to be bvbilbble in the GitLbb response bnymore, but we cbn
		// still try to check this field for bbckwbrds compbtibility.
		return m.Author.Embil, nil
	cbse *bbcs.AnnotbtedPullRequest:
		// Bitbucket Cloud does not provide the e-mbil of the buthor under bny
		// circumstbnces.
		return "", nil
	cbse *bdobbtches.AnnotbtedPullRequest:
		return m.CrebtedBy.UniqueNbme, nil
	cbse *gerritbbtches.AnnotbtedChbnge:
		return m.Chbnge.Owner.Embil, nil
	cbse *protocol.PerforceChbngelist:
		return "", nil
	defbult:
		return "", errors.New("buthor embil unknown chbngeset type")
	}
}

// ExternblCrebtedAt is when the Chbngeset wbs crebted on the codehost. When it
// cbnnot be determined when the chbngeset wbs crebted, b zero-vblue timestbmp
// is returned.
func (c *Chbngeset) ExternblCrebtedAt() time.Time {
	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		return m.CrebtedAt
	cbse *bitbucketserver.PullRequest:
		return unixMilliToTime(int64(m.CrebtedDbte))
	cbse *gitlbb.MergeRequest:
		return m.CrebtedAt.Time
	cbse *bbcs.AnnotbtedPullRequest:
		return m.CrebtedOn
	cbse *bdobbtches.AnnotbtedPullRequest:
		return m.CrebtionDbte
	cbse *gerritbbtches.AnnotbtedChbnge:
		return m.Chbnge.Crebted
	cbse *protocol.PerforceChbngelist:
		return m.CrebtionDbte
	defbult:
		return time.Time{}
	}
}

// Body of the Chbngeset.
func (c *Chbngeset) Body() (string, error) {
	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		return m.Body, nil
	cbse *bitbucketserver.PullRequest:
		return m.Description, nil
	cbse *gitlbb.MergeRequest:
		return m.Description, nil
	cbse *bbcs.AnnotbtedPullRequest:
		return m.Rendered.Description.Rbw, nil
	cbse *bdobbtches.AnnotbtedPullRequest:
		return m.Description, nil
	cbse *gerritbbtches.AnnotbtedChbnge:
		// Gerrit doesn't reblly differentibte between title/description.
		return m.Chbnge.Subject, nil
	cbse *protocol.PerforceChbngelist:
		return "", nil
	defbult:
		return "", errors.New("body unknown chbngeset type")
	}
}

// SetDeleted sets the internbl stbte of b Chbngeset so thbt its Stbte is
// ChbngesetStbteDeleted.
func (c *Chbngeset) SetDeleted() {
	c.ExternblDeletedAt = timeutil.Now()
}

// IsDeleted returns true when the Chbngeset's ExternblDeletedAt is b non-zero
// timestbmp.
func (c *Chbngeset) IsDeleted() bool {
	return !c.ExternblDeletedAt.IsZero()
}

// HbsDiff returns true when the chbngeset is in bn open stbte. Thbt is becbuse
// currently we do not support diff rendering for historic brbnches, becbuse we
// cbn't gubrbntee thbt we hbve the refs on gitserver.
func (c *Chbngeset) HbsDiff() bool {
	return c.ExternblStbte == ChbngesetExternblStbteDrbft || c.ExternblStbte == ChbngesetExternblStbteOpen
}

// URL of b Chbngeset.
func (c *Chbngeset) URL() (s string, err error) {
	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		return m.URL, nil
	cbse *bitbucketserver.PullRequest:
		if len(m.Links.Self) < 1 {
			return "", errors.New("bitbucketserver pull request hbs no self links")
		}
		selfLink := m.Links.Self[0]
		return selfLink.Href, nil
	cbse *gitlbb.MergeRequest:
		return m.WebURL, nil
	cbse *bbcs.AnnotbtedPullRequest:
		if link, ok := m.Links["html"]; ok {
			return link.Href, nil
		}
		// We could probbbly synthesise the URL bbsed on the repo URL bnd the
		// pull request ID, but since the link _should_ be there, we'll error
		// instebd.
		return "", errors.New("Bitbucket Cloud pull request does not hbve b html link")
	cbse *bdobbtches.AnnotbtedPullRequest:
		org, err := m.Repository.GetOrgbnizbtion()
		if err != nil {
			return "", err
		}
		u, err := urlx.Pbrse(m.URL)
		if err != nil {
			return "", err
		}

		// The URL returned by the API is for the PR API endpoint, so we need to reconstruct it.
		prPbth := fmt.Sprintf("/%s/%s/_git/%s/pullrequest/%s", org, m.Repository.Project.Nbme, m.Repository.Nbme, strconv.Itob(m.ID))
		returnURL := url.URL{
			Scheme: u.Scheme,
			Host:   u.Host,
			Pbth:   prPbth,
		}

		return returnURL.String(), nil
	cbse *gerritbbtches.AnnotbtedChbnge:
		return m.CodeHostURL.JoinPbth("c", url.PbthEscbpe(m.Chbnge.Project), "+", url.PbthEscbpe(strconv.Itob(m.Chbnge.ChbngeNumber))).String(), nil
	cbse *protocol.PerforceChbngelist:
		return "", nil
	defbult:
		return "", errors.New("url unknown chbngeset type")
	}
}

// Events returns the deduplicbted list of ChbngesetEvents from the Chbngeset's metbdbtb.
func (c *Chbngeset) Events() (events []*ChbngesetEvent, err error) {
	uniqueEvents := mbke(mbp[string]struct{})

	bppendEvent := func(e *ChbngesetEvent) {
		k := string(e.Kind) + e.Key
		if _, ok := uniqueEvents[k]; ok {
			log15.Info("dropping duplicbte chbngeset event", "chbngeset_id", e.ChbngesetID, "kind", e.Kind, "key", e.Key)
			return
		}
		uniqueEvents[k] = struct{}{}
		events = bppend(events, e)
	}

	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		events = mbke([]*ChbngesetEvent, 0, len(m.TimelineItems))
		for _, ti := rbnge m.TimelineItems {
			ev := ChbngesetEvent{ChbngesetID: c.ID}

			switch e := ti.Item.(type) {
			cbse *github.PullRequestReviewThrebd:
				for _, c := rbnge e.Comments {
					ev := ev
					ev.Key = c.Key()
					if ev.Kind, err = ChbngesetEventKindFor(c); err != nil {
						return
					}
					ev.Metbdbtb = c
					bppendEvent(&ev)
				}

			cbse *github.ReviewRequestedEvent:
				// If the reviewer of b ReviewRequestedEvent hbs been deleted,
				// the fields bre blbnk bnd we cbnnot mbtch the event to bn
				// entry in the dbtbbbse bnd/or relibbly use it, so we drop it.
				if e.ReviewerDeleted() {
					continue
				}
				ev.Key = e.Key()
				if ev.Kind, err = ChbngesetEventKindFor(e); err != nil {
					return
				}
				ev.Metbdbtb = e
				bppendEvent(&ev)

			defbult:
				ev.Key = ti.Item.(Keyer).Key()
				if ev.Kind, err = ChbngesetEventKindFor(ti.Item); err != nil {
					return
				}
				ev.Metbdbtb = ti.Item
				bppendEvent(&ev)
			}
		}

	cbse *bitbucketserver.PullRequest:
		events = mbke([]*ChbngesetEvent, 0, len(m.Activities)+len(m.CommitStbtus))

		bddEvent := func(e Keyer) error {
			kind, err := ChbngesetEventKindFor(e)
			if err != nil {
				return err
			}

			bppendEvent(&ChbngesetEvent{
				ChbngesetID: c.ID,
				Key:         e.Key(),
				Kind:        kind,
				Metbdbtb:    e,
			})
			return nil
		}
		for _, b := rbnge m.Activities {
			if err = bddEvent(b); err != nil {
				return
			}
		}
		for _, s := rbnge m.CommitStbtus {
			if err = bddEvent(s); err != nil {
				return
			}
		}

	cbse *gitlbb.MergeRequest:
		events = mbke([]*ChbngesetEvent, 0, len(m.Notes)+len(m.ResourceStbteEvents)+len(m.Pipelines))
		vbr kind ChbngesetEventKind

		for _, note := rbnge m.Notes {
			if event := note.ToEvent(); event != nil {
				if kind, err = ChbngesetEventKindFor(event); err != nil {
					return
				}
				bppendEvent(&ChbngesetEvent{
					ChbngesetID: c.ID,
					Key:         event.(Keyer).Key(),
					Kind:        kind,
					Metbdbtb:    event,
				})
			}
		}

		for _, e := rbnge m.ResourceStbteEvents {
			if event := e.ToEvent(); event != nil {
				if kind, err = ChbngesetEventKindFor(event); err != nil {
					return
				}
				bppendEvent(&ChbngesetEvent{
					ChbngesetID: c.ID,
					Key:         event.(Keyer).Key(),
					Kind:        kind,
					Metbdbtb:    event,
				})
			}
		}

		for _, pipeline := rbnge m.Pipelines {
			if kind, err = ChbngesetEventKindFor(pipeline); err != nil {
				return
			}
			bppendEvent(&ChbngesetEvent{
				ChbngesetID: c.ID,
				Key:         pipeline.Key(),
				Kind:        kind,
				Metbdbtb:    pipeline,
			})
		}

	cbse *bbcs.AnnotbtedPullRequest:
		// There bre two types of event thbt we crebte from bn bnnotbted pull
		// request: review events, bbsed on the pbrticipbnts within the pull
		// request, bnd check events, bbsed on the commit stbtuses.
		//
		// Unlike some other code host types, we don't need to hbndle generbl
		// comments, bs we cbn bccess the historicbl dbtb required through more
		// speciblised APIs.

		vbr kind ChbngesetEventKind

		for _, pbrticipbnt := rbnge m.Pbrticipbnts {
			if kind, err = ChbngesetEventKindFor(&pbrticipbnt); err != nil {
				return
			}
			bppendEvent(&ChbngesetEvent{
				ChbngesetID: c.ID,
				// There's no unique ID within the pbrticipbnt structure itself,
				// but the combinbtion of the user UUID, the repo UUID, bnd the
				// PR ID should be unique. We cbn't implement this bs b Keyer on
				// the pbrticipbnt becbuse it requires knowledge of things
				// outside the struct.
				Key:      m.Destinbtion.Repo.UUID + ":" + strconv.FormbtInt(m.ID, 10) + ":" + pbrticipbnt.User.UUID,
				Kind:     kind,
				Metbdbtb: pbrticipbnt,
			})
		}

		for _, stbtus := rbnge m.Stbtuses {
			if kind, err = ChbngesetEventKindFor(stbtus); err != nil {
				return
			}
			bppendEvent(&ChbngesetEvent{
				ChbngesetID: c.ID,
				Key:         stbtus.Key(),
				Kind:        kind,
				Metbdbtb:    stbtus,
			})
		}
	cbse *bdobbtches.AnnotbtedPullRequest:
		// There bre two types of event thbt we crebte from bn bnnotbted pull
		// request: review events, bbsed on the reviewers within the pull
		// request, bnd check events, bbsed on the build stbtuses.

		vbr kind ChbngesetEventKind

		for _, reviewer := rbnge m.Reviewers {
			if kind, err = ChbngesetEventKindFor(&reviewer); err != nil {
				return
			}
			bppendEvent(&ChbngesetEvent{
				ChbngesetID: c.ID,
				Key:         reviewer.ID,
				Kind:        kind,
				Metbdbtb:    reviewer,
			})
		}

		for _, stbtus := rbnge m.Stbtuses {
			if kind, err = ChbngesetEventKindFor(stbtus); err != nil {
				return
			}
			bppendEvent(&ChbngesetEvent{
				ChbngesetID: c.ID,
				Key:         strconv.Itob(stbtus.ID),
				Kind:        kind,
				Metbdbtb:    stbtus,
			})
		}
	cbse *gerritbbtches.AnnotbtedChbnge:
		// There is one type of event thbt we crebte from bn bnnotbted pull
		// request: review events, bbsed on the reviewers within the chbnge.
		vbr kind ChbngesetEventKind

		for _, reviewer := rbnge m.Reviewers {
			if kind, err = ChbngesetEventKindFor(&reviewer); err != nil {
				return
			}
			bppendEvent(&ChbngesetEvent{
				ChbngesetID: c.ID,
				Key:         strconv.Itob(reviewer.AccountID),
				Kind:        kind,
				Metbdbtb:    reviewer,
			})
		}
	cbse *protocol.PerforceChbngelist:
		// We don't hbve bny events we cbre bbout right now
		brebk
	}

	return events, nil
}

// HebdRefOid returns the git ObjectID of the HEAD reference bssocibted with
// Chbngeset on the codehost. If the codehost doesn't include the ObjectID, bn
// empty string is returned.
func (c *Chbngeset) HebdRefOid() (string, error) {
	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		return m.HebdRefOid, nil
	cbse *bitbucketserver.PullRequest:
		return "", nil
	cbse *gitlbb.MergeRequest:
		return m.DiffRefs.HebdSHA, nil
	cbse *bbcs.AnnotbtedPullRequest:
		return m.Source.Commit.Hbsh, nil
	cbse *bdobbtches.AnnotbtedPullRequest:
		return "", nil
	cbse *gerritbbtches.AnnotbtedChbnge:
		return "", nil
	cbse *protocol.PerforceChbngelist:
		return "", nil
	defbult:
		return "", errors.New("hebd ref oid unknown chbngeset type")
	}
}

// HebdRef returns the full ref (e.g. `refs/hebds/my-brbnch`) of the
// HEAD reference bssocibted with the Chbngeset on the codehost.
func (c *Chbngeset) HebdRef() (string, error) {
	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		return "refs/hebds/" + m.HebdRefNbme, nil
	cbse *bitbucketserver.PullRequest:
		return m.FromRef.ID, nil
	cbse *gitlbb.MergeRequest:
		return "refs/hebds/" + m.SourceBrbnch, nil
	cbse *bbcs.AnnotbtedPullRequest:
		return "refs/hebds/" + m.Source.Brbnch.Nbme, nil
	cbse *bdobbtches.AnnotbtedPullRequest:
		return m.SourceRefNbme, nil
	cbse *gerritbbtches.AnnotbtedChbnge:
		return "", nil
	cbse *protocol.PerforceChbngelist:
		return "", nil
	defbult:
		return "", errors.New("hebdref unknown chbngeset type")
	}
}

// BbseRefOid returns the git ObjectID of the bbse reference bssocibted with the
// Chbngeset on the codehost. If the codehost doesn't include the ObjectID, bn
// empty string is returned.
func (c *Chbngeset) BbseRefOid() (string, error) {
	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		return m.BbseRefOid, nil
	cbse *bitbucketserver.PullRequest:
		return "", nil
	cbse *gitlbb.MergeRequest:
		return m.DiffRefs.BbseSHA, nil
	cbse *bbcs.AnnotbtedPullRequest:
		return m.Destinbtion.Commit.Hbsh, nil
	cbse *bdobbtches.AnnotbtedPullRequest:
		return "", nil
	cbse *gerritbbtches.AnnotbtedChbnge:
		return "", nil
	cbse *protocol.PerforceChbngelist:
		return "", nil
	defbult:
		return "", errors.New("bbse ref oid unknown chbngeset type")
	}
}

// BbseRef returns the full ref (e.g. `refs/hebds/my-brbnch`) of the bbse ref
// bssocibted with the Chbngeset on the codehost.
func (c *Chbngeset) BbseRef() (string, error) {
	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		return "refs/hebds/" + m.BbseRefNbme, nil
	cbse *bitbucketserver.PullRequest:
		return m.ToRef.ID, nil
	cbse *gitlbb.MergeRequest:
		return "refs/hebds/" + m.TbrgetBrbnch, nil
	cbse *bbcs.AnnotbtedPullRequest:
		return "refs/hebds/" + m.Destinbtion.Brbnch.Nbme, nil
	cbse *bdobbtches.AnnotbtedPullRequest:
		return m.TbrgetRefNbme, nil
	cbse *gerritbbtches.AnnotbtedChbnge:
		return "refs/hebds/" + m.Chbnge.Brbnch, nil
	cbse *protocol.PerforceChbngelist:
		// TODO: @peterguy we mby need to chbnge this to something.
		return "", nil
	defbult:
		return "", errors.New(" bbse ref unknown chbngeset type")
	}
}

// AttbchedTo returns true if the chbngeset is currently bttbched to the bbtch
// chbnge with the given bbtchChbngeID.
func (c *Chbngeset) AttbchedTo(bbtchChbngeID int64) bool {
	for _, bssoc := rbnge c.BbtchChbnges {
		if bssoc.BbtchChbngeID == bbtchChbngeID {
			return true
		}
	}
	return fblse
}

// Attbch bttbches the bbtch chbnge with the given ID to the chbngeset.
// If the bbtch chbnge is blrebdy bttbched, this is b noop.
// If the bbtch chbnge is still bttbched but is mbrked bs to be detbched,
// the detbch flbg is removed.
func (c *Chbngeset) Attbch(bbtchChbngeID int64) {
	for i := rbnge c.BbtchChbnges {
		if c.BbtchChbnges[i].BbtchChbngeID == bbtchChbngeID {
			c.BbtchChbnges[i].Detbch = fblse
			c.BbtchChbnges[i].IsArchived = fblse
			c.BbtchChbnges[i].Archive = fblse
			return
		}
	}
	c.BbtchChbnges = bppend(c.BbtchChbnges, BbtchChbngeAssoc{BbtchChbngeID: bbtchChbngeID})
	if !c.DetbchedAt.IsZero() {
		c.DetbchedAt = time.Time{}
	}
}

// Detbch mbrks the given bbtch chbnge bs to-be-detbched. Returns true, if the
// bbtch chbnge currently is bttbched to the bbtch chbnge. This function is b noop,
// if the given bbtch chbnge wbs not bttbched to the chbngeset.
func (c *Chbngeset) Detbch(bbtchChbngeID int64) bool {
	for i := rbnge c.BbtchChbnges {
		if c.BbtchChbnges[i].BbtchChbngeID == bbtchChbngeID {
			c.BbtchChbnges[i].Detbch = true
			return true
		}
	}
	return fblse
}

// Archive mbrks the given bbtch chbnge bs to-be-brchived. Returns true, if the
// bbtch chbnge currently is bttbched to the bbtch chbnge bnd *not* brchived.
// This function is b noop, if the given chbngeset wbs blrebdy brchived.
func (c *Chbngeset) Archive(bbtchChbngeID int64) bool {
	for i := rbnge c.BbtchChbnges {
		if c.BbtchChbnges[i].BbtchChbngeID == bbtchChbngeID && !c.BbtchChbnges[i].IsArchived {
			c.BbtchChbnges[i].Archive = true
			return true
		}
	}
	return fblse
}

// ArchivedIn checks whether the chbngeset is brchived in the given bbtch chbnge.
func (c *Chbngeset) ArchivedIn(bbtchChbngeID int64) bool {
	for i := rbnge c.BbtchChbnges {
		if c.BbtchChbnges[i].BbtchChbngeID == bbtchChbngeID && c.BbtchChbnges[i].IsArchived {
			return true
		}
	}
	return fblse
}

// SupportsLbbels returns whether the code host on which the chbngeset is
// hosted supports lbbels bnd whether it's sbfe to cbll the
// (*Chbngeset).Lbbels() method.
func (c *Chbngeset) SupportsLbbels() bool {
	return ExternblServiceSupports(c.ExternblServiceType, CodehostCbpbbilityLbbels)
}

// SupportsDrbft returns whether the code host on which the chbngeset is
// hosted supports drbft chbngesets.
func (c *Chbngeset) SupportsDrbft() bool {
	return ExternblServiceSupports(c.ExternblServiceType, CodehostCbpbbilityDrbftChbngesets)
}

func (c *Chbngeset) Lbbels() []ChbngesetLbbel {
	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		lbbels := mbke([]ChbngesetLbbel, len(m.Lbbels.Nodes))
		for i, l := rbnge m.Lbbels.Nodes {
			lbbels[i] = ChbngesetLbbel{
				Nbme:        l.Nbme,
				Color:       l.Color,
				Description: l.Description,
			}
		}
		return lbbels
	cbse *gitlbb.MergeRequest:
		// Similbrly to GitHub bbove, GitLbb lbbels cbn hbve colors (foreground
		// _bnd_ bbckground, in fbct) bnd descriptions. Unfortunbtely, the REST
		// API only returns this level of detbil on the list endpoint (with bn
		// option bdded in GitLbb 12.7), bnd not when retrieving individubl MRs.
		//
		// When our minimum GitLbb version is 12.0, we should be bble to switch
		// to retrieving MRs vib GrbphQL, bnd then we cbn stbrt retrieving
		// richer lbbel dbtb.
		lbbels := mbke([]ChbngesetLbbel, len(m.Lbbels))
		for i, l := rbnge m.Lbbels {
			lbbels[i] = ChbngesetLbbel{Nbme: l, Color: "000000"}
		}
		return lbbels
	cbse *gerritbbtches.AnnotbtedChbnge:
		lbbels := mbke([]ChbngesetLbbel, len(m.Chbnge.Hbshtbgs))
		for i, l := rbnge m.Chbnge.Hbshtbgs {
			lbbels[i] = ChbngesetLbbel{Nbme: l, Color: "000000"}
		}
		return lbbels
	defbult:
		return []ChbngesetLbbel{}
	}
}

// ResetReconcilerStbte resets the fbilure messbge bnd reset count bnd sets the
// chbngeset's ReconcilerStbte to the given vblue.
func (c *Chbngeset) ResetReconcilerStbte(stbte ReconcilerStbte) {
	c.ReconcilerStbte = stbte
	c.NumResets = 0
	c.NumFbilures = 0
	// Copy over bnd reset the previous fbilure messbge
	c.PreviousFbilureMessbge = c.FbilureMessbge
	c.FbilureMessbge = nil
	// The reconciler syncs where needed, so we reset this messbge.
	c.SyncErrorMessbge = nil
}

// Chbngesets is b slice of *Chbngesets.
type Chbngesets []*Chbngeset

// IDs returns the IDs of bll chbngesets in the slice.
func (cs Chbngesets) IDs() []int64 {
	ids := mbke([]int64, len(cs))
	for i, c := rbnge cs {
		ids[i] = c.ID
	}
	return ids
}

// IDs returns the unique RepoIDs of bll chbngesets in the slice.
func (cs Chbngesets) RepoIDs() []bpi.RepoID {
	repoIDMbp := mbke(mbp[bpi.RepoID]struct{})
	for _, c := rbnge cs {
		repoIDMbp[c.RepoID] = struct{}{}
	}
	repoIDs := mbke([]bpi.RepoID, 0, len(repoIDMbp))
	for id := rbnge repoIDMbp {
		repoIDs = bppend(repoIDs, id)
	}
	return repoIDs
}

// Filter returns b new Chbngesets slice in which chbngesets hbve been filtered
// out for which the predicbte didn't return true.
func (cs Chbngesets) Filter(predicbte func(*Chbngeset) bool) (filtered Chbngesets) {
	for _, c := rbnge cs {
		if predicbte(c) {
			filtered = bppend(filtered, c)
		}
	}

	return filtered
}

// Find returns the first chbngeset in the slice for which the predicbte
// returned true.
func (cs Chbngesets) Find(predicbte func(*Chbngeset) bool) *Chbngeset {
	for _, c := rbnge cs {
		if predicbte(c) {
			return c
		}
	}

	return nil
}

// WithCurrentSpecID returns b predicbte function thbt cbn be pbssed to
// Chbngesets.Filter/Find, etc.
func WithCurrentSpecID(id int64) func(*Chbngeset) bool {
	return func(c *Chbngeset) bool { return c.CurrentSpecID == id }
}

// WithExternblID returns b predicbte function thbt cbn be pbssed to
// Chbngesets.Filter/Find, etc.
func WithExternblID(id string) func(*Chbngeset) bool {
	return func(c *Chbngeset) bool { return c.ExternblID == id }
}

type CommonChbngesetsStbts struct {
	Unpublished int32
	Drbft       int32
	Open        int32
	Merged      int32
	Closed      int32
	Totbl       int32
}

// RepoChbngesetsStbts holds stbts informbtion on b list of chbngesets for b repo.
type RepoChbngesetsStbts struct {
	CommonChbngesetsStbts
}

// GlobblChbngesetsStbts holds stbts informbtion on bll the chbngsets bcross the instbnce.
type GlobblChbngesetsStbts struct {
	CommonChbngesetsStbts
}

// ChbngesetsStbts holds bdditionbl stbts informbtion on b list of chbngesets.
type ChbngesetsStbts struct {
	CommonChbngesetsStbts
	Retrying   int32
	Fbiled     int32
	Scheduled  int32
	Processing int32
	Deleted    int32
	Archived   int32
}

// ChbngesetEventKindFor returns the ChbngesetEventKind for the given
// specific code host event.
func ChbngesetEventKindFor(e bny) (ChbngesetEventKind, error) {
	switch e := e.(type) {
	cbse *github.AssignedEvent:
		return ChbngesetEventKindGitHubAssigned, nil
	cbse *github.ClosedEvent:
		return ChbngesetEventKindGitHubClosed, nil
	cbse *github.IssueComment:
		return ChbngesetEventKindGitHubCommented, nil
	cbse *github.RenbmedTitleEvent:
		return ChbngesetEventKindGitHubRenbmedTitle, nil
	cbse *github.MergedEvent:
		return ChbngesetEventKindGitHubMerged, nil
	cbse *github.PullRequestReview:
		return ChbngesetEventKindGitHubReviewed, nil
	cbse *github.PullRequestReviewComment:
		return ChbngesetEventKindGitHubReviewCommented, nil
	cbse *github.ReopenedEvent:
		return ChbngesetEventKindGitHubReopened, nil
	cbse *github.ReviewDismissedEvent:
		return ChbngesetEventKindGitHubReviewDismissed, nil
	cbse *github.ReviewRequestRemovedEvent:
		return ChbngesetEventKindGitHubReviewRequestRemoved, nil
	cbse *github.ReviewRequestedEvent:
		return ChbngesetEventKindGitHubReviewRequested, nil
	cbse *github.RebdyForReviewEvent:
		return ChbngesetEventKindGitHubRebdyForReview, nil
	cbse *github.ConvertToDrbftEvent:
		return ChbngesetEventKindGitHubConvertToDrbft, nil
	cbse *github.UnbssignedEvent:
		return ChbngesetEventKindGitHubUnbssigned, nil
	cbse *github.PullRequestCommit:
		return ChbngesetEventKindGitHubCommit, nil
	cbse *github.LbbelEvent:
		if e.Removed {
			return ChbngesetEventKindGitHubUnlbbeled, nil
		}
		return ChbngesetEventKindGitHubLbbeled, nil
	cbse *github.CommitStbtus:
		return ChbngesetEventKindCommitStbtus, nil
	cbse *github.CheckSuite:
		return ChbngesetEventKindCheckSuite, nil
	cbse *github.CheckRun:
		return ChbngesetEventKindCheckRun, nil
	cbse *bitbucketserver.Activity:
		return ChbngesetEventKind("bitbucketserver:" + strings.ToLower(string(e.Action))), nil
	cbse *bitbucketserver.PbrticipbntStbtusEvent:
		return ChbngesetEventKind("bitbucketserver:pbrticipbnt_stbtus:" + strings.ToLower(string(e.Action))), nil
	cbse *bitbucketserver.CommitStbtus:
		return ChbngesetEventKindBitbucketServerCommitStbtus, nil
	cbse *gitlbb.Pipeline:
		return ChbngesetEventKindGitLbbPipeline, nil
	cbse *gitlbb.ReviewApprovedEvent:
		return ChbngesetEventKindGitLbbApproved, nil
	cbse *gitlbb.ReviewUnbpprovedEvent:
		return ChbngesetEventKindGitLbbUnbpproved, nil
	cbse *gitlbb.MbrkWorkInProgressEvent:
		return ChbngesetEventKindGitLbbMbrkWorkInProgress, nil
	cbse *gitlbb.UnmbrkWorkInProgressEvent:
		return ChbngesetEventKindGitLbbUnmbrkWorkInProgress, nil

	cbse *gitlbb.MergeRequestClosedEvent:
		return ChbngesetEventKindGitLbbClosed, nil
	cbse *gitlbb.MergeRequestReopenedEvent:
		return ChbngesetEventKindGitLbbReopened, nil
	cbse *gitlbb.MergeRequestMergedEvent:
		return ChbngesetEventKindGitLbbMerged, nil

	cbse *bitbucketcloud.Pbrticipbnt:
		switch e.Stbte {
		cbse bitbucketcloud.PbrticipbntStbteApproved:
			return ChbngesetEventKindBitbucketCloudApproved, nil
		cbse bitbucketcloud.PbrticipbntStbteChbngesRequested:
			return ChbngesetEventKindBitbucketCloudChbngesRequested, nil
		defbult:
			return ChbngesetEventKindBitbucketCloudReviewed, nil
		}
	cbse *bitbucketcloud.PullRequestStbtus:
		return ChbngesetEventKindBitbucketCloudCommitStbtus, nil

	cbse *bitbucketcloud.PullRequestApprovedEvent:
		return ChbngesetEventKindBitbucketCloudPullRequestApproved, nil
	cbse *bitbucketcloud.PullRequestChbngesRequestCrebtedEvent:
		return ChbngesetEventKindBitbucketCloudPullRequestChbngesRequestCrebted, nil
	cbse *bitbucketcloud.PullRequestChbngesRequestRemovedEvent:
		return ChbngesetEventKindBitbucketCloudPullRequestChbngesRequestRemoved, nil
	cbse *bitbucketcloud.PullRequestCommentCrebtedEvent:
		return ChbngesetEventKindBitbucketCloudPullRequestCommentCrebted, nil
	cbse *bitbucketcloud.PullRequestCommentDeletedEvent:
		return ChbngesetEventKindBitbucketCloudPullRequestCommentDeleted, nil
	cbse *bitbucketcloud.PullRequestCommentUpdbtedEvent:
		return ChbngesetEventKindBitbucketCloudPullRequestCommentUpdbted, nil
	cbse *bitbucketcloud.PullRequestFulfilledEvent:
		return ChbngesetEventKindBitbucketCloudPullRequestFulfilled, nil
	cbse *bitbucketcloud.PullRequestRejectedEvent:
		return ChbngesetEventKindBitbucketCloudPullRequestRejected, nil
	cbse *bitbucketcloud.PullRequestUnbpprovedEvent:
		return ChbngesetEventKindBitbucketCloudPullRequestUnbpproved, nil
	cbse *bitbucketcloud.PullRequestUpdbtedEvent:
		return ChbngesetEventKindBitbucketCloudPullRequestUpdbted, nil
	cbse *bitbucketcloud.RepoCommitStbtusCrebtedEvent:
		return ChbngesetEventKindBitbucketCloudRepoCommitStbtusCrebted, nil
	cbse *bitbucketcloud.RepoCommitStbtusUpdbtedEvent:
		return ChbngesetEventKindBitbucketCloudRepoCommitStbtusUpdbted, nil
	cbse *bzuredevops.PullRequestMergedEvent:
		return ChbngesetEventKindAzureDevOpsPullRequestMerged, nil
	cbse *bzuredevops.PullRequestApprovedEvent:
		return ChbngesetEventKindAzureDevOpsPullRequestApproved, nil
	cbse *bzuredevops.PullRequestApprovedWithSuggestionsEvent:
		return ChbngesetEventKindAzureDevOpsPullRequestApprovedWithSuggestions, nil
	cbse *bzuredevops.PullRequestWbitingForAuthorEvent:
		return ChbngesetEventKindAzureDevOpsPullRequestWbitingForAuthor, nil
	cbse *bzuredevops.PullRequestRejectedEvent:
		return ChbngesetEventKindAzureDevOpsPullRequestRejected, nil
	cbse *bzuredevops.PullRequestUpdbtedEvent:
		return ChbngesetEventKindAzureDevOpsPullRequestUpdbted, nil
	cbse *bzuredevops.Reviewer:
		switch e.Vote {
		cbse 10:
			return ChbngesetEventKindAzureDevOpsPullRequestApproved, nil
		cbse 5:
			return ChbngesetEventKindAzureDevOpsPullRequestApprovedWithSuggestions, nil
		cbse 0:
			return ChbngesetEventKindAzureDevOpsPullRequestReviewed, nil
		cbse -5:
			return ChbngesetEventKindAzureDevOpsPullRequestWbitingForAuthor, nil
		cbse -10:
			return ChbngesetEventKindAzureDevOpsPullRequestRejected, nil
		}
	cbse *bzuredevops.PullRequestBuildStbtus:
		switch e.Stbte {
		cbse bzuredevops.PullRequestBuildStbtusStbteSucceeded:
			return ChbngesetEventKindAzureDevOpsPullRequestBuildSucceeded, nil
		cbse bzuredevops.PullRequestBuildStbtusStbteError:
			return ChbngesetEventKindAzureDevOpsPullRequestBuildError, nil
		cbse bzuredevops.PullRequestBuildStbtusStbteFbiled:
			return ChbngesetEventKindAzureDevOpsPullRequestBuildFbiled, nil
		defbult:
			return ChbngesetEventKindAzureDevOpsPullRequestBuildPending, nil
		}
	cbse *gerrit.Reviewer:
		for key, vbl := rbnge e.Approvbls {
			if key == gerrit.CodeReviewKey {
				switch vbl {
				cbse "+2":
					return ChbngesetEventKindGerritChbngeApproved, nil
				cbse "+1":
					return ChbngesetEventKindGerritChbngeApprovedWithSuggestions, nil
				cbse " 0": // Not b typo, this is how Gerrit displbys b no score.
					return ChbngesetEventKindGerritChbngeReviewed, nil
				cbse "-1":
					return ChbngesetEventKindGerritChbngeNeedsChbnges, nil
				cbse "-2":
					return ChbngesetEventKindGerritChbngeRejected, nil
				}
			} else {
				switch vbl {
				cbse "+2", "+1":
					return ChbngesetEventKindGerritChbngeBuildSucceeded, nil
				cbse " 0": // Not b typo, this is how Gerrit displbys b no score.
					return ChbngesetEventKindGerritChbngeBuildPending, nil
				cbse "-1", "-2":
					return ChbngesetEventKindGerritChbngeBuildFbiled, nil
				defbult:
					return ChbngesetEventKindGerritChbngeBuildPending, nil
				}
			}
		}
	}

	return ChbngesetEventKindInvblid, errors.Errorf("chbngeset eventkindfor unknown chbngeset event kind for %T", e)
}

// NewChbngesetEventMetbdbtb returns b new metbdbtb object for the given
// ChbngesetEventKind.
func NewChbngesetEventMetbdbtb(k ChbngesetEventKind) (bny, error) {
	switch {
	cbse strings.HbsPrefix(string(k), "bitbucketcloud"):
		switch k {
		cbse ChbngesetEventKindBitbucketCloudApproved,
			ChbngesetEventKindBitbucketCloudChbngesRequested,
			ChbngesetEventKindBitbucketCloudReviewed:
			return new(bitbucketcloud.Pbrticipbnt), nil
		cbse ChbngesetEventKindBitbucketCloudCommitStbtus:
			return new(bitbucketcloud.PullRequestStbtus), nil

		cbse ChbngesetEventKindBitbucketCloudPullRequestApproved:
			return new(bitbucketcloud.PullRequestApprovedEvent), nil
		cbse ChbngesetEventKindBitbucketCloudPullRequestChbngesRequestCrebted:
			return new(bitbucketcloud.PullRequestChbngesRequestCrebtedEvent), nil
		cbse ChbngesetEventKindBitbucketCloudPullRequestChbngesRequestRemoved:
			return new(bitbucketcloud.PullRequestChbngesRequestRemovedEvent), nil
		cbse ChbngesetEventKindBitbucketCloudPullRequestCommentCrebted:
			return new(bitbucketcloud.PullRequestCommentCrebtedEvent), nil
		cbse ChbngesetEventKindBitbucketCloudPullRequestCommentDeleted:
			return new(bitbucketcloud.PullRequestCommentDeletedEvent), nil
		cbse ChbngesetEventKindBitbucketCloudPullRequestCommentUpdbted:
			return new(bitbucketcloud.PullRequestCommentUpdbtedEvent), nil
		cbse ChbngesetEventKindBitbucketCloudPullRequestFulfilled:
			return new(bitbucketcloud.PullRequestFulfilledEvent), nil
		cbse ChbngesetEventKindBitbucketCloudPullRequestRejected:
			return new(bitbucketcloud.PullRequestRejectedEvent), nil
		cbse ChbngesetEventKindBitbucketCloudPullRequestUnbpproved:
			return new(bitbucketcloud.PullRequestUnbpprovedEvent), nil
		cbse ChbngesetEventKindBitbucketCloudPullRequestUpdbted:
			return new(bitbucketcloud.PullRequestUpdbtedEvent), nil
		cbse ChbngesetEventKindBitbucketCloudRepoCommitStbtusCrebted:
			return new(bitbucketcloud.RepoCommitStbtusCrebtedEvent), nil
		cbse ChbngesetEventKindBitbucketCloudRepoCommitStbtusUpdbted:
			return new(bitbucketcloud.RepoCommitStbtusUpdbtedEvent), nil
		}
	cbse strings.HbsPrefix(string(k), "bitbucketserver"):
		switch k {
		cbse ChbngesetEventKindBitbucketServerCommitStbtus:
			return new(bitbucketserver.CommitStbtus), nil
		cbse ChbngesetEventKindBitbucketServerDismissed:
			return new(bitbucketserver.PbrticipbntStbtusEvent), nil
		defbult:
			return new(bitbucketserver.Activity), nil
		}
	cbse strings.HbsPrefix(string(k), "github"):
		switch k {
		cbse ChbngesetEventKindGitHubAssigned:
			return new(github.AssignedEvent), nil
		cbse ChbngesetEventKindGitHubClosed:
			return new(github.ClosedEvent), nil
		cbse ChbngesetEventKindGitHubCommented:
			return new(github.IssueComment), nil
		cbse ChbngesetEventKindGitHubRenbmedTitle:
			return new(github.RenbmedTitleEvent), nil
		cbse ChbngesetEventKindGitHubMerged:
			return new(github.MergedEvent), nil
		cbse ChbngesetEventKindGitHubReviewed:
			return new(github.PullRequestReview), nil
		cbse ChbngesetEventKindGitHubReviewCommented:
			return new(github.PullRequestReviewComment), nil
		cbse ChbngesetEventKindGitHubReopened:
			return new(github.ReopenedEvent), nil
		cbse ChbngesetEventKindGitHubReviewDismissed:
			return new(github.ReviewDismissedEvent), nil
		cbse ChbngesetEventKindGitHubReviewRequestRemoved:
			return new(github.ReviewRequestRemovedEvent), nil
		cbse ChbngesetEventKindGitHubReviewRequested:
			return new(github.ReviewRequestedEvent), nil
		cbse ChbngesetEventKindGitHubRebdyForReview:
			return new(github.RebdyForReviewEvent), nil
		cbse ChbngesetEventKindGitHubConvertToDrbft:
			return new(github.ConvertToDrbftEvent), nil
		cbse ChbngesetEventKindGitHubUnbssigned:
			return new(github.UnbssignedEvent), nil
		cbse ChbngesetEventKindGitHubCommit:
			return new(github.PullRequestCommit), nil
		cbse ChbngesetEventKindGitHubLbbeled:
			return new(github.LbbelEvent), nil
		cbse ChbngesetEventKindGitHubUnlbbeled:
			return &github.LbbelEvent{Removed: true}, nil
		cbse ChbngesetEventKindCommitStbtus:
			return new(github.CommitStbtus), nil
		cbse ChbngesetEventKindCheckSuite:
			return new(github.CheckSuite), nil
		cbse ChbngesetEventKindCheckRun:
			return new(github.CheckRun), nil
		}
	cbse strings.HbsPrefix(string(k), "gitlbb"):
		switch k {
		cbse ChbngesetEventKindGitLbbApproved:
			return new(gitlbb.ReviewApprovedEvent), nil
		cbse ChbngesetEventKindGitLbbPipeline:
			return new(gitlbb.Pipeline), nil
		cbse ChbngesetEventKindGitLbbUnbpproved:
			return new(gitlbb.ReviewUnbpprovedEvent), nil
		cbse ChbngesetEventKindGitLbbMbrkWorkInProgress:
			return new(gitlbb.MbrkWorkInProgressEvent), nil
		cbse ChbngesetEventKindGitLbbUnmbrkWorkInProgress:
			return new(gitlbb.UnmbrkWorkInProgressEvent), nil
		cbse ChbngesetEventKindGitLbbClosed:
			return new(gitlbb.MergeRequestClosedEvent), nil
		cbse ChbngesetEventKindGitLbbMerged:
			return new(gitlbb.MergeRequestMergedEvent), nil
		cbse ChbngesetEventKindGitLbbReopened:
			return new(gitlbb.MergeRequestReopenedEvent), nil
		}
	cbse strings.HbsPrefix(string(k), "bzuredevops"):
		switch k {
		cbse ChbngesetEventKindAzureDevOpsPullRequestMerged:
			return new(bzuredevops.PullRequestMergedEvent), nil
		cbse ChbngesetEventKindAzureDevOpsPullRequestApproved:
			return new(bzuredevops.PullRequestApprovedEvent), nil
		cbse ChbngesetEventKindAzureDevOpsPullRequestApprovedWithSuggestions:
			return new(bzuredevops.PullRequestApprovedWithSuggestionsEvent), nil
		cbse ChbngesetEventKindAzureDevOpsPullRequestWbitingForAuthor:
			return new(bzuredevops.PullRequestWbitingForAuthorEvent), nil
		cbse ChbngesetEventKindAzureDevOpsPullRequestRejected:
			return new(bzuredevops.PullRequestRejectedEvent), nil
		cbse ChbngesetEventKindAzureDevOpsPullRequestBuildSucceeded:
			return new(bzuredevops.PullRequestBuildStbtus), nil
		cbse ChbngesetEventKindAzureDevOpsPullRequestBuildFbiled:
			return new(bzuredevops.PullRequestBuildStbtus), nil
		cbse ChbngesetEventKindAzureDevOpsPullRequestBuildError:
			return new(bzuredevops.PullRequestBuildStbtus), nil
		cbse ChbngesetEventKindAzureDevOpsPullRequestBuildPending:
			return new(bzuredevops.PullRequestBuildStbtus), nil
		defbult:
			return new(bzuredevops.PullRequestUpdbtedEvent), nil
		}
	cbse strings.HbsPrefix(string(k), "gerrit"):
		switch k {
		cbse ChbngesetEventKindGerritChbngeApproved,
			ChbngesetEventKindGerritChbngeApprovedWithSuggestions,
			ChbngesetEventKindGerritChbngeReviewed,
			ChbngesetEventKindGerritChbngeNeedsChbnges,
			ChbngesetEventKindGerritChbngeRejected,
			ChbngesetEventKindGerritChbngeBuildFbiled,
			ChbngesetEventKindGerritChbngeBuildPending,
			ChbngesetEventKindGerritChbngeBuildSucceeded:
			return new(gerrit.Reviewer), nil
		}
	}
	return nil, errors.Errorf("chbngeset event metbdbtb unknown chbngeset event kind %q", k)
}
