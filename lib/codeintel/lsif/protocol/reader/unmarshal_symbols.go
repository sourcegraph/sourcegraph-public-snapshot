package reader

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"
)

func init() {
	vertexUnmarshalersWithInterner["documentSymbolResult"] = unmarshalDocumentSymbolResult
}

func unmarshalDocumentSymbolResult(interner *Interner, line []byte) (interface{}, error) {
	// Like RangeBasedDocumentSymbol, but accepts numeric *and* string IDs.
	type _rangeBasedDocumentSymbol struct {
		ID       json.RawMessage             `json:"id"`
		Children []_rangeBasedDocumentSymbol `json:"children"`
	}
	var payload struct {
		Result []_rangeBasedDocumentSymbol `json:"result"`
	}
	if err := unmarshaller.Unmarshal(line, &payload); err != nil {
		return nil, err
	}

	var toRangeBasedDocumentSymbol func(item _rangeBasedDocumentSymbol) (protocol.RangeBasedDocumentSymbol, error)
	toRangeBasedDocumentSymbol = func(item _rangeBasedDocumentSymbol) (protocol.RangeBasedDocumentSymbol, error) {
		var children []protocol.RangeBasedDocumentSymbol
		if len(item.Children) > 0 {
			children = make([]protocol.RangeBasedDocumentSymbol, len(item.Children))
		}
		for i, child := range item.Children {
			var err error
			children[i], err = toRangeBasedDocumentSymbol(child)
			if err != nil {
				return protocol.RangeBasedDocumentSymbol{}, err
			}
		}

		id, err := internRaw(interner, item.ID)
		if err != nil {
			return protocol.RangeBasedDocumentSymbol{}, err
		}

		return protocol.RangeBasedDocumentSymbol{
			ID:       uint64(id), // TODO(sqs): sketchy conversion
			Children: children,
		}, nil
	}

	results := make([]protocol.RangeBasedDocumentSymbol, len(payload.Result))
	for i, result := range payload.Result {
		var err error
		results[i], err = toRangeBasedDocumentSymbol(result)
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}
