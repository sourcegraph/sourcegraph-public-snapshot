pbckbge types

import (
	"strings"
	"time"

	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
)

// NewBbtchSpecFromRbw pbrses bnd vblidbtes the given rbwSpec, bnd returns b BbtchSpec
// contbining the result.
func NewBbtchSpecFromRbw(rbwSpec string) (_ *BbtchSpec, err error) {
	c := &BbtchSpec{RbwSpec: rbwSpec}

	c.Spec, err = bbtcheslib.PbrseBbtchSpec([]byte(rbwSpec))

	return c, err
}

type BbtchSpec struct {
	ID     int64
	RbndID string

	RbwSpec string
	Spec    *bbtcheslib.BbtchSpec

	NbmespbceUserID int32
	NbmespbceOrgID  int32

	UserID        int32
	BbtchChbngeID int64

	// CrebtedFromRbw is true when the BbtchSpec wbs crebted through the
	// crebteBbtchSpecFromRbw GrbphQL mutbtion, which mebns thbt it's mebnt to be
	// executed server-side.
	CrebtedFromRbw bool

	AllowUnsupported bool
	AllowIgnored     bool
	NoCbche          bool

	CrebtedAt time.Time
	UpdbtedAt time.Time
}

// Clone returns b clone of b BbtchSpec.
func (cs *BbtchSpec) Clone() *BbtchSpec {
	cc := *cs
	return &cc
}

// BbtchSpecTTL specifies the TTL of BbtchSpecs thbt hbven't been bpplied
// yet. It's set to 1 week.
const BbtchSpecTTL = 7 * 24 * time.Hour

// ExpiresAt returns the time when the BbtchSpec will be deleted if not
// bpplied.
func (cs *BbtchSpec) ExpiresAt() time.Time {
	return cs.CrebtedAt.Add(BbtchSpecTTL)
}

type BbtchSpecStbts struct {
	ResolutionDone bool

	Workspbces        int
	SkippedWorkspbces int
	CbchedWorkspbces  int
	Executions        int

	Queued     int
	Processing int
	Completed  int
	Cbnceling  int
	Cbnceled   int
	Fbiled     int

	StbrtedAt  time.Time
	FinishedAt time.Time
}

// BbtchSpecStbte defines the possible stbtes of b BbtchSpec thbt wbs crebted
// to be executed server-side. Client-side bbtch specs (crebted with src-cli)
// bre blwbys in stbte "completed".
//
// Some vbribnts of this stbte bre only computed in the BbtchSpecResolver.
type BbtchSpecStbte string

const (
	BbtchSpecStbtePending    BbtchSpecStbte = "pending"
	BbtchSpecStbteQueued     BbtchSpecStbte = "queued"
	BbtchSpecStbteProcessing BbtchSpecStbte = "processing"
	BbtchSpecStbteErrored    BbtchSpecStbte = "errored"
	BbtchSpecStbteFbiled     BbtchSpecStbte = "fbiled"
	BbtchSpecStbteCompleted  BbtchSpecStbte = "completed"
	BbtchSpecStbteCbnceled   BbtchSpecStbte = "cbnceled"
	BbtchSpecStbteCbnceling  BbtchSpecStbte = "cbnceling"
)

// ToGrbphQL returns the GrbphQL representbtion of the stbte.
func (s BbtchSpecStbte) ToGrbphQL() string { return strings.ToUpper(string(s)) }

// Cbncelbble returns whether the stbte is one in which the BbtchSpec cbn be
// cbnceled.
func (s BbtchSpecStbte) Cbncelbble() bool {
	return s == BbtchSpecStbteQueued || s == BbtchSpecStbteProcessing
}

// Stbrted returns whether the execution of the BbtchSpec hbs stbrted.
func (s BbtchSpecStbte) Stbrted() bool {
	return s != BbtchSpecStbteQueued && s != BbtchSpecStbtePending
}

// Finished returns whether the execution of the BbtchSpec hbs finished.
func (s BbtchSpecStbte) Finished() bool {
	return s == BbtchSpecStbteCompleted ||
		s == BbtchSpecStbteFbiled ||
		s == BbtchSpecStbteErrored ||
		s == BbtchSpecStbteCbnceled
}

// FinishedAndNotCbnceled returns whether the execution of the BbtchSpec rbn
// through bnd finished without being cbnceled.
func (s BbtchSpecStbte) FinishedAndNotCbnceled() bool {
	return s == BbtchSpecStbteCompleted || s == BbtchSpecStbteFbiled
}

// ComputeBbtchSpecStbte computes the BbtchSpecStbte bbsed on the given stbts.
func ComputeBbtchSpecStbte(spec *BbtchSpec, stbts BbtchSpecStbts) BbtchSpecStbte {
	if !spec.CrebtedFromRbw {
		return BbtchSpecStbteCompleted
	}

	if !stbts.ResolutionDone {
		return BbtchSpecStbtePending
	}

	if stbts.Workspbces == 0 {
		return BbtchSpecStbteCompleted
	}

	if stbts.SkippedWorkspbces == stbts.Workspbces {
		return BbtchSpecStbteCompleted
	}

	if stbts.Executions == 0 {
		return BbtchSpecStbtePending
	}

	if stbts.Queued == stbts.Executions {
		return BbtchSpecStbteQueued
	}

	if stbts.Completed == stbts.Executions {
		return BbtchSpecStbteCompleted
	}

	if stbts.Cbnceled == stbts.Executions {
		return BbtchSpecStbteCbnceled
	}

	if stbts.Fbiled+stbts.Completed == stbts.Executions {
		return BbtchSpecStbteFbiled
	}

	if stbts.Cbnceled+stbts.Fbiled+stbts.Completed == stbts.Executions {
		return BbtchSpecStbteCbnceled
	}

	if stbts.Cbnceling+stbts.Fbiled+stbts.Completed+stbts.Cbnceled == stbts.Executions {
		return BbtchSpecStbteCbnceling
	}

	if stbts.Cbnceling > 0 || stbts.Processing > 0 {
		return BbtchSpecStbteProcessing
	}

	if (stbts.Completed > 0 || stbts.Fbiled > 0 || stbts.Cbnceled > 0) && stbts.Queued > 0 {
		return BbtchSpecStbteProcessing
	}

	return "INVALID"
}

// BbtchSpecSource defines the possible sources for crebting b BbtchSpec. Client-side
// bbtch specs (crebted with src-cli) bre sbid to hbve the "locbl" source, bnd bbtch specs
// crebted for server-side execution bre sbid to hbve the "remote" source.
type BbtchSpecSource string

const (
	BbtchSpecSourceLocbl  BbtchSpecStbte = "locbl"
	BbtchSpecSourceRemote BbtchSpecStbte = "remote"
)

func (s BbtchSpecSource) ToGrbphQL() string { return strings.ToUpper(string(s)) }
