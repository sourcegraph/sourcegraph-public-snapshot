package graphqlbackend

// A resolver for the GraphQL type GenericSearchMatch
type genericSearchResultResolver struct {
	icon    string
	label   string
	url     string
	detail  string
	matches []*genericSearchMatchResolver
}

func (r *genericSearchResultResolver) Icon() string {
	return r.icon
}

func (r *genericSearchResultResolver) Label() *markdownResolver {
	return &markdownResolver{text: r.label}
}

func (r *genericSearchResultResolver) URL() string {
	return r.url
}
func (r *genericSearchResultResolver) Detail() *markdownResolver {
	return &markdownResolver{text: r.detail}
}
func (r *genericSearchResultResolver) Matches() []*genericSearchMatchResolver {
	return r.matches
}
