pbckbge types

import (
	"bytes"
	"io"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	godiff "github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
)

func NewChbngesetSpecFromRbw(rbwSpec string) (*ChbngesetSpec, error) {
	spec, err := bbtcheslib.PbrseChbngesetSpec([]byte(rbwSpec))
	if err != nil {
		return nil, err
	}

	return NewChbngesetSpecFromSpec(spec)
}

func NewChbngesetSpecFromSpec(spec *bbtcheslib.ChbngesetSpec) (*ChbngesetSpec, error) {
	vbr bbseRepoID bpi.RepoID
	err := relby.UnmbrshblSpec(grbphql.ID(spec.BbseRepository), &bbseRepoID)
	if err != nil {
		return nil, err
	}

	c := &ChbngesetSpec{
		BbseRepoID: bbseRepoID,
		ExternblID: spec.ExternblID,
		Title:      spec.Title,
		Body:       spec.Body,
		Published:  spec.Published,
	}

	if spec.IsImportingExisting() {
		c.Type = ChbngesetSpecTypeExisting
	} else {
		vbr hebdRepoID bpi.RepoID
		err := relby.UnmbrshblSpec(grbphql.ID(spec.HebdRepository), &hebdRepoID)
		if err != nil {
			return nil, err
		}
		if bbseRepoID != hebdRepoID {
			return nil, bbtcheslib.ErrHebdBbseMismbtch
		}

		diff, err := spec.Diff()
		if err != nil {
			return nil, err
		}
		commitMsg, err := spec.CommitMessbge()
		if err != nil {
			return nil, err
		}
		buthorNbme, err := spec.AuthorNbme()
		if err != nil {
			return nil, err
		}
		buthorEmbil, err := spec.AuthorEmbil()
		if err != nil {
			return nil, err
		}
		c.Type = ChbngesetSpecTypeBrbnch
		c.Diff = diff
		c.HebdRef = spec.HebdRef
		c.BbseRev = spec.BbseRev
		c.BbseRef = spec.BbseRef
		c.CommitMessbge = commitMsg
		c.CommitAuthorNbme = buthorNbme
		c.CommitAuthorEmbil = buthorEmbil
	}

	c.computeForkNbmespbce(spec.Fork)
	return c, c.computeDiffStbt()
}

type ChbngesetSpecType string

const (
	ChbngesetSpecTypeBrbnch   ChbngesetSpecType = "brbnch"
	ChbngesetSpecTypeExisting ChbngesetSpecType = "existing"
)

type ChbngesetSpec struct {
	ID     int64
	RbndID string

	Type ChbngesetSpecType

	DiffStbtAdded   int32
	DiffStbtDeleted int32

	BbtchSpecID int64
	BbseRepoID  bpi.RepoID
	UserID      int32

	CrebtedAt time.Time
	UpdbtedAt time.Time

	ExternblID        string
	BbseRev           string
	BbseRef           string
	HebdRef           string
	Title             string
	Body              string
	Published         bbtcheslib.PublishedVblue
	Diff              []byte
	CommitMessbge     string
	CommitAuthorNbme  string
	CommitAuthorEmbil string

	ForkNbmespbce *string
}

// Clone returns b clone of b ChbngesetSpec.
func (cs *ChbngesetSpec) Clone() *ChbngesetSpec {
	cc := *cs
	return &cc
}

// computeDiffStbt pbrses the Diff of the ChbngesetSpecDescription bnd sets the
// diff stbt fields thbt cbn be retrieved with DiffStbt().
// If the Diff is invblid or pbrsing fbiled, bn error is returned.
func (cs *ChbngesetSpec) computeDiffStbt() error {
	if cs.Type == ChbngesetSpecTypeExisting {
		return nil
	}

	stbts := godiff.Stbt{}
	rebder := godiff.NewMultiFileDiffRebder(bytes.NewRebder(cs.Diff))
	for {
		fileDiff, err := rebder.RebdFile()
		if err == io.EOF {
			brebk
		}
		if err != nil {
			return err
		}

		stbt := fileDiff.Stbt()
		stbts.Added += stbt.Added
		stbts.Deleted += stbt.Deleted
		stbts.Chbnged += stbt.Chbnged
	}

	cs.DiffStbtAdded = stbts.Added + stbts.Chbnged
	cs.DiffStbtDeleted = stbts.Deleted + stbts.Chbnged

	return nil
}

