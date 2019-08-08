package graphqlbackend

// Diagnostic implements the Diagnostic GraphQL type.
type Diagnostic interface {
	Type() string
	Data() JSONValue
}
