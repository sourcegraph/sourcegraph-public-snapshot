package graphql

const NamespaceFieldsFragment = `
fragment namespaceFields on Namespace {
    id
    namespaceName
    url
}
`

type Namespace struct {
	ID            string
	NamespaceName string
	URL           string
}
