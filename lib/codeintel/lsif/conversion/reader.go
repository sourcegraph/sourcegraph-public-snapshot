package conversion

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
)

type Pair struct {
	Element Element
	Err     error
}

// Read reads the given content as line-separated JSON objects and returns a channel of Pair values
// for each non-empty line.
func Read(ctx context.Context, input reader.Dump) <-chan Pair {
	elements := make(chan Pair)

	go func() {
		defer close(elements)

		for e := range reader.Read(ctx, input) {
			element := Element{
				ID:      e.Element.ID,
				Type:    e.Element.Type,
				Label:   e.Element.Label,
				Payload: translatePayload(e.Element.Payload),
			}

			elements <- Pair{Element: element, Err: e.Err}
		}
	}()

	return elements
}

func translatePayload(payload interface{}) interface{} {
	switch v := payload.(type) {
	case reader.Edge:
		return Edge(v)

	case reader.MetaData:
		return MetaData(v)

	case reader.PackageInformation:
		return PackageInformation(v)

	case reader.Diagnostic:
		return Diagnostic(v)

	case reader.Range:
		return Range{Range: v}

	case reader.ResultSet:
		return ResultSet{ResultSet: v}

	case reader.Moniker:
		return Moniker{Moniker: v}

	case []reader.Diagnostic:
		diagnostics := make([]Diagnostic, 0, len(v))
		for _, v := range v {
			diagnostics = append(diagnostics, Diagnostic(v))
		}

		return diagnostics
	}

	return payload
}
