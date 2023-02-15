package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/markdown"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RootResolver interface {
	CodeNavServiceResolver
	AutoindexingServiceResolver
	UploadsServiceResolver
	PoliciesServiceResolver
}

type CodeNavServiceResolver interface {
	GitBlobLSIFData(ctx context.Context, args *GitBlobLSIFDataArgs) (GitBlobLSIFDataResolver, error)
}

type AutoindexingServiceResolver interface {
	RequestLanguageSupport(ctx context.Context, args *RequestLanguageSupportArgs) (*EmptyResponse, error)
	RequestedLanguageSupport(ctx context.Context) ([]string, error)
	GitBlobCodeIntelInfo(ctx context.Context, args *GitTreeEntryCodeIntelInfoArgs) (_ GitBlobCodeIntelSupportResolver, err error)
	GitTreeCodeIntelInfo(ctx context.Context, args *GitTreeEntryCodeIntelInfoArgs) (resolver GitTreeCodeIntelSupportResolver, err error)
	IndexConfiguration(ctx context.Context, id graphql.ID) (IndexConfigurationResolver, error) // TODO - rename ...ForRepo
	DeleteLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error)
	DeleteLSIFIndexes(ctx context.Context, args *DeleteLSIFIndexesArgs) (*EmptyResponse, error)
	ReindexLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error)
	ReindexLSIFIndexes(ctx context.Context, args *ReindexLSIFIndexesArgs) (*EmptyResponse, error)
	LSIFIndexByID(ctx context.Context, id graphql.ID) (_ LSIFIndexResolver, err error)
	LSIFIndexes(ctx context.Context, args *LSIFIndexesQueryArgs) (LSIFIndexConnectionResolver, error)
	LSIFIndexesByRepo(ctx context.Context, args *LSIFRepositoryIndexesQueryArgs) (LSIFIndexConnectionResolver, error)
	QueueAutoIndexJobsForRepo(ctx context.Context, args *QueueAutoIndexJobsForRepoArgs) ([]LSIFIndexResolver, error)
	UpdateRepositoryIndexConfiguration(ctx context.Context, args *UpdateRepositoryIndexConfigurationArgs) (*EmptyResponse, error)
	RepositorySummary(ctx context.Context, id graphql.ID) (CodeIntelRepositorySummaryResolver, error)
	CodeIntelligenceInferenceScript(ctx context.Context) (string, error)
	UpdateCodeIntelligenceInferenceScript(ctx context.Context, args *UpdateCodeIntelligenceInferenceScriptArgs) (*EmptyResponse, error)

	PreciseIndexByID(ctx context.Context, id graphql.ID) (PreciseIndexResolver, error)
	PreciseIndexes(ctx context.Context, args *PreciseIndexesQueryArgs) (PreciseIndexConnectionResolver, error)
	DeletePreciseIndex(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error)
	DeletePreciseIndexes(ctx context.Context, args *DeletePreciseIndexesArgs) (*EmptyResponse, error)
	ReindexPreciseIndex(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error)
	ReindexPreciseIndexes(ctx context.Context, args *ReindexPreciseIndexesArgs) (*EmptyResponse, error)
}

type UploadsServiceResolver interface {
	CommitGraph(ctx context.Context, id graphql.ID) (CodeIntelligenceCommitGraphResolver, error)
	LSIFUploadByID(ctx context.Context, id graphql.ID) (LSIFUploadResolver, error)
	LSIFUploads(ctx context.Context, args *LSIFUploadsQueryArgs) (LSIFUploadConnectionResolver, error)
	LSIFUploadsByRepo(ctx context.Context, args *LSIFRepositoryUploadsQueryArgs) (LSIFUploadConnectionResolver, error)
	DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error)
	DeleteLSIFUploads(ctx context.Context, args *DeleteLSIFUploadsArgs) (*EmptyResponse, error)
}

