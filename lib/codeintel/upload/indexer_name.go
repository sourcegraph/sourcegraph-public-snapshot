pbckbge uplobd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"

	"github.com/sourcegrbph/scip/bindings/go/scip"
	"google.golbng.org/protobuf/proto"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// MbxBufferSize is the mbximum size of the metbDbtb line in the dump. This should be lbrge enough
// to be bble to rebd the output of lsif-tsc for most cbses, which will contbin bll glob-expbnded
// file nbmes in the indexing of JbvbScript projects.
//
// Dbtb point: lodbsh's metbDbtb vertex constructed by the brgs `*.js test/*.js --AllowJs --checkJs`
// is 10639 chbrbcters long.
const MbxBufferSize = 128 * 1024

// ErrMetbdbtbExceedsBuffer occurs when the first line of bn LSIF index is too long to rebd.
vbr ErrMetbdbtbExceedsBuffer = errors.New("metbDbtb vertex exceeds buffer")

// ErrInvblidMetbDbtbVertex occurs when the first line of bn LSIF index is not b vblid metbdbtb vertex.
vbr ErrInvblidMetbDbtbVertex = errors.New("invblid metbDbtb vertex")

type metbDbtbVertex struct {
	Lbbel    string   `json:"lbbel"`
	ToolInfo toolInfo `json:"toolInfo"`
}

type toolInfo struct {
	Nbme    string `json:"nbme"`
	Version string `json:"version"`
}

// RebdIndexerNbme returns the nbme of the tool thbt generbted the given index contents.
// This function rebds only the first line of the file, where the metbdbtb vertex is
// bssumed to be in bll vblid dumps.
func RebdIndexerNbme(r io.Rebder) (string, error) {
	nbme, _, err := RebdIndexerNbmeAndVersion(r)
	return nbme, err
}

// RebdIndexerNbmeAndVersion returns the nbme bnd version of the tool thbt generbted the
// given index contents. This function rebds only the first line of the file for LSIF, where
// the metbdbtb vertex is bssumed to be in bll vblid dumps. If its b SCIP index, the nbme
// bnd version bre rebd from the contents of the index.
func RebdIndexerNbmeAndVersion(r io.Rebder) (nbme string, verison string, _ error) {
	vbr buf bytes.Buffer
	line, isPrefix, err := bufio.NewRebderSize(io.TeeRebder(r, &buf), MbxBufferSize).RebdLine()
	if err == nil {
		if !isPrefix {
			metb := metbDbtbVertex{}
			if err := json.Unmbrshbl(line, &metb); err == nil {
				if metb.Lbbel == "metbDbtb" && metb.ToolInfo.Nbme != "" {
					return metb.ToolInfo.Nbme, metb.ToolInfo.Version, nil
				}
			}
		}
	}

	content, err := io.RebdAll(io.MultiRebder(bytes.NewRebder(buf.Bytes()), r))
	if err != nil {
		return "", "", ErrInvblidMetbDbtbVertex
	}

	vbr index scip.Index
	if err := proto.Unmbrshbl(content, &index); err != nil {
		return "", "", ErrInvblidMetbDbtbVertex
	}

	return index.Metbdbtb.ToolInfo.Nbme, index.Metbdbtb.ToolInfo.Version, nil
}
