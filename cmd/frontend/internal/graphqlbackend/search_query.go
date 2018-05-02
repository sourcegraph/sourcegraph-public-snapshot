package graphqlbackend

type searchQuery struct {
	query string
}

func (q searchQuery) Query() string { return q.query }

type searchQueryDescription struct {
	description string
	query       string
}

func (q searchQueryDescription) Query() string { return q.query }
func (q searchQueryDescription) Description() *string {
	if q.description == "" {
		return nil
	}
	return &q.description
}
