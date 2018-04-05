package graphqlbackend

type diffResolver struct {
	repo          *repositoryResolver
	revisionRange *gitRevisionRange
}

func (d *diffResolver) Repository() *repositoryResolver { return d.repo }

func (d *diffResolver) Range() *gitRevisionRange { return d.revisionRange }
