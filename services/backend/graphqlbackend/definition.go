package graphqlbackend

type definitionResolver struct {
	globalReferences []*globalReferencesResolver
}

func (r *definitionResolver) GlobalReferences() []*globalReferencesResolver {
	return r.globalReferences
}
