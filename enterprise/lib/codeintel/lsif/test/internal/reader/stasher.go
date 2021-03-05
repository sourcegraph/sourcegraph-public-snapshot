package reader

import reader "github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/protocol/reader"

// Stasher maintains a mapping from identifiers to vertex and edge elements.
type Stasher struct {
	vertices map[int]LineContext
	edges    map[int]LineContext
}

// NewStasher creates a new empty Stasher.
func NewStasher() *Stasher {
	return &Stasher{
		vertices: map[int]LineContext{},
		edges:    map[int]LineContext{},
	}
}

// Vertices invokes the given function on each registered vertex. If any invocation returns false,
// iteration of the vertices will not complete and false will be returned immediately.
func (s *Stasher) Vertices(f func(lineContext LineContext) bool) bool {
	for _, lineContext := range s.vertices {
		if !f(lineContext) {
			return false
		}
	}

	return true
}

// Edges invokes the given function on each registered edge. If any invocation returns false,
// iteration of the edges will not complete and false will be returned immediately.
func (s *Stasher) Edges(f func(lineContext LineContext, edge reader.Edge) bool) bool {
	for _, lineContext := range s.edges {
		edge, ok := lineContext.Element.Payload.(reader.Edge)
		if !ok {
			continue
		}

		if !f(lineContext, edge) {
			return false
		}
	}

	return true
}

// Vertex returns a vertex element by its identifier.
func (s *Stasher) Vertex(id int) (LineContext, bool) {
	v, ok := s.vertices[id]
	return v, ok
}

// Edge returns a edge element by its identifier.
func (s *Stasher) Edge(id int) (LineContext, bool) {
	v, ok := s.edges[id]
	return v, ok
}

// StashVertex registers a vertex element. This method may fail if another vertex or edge has already
// been registered with the same identifier.
func (s *Stasher) StashVertex(lineContext LineContext) *ValidationError {
	if err := s.checkIdentifier(lineContext); err != nil {
		return err
	}

	s.vertices[lineContext.Element.ID] = lineContext
	return nil
}

// StashEdge registers an edge element. This method may fail if another vertex or edge has already
// been registered with the same identifier.
func (s *Stasher) StashEdge(lineContext LineContext) *ValidationError {
	if err := s.checkIdentifier(lineContext); err != nil {
		return err
	}

	s.edges[lineContext.Element.ID] = lineContext
	return nil
}

func (s *Stasher) checkIdentifier(lineContext LineContext) *ValidationError {
	if other, ok := s.vertices[lineContext.Element.ID]; ok {
		return NewValidationError("identifier already exists").AddContext(lineContext, other)
	}
	if other, ok := s.edges[lineContext.Element.ID]; ok {
		return NewValidationError("identifier already exists").AddContext(lineContext, other)
	}

	return nil
}
