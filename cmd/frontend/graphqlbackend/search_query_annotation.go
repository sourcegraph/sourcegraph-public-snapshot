package graphqlbackend

type searchQueryAnnotationResolver struct {
	name  string
	value string
}

func (a searchQueryAnnotationResolver) Name() string {
	return a.name
}

func (a searchQueryAnnotationResolver) Value() string {
	return a.value
}
