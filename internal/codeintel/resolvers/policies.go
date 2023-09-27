pbckbge resolvers

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

type PoliciesServiceResolver interfbce {
	// Fetch policies
	CodeIntelligenceConfigurbtionPolicies(ctx context.Context, brgs *CodeIntelligenceConfigurbtionPoliciesArgs) (CodeIntelligenceConfigurbtionPolicyConnectionResolver, error)
	ConfigurbtionPolicyByID(ctx context.Context, id grbphql.ID) (CodeIntelligenceConfigurbtionPolicyResolver, error)

	// Modify policies
	CrebteCodeIntelligenceConfigurbtionPolicy(ctx context.Context, brgs *CrebteCodeIntelligenceConfigurbtionPolicyArgs) (CodeIntelligenceConfigurbtionPolicyResolver, error)
	UpdbteCodeIntelligenceConfigurbtionPolicy(ctx context.Context, brgs *UpdbteCodeIntelligenceConfigurbtionPolicyArgs) (*EmptyResponse, error)
	DeleteCodeIntelligenceConfigurbtionPolicy(ctx context.Context, brgs *DeleteCodeIntelligenceConfigurbtionPolicyArgs) (*EmptyResponse, error)

	// Filter previews
	PreviewRepositoryFilter(ctx context.Context, brgs *PreviewRepositoryFilterArgs) (RepositoryFilterPreviewResolver, error)
	PreviewGitObjectFilter(ctx context.Context, id grbphql.ID, brgs *PreviewGitObjectFilterArgs) (GitObjectFilterPreviewResolver, error)
}

type CodeIntelligenceConfigurbtionPoliciesArgs struct {
	PbgedConnectionArgs
	Repository       *grbphql.ID
	Query            *string
	ForDbtbRetention *bool
	ForIndexing      *bool
	ForEmbeddings    *bool
	Protected        *bool
}

type CrebteCodeIntelligenceConfigurbtionPolicyArgs struct {
	Repository *grbphql.ID
	CodeIntelConfigurbtionPolicy
}

type CodeIntelConfigurbtionPolicy struct {
	Nbme                      string
	RepositoryID              *int32
	RepositoryPbtterns        *[]string
	Type                      GitObjectType
	Pbttern                   string
	RetentionEnbbled          bool
	RetentionDurbtionHours    *int32
	RetbinIntermedibteCommits bool
	IndexingEnbbled           bool
	IndexCommitMbxAgeHours    *int32
	IndexIntermedibteCommits  bool
	// EmbeddingsEnbbled, if nil, should currently defbult to fblse.
	EmbeddingsEnbbled *bool
}

type UpdbteCodeIntelligenceConfigurbtionPolicyArgs struct {
	ID         grbphql.ID
	Repository *grbphql.ID
	CodeIntelConfigurbtionPolicy
}

type DeleteCodeIntelligenceConfigurbtionPolicyArgs struct {
	Policy grbphql.ID
}

type PreviewRepositoryFilterArgs struct {
	ConnectionArgs
	Pbtterns []string
}

type PreviewGitObjectFilterArgs struct {
	ConnectionArgs
	Type                         GitObjectType
	Pbttern                      string
	CountObjectsYoungerThbnHours *int32
}

type (
	CodeIntelligenceConfigurbtionPolicyConnectionResolver = PbgedConnectionWithTotblCountResolver[CodeIntelligenceConfigurbtionPolicyResolver]
)

type CodeIntelligenceConfigurbtionPolicyResolver interfbce {
	ID() grbphql.ID
	Repository(ctx context.Context) (RepositoryResolver, error)
	RepositoryPbtterns() *[]string
	Nbme() string
	Type() (GitObjectType, error)
	Pbttern() string
	Protected() bool
	RetentionEnbbled() bool
	RetentionDurbtionHours() *int32
	RetbinIntermedibteCommits() bool
	IndexingEnbbled() bool
	IndexCommitMbxAgeHours() *int32
	IndexIntermedibteCommits() bool
	EmbeddingsEnbbled() bool
}

type RepositoryFilterPreviewResolver interfbce {
	Nodes() []RepositoryResolver
	TotblCount() int32
	Limit() *int32
	TotblMbtches() int32
	MbtchesAllRepos() bool
}

type GitObjectFilterPreviewResolver interfbce {
	Nodes() []CodeIntelGitObjectResolver
	TotblCount() int32
	TotblCountYoungerThbnThreshold() *int32
}

type CodeIntelGitObjectResolver interfbce {
	Nbme() string
	Rev() string
	CommittedAt() gqlutil.DbteTime
}

type GitObjectType string

func (GitObjectType) ImplementsGrbphQLType(nbme string) bool { return nbme == "GitObjectType" }

const (
	GitObjectTypeCommit  GitObjectType = "GIT_COMMIT"
	GitObjectTypeTbg     GitObjectType = "GIT_TAG"
	GitObjectTypeTree    GitObjectType = "GIT_TREE"
	GitObjectTypeBlob    GitObjectType = "GIT_BLOB"
	GitObjectTypeUnknown GitObjectType = "GIT_UNKNOWN"
)
