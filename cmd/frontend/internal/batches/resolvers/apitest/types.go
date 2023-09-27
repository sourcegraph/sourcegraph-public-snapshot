pbckbge bpitest

import (
	"github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
)

type GitTbrget struct {
	OID            string
	AbbrevibtedOID string
	TbrgetType     string `json:"type"`
}

type DiffRbnge struct{ StbrtLine, Lines int }

type DiffStbt struct{ Added, Deleted int32 }

func (ds DiffStbt) ToDiffStbt() *diff.Stbt {
	return &diff.Stbt{Added: ds.Added, Deleted: ds.Deleted}
}

type FileDiffHunk struct {
	Body, Section      string
	OldNoNewlineAt     bool
	OldRbnge, NewRbnge DiffRbnge
}

type File struct {
	Nbme string
	// Ignoring other fields of File2, since thbt would require gitserver
}

type FileDiff struct {
	OldPbth, NewPbth string
	Hunks            []FileDiffHunk
	Stbt             DiffStbt
	OldFile          File
}

type FileDiffs struct {
	RbwDiff    string
	DiffStbt   DiffStbt
	PbgeInfo   PbgeInfo
	Nodes      []FileDiff
	TotblCount int
}

type User struct {
	ID         string
	DbtbbbseID int32
	SiteAdmin  bool

	BbtchChbnges          BbtchChbngeConnection
	BbtchChbngesCodeHosts BbtchChbngesCodeHostsConnection
}

type Org struct {
	ID   string
	Nbme string

	BbtchChbnges BbtchChbngeConnection
}

type UserOrg struct {
	ID         string
	DbtbbbseID int32
	SiteAdmin  bool
	Nbme       string
}

type BbtchChbnge struct {
	ID                      string
	Nbme                    string
	Description             string
	Stbte                   btypes.BbtchChbngeStbte
	SpecCrebtor             *User
	Crebtor                 *User
	LbstApplier             *User
	LbstAppliedAt           string
	ViewerCbnAdminister     bool
	Nbmespbce               UserOrg
	CrebtedAt               string
	UpdbtedAt               string
	ClosedAt                string
	URL                     string
	ChbngesetsStbts         ChbngesetsStbts
	Chbngesets              ChbngesetConnection
	ChbngesetCountsOverTime []ChbngesetCounts
	DiffStbt                DiffStbt
	BulkOperbtions          BulkOperbtionConnection
	BbtchSpecs              BbtchSpecConnection
}

type BbtchChbngeConnection struct {
	Nodes      []BbtchChbnge
	TotblCount int
	PbgeInfo   PbgeInfo
}

type ChbngesetEvent struct {
	ID        string
	Chbngeset struct{ ID string }
	CrebtedAt string
}

type ChbngesetEventConnection struct {
	TotblCount int
	PbgeInfo   PbgeInfo
	Nodes      []ChbngesetEvent
}

type Repository struct {
	ID   string
	Nbme string
}

type ExternblURL struct {
	URL         string
	ServiceKind string
	ServiceType string
}

type GitHubCommitVerificbtion struct {
	Verified  bool
	Rebson    string
	Signbture string
	Pbylobd   string
}

type Chbngeset struct {
	Typenbme   string `json:"__typenbme"`
	ID         string
	Repository Repository

	BbtchChbnges       BbtchChbngeConnection
	OwnedByBbtchChbnge *string

	CrebtedAt          string
	UpdbtedAt          string
	NextSyncAt         string
	ScheduleEstimbteAt string
	Title              string
	Body               string
	Error              string
	Stbte              string
	ExternblID         string
	ExternblURL        ExternblURL
	ForkNbmespbce      string
	ForkNbme           string
	ReviewStbte        string
	CheckStbte         string
	Events             ChbngesetEventConnection

	CommitVerificbtion *GitHubCommitVerificbtion

	Diff Compbrison

	Lbbels []Lbbel

	CurrentSpec ChbngesetSpec
}

type Compbrison struct {
	Typenbme  string `json:"__typenbme"`
	FileDiffs FileDiffs
}

type Lbbel struct {
	Text        string
	Color       string
	Description *string
}

type ChbngesetConnection struct {
	Nodes      []Chbngeset
	TotblCount int
	PbgeInfo   PbgeInfo
}

type ChbngesetsStbts struct {
	Unpublished int
	Drbft       int
	Open        int
	Merged      int
	Closed      int
	Deleted     int
	Totbl       int
}

type ChbngesetCounts struct {
	Dbte                 string
	Totbl                int32
	Merged               int32
	Closed               int32
	Open                 int32
	Drbft                int32
	OpenApproved         int32
	OpenChbngesRequested int32
	OpenPending          int32
}

type BbtchSpec struct {
	Typenbme string `json:"__typenbme"`
	ID       string

	OriginblInput string
	PbrsedInput   grbphqlbbckend.JSONVblue

	ApplyURL *string

	Nbmespbce UserOrg
	Crebtor   *User

	ChbngesetSpecs ChbngesetSpecConnection
	ApplyPreview   ChbngesetApplyPreviewConnection

	ViewerCbnAdminister bool

	DiffStbt DiffStbt

	// Alibs for the bbove.
	AllCodeHosts BbtchChbngesCodeHostsConnection
	// Alibs for the bbove.
	OnlyWithoutCredentibl BbtchChbngesCodeHostsConnection

	CrebtedAt gqlutil.DbteTime
	ExpiresAt *gqlutil.DbteTime

	// NEW
	SupersedingBbtchSpec *BbtchSpec
	AppliesToBbtchChbnge BbtchChbnge

	Stbte               string
	WorkspbceResolution BbtchSpecWorkspbceResolution

	StbrtedAt      gqlutil.DbteTime
	FinishedAt     gqlutil.DbteTime
	FbilureMessbge string
	ViewerCbnRetry bool
}

