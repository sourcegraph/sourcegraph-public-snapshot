package graphqlbackend

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type repositoryContributorResolver struct {
	db    database.DB
	name  string
	email string
	count int32

	repo *RepositoryResolver
	args repositoryContributorsArgs

	// For use with RepositoryResolver only
	index int
}

// gitContributorGQLID is a type used for marshaling and unmarshaling a Git contributor's
// GraphQL ID.
type gitContributorGQLID struct {
	Repository graphql.ID `json:"r"`
	Email      string     `json:"e"`
	Name       string     `json:"n"`
}

func (r repositoryContributorResolver) ID() graphql.ID {
	return relay.MarshalID("RepositoryContributor", gitContributorGQLID{Repository: r.repo.ID(), Email: r.email, Name: r.name})
}

func (r *repositoryContributorResolver) Person() *PersonResolver {
	return &PersonResolver{db: r.db, name: r.name, email: r.email, includeUserInfo: true}
}

func (r *repositoryContributorResolver) Count() int32 { return r.count }

func (r *repositoryContributorResolver) Repository() *RepositoryResolver { return r.repo }

func (r *repositoryContributorResolver) Commits(args *struct {
	First *int32
}) *gitCommitConnectionResolver {
	var revisionRange string
	if r.args.RevisionRange != nil {
		revisionRange = *r.args.RevisionRange
	}
	return &gitCommitConnectionResolver{
		db:              r.db,
		gitserverClient: r.repo.gitserverClient,
		revisionRange:   revisionRange,
		path:            r.args.Path,
		author:          &r.email, // TODO(sqs): support when contributor resolves to user, and user has multiple emails
		after:           r.args.AfterDate,
		first:           args.First,
		repo:            r.repo,
	}
}
