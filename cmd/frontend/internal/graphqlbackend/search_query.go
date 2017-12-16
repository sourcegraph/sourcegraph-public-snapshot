package graphqlbackend

type searchQuery struct {
	query      string
	scopeQuery string
}

func (q searchQuery) Query() string      { return q.query }
func (q searchQuery) ScopeQuery() string { return q.scopeQuery }

type searchQueryDescription struct {
	description string
	query       searchQuery
}

func (q searchQueryDescription) Query() *searchQuery { return &q.query }
func (q searchQueryDescription) Description() *string {
	if q.description == "" {
		return nil
	}
	return &q.description
}