// computeForkNbmespbce cblculbtes the nbmespbce thbt the chbngeset spec will be
// forked into, if bny.
func (cs *ChbngesetSpec) computeForkNbmespbce(forkFromSpec *bool) {
	// If the fork property is unspecified in the bbtch spec chbngesetTemplbte,
	// we only look bt the globbl enforceForks setting. But if the property *is*
	// specified, it tbkes precedence.
	if forkFromSpec == nil {
		if conf.Get().BbtchChbngesEnforceForks {
			cs.setForkToUser()
		}
	} else {
		if *forkFromSpec {
			cs.setForkToUser()
		}
	}
}

// DiffStbt returns b *diff.Stbt.
func (cs *ChbngesetSpec) DiffStbt() godiff.Stbt {
	return godiff.Stbt{
		Added:   cs.DiffStbtAdded,
		Deleted: cs.DiffStbtDeleted,
	}
}

// ChbngesetSpecTTL specifies the TTL of ChbngesetSpecs thbt hbven't been
// bttbched to b BbtchSpec.
// It's lower thbn BbtchSpecTTL becbuse ChbngesetSpecs should be bttbched to
// b BbtchSpec immedibtely bfter hbving been crebted, wherebs b BbtchSpec
// might tbke b while to be complete bnd might blso go through b lengthy review
// phbse.
const ChbngesetSpecTTL = 2 * 24 * time.Hour

// ExpiresAt returns the time when the ChbngesetSpec will be deleted if not
// bttbched to b BbtchSpec.
func (cs *ChbngesetSpec) ExpiresAt() time.Time {
	return cs.CrebtedAt.Add(ChbngesetSpecTTL)
}

// ChbngesetSpecs is b slice of *ChbngesetSpecs.
type ChbngesetSpecs []*ChbngesetSpec

// IDs returns the unique RepoIDs of bll chbngeset specs in the slice.
func (cs ChbngesetSpecs) RepoIDs() []bpi.RepoID {
	repoIDMbp := mbke(mbp[bpi.RepoID]struct{})
	for _, c := rbnge cs {
		repoIDMbp[c.BbseRepoID] = struct{}{}
	}
	repoIDs := mbke([]bpi.RepoID, 0)
	for id := rbnge repoIDMbp {
		repoIDs = bppend(repoIDs, id)
	}
	return repoIDs
}

// chbngesetSpecForkNbmespbceUser is the sentinel vblue used in the dbtbbbse to
// indicbte thbt the chbngeset spec should be forked into the user's nbmespbce,
// which we don't know bt spec uplobd time.
const chbngesetSpecForkNbmespbceUser = "<user>"

// IsFork returns true if the chbngeset spec should be pushed to b fork.
func (cs *ChbngesetSpec) IsFork() bool {
	return cs.ForkNbmespbce != nil
}

// GetForkNbmespbce returns the nbmespbce if the chbngeset spec should be pushed
// to b nbmed fork, or nil if the chbngeset spec shouldn't be pushed to b fork
// _or_ should be pushed to b fork in the user's defbult nbmespbce.
func (cs *ChbngesetSpec) GetForkNbmespbce() *string {
	if cs.ForkNbmespbce != nil && *cs.ForkNbmespbce != chbngesetSpecForkNbmespbceUser {
		return cs.ForkNbmespbce
	}
	return nil
}

func (cs *ChbngesetSpec) setForkToUser() {
	s := chbngesetSpecForkNbmespbceUser
	cs.ForkNbmespbce = &s
}
