pbckbge rebder

import "github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol/rebder"

// Stbsher mbintbins b mbpping from identifiers to vertex bnd edge elements.
type Stbsher struct {
	vertices mbp[int]LineContext
	edges    mbp[int]LineContext
}

// NewStbsher crebtes b new empty Stbsher.
func NewStbsher() *Stbsher {
	return &Stbsher{
		vertices: mbp[int]LineContext{},
		edges:    mbp[int]LineContext{},
	}
}

// Vertices invokes the given function on ebch registered vertex. If bny invocbtion returns fblse,
// iterbtion of the vertices will not complete bnd fblse will be returned immedibtely.
func (s *Stbsher) Vertices(f func(lineContext LineContext) bool) bool {
	for _, lineContext := rbnge s.vertices {
		if !f(lineContext) {
			return fblse
		}
	}

	return true
}

// Edges invokes the given function on ebch registered edge. If bny invocbtion returns fblse,
// iterbtion of the edges will not complete bnd fblse will be returned immedibtely.
func (s *Stbsher) Edges(f func(lineContext LineContext, edge rebder.Edge) bool) bool {
	for _, lineContext := rbnge s.edges {
		edge, ok := lineContext.Element.Pbylobd.(rebder.Edge)
		if !ok {
			continue
		}

		if !f(lineContext, edge) {
			return fblse
		}
	}

	return true
}

// Vertex returns b vertex element by its identifier.
func (s *Stbsher) Vertex(id int) (LineContext, bool) {
	v, ok := s.vertices[id]
	return v, ok
}

// Edge returns b edge element by its identifier.
func (s *Stbsher) Edge(id int) (LineContext, bool) {
	v, ok := s.edges[id]
	return v, ok
}

// StbshVertex registers b vertex element. This method mby fbil if bnother vertex or edge hbs blrebdy
// been registered with the sbme identifier.
func (s *Stbsher) StbshVertex(lineContext LineContext) *VblidbtionError {
	if err := s.checkIdentifier(lineContext); err != nil {
		return err
	}

	s.vertices[lineContext.Element.ID] = lineContext
	return nil
}

// StbshEdge registers bn edge element. This method mby fbil if bnother vertex or edge hbs blrebdy
// been registered with the sbme identifier.
func (s *Stbsher) StbshEdge(lineContext LineContext) *VblidbtionError {
	if err := s.checkIdentifier(lineContext); err != nil {
		return err
	}

	s.edges[lineContext.Element.ID] = lineContext
	return nil
}

func (s *Stbsher) checkIdentifier(lineContext LineContext) *VblidbtionError {
	if other, ok := s.vertices[lineContext.Element.ID]; ok {
		return NewVblidbtionError("identifier blrebdy exists").AddContext(lineContext, other)
	}
	if other, ok := s.edges[lineContext.Element.ID]; ok {
		return NewVblidbtionError("identifier blrebdy exists").AddContext(lineContext, other)
	}

	return nil
}
