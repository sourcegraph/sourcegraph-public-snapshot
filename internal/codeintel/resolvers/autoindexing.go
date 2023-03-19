package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type AutoindexingServiceResolver interface {
	// Fetch precise indexes
	PreciseIndexes(ctx context.Context, args *PreciseIndexesQueryArgs) (PreciseIndexConnectionResolver, error)
	PreciseIndexByID(ctx context.Context, id graphql.ID) (PreciseIndexResolver, error)
	IndexerKeys(ctx context.Context, args *IndexerKeyQueryArgs) ([]string, error)

	// Modify precise indexes
	DeletePreciseIndex(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error)
	DeletePreciseIndexes(ctx context.Context, args *DeletePreciseIndexesArgs) (*EmptyResponse, error)
	ReindexPreciseIndex(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error)
	ReindexPreciseIndexes(ctx context.Context, args *ReindexPreciseIndexesArgs) (*EmptyResponse, error)

	// Inference configuration
	CodeIntelligenceInferenceScript(ctx context.Context) (string, error)
	UpdateCodeIntelligenceInferenceScript(ctx context.Context, args *UpdateCodeIntelligenceInferenceScriptArgs) (*EmptyResponse, error)

	// Repository configuration
	IndexConfiguration(ctx context.Context, id graphql.ID) (IndexConfigurationResolver, error)
	UpdateRepositoryIndexConfiguration(ctx context.Context, args *UpdateRepositoryIndexConfigurationArgs) (*EmptyResponse, error)

	// Inference
	InferAutoIndexJobsForRepo(ctx context.Context, args *InferAutoIndexJobsForRepoArgs) ([]AutoIndexJobDescriptionResolver, error)
	QueueAutoIndexJobsForRepo(ctx context.Context, args *QueueAutoIndexJobsForRepoArgs) ([]PreciseIndexResolver, error)

	// Coverage
	CodeIntelSummary(ctx context.Context) (CodeIntelSummaryResolver, error)
	RepositorySummary(ctx context.Context, id graphql.ID) (CodeIntelRepositorySummaryResolver, error)
	GitBlobCodeIntelInfo(ctx context.Context, args *GitTreeEntryCodeIntelInfoArgs) (GitBlobCodeIntelSupportResolver, error)
	GitTreeCodeIntelInfo(ctx context.Context, args *GitTreeEntryCodeIntelInfoArgs) (GitTreeCodeIntelSupportResolver, error)
}

type PreciseIndexesQueryArgs struct {
	PagedConnectionArgs
	Repo           *graphql.ID
	Query          *string
	States         *[]string
	IndexerKey     *string
	DependencyOf   *string
	DependentOf    *string
	IncludeDeleted *bool
}

type IndexerKeyQueryArgs struct {
	Repo *graphql.ID
}

type DeletePreciseIndexesArgs struct {
	Query           *string
	States          *[]string
	IndexerKey      *string
	Repository      *graphql.ID
	IsLatestForRepo *bool
}

type ReindexPreciseIndexesArgs struct {
	Query           *string
	States          *[]string
	IndexerKey      *string
	Repository      *graphql.ID
	IsLatestForRepo *bool
}

type UpdateCodeIntelligenceInferenceScriptArgs struct {
	Script string
}

type UpdateRepositoryIndexConfigurationArgs struct {
	Repository    graphql.ID
	Configuration string
}

type InferAutoIndexJobsForRepoArgs struct {
	Repository graphql.ID
	Rev        *string
	Script     *string
}

type QueueAutoIndexJobsForRepoArgs struct {
	Repository    graphql.ID
	Rev           *string
	Configuration *string
}

type GitTreeEntryCodeIntelInfoArgs struct {
	Repo   *types.Repo
	Path   string
	Commit string
}

type (
	PreciseIndexConnectionResolver = PagedConnectionResolverWithCount[PreciseIndexResolver]
)

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

type LSIFUploadRetentionPolicyMatchesArgs struct {
	MatchesOnly bool
	PagedConnectionArgs
	Query *string
}

type CodeIntelligenceRetentionPolicyMatchesConnectionResolver = PagedConnectionResolverWithCount[CodeIntelligenceRetentionPolicyMatchResolver]

type CodeIntelligenceRetentionPolicyMatchResolver interface {
	ConfigurationPolicy() CodeIntelligenceConfigurationPolicyResolver
	Matches() bool
	ProtectingCommits() *[]string
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
}

