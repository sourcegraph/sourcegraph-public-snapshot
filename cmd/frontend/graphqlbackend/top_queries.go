package graphqlbackend

import "context"

func (s *schemaResolver) TopQueries(ctx context.Context) ([]string, error) {
	return []string {"foo", "bar"}, nil
}
