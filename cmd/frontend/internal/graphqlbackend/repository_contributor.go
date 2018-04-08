package graphqlbackend

type repositoryContributorResolver struct {
	name  string
	email string
	count int32

	repo *repositoryResolver
}

func (r *repositoryContributorResolver) Person() *personResolver {
	return &personResolver{name: r.name, email: r.email}
}

func (r *repositoryContributorResolver) Count() int32 { return r.count }

func (r *repositoryContributorResolver) Repository() *repositoryResolver { return r.repo }
