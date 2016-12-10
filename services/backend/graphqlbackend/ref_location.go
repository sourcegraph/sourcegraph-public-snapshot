package graphqlbackend

type refLocationResolver struct {
	startLineNumber int32
	startColumn     int32
	endLineNumber   int32
	endColumn       int32
}

func (r *refLocationResolver) StartLineNumber() int32 {
	return r.startLineNumber
}

func (r *refLocationResolver) StartColumn() int32 {
	return r.startColumn
}

func (r *refLocationResolver) EndLineNumber() int32 {
	return r.endLineNumber
}

func (r *refLocationResolver) EndColumn() int32 {
	return r.endColumn
}
