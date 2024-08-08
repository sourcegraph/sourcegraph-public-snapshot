package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type ListOwnershipArgs struct {
	First   *int32
	After   *string
	Reasons *[]OwnershipReasonType
}

type OwnershipReasonType string

const (
	CodeownersFileEntry              OwnershipReasonType = "CODEOWNERS_FILE_ENTRY"
	AssignedOwner                    OwnershipReasonType = "ASSIGNED_OWNER"
	RecentContributorOwnershipSignal OwnershipReasonType = "RECENT_CONTRIBUTOR_OWNERSHIP_SIGNAL"
	RecentViewOwnershipSignal        OwnershipReasonType = "RECENT_VIEW_OWNERSHIP_SIGNAL"
)

func (args *ListOwnershipArgs) IncludeReason(reason OwnershipReasonType) bool {
	rs := args.Reasons
	// When the reasons list is empty, we do not filter - the result
	// contains all the reasons, so for every reason we return true.
	if rs == nil || len(*rs) == 0 {
		return true
	}
	for _, r := range *rs {
		if r == reason {
			return true
		}
	}
	return false
}

type OwnResolver interface {
	GitBlobOwnership(ctx context.Context, blob *GitTreeEntryResolver, args ListOwnershipArgs) (OwnershipConnectionResolver, error)
	GitCommitOwnership(ctx context.Context, commit *GitCommitResolver, args ListOwnershipArgs) (OwnershipConnectionResolver, error)
	GitTreeOwnership(ctx context.Context, tree *GitTreeEntryResolver, args ListOwnershipArgs) (OwnershipConnectionResolver, error)

	GitTreeOwnershipStats(ctx context.Context, tree *GitTreeEntryResolver) (OwnershipStatsResolver, error)
	InstanceOwnershipStats(ctx context.Context) (OwnershipStatsResolver, error)

	PersonOwnerField(person *PersonResolver) string
	UserOwnerField(user *UserResolver) string
	TeamOwnerField(team *TeamResolver) string

	NodeResolvers() map[string]NodeByIDFunc

	// Codeowners queries.
	CodeownersIngestedFiles(context.Context, *CodeownersIngestedFilesArgs) (CodeownersIngestedFileConnectionResolver, error)
	RepoIngestedCodeowners(context.Context, api.RepoID) (CodeownersIngestedFileResolver, error)

	// Codeowners mutations.
	AddCodeownersFile(context.Context, *CodeownersFileArgs) (CodeownersIngestedFileResolver, error)
	UpdateCodeownersFile(context.Context, *CodeownersFileArgs) (CodeownersIngestedFileResolver, error)
	DeleteCodeownersFiles(context.Context, *DeleteCodeownersFileArgs) (*EmptyResponse, error)

	// Assigned ownership mutations.
	AssignOwner(context.Context, *AssignOwnerOrTeamArgs) (*EmptyResponse, error)
	RemoveAssignedOwner(context.Context, *AssignOwnerOrTeamArgs) (*EmptyResponse, error)
	AssignTeam(context.Context, *AssignOwnerOrTeamArgs) (*EmptyResponse, error)
	RemoveAssignedTeam(context.Context, *AssignOwnerOrTeamArgs) (*EmptyResponse, error)

	// Config.
	OwnSignalConfigurations(ctx context.Context) ([]SignalConfigurationResolver, error)
	UpdateOwnSignalConfigurations(ctx context.Context, configurationsArgs UpdateSignalConfigurationsArgs) ([]SignalConfigurationResolver, error)
}

type OwnershipConnectionResolver interface {
	TotalCount(context.Context) (int32, error)
	TotalOwners(context.Context) (int32, error)
	PageInfo(context.Context) (*gqlutil.PageInfo, error)
	Nodes(context.Context) ([]OwnershipResolver, error)
}

type OwnershipStatsResolver interface {
	TotalFiles(context.Context) (int32, error)
	TotalCodeownedFiles(context.Context) (int32, error)
	TotalOwnedFiles(context.Context) (int32, error)
	TotalAssignedOwnershipFiles(context.Context) (int32, error)
	UpdatedAt(ctx context.Context) (*gqlutil.DateTime, error)
}

type Ownable interface {
	ToGitBlob(context.Context) (*GitTreeEntryResolver, bool)
}

type OwnershipResolver interface {
	Owner(context.Context) (OwnerResolver, error)
	Reasons(context.Context) ([]OwnershipReasonResolver, error)
}

type OwnerResolver interface {
	OwnerField(context.Context) (string, error)

	ToPerson() (*PersonResolver, bool)
	ToTeam() (*TeamResolver, bool)
}

type OwnershipReasonResolver interface {
	SimpleOwnReasonResolver
	ToCodeownersFileEntry() (CodeownersFileEntryResolver, bool)
	ToRecentContributorOwnershipSignal() (RecentContributorOwnershipSignalResolver, bool)
	ToRecentViewOwnershipSignal() (RecentViewOwnershipSignalResolver, bool)
	ToAssignedOwner() (AssignedOwnerResolver, bool)
}

type SimpleOwnReasonResolver interface {
	Title() (string, error)
	Description() (string, error)
}

type CodeownersFileEntryResolver interface {
	Title() (string, error)
	Description() (string, error)
	CodeownersFile(context.Context) (FileResolver, error)
	RuleLineMatch(context.Context) (int32, error)
}

type RecentContributorOwnershipSignalResolver interface {
	Title() (string, error)
	Description() (string, error)
}

type RecentViewOwnershipSignalResolver interface {
	Title() (string, error)
	Description() (string, error)
}

type AssignedOwnerResolver interface {
	Title() (string, error)
	Description() (string, error)
	IsDirectMatch() bool
}

type AssignedTeamResolver interface {
	Title() (string, error)
	Description() (string, error)
}

type CodeownersFileArgs struct {
	Input CodeownersFileInput
}

type CodeownersFileInput struct {
	FileContents string
	RepoID       *graphql.ID
	RepoName     *string
}

type DeleteCodeownersFilesInput struct {
	RepoID   *graphql.ID
	RepoName *string
}

type AssignOwnerOrTeamArgs struct {
	Input AssignOwnerOrTeamInput
}

type AssignOwnerOrTeamInput struct {
	// AssignedOwnerID is an ID of a user or a team who is assigned as an owner.
	AssignedOwnerID graphql.ID
	RepoID          graphql.ID
	AbsolutePath    string
}

type DeleteCodeownersFileArgs struct {
	Repositories []DeleteCodeownersFilesInput
}

type CodeownersIngestedFilesArgs struct {
	First *int32
	After *string
}

type CodeownersIngestedFileResolver interface {
	ID() graphql.ID
	Contents() string
	Repository() *RepositoryResolver
	CreatedAt() gqlutil.DateTime
	UpdatedAt() gqlutil.DateTime
}

type CodeownersIngestedFileConnectionResolver interface {
	Nodes(ctx context.Context) ([]CodeownersIngestedFileResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*gqlutil.PageInfo, error)
}

type SignalConfigurationResolver interface {
	Name() string
	Description() string
	IsEnabled() bool
	ExcludedRepoPatterns() []string
}

type UpdateSignalConfigurationsArgs struct {
	Input UpdateSignalConfigurationsInput
}

type UpdateSignalConfigurationsInput struct {
	Configs []SignalConfigurationUpdate
}

type SignalConfigurationUpdate struct {
	Name                 string
	ExcludedRepoPatterns []string
	Enabled              bool
}