type PoliciesServiceResolver interface {
	CodeIntelligenceConfigurationPolicies(ctx context.Context, args *CodeIntelligenceConfigurationPoliciesArgs) (CodeIntelligenceConfigurationPolicyConnectionResolver, error)
	ConfigurationPolicyByID(ctx context.Context, id graphql.ID) (CodeIntelligenceConfigurationPolicyResolver, error)
	CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *CreateCodeIntelligenceConfigurationPolicyArgs) (CodeIntelligenceConfigurationPolicyResolver, error)
	DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *DeleteCodeIntelligenceConfigurationPolicyArgs) (*EmptyResponse, error)
	PreviewGitObjectFilter(ctx context.Context, id graphql.ID, args *PreviewGitObjectFilterArgs) (GitObjectFilterPreviewResolver, error)
	PreviewRepositoryFilter(ctx context.Context, args *PreviewRepositoryFilterArgs) (RepositoryFilterPreviewResolver, error)
	UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *UpdateCodeIntelligenceConfigurationPolicyArgs) (*EmptyResponse, error)
}

type CodeIntelRepositorySummaryResolver interface {
	RecentUploads() []LSIFUploadsWithRepositoryNamespaceResolver
	RecentIndexes() []LSIFIndexesWithRepositoryNamespaceResolver
	LastUploadRetentionScan() *gqlutil.DateTime
	LastIndexScan() *gqlutil.DateTime
	AvailableIndexers() []InferredAvailableIndexersResolver
}

type PreciseIndexesQueryArgs struct {
	graphqlutil.ConnectionArgs
	After        *string
	Repo         *graphql.ID
	Query        *string
	States       *[]string
	DependencyOf *string
	DependentOf  *string
}

