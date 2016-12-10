package graphqlbackend

type globalReferencesResolver struct {
	refLocation *refLocationResolver
	uri         *uriResolver
}

func (r *globalReferencesResolver) URI() *uriResolver {
	return r.uri
}

func (r *globalReferencesResolver) RefLocation() *refLocationResolver {
	return r.refLocation
}
