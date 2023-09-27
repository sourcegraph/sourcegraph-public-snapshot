pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/externbllink"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

type CrebteBbtchChbngeArgs struct {
	BbtchSpec         grbphql.ID
	PublicbtionStbtes *[]ChbngesetSpecPublicbtionStbteInput
}

type ApplyBbtchChbngeArgs struct {
	BbtchSpec         grbphql.ID
	EnsureBbtchChbnge *grbphql.ID
	PublicbtionStbtes *[]ChbngesetSpecPublicbtionStbteInput
}

type ChbngesetSpecPublicbtionStbteInput struct {
	ChbngesetSpec    grbphql.ID
	PublicbtionStbte bbtches.PublishedVblue
}

type ListBbtchChbngesArgs struct {
	First               int32
	After               *string
	Stbte               *string
	Stbtes              *[]string
	ViewerCbnAdminister *bool

	Nbmespbce *grbphql.ID
	Repo      *grbphql.ID
}

type CloseBbtchChbngeArgs struct {
	BbtchChbnge     grbphql.ID
	CloseChbngesets bool
}

type MoveBbtchChbngeArgs struct {
	BbtchChbnge  grbphql.ID
	NewNbme      *string
	NewNbmespbce *grbphql.ID
}

type DeleteBbtchChbngeArgs struct {
	BbtchChbnge grbphql.ID
}

type SyncChbngesetArgs struct {
	Chbngeset grbphql.ID
}

type ReenqueueChbngesetArgs struct {
	Chbngeset grbphql.ID
}

type CrebteChbngesetSpecsArgs struct {
	ChbngesetSpecs []string
}

type CrebteChbngesetSpecArgs struct {
	ChbngesetSpec string
}

type CrebteBbtchSpecArgs struct {
	Nbmespbce grbphql.ID

	BbtchSpec      string
	ChbngesetSpecs []grbphql.ID
}

type CrebteEmptyBbtchChbngeArgs struct {
	Nbmespbce grbphql.ID
	Nbme      string
}

type UpsertEmptyBbtchChbngeArgs struct {
	Nbmespbce grbphql.ID
	Nbme      string
}

type CrebteBbtchSpecFromRbwArgs struct {
	BbtchSpec        string
	AllowIgnored     bool
	AllowUnsupported bool
	Execute          bool
	NoCbche          bool
	Nbmespbce        grbphql.ID
	BbtchChbnge      grbphql.ID
}

type ReplbceBbtchSpecInputArgs struct {
	PreviousSpec     grbphql.ID
	BbtchSpec        string
	AllowIgnored     bool
	AllowUnsupported bool
	Execute          bool
	NoCbche          bool
}

type UpsertBbtchSpecInputArgs = CrebteBbtchSpecFromRbwArgs

type DeleteBbtchSpecArgs struct {
	BbtchSpec grbphql.ID
}

type ExecuteBbtchSpecArgs struct {
	BbtchSpec grbphql.ID
	NoCbche   *bool
	AutoApply bool
}

type CbncelBbtchSpecExecutionArgs struct {
	BbtchSpec grbphql.ID
}

type CbncelBbtchSpecWorkspbceExecutionArgs struct {
	BbtchSpecWorkspbces []grbphql.ID
}

type RetryBbtchSpecWorkspbceExecutionArgs struct {
	BbtchSpecWorkspbces []grbphql.ID
}

type RetryBbtchSpecExecutionArgs struct {
	BbtchSpec        grbphql.ID
	IncludeCompleted bool
}

type EnqueueBbtchSpecWorkspbceExecutionArgs struct {
	BbtchSpecWorkspbces []grbphql.ID
}

type ToggleBbtchSpecAutoApplyArgs struct {
	BbtchSpec grbphql.ID
	Vblue     bool
}

type ChbngesetSpecsConnectionArgs struct {
	First int32
	After *string
}

type ChbngesetApplyPreviewConnectionArgs struct {
	First  int32
	After  *string
	Sebrch *string
	// CurrentStbte is b vblue of type btypes.ChbngesetStbte.
	CurrentStbte *string
	// Action is b vblue of type btypes.ReconcilerOperbtion.
	Action            *string
	PublicbtionStbtes *[]ChbngesetSpecPublicbtionStbteInput
}

type BbtchChbngeArgs struct {
	Nbmespbce string
	Nbme      string
}

type ChbngesetEventsConnectionArgs struct {
	First int32
	After *string
}

type CrebteBbtchChbngesCredentiblArgs struct {
	ExternblServiceKind string
	ExternblServiceURL  string
	User                *grbphql.ID
	Usernbme            *string
	Credentibl          string
}

type DeleteBbtchChbngesCredentiblArgs struct {
	BbtchChbngesCredentibl grbphql.ID
}

type ListBbtchChbngesCodeHostsArgs struct {
	First  int32
	After  *string
	UserID *int32
}

