package graphqlbackend

type symbolResolver struct {
	path      string
	line      int32
	character int32
}

func (r *symbolResolver) Path() string {
	return r.path
}

func (r *symbolResolver) Line() int32 {
	return r.line
}

func (r *symbolResolver) Character() int32 {
	return r.character
}
