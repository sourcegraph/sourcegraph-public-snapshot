package graphqlbackend

type contributorResolver struct {
	name  string
	email string
	count int32
}

func (r *contributorResolver) Person() *personResolver {
	return &personResolver{name: r.name, email: r.email}
}

func (r *contributorResolver) Count() int32 { return r.count }
