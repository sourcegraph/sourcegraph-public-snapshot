package graphqlbackend

import "github.com/sourcegraph/sourcegraph/internal/search"

// searchQueryDescriptionResolver is a type for the SearchQueryDescription resolver used
// by SearchAlert. This name is a bit of a misnomer but cannot be changed: It
// must be this way to work with the GQL definition and compatibility. We use
// our internal, resolver-agnostic alert type to do real work.
type searchQueryDescriptionResolver struct {
	query *search.ProposedQuery
}

func (q searchQueryDescriptionResolver) Query() string {
	// Do not add logic here that manipulates the query string. Do it in the QueryString() method.
	return q.query.QueryString()
}

func (q searchQueryDescriptionResolver) Description() *string {
	if q.query.Description == "" {
		return nil
	}

	return &q.query.Description
}

func (q searchQueryDescriptionResolver) Annotations() *[]searchQueryAnnotationResolver {
	if len(q.query.Annotations) == 0 {
		return nil
	}

	a := make([]searchQueryAnnotationResolver, 0, len(q.query.Annotations))
	for name, value := range q.query.Annotations {
		a = append(a, searchQueryAnnotationResolver{name: string(name), value: value})
	}
	return &a
}
