package graphqlbackend

type repositoryContributorResolver struct {
	name  string
	email string
	count int32

	repo *repositoryResolver
	args repositoryContributorsArgs
}

func (r *repositoryContributorResolver) Person() *personResolver {
	return &personResolver{name: r.name, email: r.email}
}

func (r *repositoryContributorResolver) Count() int32 { return r.count }

func (r *repositoryContributorResolver) Repository() *repositoryResolver { return r.repo }

func (r *repositoryContributorResolver) Commits(args *struct {
	First *int32
}) *gitCommitConnectionResolver {
	var range_ string
	if r.args.Range != nil {
		range_ = *r.args.Range
	}
	return &gitCommitConnectionResolver{
		range_: range_,
		path:   r.args.Path,
		author: &r.email, // TODO(sqs): support when contributor resolves to user, and user has multiple emails
		first:  args.First,
		repo:   r.repo,
	}
}
