package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type repositoryContributorResolver struct {
	db    database.DB
	name  string
	email string
	count int32

	repo *RepositoryResolver
	args repositoryContributorsArgs
}

func (r *repositoryContributorResolver) Person() *PersonResolver {
	return &PersonResolver{db: r.db, name: r.name, email: r.email}
}

func (r *repositoryContributorResolver) Count() int32 { return r.count }

func (r *repositoryContributorResolver) Repository() *RepositoryResolver { return r.repo }

func (r *repositoryContributorResolver) Commits(args *struct {
	First *int32
}) GitCommitConnectionResolver {
	var revisionRange string
	if r.args.RevisionRange != nil {
		revisionRange = *r.args.RevisionRange
	}
	return NewGitCommitConnectionResolver(r.db, r.repo, GitCommitConnectionArgs{
		RevisionRange: revisionRange,
		Path:          r.args.Path,
		Author:        &r.email, // TODO(sqs): support when contributor resolves to user, and user has multiple emails
		After:         r.args.After,
		First:         args.First,
	})
}
