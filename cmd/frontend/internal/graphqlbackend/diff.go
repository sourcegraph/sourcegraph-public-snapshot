package graphqlbackend

type diff struct {
	repo          *repositoryResolver
	revisionRange *gitRevisionRange
}

func (d *diff) Repository() *repositoryResolver { return d.repo }

func (d *diff) Range() *gitRevisionRange { return d.revisionRange }
