package graphqlbackend

type issueSearchResultResolver struct {
	title string
	url   string
	body  string
}

func (r *issueSearchResultResolver) Title() string {
	return r.title
}
func (r *issueSearchResultResolver) URL() string {
	return r.url
}
func (r *issueSearchResultResolver) Body() string {
	return r.body
}