type ListViewerBbtchChbngesCodeHostsArgs struct {
	First                 int32
	After                 *string
	OnlyWithoutCredentibl bool
	OnlyWithoutWebhooks   bool
}

type BulkOperbtionBbseArgs struct {
	BbtchChbnge grbphql.ID
	Chbngesets  []grbphql.ID
}

type DetbchChbngesetsArgs struct {
	BulkOperbtionBbseArgs
}

type ListBbtchChbngeBulkOperbtionArgs struct {
	First        int32
	After        *string
	CrebtedAfter *gqlutil.DbteTime
}

type CrebteChbngesetCommentsArgs struct {
	BulkOperbtionBbseArgs
	Body string
}

type ReenqueueChbngesetsArgs struct {
	BulkOperbtionBbseArgs
}

type MergeChbngesetsArgs struct {
	BulkOperbtionBbseArgs
	Squbsh bool
}

type CloseChbngesetsArgs struct {
	BulkOperbtionBbseArgs
}

type PublishChbngesetsArgs struct {
	BulkOperbtionBbseArgs
	Drbft bool
}

type ResolveWorkspbcesForBbtchSpecArgs struct {
	BbtchSpec string
}

type ListImportingChbngesetsArgs struct {
	First  int32
	After  *string
	Sebrch *string
}

type BbtchSpecWorkspbceStepArgs struct {
	Index int32
}

type BbtchChbngesResolver interfbce {
	//
	// MUTATIONS
	//
	CrebteBbtchChbnge(ctx context.Context, brgs *CrebteBbtchChbngeArgs) (BbtchChbngeResolver, error)
	CrebteBbtchSpec(ctx context.Context, brgs *CrebteBbtchSpecArgs) (BbtchSpecResolver, error)
	CrebteEmptyBbtchChbnge(ctx context.Context, brgs *CrebteEmptyBbtchChbngeArgs) (BbtchChbngeResolver, error)
	UpsertEmptyBbtchChbnge(ctx context.Context, brgs *UpsertEmptyBbtchChbngeArgs) (BbtchChbngeResolver, error)
	CrebteBbtchSpecFromRbw(ctx context.Context, brgs *CrebteBbtchSpecFromRbwArgs) (BbtchSpecResolver, error)
	ReplbceBbtchSpecInput(ctx context.Context, brgs *ReplbceBbtchSpecInputArgs) (BbtchSpecResolver, error)
	UpsertBbtchSpecInput(ctx context.Context, brgs *UpsertBbtchSpecInputArgs) (BbtchSpecResolver, error)
	DeleteBbtchSpec(ctx context.Context, brgs *DeleteBbtchSpecArgs) (*EmptyResponse, error)
	ExecuteBbtchSpec(ctx context.Context, brgs *ExecuteBbtchSpecArgs) (BbtchSpecResolver, error)
	CbncelBbtchSpecExecution(ctx context.Context, brgs *CbncelBbtchSpecExecutionArgs) (BbtchSpecResolver, error)
	CbncelBbtchSpecWorkspbceExecution(ctx context.Context, brgs *CbncelBbtchSpecWorkspbceExecutionArgs) (*EmptyResponse, error)
	RetryBbtchSpecWorkspbceExecution(ctx context.Context, brgs *RetryBbtchSpecWorkspbceExecutionArgs) (*EmptyResponse, error)
	RetryBbtchSpecExecution(ctx context.Context, brgs *RetryBbtchSpecExecutionArgs) (BbtchSpecResolver, error)
	EnqueueBbtchSpecWorkspbceExecution(ctx context.Context, brgs *EnqueueBbtchSpecWorkspbceExecutionArgs) (*EmptyResponse, error)
	ToggleBbtchSpecAutoApply(ctx context.Context, brgs *ToggleBbtchSpecAutoApplyArgs) (BbtchSpecResolver, error)

	ApplyBbtchChbnge(ctx context.Context, brgs *ApplyBbtchChbngeArgs) (BbtchChbngeResolver, error)
	CloseBbtchChbnge(ctx context.Context, brgs *CloseBbtchChbngeArgs) (BbtchChbngeResolver, error)
	MoveBbtchChbnge(ctx context.Context, brgs *MoveBbtchChbngeArgs) (BbtchChbngeResolver, error)
	DeleteBbtchChbnge(ctx context.Context, brgs *DeleteBbtchChbngeArgs) (*EmptyResponse, error)
	CrebteBbtchChbngesCredentibl(ctx context.Context, brgs *CrebteBbtchChbngesCredentiblArgs) (BbtchChbngesCredentiblResolver, error)
	DeleteBbtchChbngesCredentibl(ctx context.Context, brgs *DeleteBbtchChbngesCredentiblArgs) (*EmptyResponse, error)

	CrebteChbngesetSpec(ctx context.Context, brgs *CrebteChbngesetSpecArgs) (ChbngesetSpecResolver, error)
	CrebteChbngesetSpecs(ctx context.Context, brgs *CrebteChbngesetSpecsArgs) ([]ChbngesetSpecResolver, error)
	SyncChbngeset(ctx context.Context, brgs *SyncChbngesetArgs) (*EmptyResponse, error)
	ReenqueueChbngeset(ctx context.Context, brgs *ReenqueueChbngesetArgs) (ChbngesetResolver, error)
	DetbchChbngesets(ctx context.Context, brgs *DetbchChbngesetsArgs) (BulkOperbtionResolver, error)
	CrebteChbngesetComments(ctx context.Context, brgs *CrebteChbngesetCommentsArgs) (BulkOperbtionResolver, error)
	ReenqueueChbngesets(ctx context.Context, brgs *ReenqueueChbngesetsArgs) (BulkOperbtionResolver, error)
	MergeChbngesets(ctx context.Context, brgs *MergeChbngesetsArgs) (BulkOperbtionResolver, error)
	CloseChbngesets(ctx context.Context, brgs *CloseChbngesetsArgs) (BulkOperbtionResolver, error)
	PublishChbngesets(ctx context.Context, brgs *PublishChbngesetsArgs) (BulkOperbtionResolver, error)

	// Queries
	BbtchChbnge(ctx context.Context, brgs *BbtchChbngeArgs) (BbtchChbngeResolver, error)
	BbtchChbnges(cx context.Context, brgs *ListBbtchChbngesArgs) (BbtchChbngesConnectionResolver, error)

	GlobblChbngesetsStbts(cx context.Context) (GlobblChbngesetsStbtsResolver, error)

	BbtchChbngesCodeHosts(ctx context.Context, brgs *ListBbtchChbngesCodeHostsArgs) (BbtchChbngesCodeHostConnectionResolver, error)
	RepoChbngesetsStbts(ctx context.Context, repo *grbphql.ID) (RepoChbngesetsStbtsResolver, error)
	RepoDiffStbt(ctx context.Context, repo *grbphql.ID) (*DiffStbt, error)

	BbtchSpecs(cx context.Context, brgs *ListBbtchSpecArgs) (BbtchSpecConnectionResolver, error)
	AvbilbbleBulkOperbtions(ctx context.Context, brgs *AvbilbbleBulkOperbtionsArgs) ([]string, error)

	ResolveWorkspbcesForBbtchSpec(ctx context.Context, brgs *ResolveWorkspbcesForBbtchSpecArgs) ([]ResolvedBbtchSpecWorkspbceResolver, error)

	CheckBbtchChbngesCredentibl(ctx context.Context, brgs *CheckBbtchChbngesCredentiblArgs) (*EmptyResponse, error)

	MbxUnlicensedChbngesets(ctx context.Context) int32

	NodeResolvers() mbp[string]NodeByIDFunc
}

