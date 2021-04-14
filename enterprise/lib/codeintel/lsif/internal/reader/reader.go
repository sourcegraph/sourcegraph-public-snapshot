package reader

import (
	"context"
	"io"

	reader "github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/protocol/reader"
)

// ElementMapper is the type of function that is invoked for each parsed element.
type ElementMapper func(lineContext LineContext)

// Read consumes the given reader as newline-delimited JSON-encoded LSIF. Each parsed vertex and each
// parsed edge element is registered to the given Stasher. If vertex or edge mappers are supplied, they
// are invoked on each parsed element.
func Read(r io.Reader, stasher *Stasher, vertexMapper, edgeMapper ElementMapper) error {
	index := 0
	for pair := range reader.Read(context.Background(), r) {
		if pair.Err != nil {
			return pair.Err
		}

		index++
		lineContext := LineContext{
			Index:   index,
			Element: pair.Element,
		}

		if pair.Element.Type == "vertex" {
			if vertexMapper != nil {
				vertexMapper(lineContext)
			}

			stasher.StashVertex(lineContext)
		}

		if pair.Element.Type == "edge" {
			if edgeMapper != nil {
				edgeMapper(lineContext)
			}

			stasher.StashEdge(lineContext)
		}
	}

	return nil
}
