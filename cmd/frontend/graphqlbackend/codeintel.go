package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type CodeIntelResolver interface {
	LSIFUploadByID(ctx context.Context, id graphql.ID) (LSIFUploadResolver, error)
	LSIFUploads(ctx context.Context, args *LSIFUploadsQueryArgs) (LSIFUploadConnectionResolver, error)
	LSIFUploadsByRepo(ctx context.Context, args *LSIFRepositoryUploadsQueryArgs) (LSIFUploadConnectionResolver, error)
	DeleteLSIFUpload(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error)
	LSIFIndexByID(ctx context.Context, id graphql.ID) (LSIFIndexResolver, error)
	LSIFIndexes(ctx context.Context, args *LSIFIndexesQueryArgs) (LSIFIndexConnectionResolver, error)
	LSIFIndexesByRepo(ctx context.Context, args *LSIFRepositoryIndexesQueryArgs) (LSIFIndexConnectionResolver, error)
	DeleteLSIFIndex(ctx context.Context, args *struct{ ID graphql.ID }) (*EmptyResponse, error)
	CommitGraph(ctx context.Context, id graphql.ID) (CodeIntelligenceCommitGraphResolver, error)
	QueueAutoIndexJobsForRepo(ctx context.Context, args *QueueAutoIndexJobsForRepoArgs) ([]LSIFIndexResolver, error)
	GitBlobLSIFData(ctx context.Context, args *GitBlobLSIFDataArgs) (GitBlobLSIFDataResolver, error)
	CodeIntelligenceConfigurationPolicies(ctx context.Context, args *CodeIntelligenceConfigurationPoliciesArgs) ([]CodeIntelligenceConfigurationPolicyResolver, error)
	CreateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *CreateCodeIntelligenceConfigurationPolicyArgs) (CodeIntelligenceConfigurationPolicyResolver, error)
	UpdateCodeIntelligenceConfigurationPolicy(ctx context.Context, args *UpdateCodeIntelligenceConfigurationPolicyArgs) (*EmptyResponse, error)
	DeleteCodeIntelligenceConfigurationPolicy(ctx context.Context, args *DeleteCodeIntelligenceConfigurationPolicyArgs) (*EmptyResponse, error)
	IndexConfiguration(ctx context.Context, id graphql.ID) (IndexConfigurationResolver, error) // TODO - rename ...ForRepo
	UpdateRepositoryIndexConfiguration(ctx context.Context, args *UpdateRepositoryIndexConfigurationArgs) (*EmptyResponse, error)
	PreviewGitObjectFilter(ctx context.Context, id graphql.ID, args *PreviewGitObjectFilterArgs) ([]GitObjectFilterPreviewResolver, error)
	NodeResolvers() map[string]NodeByIDFunc
	DocumentationSearch(ctx context.Context, args *DocumentationSearchArgs) (DocumentationSearchResultsResolver, error)
}

type LSIFUploadsQueryArgs struct {
	graphqlutil.ConnectionArgs
	Query           *string
	State           *string
	IsLatestForRepo *bool
	DependencyOf    *graphql.ID
	DependentOf     *graphql.ID
	After           *string
}

type LSIFRepositoryUploadsQueryArgs struct {
	*LSIFUploadsQueryArgs
	RepositoryID graphql.ID
}

type LSIFUploadResolver interface {
	ID() graphql.ID
	InputCommit() string
	InputRoot() string
	IsLatestForRepo() bool
	UploadedAt() DateTime
	State() string
	Failure() *string
	StartedAt() *DateTime
	FinishedAt() *DateTime
	InputIndexer() string
	PlaceInQueue() *int32
	AssociatedIndex(ctx context.Context) (LSIFIndexResolver, error)
	ProjectRoot(ctx context.Context) (*GitTreeEntryResolver, error)
}

