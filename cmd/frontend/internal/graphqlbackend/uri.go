package graphqlbackend

type uriResolver struct {
	host     string
	fragment string
	path     string
	query    string
	scheme   string
}

func (r *uriResolver) Host() string {
	return r.host
}

func (r *uriResolver) Fragment() string {
	return r.fragment
}

func (r *uriResolver) Path() string {
	return r.path
}

func (r *uriResolver) Query() string {
	return r.query
}

func (r *uriResolver) Scheme() string {
	return r.scheme
}
