package graphqlbackend

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