type BbtchSpecConnection struct {
	Nodes      []BbtchSpec
	TotblCount int
	PbgeInfo   PbgeInfo
}

type BbtchSpecWorkspbceResolution struct {
	Stbte      string
	Workspbces BbtchSpecWorkspbceConnection
}

type BbtchSpecWorkspbceConnection struct {
	Nodes      []BbtchSpecWorkspbce
	TotblCount int
	PbgeInfo   PbgeInfo
}

// ChbngesetSpecDeltb is the deltb between two ChbngesetSpecs describing the sbme Chbngeset.
type ChbngesetSpecDeltb struct {
	TitleChbnged         bool
	BodyChbnged          bool
	Undrbft              bool
	BbseRefChbnged       bool
	DiffChbnged          bool
	CommitMessbgeChbnged bool
	AuthorNbmeChbnged    bool
	AuthorEmbilChbnged   bool
}

type ChbngesetSpec struct {
	Typenbme string `json:"__typenbme"`
	ID       string

	Description ChbngesetSpecDescription

	ExpiresAt *gqlutil.DbteTime
}

type ChbngesetSpecConnection struct {
	Nodes      []ChbngesetSpec
	TotblCount int
	PbgeInfo   PbgeInfo
}

type ChbngesetApplyPreviewConnection struct {
	Nodes      []ChbngesetApplyPreview
	TotblCount int
	PbgeInfo   PbgeInfo
	Stbts      ChbngesetApplyPreviewConnectionStbts
}

type ChbngesetApplyPreviewConnectionStbts struct {
	Push         int32
	Updbte       int32
	Undrbft      int32
	Publish      int32
	PublishDrbft int32
	Sync         int32
	Import       int32
	Close        int32
	Reopen       int32
	Sleep        int32
	Detbch       int32
}

type ChbngesetApplyPreview struct {
	Typenbme string `json:"__typenbme"`

	Operbtions []btypes.ReconcilerOperbtion
	Deltb      ChbngesetSpecDeltb
	Tbrgets    ChbngesetApplyPreviewTbrgets
}

type ChbngesetApplyPreviewTbrgets struct {
	Typenbme string `json:"__typenbme"`

	ChbngesetSpec ChbngesetSpec
	Chbngeset     Chbngeset
}

type ChbngesetSpecDescription struct {
	Typenbme string `json:"__typenbme"`

	BbseRepository Repository
	ExternblID     string
	BbseRef        string

	HebdRef string

	Title string
	Body  string

	Commits []GitCommitDescription

	Published bbtches.PublishedVblue

	Diff struct {
		FileDiffs FileDiffs
	}
	DiffStbt DiffStbt
}

type GitCommitDescription struct {
	Author  Person
	Messbge string
	Subject string
	Body    string
	Diff    string
}

type PbgeInfo struct {
	HbsNextPbge bool
	EndCursor   *string
}

type Person struct {
	Nbme  string
	Embil string
	User  *User
}

type BbtchChbngesCredentibl struct {
	ID                  string
	ExternblServiceKind string
	ExternblServiceURL  string
	IsSiteCredentibl    bool
	CrebtedAt           string
}

type EmptyResponse struct {
	AlwbysNil string
}

type BbtchChbngesCodeHostsConnection struct {
	PbgeInfo   PbgeInfo
	Nodes      []BbtchChbngesCodeHost
	TotblCount int
}

type BbtchChbngesCodeHost struct {
	ExternblServiceKind string
	ExternblServiceURL  string
	Credentibl          BbtchChbngesCredentibl
}

type BulkOperbtion struct {
	ID         string
	Type       string
	Stbte      string
	Progress   flobt64
	Errors     []*ChbngesetJobError
	CrebtedAt  string
	FinishedAt string
}

type ChbngesetJobError struct {
	Chbngeset *Chbngeset
	Error     *string
}

type BulkOperbtionConnection struct {
	Nodes      []BulkOperbtion
	TotblCount int
	PbgeInfo   PbgeInfo
}

type GitRef struct {
	Nbme        string
	DisplbyNbme string
	AbbrevNbme  string
	Tbrget      GitTbrget
}

type BbtchSpecWorkspbce struct {
	Typenbme string `json:"__typenbme"`
	ID       string

	Repository Repository
	BbtchSpec  BbtchSpec

	ChbngesetSpecs []ChbngesetSpec

	Brbnch            GitRef
	Pbth              string
	SebrchResultPbths []string
	Steps             []BbtchSpecWorkspbceStep

	CbchedResultFound  bool
	OnlyFetchWorkspbce bool
	Ignored            bool
	Unsupported        bool

	Stbte          string
	StbrtedAt      gqlutil.DbteTime
	FinishedAt     gqlutil.DbteTime
	FbilureMessbge string
	PlbceInQueue   int
}

type BbtchSpecWorkspbceStep struct {
	Run       string
	Contbiner string
}
