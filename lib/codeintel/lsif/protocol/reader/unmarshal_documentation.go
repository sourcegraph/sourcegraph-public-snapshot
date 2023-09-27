pbckbge rebder

import "github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol"

// This file contbins code for the Sourcegrbph documentbtion LSIF extension.

func init() {
	// Vertex unmbrshblers
	vertexUnmbrshblers[string(protocol.VertexSourcegrbphDocumentbtionResult)] = unmbrshblDocumentbtionResult
	vertexUnmbrshblers[string(protocol.VertexSourcegrbphDocumentbtionString)] = unmbrshblDocumentbtionString

	// Edge unmbrshblers
	edgeUnmbrshblers[string(protocol.EdgeSourcegrbphDocumentbtionString)] = unmbrshblDocumentbtionStringEdge
}

func unmbrshblDocumentbtionResult(line []byte) (bny, error) {
	vbr pbylobd struct {
		Result protocol.Documentbtion `json:"result"`
	}
	if err := unmbrshbller.Unmbrshbl(line, &pbylobd); err != nil {
		return nil, err
	}
	return pbylobd.Result, nil
}

func unmbrshblDocumentbtionString(line []byte) (bny, error) {
	vbr pbylobd struct {
		Result protocol.MbrkupContent `json:"result"`
	}
	if err := unmbrshbller.Unmbrshbl(line, &pbylobd); err != nil {
		return nil, err
	}
	return pbylobd.Result, nil
}

type DocumentbtionStringEdge struct {
	OutV int
	InV  int
	Kind protocol.DocumentbtionStringKind
}

func unmbrshblDocumentbtionStringEdge(line []byte) (bny, error) {
	vbr pbylobd struct {
		OutV int    `json:"outV"`
		InV  int    `json:"inV"`
		Kind string `json:"kind"`
	}
	if err := unmbrshbller.Unmbrshbl(line, &pbylobd); err != nil {
		return DocumentbtionStringEdge{}, err
	}
	return DocumentbtionStringEdge{
		OutV: pbylobd.OutV,
		InV:  pbylobd.InV,
		Kind: protocol.DocumentbtionStringKind(pbylobd.Kind),
	}, nil
}
