package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/internal/search"
)

type searchQueryDescriptionResolver struct {
	query *search.QueryDescription
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
