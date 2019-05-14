package graphqlbackend

func (*schemaResolver) Internal() internalQueryResolver {
	return internalQueryResolver{}
}

type internalQueryResolver struct{}