type LSIFUploadConnectionResolver interface {
	Nodes(ctx context.Context) ([]LSIFUploadResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type LSIFIndexesQueryArgs struct {
	graphqlutil.ConnectionArgs
	Query *string
	State *string
	After *string
}

type LSIFRepositoryIndexesQueryArgs struct {
	*LSIFIndexesQueryArgs
	RepositoryID graphql.ID
}

type LSIFIndexResolver interface {
	ID() graphql.ID
	InputCommit() string
	InputRoot() string
	InputIndexer() string
	QueuedAt() DateTime
	State() string
	Failure() *string
	StartedAt() *DateTime
	FinishedAt() *DateTime
	Steps() IndexStepsResolver
	PlaceInQueue() *int32
	AssociatedUpload(ctx context.Context) (LSIFUploadResolver, error)
	ProjectRoot(ctx context.Context) (*GitTreeEntryResolver, error)
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

type IndexStepResolver interface {
	IndexerArgs() []string
	Outfile() *string
	LogEntry() ExecutionLogEntryResolver
}

type LSIFIndexConnectionResolver interface {
	Nodes(ctx context.Context) ([]LSIFIndexResolver, error)
	TotalCount(ctx context.Context) (*int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type QueueAutoIndexJobsForRepoArgs struct {
	Repository    graphql.ID
	Rev           *string
	Configuration *string
}

type GitTreeLSIFDataResolver interface {
	Diagnostics(ctx context.Context, args *LSIFDiagnosticsArgs) (DiagnosticConnectionResolver, error)
	DocumentationPage(ctx context.Context, args *LSIFDocumentationPageArgs) (DocumentationPageResolver, error)
	DocumentationPathInfo(ctx context.Context, args *LSIFDocumentationPathInfoArgs) (JSONValue, error)
	DocumentationDefinitions(ctx context.Context, args *LSIFQueryDocumentationArgs) (LocationConnectionResolver, error)
	DocumentationReferences(ctx context.Context, args *LSIFPagedQueryDocumentationArgs) (LocationConnectionResolver, error)
}

type CodeIntelligenceCommitGraphResolver interface {
	Stale(ctx context.Context) (bool, error)
	UpdatedAt(ctx context.Context) (*DateTime, error)
}

type GitBlobLSIFDataResolver interface {
	GitTreeLSIFDataResolver
	ToGitTreeLSIFData() (GitTreeLSIFDataResolver, bool)
	ToGitBlobLSIFData() (GitBlobLSIFDataResolver, bool)

	Stencil(ctx context.Context) ([]RangeResolver, error)
	Ranges(ctx context.Context, args *LSIFRangesArgs) (CodeIntelligenceRangeConnectionResolver, error)
	Definitions(ctx context.Context, args *LSIFQueryPositionArgs) (LocationConnectionResolver, error)
	References(ctx context.Context, args *LSIFPagedQueryPositionArgs) (LocationConnectionResolver, error)
	Hover(ctx context.Context, args *LSIFQueryPositionArgs) (HoverResolver, error)
	Documentation(ctx context.Context, args *LSIFQueryPositionArgs) (DocumentationResolver, error)
}

type GitBlobLSIFDataArgs struct {
	Repo      *types.Repo
	Commit    api.CommitID
	Path      string
	ExactPath bool
	ToolName  string
}

type LSIFRangesArgs struct {
	StartLine int32
	EndLine   int32
}

type LSIFQueryPositionArgs struct {
	Line      int32
	Character int32
}

type LSIFPagedQueryPositionArgs struct {
	LSIFQueryPositionArgs
	graphqlutil.ConnectionArgs
	After *string
}

type LSIFQueryDocumentationArgs struct {
	PathID string
}

type LSIFPagedQueryDocumentationArgs struct {
	PathID string
	graphqlutil.ConnectionArgs
	After *string
}

type LSIFDiagnosticsArgs struct {
	graphqlutil.ConnectionArgs
}

type CodeIntelligenceRangeConnectionResolver interface {
	Nodes(ctx context.Context) ([]CodeIntelligenceRangeResolver, error)
}

type CodeIntelligenceRangeResolver interface {
	Range(ctx context.Context) (RangeResolver, error)
	Definitions(ctx context.Context) (LocationConnectionResolver, error)
	References(ctx context.Context) (LocationConnectionResolver, error)
	Hover(ctx context.Context) (HoverResolver, error)
	Documentation(ctx context.Context) (DocumentationResolver, error)
}

type LocationConnectionResolver interface {
	Nodes(ctx context.Context) ([]LocationResolver, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type HoverResolver interface {
	Markdown() Markdown
	Range() RangeResolver
}

type DocumentationResolver interface {
	PathID() string
}

type DiagnosticConnectionResolver interface {
	Nodes(ctx context.Context) ([]DiagnosticResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type DiagnosticResolver interface {
	Severity() (*string, error)
	Code() (*string, error)
	Source() (*string, error)
	Message() (*string, error)
	Location(ctx context.Context) (LocationResolver, error)
}

type CodeIntelConfigurationPolicy struct {
	Name                      string
	Type                      GitObjectType
	Pattern                   string
	RetentionEnabled          bool
	RetentionDurationHours    *int32
	RetainIntermediateCommits bool
	IndexingEnabled           bool
	IndexCommitMaxAgeHours    *int32
	IndexIntermediateCommits  bool
}

type CodeIntelligenceConfigurationPoliciesArgs struct {
	Repository *graphql.ID
}

type CreateCodeIntelligenceConfigurationPolicyArgs struct {
	Repository *graphql.ID
	CodeIntelConfigurationPolicy
}

type UpdateCodeIntelligenceConfigurationPolicyArgs struct {
	ID graphql.ID
	CodeIntelConfigurationPolicy
}

type DeleteCodeIntelligenceConfigurationPolicyArgs struct {
	Policy graphql.ID
}

type IndexConfigurationResolver interface {
	Configuration(ctx context.Context) (*string, error)
	InferredConfiguration(ctx context.Context) (*string, error)
}

type UpdateRepositoryIndexConfigurationArgs struct {
	Repository    graphql.ID
	Configuration string
}

type PreviewGitObjectFilterArgs struct {
	Type    GitObjectType
	Pattern string
}

type GitObjectFilterPreviewResolver interface {
	Name() string
	Rev() string
}

type CodeIntelligenceConfigurationPolicyResolver interface {
	ID() graphql.ID
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