type LSIFIndexConnectionResolver interface {
	Nodes(ctx context.Context) ([]LSIFIndexResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (PageInfo, error)
}

type LSIFUploadConnectionResolver interface {
	Nodes(ctx context.Context) ([]LSIFUploadResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (PageInfo, error)
}

type PreciseIndexConnectionResolver interface {
	Nodes(ctx context.Context) ([]PreciseIndexResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (PageInfo, error)
}

type PreciseIndexResolver interface {
	ID() graphql.ID
	ProjectRoot(ctx context.Context) (GitTreeEntryResolver, error)
	InputCommit() string
	Tags(ctx context.Context) ([]string, error)
	InputRoot() string
	InputIndexer() string
	Indexer() CodeIntelIndexerResolver
	State() string
	QueuedAt() *gqlutil.DateTime
	UploadedAt() *gqlutil.DateTime
	IndexingStartedAt() *gqlutil.DateTime
	ProcessingStartedAt() *gqlutil.DateTime
	IndexingFinishedAt() *gqlutil.DateTime
	ProcessingFinishedAt() *gqlutil.DateTime
	Steps() IndexStepsResolver
	Failure() *string
	PlaceInQueue() *int32
	ShouldReindex(ctx context.Context) bool
	IsLatestForRepo() bool
	RetentionPolicyOverview(ctx context.Context, args *LSIFUploadRetentionPolicyMatchesArgs) (CodeIntelligenceRetentionPolicyMatchesConnectionResolver, error)
	AuditLogs(ctx context.Context) (*[]LSIFUploadsAuditLogsResolver, error)
}

type PageInfo interface {
	EndCursor() *string
	HasNextPage() bool
}

type RepositoryResolver interface {
	ID() graphql.ID
	Name() string
	Type(ctx context.Context) (*types.Repo, error)
	CommitFromID(ctx context.Context, args *RepositoryCommitArgs, commitID api.CommitID) (GitCommitResolver, error)
	URL() string
	URI(ctx context.Context) (string, error)
	ExternalRepository() ExternalRepositoryResolver
}

type ExternalRepositoryResolver interface {
	ServiceType() string
	ServiceID() string
}

type GitCommitResolver interface {
	ID() graphql.ID
	Repository() RepositoryResolver
	OID() GitObjectID
	AbbreviatedOID() string
	URL() string
}

//
//

type GitObjectType string

func (GitObjectType) ImplementsGraphQLType(name string) bool { return name == "GitObjectType" }

const (
	GitObjectTypeCommit  GitObjectType = "GIT_COMMIT"
	GitObjectTypeTag     GitObjectType = "GIT_TAG"
	GitObjectTypeTree    GitObjectType = "GIT_TREE"
	GitObjectTypeBlob    GitObjectType = "GIT_BLOB"
	GitObjectTypeUnknown GitObjectType = "GIT_UNKNOWN"
)

type GitObjectID string

func (GitObjectID) ImplementsGraphQLType(name string) bool {
	return name == "GitObjectID"
}

func (id *GitObjectID) UnmarshalGraphQL(input any) error {
	if input, ok := input.(string); ok && gitserver.IsAbsoluteRevision(input) {
		*id = GitObjectID(input)
		return nil
	}
	return errors.New("GitObjectID: expected 40-character string (SHA-1 hash)")
}

//
//

type RepositoryCommitArgs struct {
	Rev          string
	InputRevspec *string
}

type RepositoryFilterPreviewResolver interface {
	Nodes() []RepositoryResolver
	TotalCount() int32
	Limit() *int32
	TotalMatches() int32
	MatchesAllRepos() bool
}

type CodeIntelligenceCommitGraphResolver interface {
	Stale(ctx context.Context) (bool, error)
	UpdatedAt(ctx context.Context) (*gqlutil.DateTime, error)
}

type GitObjectFilterPreviewResolver interface {
	Nodes() []CodeIntelGitObjectResolver
	TotalCount() int32
	TotalCountYoungerThanThreshold() *int32
}

type CodeIntelGitObjectResolver interface {
	Name() string
	Rev() string
	CommittedAt() gqlutil.DateTime
}

type GitBlobCodeIntelSupportResolver interface {
	SearchBasedSupport(context.Context) (SearchBasedSupportResolver, error)
	PreciseSupport(context.Context) (PreciseSupportResolver, error)
}

type SearchBasedSupportResolver interface {
	SupportLevel() string
	Language() string
}

type PreciseSupportResolver interface {
	SupportLevel() string
	Indexers() *[]CodeIntelIndexerResolver
}

type CodeIntelIndexerResolver interface {
	Name() string
	URL() string
}

type IndexConfigurationResolver interface {
	Configuration(ctx context.Context) (*string, error)
	InferredConfiguration(ctx context.Context) (*string, error)
}

type GitTreeCodeIntelSupportResolver interface {
	SearchBasedSupport(context.Context) (*[]GitTreeSearchBasedCoverage, error)
	PreciseSupport(context.Context) (*[]GitTreePreciseCoverage, error)
}

type GitTreeSearchBasedCoverage interface {
	CoveredPaths() []string
	Support() SearchBasedSupportResolver
}

type GitTreePreciseCoverage interface {
	Support() PreciseSupportResolver
	Confidence() string
}

type GitBlobLSIFDataResolver interface {
	GitTreeLSIFDataResolver
	ToGitTreeLSIFData() (GitTreeLSIFDataResolver, bool)
	ToGitBlobLSIFData() (GitBlobLSIFDataResolver, bool)

	Stencil(ctx context.Context) ([]RangeResolver, error)
	Ranges(ctx context.Context, args *LSIFRangesArgs) (CodeIntelligenceRangeConnectionResolver, error)
	Definitions(ctx context.Context, args *LSIFQueryPositionArgs) (LocationConnectionResolver, error)
	References(ctx context.Context, args *LSIFPagedQueryPositionArgs) (LocationConnectionResolver, error)
	Implementations(ctx context.Context, args *LSIFPagedQueryPositionArgs) (LocationConnectionResolver, error)
	Hover(ctx context.Context, args *LSIFQueryPositionArgs) (HoverResolver, error)
}

type GitTreeLSIFDataResolver interface {
	LSIFUploads(ctx context.Context) ([]LSIFUploadResolver, error)
	Diagnostics(ctx context.Context, args *LSIFDiagnosticsArgs) (DiagnosticConnectionResolver, error)
}

type CodeIntelligenceRangeConnectionResolver interface {
	Nodes(ctx context.Context) ([]CodeIntelligenceRangeResolver, error)
}

type LocationConnectionResolver interface {
	Nodes(ctx context.Context) ([]LocationResolver, error)
	PageInfo(ctx context.Context) (PageInfo, error)
}

type LSIFDiagnosticsArgs struct {
	graphqlutil.ConnectionArgs
}

type LSIFRangesArgs struct {
	StartLine int32
	EndLine   int32
}

type LSIFPagedQueryPositionArgs struct {
	LSIFQueryPositionArgs
	graphqlutil.ConnectionArgs
	After  *string
	Filter *string
}

type LSIFQueryPositionArgs struct {
	Line      int32
	Character int32
	Filter    *string
}

type RangeResolver interface {
	Start() PositionResolver
	End() PositionResolver
}

type PositionResolver interface {
	Line() int32
	Character() int32
}

type GitTreeContentPageArgs struct {
	StartLine *int32
	EndLine   *int32
}

type GitTreeEntryResolver interface {
	Path() string
	Name() string
	ToGitTree() (GitTreeEntryResolver, bool)
	ToGitBlob() (GitTreeEntryResolver, bool)
	ByteSize(ctx context.Context) (int32, error)
	Content(ctx context.Context, args *GitTreeContentPageArgs) (string, error)
	Commit() GitCommitResolver
	Repository() RepositoryResolver
	CanonicalURL() string
	IsRoot() bool
	IsDirectory() bool
	URL(ctx context.Context) (string, error)
	Submodule() GitSubmoduleResolver
}

type LocationResolver interface {
	Resource() GitTreeEntryResolver
	Range() RangeResolver
	URL(ctx context.Context) (string, error)
	CanonicalURL() string
}

type GitSubmoduleResolver interface {
	URL() string
	Commit() string
	Path() string
}

type Markdown string

func (m Markdown) Text() string {
	return string(m)
}

func (m Markdown) HTML() (string, error) {
	return markdown.Render(string(m))
}

type CodeIntelligenceRangeResolver interface {
	Range(ctx context.Context) (RangeResolver, error)
	Definitions(ctx context.Context) (LocationConnectionResolver, error)
	References(ctx context.Context) (LocationConnectionResolver, error)
	Implementations(ctx context.Context) (LocationConnectionResolver, error)
	Hover(ctx context.Context) (HoverResolver, error)
}

type HoverResolver interface {
	Markdown() Markdown
	Range() RangeResolver
}

type DiagnosticConnectionResolver interface {
	Nodes(ctx context.Context) ([]DiagnosticResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (PageInfo, error)
}

type CodeIntelligenceConfigurationPolicyConnectionResolver interface {
	Nodes(ctx context.Context) ([]CodeIntelligenceConfigurationPolicyResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (PageInfo, error)
}

type RetentionPolicyMatcherResolver interface {
	ConfigurationPolicy() CodeIntelligenceConfigurationPolicyResolver
	Matches() bool
	ProtectingCommits() *[]string
}

type CodeIntelligenceConfigurationPolicyResolver interface {
	ID() graphql.ID
	Repository(ctx context.Context) (RepositoryResolver, error)
	RepositoryPatterns() *[]string
	Name() string
	Type() (GitObjectType, error)
	Pattern() string
	Protected() bool
	RetentionEnabled() bool
	RetentionDurationHours() *int32
	RetainIntermediateCommits() bool
	IndexingEnabled() bool
	IndexCommitMaxAgeHours() *int32
	IndexIntermediateCommits() bool
}

type DiagnosticResolver interface {
	Severity() (*string, error)
	Code() (*string, error)
	Source() (*string, error)
	Message() (*string, error)
	Location(ctx context.Context) (LocationResolver, error)
}

type LSIFIndexResolver interface {
	ID() graphql.ID
	InputCommit() string
	Tags(ctx context.Context) ([]string, error)
	InputRoot() string
	InputIndexer() string
	Indexer() CodeIntelIndexerResolver
	QueuedAt() gqlutil.DateTime
	State() string
	Failure() *string
	StartedAt() *gqlutil.DateTime
	FinishedAt() *gqlutil.DateTime
	Steps() IndexStepsResolver
	PlaceInQueue() *int32
	AssociatedUpload(ctx context.Context) (LSIFUploadResolver, error)
	ShouldReindex(ctx context.Context) bool
	ProjectRoot(ctx context.Context) (GitTreeEntryResolver, error)
}

type IndexStepsResolver interface {
	Setup() []ExecutionLogEntryResolver
	PreIndex() []PreIndexStepResolver
	Index() IndexStepResolver
	Upload() ExecutionLogEntryResolver
	Teardown() []ExecutionLogEntryResolver
}

type PreIndexStepResolver interface {
	Root() string
	Image() string
	Commands() []string
	LogEntry() ExecutionLogEntryResolver
}

type ExecutionLogEntryResolver interface {
	Key() string
	Command() []string
	StartTime() gqlutil.DateTime
	ExitCode() *int32
	Out(ctx context.Context) (string, error)
	DurationMilliseconds() *int32
}

type UploadDocumentPathsConnectionResolver interface {
	Nodes(ctx context.Context) ([]string, error)
	TotalCount(ctx context.Context) (*int32, error)
}

type LSIFUploadResolver interface {
	ID() graphql.ID
	InputCommit() string
	Tags(ctx context.Context) ([]string, error)
	InputRoot() string
	IsLatestForRepo() bool
	UploadedAt() gqlutil.DateTime
	State() string
	Failure() *string
	StartedAt() *gqlutil.DateTime
	FinishedAt() *gqlutil.DateTime
	InputIndexer() string
	Indexer() CodeIntelIndexerResolver
	PlaceInQueue() *int32
	AssociatedIndex(ctx context.Context) (LSIFIndexResolver, error)
	ProjectRoot(ctx context.Context) (GitTreeEntryResolver, error)
	RetentionPolicyOverview(ctx context.Context, args *LSIFUploadRetentionPolicyMatchesArgs) (CodeIntelligenceRetentionPolicyMatchesConnectionResolver, error)
	DocumentPaths(ctx context.Context, args *LSIFUploadDocumentPathsQueryArgs) (LSIFUploadDocumentPathsConnectionResolver, error)
	AuditLogs(ctx context.Context) (*[]LSIFUploadsAuditLogsResolver, error)
}

type LSIFUploadRetentionPolicyMatchesArgs struct {
	MatchesOnly bool
	First       *int32
	After       *string
	Query       *string
}

type LSIFUploadDocumentPathsQueryArgs struct {
	Pattern string
}

type LSIFUploadsAuditLogsResolver interface {
	LogTimestamp() gqlutil.DateTime
	UploadDeletedAt() *gqlutil.DateTime
	Reason() *string
	ChangedColumns() []AuditLogColumnChange
	UploadID() graphql.ID
	InputCommit() string
	InputRoot() string
	InputIndexer() string
	UploadedAt() gqlutil.DateTime
	Operation() string
	// AssociatedIndex(ctx context.Context) (LSIFIndexResolver, error)
}

type IndexStepResolver interface {
	IndexerArgs() []string
	Outfile() *string
	LogEntry() ExecutionLogEntryResolver
}

type AuditLogColumnChange interface {
	Column() string
	Old() *string
	New() *string
}

type AuditLogColumnChangeResolver interface {
	Column() string
	Old() *string
	New() *string
}

type CodeIntelligenceRetentionPolicyMatchesConnectionResolver interface {
	Nodes(ctx context.Context) ([]CodeIntelligenceRetentionPolicyMatchResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (PageInfo, error)
}

type CodeIntelligenceRetentionPolicyMatchResolver interface {
	ConfigurationPolicy() CodeIntelligenceConfigurationPolicyResolver
	Matches() bool
	ProtectingCommits() *[]string
}

type LSIFUploadDocumentPathsConnectionResolver interface {
	Nodes(ctx context.Context) ([]string, error)
	TotalCount(ctx context.Context) (*int32, error)
}

type GitBlobLSIFDataArgs struct {
	Repo      *types.Repo
	Commit    api.CommitID
	Path      string
	ExactPath bool
	ToolName  string
}

type GitTreeEntryCodeIntelInfoArgs struct {
	Repo   *types.Repo
	Path   string
	Commit string
}

type RequestLanguageSupportArgs struct {
	Language string
}

// EmptyResponse is a type that can be used in the return signature for graphql queries
// that don't require a return value.
type EmptyResponse struct{}

// AlwaysNil exists since various graphql tools expect at least one field to be
// present in the schema so we provide a dummy one here that is always nil.
func (er *EmptyResponse) AlwaysNil() *string {
	return nil
}

type LSIFIndexesQueryArgs struct {
	graphqlutil.ConnectionArgs
	Query *string
	State *string
	After *string
}

type ReindexLSIFIndexesArgs struct {
	Query      *string
	State      *string
	Repository *graphql.ID
}

type DeleteLSIFIndexesArgs struct {
	Query      *string
	State      *string
	Repository *graphql.ID
}

type DeletePreciseIndexesArgs struct {
	Query           *string
	States          *[]string
	Repository      *graphql.ID
	IsLatestForRepo *bool
}

type ReindexPreciseIndexesArgs struct {
	Query           *string
	States          *[]string
	Repository      *graphql.ID
	IsLatestForRepo *bool
}

type LSIFRepositoryIndexesQueryArgs struct {
	*LSIFIndexesQueryArgs
	RepositoryID graphql.ID
}

type LSIFUploadsQueryArgs struct {
	graphqlutil.ConnectionArgs
	Query           *string
	State           *string
	IsLatestForRepo *bool
	DependencyOf    *graphql.ID
	DependentOf     *graphql.ID
	After           *string
	IncludeDeleted  *bool
}

type LSIFRepositoryUploadsQueryArgs struct {
	*LSIFUploadsQueryArgs
	RepositoryID graphql.ID
}

type DeleteLSIFUploadsArgs struct {
	Query           *string
	State           *string
	IsLatestForRepo *bool
	Repository      *graphql.ID
}

type LSIFIndexesWithRepositoryNamespaceResolver interface {
	Root() string
	Indexer() CodeIntelIndexerResolver
	Indexes() []LSIFIndexResolver
}

type QueueAutoIndexJobsForRepoArgs struct {
	Repository    graphql.ID
	Rev           *string
	Configuration *string
}

type UpdateRepositoryIndexConfigurationArgs struct {
	Repository    graphql.ID
	Configuration string
}

type UpdateCodeIntelligenceInferenceScriptArgs struct {
	Script string
}

type LSIFUploadsWithRepositoryNamespaceResolver interface {
	Root() string
	Indexer() CodeIntelIndexerResolver
	Uploads() []LSIFUploadResolver
}

type CodeIntelligenceConfigurationPoliciesArgs struct {
	graphqlutil.ConnectionArgs
	Repository       *graphql.ID
	Query            *string
	ForDataRetention *bool
	ForIndexing      *bool
	Protected        *bool
	After            *string
}

type CodeIntelConfigurationPolicy struct {
	Name                      string
	RepositoryID              *int32
	RepositoryPatterns        *[]string
	Type                      GitObjectType
	Pattern                   string
	RetentionEnabled          bool
	RetentionDurationHours    *int32
	RetainIntermediateCommits bool
	IndexingEnabled           bool
	IndexCommitMaxAgeHours    *int32
	IndexIntermediateCommits  bool
}

type CreateCodeIntelligenceConfigurationPolicyArgs struct {
	Repository *graphql.ID
	CodeIntelConfigurationPolicy
}

type UpdateCodeIntelligenceConfigurationPolicyArgs struct {
	ID         graphql.ID
	Repository *graphql.ID
	CodeIntelConfigurationPolicy
}

type DeleteCodeIntelligenceConfigurationPolicyArgs struct {
	Policy graphql.ID
}

type PreviewGitObjectFilterArgs struct {
	graphqlutil.ConnectionArgs
	Type                         GitObjectType
	Pattern                      string
	CountObjectsYoungerThanHours *int32
}

type PreviewRepositoryFilterArgs struct {
	graphqlutil.ConnectionArgs
	Patterns []string
}

type InferredAvailableIndexersResolver interface {
	Roots() []string
	Index() string
	URL() string
}

type inferredAvailableIndexersResolver struct {
	roots []string
	index string
	url   string
}

func NewInferredAvailableIndexersResolver(index string, roots []string, url string) InferredAvailableIndexersResolver {
	return &inferredAvailableIndexersResolver{
		index: index,
		roots: roots,
		url:   url,
	}
}

func (r *inferredAvailableIndexersResolver) Roots() []string {
	return r.roots
}

func (r *inferredAvailableIndexersResolver) Index() string {
	return r.index
}

func (r *inferredAvailableIndexersResolver) URL() string {
	return r.url
}
