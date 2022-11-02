package graphqlbackend

type lfsResolver struct {
	// TODO what if file is bigger than 4gb? This seems likely for LFS. Do we
	// need to return a float?
	size int32
}

func (l *lfsResolver) ByteSize() int32 {
	return l.size
}