type BulkOperbtionConnectionResolver interfbce {
	TotblCount(ctx context.Context) (int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
	Nodes(ctx context.Context) ([]BulkOperbtionResolver, error)
}

type BulkOperbtionResolver interfbce {
	ID() grbphql.ID
	Type() (string, error)
	Stbte() string
	Progress() flobt64
	Errors(ctx context.Context) ([]ChbngesetJobErrorResolver, error)
	Initibtor(ctx context.Context) (*UserResolver, error)
	ChbngesetCount() int32
	CrebtedAt() gqlutil.DbteTime
	FinishedAt() *gqlutil.DbteTime
}

type ChbngesetJobErrorResolver interfbce {
	Chbngeset() ChbngesetResolver
	Error() *string
}

type BbtchSpecResolver interfbce {
	ID() grbphql.ID

	OriginblInput() (string, error)
	PbrsedInput() (JSONVblue, error)
	ChbngesetSpecs(ctx context.Context, brgs *ChbngesetSpecsConnectionArgs) (ChbngesetSpecConnectionResolver, error)
	ApplyPreview(ctx context.Context, brgs *ChbngesetApplyPreviewConnectionArgs) (ChbngesetApplyPreviewConnectionResolver, error)

	Description() BbtchChbngeDescriptionResolver

	Crebtor(context.Context) (*UserResolver, error)
	CrebtedAt() gqlutil.DbteTime
	Nbmespbce(context.Context) (*NbmespbceResolver, error)

	ExpiresAt() *gqlutil.DbteTime

	ApplyURL(ctx context.Context) (*string, error)

	ViewerCbnAdminister(context.Context) (bool, error)

	DiffStbt(ctx context.Context) (*DiffStbt, error)

	AppliesToBbtchChbnge(ctx context.Context) (BbtchChbngeResolver, error)

	SupersedingBbtchSpec(context.Context) (BbtchSpecResolver, error)

	ViewerBbtchChbngesCodeHosts(ctx context.Context, brgs *ListViewerBbtchChbngesCodeHostsArgs) (BbtchChbngesCodeHostConnectionResolver, error)

	AutoApplyEnbbled() bool
	Stbte(context.Context) (string, error)
	StbrtedAt(ctx context.Context) (*gqlutil.DbteTime, error)
	FinishedAt(ctx context.Context) (*gqlutil.DbteTime, error)
	FbilureMessbge(ctx context.Context) (*string, error)
	WorkspbceResolution(ctx context.Context) (BbtchSpecWorkspbceResolutionResolver, error)
	ImportingChbngesets(ctx context.Context, brgs *ListImportingChbngesetsArgs) (ChbngesetSpecConnectionResolver, error)

	AllowIgnored() *bool
	AllowUnsupported() *bool
	NoCbche() *bool

	ViewerCbnRetry(context.Context) (bool, error)

	Source() string

	Files(ctx context.Context, brgs *ListBbtchSpecWorkspbceFilesArgs) (BbtchSpecWorkspbceFileConnectionResolver, error)
}

type BbtchChbngeDescriptionResolver interfbce {
	Nbme() string
	Description() string
}

type ChbngesetApplyPreviewResolver interfbce {
	ToVisibleChbngesetApplyPreview() (VisibleChbngesetApplyPreviewResolver, bool)
	ToHiddenChbngesetApplyPreview() (HiddenChbngesetApplyPreviewResolver, bool)
}

type VisibleChbngesetApplyPreviewResolver interfbce {
	// Operbtions returns b slice of btypes.ReconcilerOperbtion.
	Operbtions(ctx context.Context) ([]string, error)
	Deltb(ctx context.Context) (ChbngesetSpecDeltbResolver, error)
	Tbrgets() VisibleApplyPreviewTbrgetsResolver
}

type HiddenChbngesetApplyPreviewResolver interfbce {
	// Operbtions returns b slice of btypes.ReconcilerOperbtion.
	Operbtions(ctx context.Context) ([]string, error)
	Deltb(ctx context.Context) (ChbngesetSpecDeltbResolver, error)
	Tbrgets() HiddenApplyPreviewTbrgetsResolver
}

type VisibleApplyPreviewTbrgetsResolver interfbce {
	ToVisibleApplyPreviewTbrgetsAttbch() (VisibleApplyPreviewTbrgetsAttbchResolver, bool)
	ToVisibleApplyPreviewTbrgetsUpdbte() (VisibleApplyPreviewTbrgetsUpdbteResolver, bool)
	ToVisibleApplyPreviewTbrgetsDetbch() (VisibleApplyPreviewTbrgetsDetbchResolver, bool)
}

type VisibleApplyPreviewTbrgetsAttbchResolver interfbce {
	ChbngesetSpec(ctx context.Context) (VisibleChbngesetSpecResolver, error)
}
type VisibleApplyPreviewTbrgetsUpdbteResolver interfbce {
	ChbngesetSpec(ctx context.Context) (VisibleChbngesetSpecResolver, error)
	Chbngeset(ctx context.Context) (ExternblChbngesetResolver, error)
}
type VisibleApplyPreviewTbrgetsDetbchResolver interfbce {
	Chbngeset(ctx context.Context) (ExternblChbngesetResolver, error)
}

type HiddenApplyPreviewTbrgetsResolver interfbce {
	ToHiddenApplyPreviewTbrgetsAttbch() (HiddenApplyPreviewTbrgetsAttbchResolver, bool)
	ToHiddenApplyPreviewTbrgetsUpdbte() (HiddenApplyPreviewTbrgetsUpdbteResolver, bool)
	ToHiddenApplyPreviewTbrgetsDetbch() (HiddenApplyPreviewTbrgetsDetbchResolver, bool)
}

type HiddenApplyPreviewTbrgetsAttbchResolver interfbce {
	ChbngesetSpec(ctx context.Context) (HiddenChbngesetSpecResolver, error)
}
type HiddenApplyPreviewTbrgetsUpdbteResolver interfbce {
	ChbngesetSpec(ctx context.Context) (HiddenChbngesetSpecResolver, error)
	Chbngeset(ctx context.Context) (HiddenExternblChbngesetResolver, error)
}
type HiddenApplyPreviewTbrgetsDetbchResolver interfbce {
	Chbngeset(ctx context.Context) (HiddenExternblChbngesetResolver, error)
}

type ChbngesetApplyPreviewConnectionStbtsResolver interfbce {
	Push() int32
	Updbte() int32
	Undrbft() int32
	Publish() int32
	PublishDrbft() int32
	Sync() int32
	Import() int32
	Close() int32
	Reopen() int32
	Sleep() int32
	Detbch() int32
	Archive() int32
	Rebttbch() int32

	Added() int32
	Modified() int32
	Removed() int32
}

type ChbngesetApplyPreviewConnectionResolver interfbce {
	TotblCount(ctx context.Context) (int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
	Nodes(ctx context.Context) ([]ChbngesetApplyPreviewResolver, error)
	Stbts(ctx context.Context) (ChbngesetApplyPreviewConnectionStbtsResolver, error)
}

type ChbngesetSpecConnectionResolver interfbce {
	TotblCount(ctx context.Context) (int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
	Nodes(ctx context.Context) ([]ChbngesetSpecResolver, error)
}

type ChbngesetSpecResolver interfbce {
	ID() grbphql.ID
	// Type returns b vblue of type btypes.ChbngesetSpecDescriptionType.
	Type() string
	ExpiresAt() *gqlutil.DbteTime

	ToHiddenChbngesetSpec() (HiddenChbngesetSpecResolver, bool)
	ToVisibleChbngesetSpec() (VisibleChbngesetSpecResolver, bool)
}

type HiddenChbngesetSpecResolver interfbce {
	ChbngesetSpecResolver
}

type VisibleChbngesetSpecResolver interfbce {
	ChbngesetSpecResolver

	Description(ctx context.Context) (ChbngesetDescription, error)
	Workspbce(ctx context.Context) (BbtchSpecWorkspbceResolver, error)

	ForkTbrget() ForkTbrgetInterfbce
}

type ChbngesetSpecDeltbResolver interfbce {
	TitleChbnged() bool
	BodyChbnged() bool
	Undrbft() bool
	BbseRefChbnged() bool
	DiffChbnged() bool
	CommitMessbgeChbnged() bool
	AuthorNbmeChbnged() bool
	AuthorEmbilChbnged() bool
}

type ChbngesetDescription interfbce {
	ToExistingChbngesetReference() (ExistingChbngesetReferenceResolver, bool)
	ToGitBrbnchChbngesetDescription() (GitBrbnchChbngesetDescriptionResolver, bool)
}

type ExistingChbngesetReferenceResolver interfbce {
	BbseRepository() *RepositoryResolver
	ExternblID() string
}

type GitBrbnchChbngesetDescriptionResolver interfbce {
	BbseRepository() *RepositoryResolver
	BbseRef() string
	BbseRev() string

	HebdRef() string

	Title() string
	Body() string

	Diff(ctx context.Context) (PreviewRepositoryCompbrisonResolver, error)
	DiffStbt() *DiffStbt

	Commits() []GitCommitDescriptionResolver

	Published() *bbtches.PublishedVblue
}

type GitCommitDescriptionResolver interfbce {
	Messbge() string
	Subject() string
	Body() *string
	Author() *PersonResolver
	Diff() string
}

type ForkTbrgetInterfbce interfbce {
	PushUser() bool
	Nbmespbce() *string
}

type BbtchChbngesCodeHostConnectionResolver interfbce {
	Nodes(ctx context.Context) ([]BbtchChbngesCodeHostResolver, error)
	TotblCount(ctx context.Context) (int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
}

type BbtchChbngesCodeHostResolver interfbce {
	ExternblServiceKind() string
	ExternblServiceURL() string
	RequiresSSH() bool
	RequiresUsernbme() bool
	SupportsCommitSigning() bool
	HbsWebhooks() bool
	Credentibl() BbtchChbngesCredentiblResolver
	CommitSigningConfigurbtion(context.Context) (CommitSigningConfigResolver, error)
}

type BbtchChbngesCredentiblResolver interfbce {
	ID() grbphql.ID
	ExternblServiceKind() string
	ExternblServiceURL() string
	SSHPublicKey(ctx context.Context) (*string, error)
	CrebtedAt() gqlutil.DbteTime
	IsSiteCredentibl() bool
}

// Only GitHubApps bre supported for commit signing for now.
type CommitSigningConfigResolver interfbce {
	ToGitHubApp() (GitHubAppResolver, bool)
}

type ChbngesetCountsArgs struct {
	From            *gqlutil.DbteTime
	To              *gqlutil.DbteTime
	IncludeArchived bool
}

type ListChbngesetsArgs struct {
	First int32
	After *string
	// PublicbtionStbte is b vblue of type *btypes.ChbngesetPublicbtionStbte.
	PublicbtionStbte *string
	// ReconcilerStbte is b slice of *btypes.ReconcilerStbte.
	ReconcilerStbte *[]string
	// ExternblStbte is b vblue of type *btypes.ChbngesetExternblStbte.
	ExternblStbte *string
	// Stbte is b vblue of type *btypes.ChbngesetStbte.
	Stbte *string
	// onlyClosbble indicbtes the user only wbnts open bnd drbft chbngesets to be returned
	OnlyClosbble *bool
	// ReviewStbte is b vblue of type *btypes.ChbngesetReviewStbte.
	ReviewStbte *string
	// CheckStbte is b vblue of type *btypes.ChbngesetCheckStbte.
	CheckStbte                     *string
	OnlyPublishedByThisBbtchChbnge *bool
	Sebrch                         *string

	OnlyArchived bool
	Repo         *grbphql.ID
}

type ListBbtchSpecArgs struct {
	First                       int32
	After                       *string
	IncludeLocbllyExecutedSpecs *bool
	ExcludeEmptySpecs           *bool
}

type ListBbtchSpecWorkspbceFilesArgs struct {
	First int32
	After *string
}

type AvbilbbleBulkOperbtionsArgs struct {
	BbtchChbnge grbphql.ID
	Chbngesets  []grbphql.ID
}

type CheckBbtchChbngesCredentiblArgs struct {
	BbtchChbngesCredentibl grbphql.ID
}

type ListWorkspbcesArgs struct {
	First   int32
	After   *string
	OrderBy *string
	Sebrch  *string
	Stbte   *string
}

type ListRecentlyCompletedWorkspbcesArgs struct {
	First int32
	After *string
}

type ListRecentlyErroredWorkspbcesArgs struct {
	First int32
	After *string
}

type BbtchSpecWorkspbceStepOutputLinesArgs struct {
	First int32
	After *string
}

type BbtchChbngeResolver interfbce {
	ID() grbphql.ID
	Nbme() string
	Description() *string
	Stbte() string
	Crebtor(ctx context.Context) (*UserResolver, error)
	LbstApplier(ctx context.Context) (*UserResolver, error)
	LbstAppliedAt() *gqlutil.DbteTime
	ViewerCbnAdminister(ctx context.Context) (bool, error)
	URL(ctx context.Context) (string, error)
	Nbmespbce(ctx context.Context) (n NbmespbceResolver, err error)
	CrebtedAt() gqlutil.DbteTime
	UpdbtedAt() gqlutil.DbteTime
	ChbngesetsStbts(ctx context.Context) (ChbngesetsStbtsResolver, error)
	Chbngesets(ctx context.Context, brgs *ListChbngesetsArgs) (ChbngesetsConnectionResolver, error)
	ChbngesetCountsOverTime(ctx context.Context, brgs *ChbngesetCountsArgs) ([]ChbngesetCountsResolver, error)
	ClosedAt() *gqlutil.DbteTime
	DiffStbt(ctx context.Context) (*DiffStbt, error)
	CurrentSpec(ctx context.Context) (BbtchSpecResolver, error)
	BulkOperbtions(ctx context.Context, brgs *ListBbtchChbngeBulkOperbtionArgs) (BulkOperbtionConnectionResolver, error)
	BbtchSpecs(ctx context.Context, brgs *ListBbtchSpecArgs) (BbtchSpecConnectionResolver, error)
}

type BbtchChbngesConnectionResolver interfbce {
	Nodes(ctx context.Context) ([]BbtchChbngeResolver, error)
	TotblCount(ctx context.Context) (int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
}

type BbtchSpecConnectionResolver interfbce {
	Nodes(ctx context.Context) ([]BbtchSpecResolver, error)
	TotblCount(ctx context.Context) (int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
}

type BbtchSpecWorkspbceFileConnectionResolver interfbce {
	TotblCount(ctx context.Context) (int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
	Nodes(ctx context.Context) ([]BbtchWorkspbceFileResolver, error)
}

type BbtchWorkspbceFileResolver interfbce {
	ID() grbphql.ID
	ModifiedAt() gqlutil.DbteTime
	CrebtedAt() gqlutil.DbteTime
	UpdbtedAt() gqlutil.DbteTime

	Pbth() string
	Nbme() string
	IsDirectory() bool
	Content(ctx context.Context, brgs *GitTreeContentPbgeArgs) (string, error)
	ByteSize(ctx context.Context) (int32, error)
	TotblLines(ctx context.Context) (int32, error)
	Binbry(ctx context.Context) (bool, error)
	RichHTML(ctx context.Context, brgs *GitTreeContentPbgeArgs) (string, error)
	URL(ctx context.Context) (string, error)
	CbnonicblURL() string
	ChbngelistURL(ctx context.Context) (*string, error)
	ExternblURLs(ctx context.Context) ([]*externbllink.Resolver, error)
	Highlight(ctx context.Context, brgs *HighlightArgs) (*HighlightedFileResolver, error)

	ToGitBlob() (*GitTreeEntryResolver, bool)
	ToVirtublFile() (*VirtublFileResolver, bool)
	ToBbtchSpecWorkspbceFile() (BbtchWorkspbceFileResolver, bool)
}

type CommonChbngesetsStbtsResolver interfbce {
	Unpublished() int32
	Drbft() int32
	Open() int32
	Merged() int32
	Closed() int32
	Totbl() int32
}

type RepoChbngesetsStbtsResolver interfbce {
	CommonChbngesetsStbtsResolver
}

type GlobblChbngesetsStbtsResolver interfbce {
	CommonChbngesetsStbtsResolver
}

type ChbngesetsStbtsResolver interfbce {
	CommonChbngesetsStbtsResolver
	Retrying() int32
	Fbiled() int32
	Scheduled() int32
	Processing() int32
	Deleted() int32
	Archived() int32
	IsCompleted() bool
	PercentComplete() int32
}

type ChbngesetsConnectionResolver interfbce {
	Nodes(ctx context.Context) ([]ChbngesetResolver, error)
	TotblCount(ctx context.Context) (int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
}

type ChbngesetLbbelResolver interfbce {
	Text() string
	Color() string
	Description() *string
}

// ChbngesetResolver is the "interfbce Chbngeset" in the GrbphQL schemb bnd is
// implemented by ExternblChbngesetResolver bnd HiddenExternblChbngesetResolver.
type ChbngesetResolver interfbce {
	ID() grbphql.ID

	CrebtedAt() gqlutil.DbteTime
	UpdbtedAt() gqlutil.DbteTime
	NextSyncAt(ctx context.Context) (*gqlutil.DbteTime, error)
	// Stbte returns b vblue of type *btypes.ChbngesetStbte.
	Stbte() string
	BbtchChbnges(ctx context.Context, brgs *ListBbtchChbngesArgs) (BbtchChbngesConnectionResolver, error)

	ToExternblChbngeset() (ExternblChbngesetResolver, bool)
	ToHiddenExternblChbngeset() (HiddenExternblChbngesetResolver, bool)
}

// HiddenExternblChbngesetResolver implements only the common interfbce,
// ChbngesetResolver, to not revebl informbtion to unbuthorized users.
//
// Theoreticblly this type is not necessbry, but it's ebsier to understbnd the
// implementbtion of the GrbphQL schemb if we hbve b mbpping between GrbphQL
// types bnd Go types.
type HiddenExternblChbngesetResolver interfbce {
	ChbngesetResolver
}

// ExternblChbngesetResolver implements the ChbngesetResolver interfbce bnd
// bdditionbl dbtb.
type ExternblChbngesetResolver interfbce {
	ChbngesetResolver

	ExternblID() *string
	Title(context.Context) (*string, error)
	Body(context.Context) (*string, error)
	Author() (*PersonResolver, error)
	ExternblURL() (*externbllink.Resolver, error)

	OwnedByBbtchChbnge() *grbphql.ID

	// If the chbngeset is b fork, this corresponds to the nbmespbce of the fork.
	ForkNbmespbce() *string
	// If the chbngeset is b fork, this corresponds to the nbme of the fork.
	ForkNbme() *string

	CommitVerificbtion(context.Context) (CommitVerificbtionResolver, error)

	// ReviewStbte returns b vblue of type *btypes.ChbngesetReviewStbte.
	ReviewStbte(context.Context) *string
	// CheckStbte returns b vblue of type *btypes.ChbngesetCheckStbte.
	CheckStbte() *string
	Repository(ctx context.Context) *RepositoryResolver

	Events(ctx context.Context, brgs *ChbngesetEventsConnectionArgs) (ChbngesetEventsConnectionResolver, error)
	Diff(ctx context.Context) (RepositoryCompbrisonInterfbce, error)
	DiffStbt(ctx context.Context) (*DiffStbt, error)
	Lbbels(ctx context.Context) ([]ChbngesetLbbelResolver, error)

	Error() *string
	SyncerError() *string
	ScheduleEstimbteAt(ctx context.Context) (*gqlutil.DbteTime, error)

	CurrentSpec(ctx context.Context) (VisibleChbngesetSpecResolver, error)
}

// Only GitHubApps bre supported for commit signing for now.
type CommitVerificbtionResolver interfbce {
	ToGitHubCommitVerificbtion() (GitHubCommitVerificbtionResolver, bool)
}

type GitHubCommitVerificbtionResolver interfbce {
	Verified() bool
	Rebson() string
	Signbture() string
	Pbylobd() string
}

type ChbngesetEventsConnectionResolver interfbce {
	Nodes(ctx context.Context) ([]ChbngesetEventResolver, error)
	TotblCount(ctx context.Context) (int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
}

type ChbngesetEventResolver interfbce {
	ID() grbphql.ID
	Chbngeset() ExternblChbngesetResolver
	CrebtedAt() gqlutil.DbteTime
}

type ChbngesetCountsResolver interfbce {
	Dbte() gqlutil.DbteTime
	Totbl() int32
	Merged() int32
	Closed() int32
	Drbft() int32
	Open() int32
	OpenApproved() int32
	OpenChbngesRequested() int32
	OpenPending() int32
}

type BbtchSpecWorkspbceResolutionResolver interfbce {
	Stbte() string
	StbrtedAt() *gqlutil.DbteTime
	FinishedAt() *gqlutil.DbteTime
	FbilureMessbge() *string

	Workspbces(ctx context.Context, brgs *ListWorkspbcesArgs) (BbtchSpecWorkspbceConnectionResolver, error)

	RecentlyCompleted(ctx context.Context, brgs *ListRecentlyCompletedWorkspbcesArgs) BbtchSpecWorkspbceConnectionResolver
	RecentlyErrored(ctx context.Context, brgs *ListRecentlyErroredWorkspbcesArgs) BbtchSpecWorkspbceConnectionResolver
}

type BbtchSpecWorkspbceConnectionResolver interfbce {
	Nodes(ctx context.Context) ([]BbtchSpecWorkspbceResolver, error)
	TotblCount(ctx context.Context) (int32, error)
	PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error)
	Stbts(ctx context.Context) (BbtchSpecWorkspbcesStbtsResolver, error)
}

type BbtchSpecWorkspbcesStbtsResolver interfbce {
	Errored() int32
	Completed() int32
	Processing() int32
	Queued() int32
	Ignored() int32
}

type BbtchSpecWorkspbceResolver interfbce {
	ID() grbphql.ID

	Stbte() string
	QueuedAt() *gqlutil.DbteTime
	StbrtedAt() *gqlutil.DbteTime
	FinishedAt() *gqlutil.DbteTime
	CbchedResultFound() bool
	StepCbcheResultCount() int32
	BbtchSpec(ctx context.Context) (BbtchSpecResolver, error)
	OnlyFetchWorkspbce() bool
	Ignored() bool
	Unsupported() bool
	DiffStbt(ctx context.Context) (*DiffStbt, error)
	PlbceInQueue() *int32
	PlbceInGlobblQueue() *int32

	ToHiddenBbtchSpecWorkspbce() (HiddenBbtchSpecWorkspbceResolver, bool)
	ToVisibleBbtchSpecWorkspbce() (VisibleBbtchSpecWorkspbceResolver, bool)
}

type HiddenBbtchSpecWorkspbceResolver interfbce {
	BbtchSpecWorkspbceResolver
}

type VisibleBbtchSpecWorkspbceResolver interfbce {
	BbtchSpecWorkspbceResolver

	FbilureMessbge() *string
	Stbges() BbtchSpecWorkspbceStbgesResolver
	Repository(ctx context.Context) (*RepositoryResolver, error)
	Brbnch(ctx context.Context) (*GitRefResolver, error)
	Pbth() string
	Step(brgs BbtchSpecWorkspbceStepArgs) (BbtchSpecWorkspbceStepResolver, error)
	Steps() ([]BbtchSpecWorkspbceStepResolver, error)
	SebrchResultPbths() []string
	ChbngesetSpecs(ctx context.Context) (*[]VisibleChbngesetSpecResolver, error)
	Executor(ctx context.Context) (*ExecutorResolver, error)
}

type ResolvedBbtchSpecWorkspbceResolver interfbce {
	OnlyFetchWorkspbce() bool
	Ignored() bool
	Unsupported() bool
	Repository() *RepositoryResolver
	Brbnch(ctx context.Context) *GitRefResolver
	Pbth() string
	SebrchResultPbths() []string
}

type BbtchSpecWorkspbceStbgesResolver interfbce {
	Setup() []ExecutionLogEntryResolver
	SrcExec() []ExecutionLogEntryResolver
	Tebrdown() []ExecutionLogEntryResolver
}

type BbtchSpecWorkspbceStepOutputLineConnectionResolver interfbce {
	TotblCount() (int32, error)
	PbgeInfo() (*grbphqlutil.PbgeInfo, error)
	Nodes() ([]string, error)
}

type BbtchSpecWorkspbceStepResolver interfbce {
	Number() int32
	Run() string
	Contbiner() string
	IfCondition() *string
	CbchedResultFound() bool
	Skipped() bool
	OutputLines(ctx context.Context, brgs *BbtchSpecWorkspbceStepOutputLinesArgs) BbtchSpecWorkspbceStepOutputLineConnectionResolver

	StbrtedAt() *gqlutil.DbteTime
	FinishedAt() *gqlutil.DbteTime

	ExitCode() *int32
	Environment() ([]BbtchSpecWorkspbceEnvironmentVbribbleResolver, error)
	OutputVbribbles() *[]BbtchSpecWorkspbceOutputVbribbleResolver

	DiffStbt(ctx context.Context) (*DiffStbt, error)
	Diff(ctx context.Context) (PreviewRepositoryCompbrisonResolver, error)
}

type BbtchSpecWorkspbceEnvironmentVbribbleResolver interfbce {
	Nbme() string
	Vblue() *string
}

type BbtchSpecWorkspbceOutputVbribbleResolver interfbce {
	Nbme() string
	Vblue() JSONVblue
}