type AuditLogColumnChange interface {
	Column() string
	Old() *string
	New() *string
}

type IndexConfigurationResolver interface {
	Configuration(ctx context.Context) (*string, error)
	ParsedConfiguration(ctx context.Context) (*[]AutoIndexJobDescriptionResolver, error)
	InferredConfiguration(ctx context.Context) (InferredConfigurationResolver, error)
}

type AutoIndexJobDescriptionResolver interface {
	Root() string
	Indexer() CodeIntelIndexerResolver
	ComparisonKey() string
	Steps() IndexStepsResolver
}

type CodeIntelIndexerResolver interface {
	Key() string
	Name() string
	URL() string
	ImageName() *string
}

type IndexStepsResolver interface {
	Setup() []ExecutionLogEntryResolver
	PreIndex() []PreIndexStepResolver
	Index() IndexStepResolver
	Upload() ExecutionLogEntryResolver
	Teardown() []ExecutionLogEntryResolver
}

type ExecutionLogEntryResolver interface {
	Key() string
	Command() []string
	StartTime() gqlutil.DateTime
	ExitCode() *int32
	Out(ctx context.Context) (string, error)
	DurationMilliseconds() *int32
}

type PreIndexStepResolver interface {
	Root() string
	Image() string
	Commands() []string
	LogEntry() ExecutionLogEntryResolver
}

type IndexStepResolver interface {
	Commands() []string
	IndexerArgs() []string
	Outfile() *string
	RequestedEnvVars() *[]string
	LogEntry() ExecutionLogEntryResolver
}

type InferredConfigurationResolver interface {
	Configuration() string
	ParsedConfiguration(ctx context.Context) (*[]AutoIndexJobDescriptionResolver, error)
	LimitError() *string
}

type CodeIntelSummaryResolver interface {
	NumRepositoriesWithCodeIntelligence(ctx context.Context) (int32, error)
	RepositoriesWithErrors(ctx context.Context, args *RepositoriesWithErrorsArgs) (CodeIntelRepositoryWithErrorConnectionResolver, error)
	RepositoriesWithConfiguration(ctx context.Context, args *RepositoriesWithConfigurationArgs) (CodeIntelRepositoryWithConfigurationConnectionResolver, error)
}

type (
	RepositoriesWithErrorsArgs                             = PagedConnectionArgs
	RepositoriesWithConfigurationArgs                      = PagedConnectionArgs
	CodeIntelRepositoryWithErrorConnectionResolver         = PagedConnectionResolverWithCount[CodeIntelRepositoryWithErrorResolver]
	CodeIntelRepositoryWithConfigurationConnectionResolver = PagedConnectionResolverWithCount[CodeIntelRepositoryWithConfigurationResolver]
)

type CodeIntelRepositoryWithErrorResolver interface {
	Repository() RepositoryResolver
	Count() int32
}

type CodeIntelRepositoryWithConfigurationResolver interface {
	Repository() RepositoryResolver
	Indexers() []IndexerWithCountResolver
}

type IndexerWithCountResolver interface {
	Indexer() CodeIntelIndexerResolver
	Count() int32
}

type CodeIntelRepositorySummaryResolver interface {
	RecentActivity(ctx context.Context) ([]PreciseIndexResolver, error)
	LastUploadRetentionScan() *gqlutil.DateTime
	LastIndexScan() *gqlutil.DateTime
	AvailableIndexers() []InferredAvailableIndexersResolver
	LimitError() *string
}

type InferredAvailableIndexersResolver interface {
	Indexer() CodeIntelIndexerResolver
	Roots() []string
	RootsWithKeys() []RootsWithKeyResolver
}

type RootsWithKeyResolver interface {
	Root() string
	ComparisonKey() string
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

type GitTreeCodeIntelSupportResolver interface {
	SearchBasedSupport(context.Context) (*[]GitTreeSearchBasedCoverage, error)
	PreciseSupport(context.Context) (GitTreePreciseCoverageErrorResolver, error)
}

type GitTreeSearchBasedCoverage interface {
	CoveredPaths() []string
	Support() SearchBasedSupportResolver
}

type GitTreePreciseCoverageErrorResolver interface {
	Coverage() []GitTreePreciseCoverage
	LimitError() *string
}

type GitTreePreciseCoverage interface {
	Support() PreciseSupportResolver
	Confidence() string
}
