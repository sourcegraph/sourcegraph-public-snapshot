package graphqlbackend

type dependencyReferencesResolver struct {
	data string
}

func (r *dependencyReferencesResolver) Data() string {
	return r.data
}
