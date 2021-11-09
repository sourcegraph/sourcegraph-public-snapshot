package database

// Cursors is a slice of Cursors which is needed when a single Cursor isn't specific
// enough to paginate through unique records. Example: (repos.stars, repo.name)
type Cursors []*Cursor

// A Cursor for efficient index based pagination through large result sets.
type Cursor struct {
	// Columns contains the relevant columns for cursor-based pagination (e.g. "name")
	Column string
	// Value contains the relevant value for cursor-based pagination (e.g. "Zaphod").
	Value string
	// Direction contains the comparison for cursor-based pagination, all possible values are: next, prev.
	Direction string
}
