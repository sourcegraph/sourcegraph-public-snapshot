package graphqlbackend

// connectionArgs is the common set of args to fields that return connections (lists).
type connectionArgs struct {
	First *int32 // return the first n items
}

// connectionResolverCommon is the common information that connection resolvers need,
// obtained from connectionArgs.
type connectionResolverCommon struct {
	first int32 // return the first n items
}

func newConnectionResolverCommon(args connectionArgs) connectionResolverCommon {
	const defaultFirst = 10

	var c connectionResolverCommon

	if args.First == nil {
		c.first = defaultFirst
	} else {
		c.first = *args.First
	}

	return c
}
