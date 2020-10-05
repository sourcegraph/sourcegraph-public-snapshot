package graphs

import "time"

// Graph is a subset of repositories.
type Graph struct {
	ID int64

	OwnerUserID int32
	OwnerOrgID  int32

	Name        string
	Description *string
	Spec        string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Clone returns a clone of the graph.
func (g *Graph) Clone() *Graph {
	gg := *g
	return &gg
}
