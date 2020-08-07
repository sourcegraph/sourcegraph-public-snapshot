package api

import "encoding/json"

// graphqlError wraps a raw JSON error returned from a GraphQL endpoint.
type graphqlError struct{ v interface{} }

func (g *graphqlError) Error() string {
	j, _ := json.MarshalIndent(g.v, "", "  ")
	return string(j)
}
