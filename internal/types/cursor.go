package types

// MultiCursor is a slice of Cursors which is needed when a single Cursor isn't specific
// enough to paginate through unique records. Example: (repos.stars, repo.id)
type MultiCursor []*Cursor

// A Cursor for efficient index based pagination through large result sets.
type Cursor struct {
	// Columns contains the relevant columns for cursor-based pagination (e.g. "name")
	Column string `json:"c"`
	// Value contains the relevant value for cursor-based pagination (e.g. "Zaphod").
	Value string `json:"v"`
	// Direction contains the comparison for cursor-based pagination, all possible values are: next, prev.
	Direction string `json:"d"`
}
