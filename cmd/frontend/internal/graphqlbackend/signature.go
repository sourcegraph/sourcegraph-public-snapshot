package graphqlbackend

type signatureResolver struct {
	person *personResolver
	date   string
}

func (r *signatureResolver) Person() *personResolver {
	return r.person
}

func (r *signatureResolver) Date() string {
	return r.date
}
