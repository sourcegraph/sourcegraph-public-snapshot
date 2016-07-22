package golang

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"golang.org/x/tools/cmd/guru/serial"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func (h *Session) handleHover(req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
	// Find the range of the symbol
	contents, err := h.readFile(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	r, err := rangeAtPosition(params.Position, contents)
	if err != nil {
		return nil, err
	}

	// guru describe for symbol info
	ofs, valid := offsetForPosition(contents, params.Position)
	if !valid {
		return nil, errors.New("invalid position")
	}
	desc, err := guruDescribe(h.filePath(params.TextDocument.URI), int(ofs))
	if err != nil {
		return nil, err
	}
	var s string
	switch desc.Detail {
	case "package":
		s = "package " + desc.Package.Path
	case "type":
		s = "type " + desc.Type.Type
	case "value":
		s = desc.Value.Type
	case "":
		s = desc.Desc
	default:
		return nil, fmt.Errorf("unexpected guru describe detail %s", desc.Detail)
	}

	return &lsp.Hover{
		Contents: []lsp.MarkedString{{Language: "go", Value: s}},
		Range:    r,
	}, nil
}

func guruDescribe(path string, offset int) (serial.Describe, error) {
	var d serial.Describe
	c := exec.Command("guru", "-json", "describe", fmt.Sprintf("%s:#%d", path, offset))
	b, err := c.Output()
	if err != nil {
		return d, err
	}
	err = json.Unmarshal(b, &d)
	return d, err
}
