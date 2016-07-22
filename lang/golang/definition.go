package golang

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/tools/cmd/guru/serial"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func (h *Session) handleDefinition(req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) ([]lsp.Location, error) {
	// Find start of definition using guru
	contents, err := h.readFile(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	ofs, valid := offsetForPosition(contents, params.Position)
	if !valid {
		return nil, errors.New("invalid position")
	}
	def, err := guruDefinition(h.filePath(params.TextDocument.URI), int(ofs))
	if err != nil {
		return nil, err
	}
	filename, line, column := guruPos(def.ObjPos)

	uri, err := h.fileURI(filename)
	if err != nil {
		return nil, err
	}
	if uri != params.TextDocument.URI {
		// different file to input
		contents, err = ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
	}
	r, err := rangeAtPosition(lsp.Position{Line: line - 1, Character: column - 1}, contents)
	if err != nil {
		return nil, err
	}

	var locs []lsp.Location
	locs = append(locs, lsp.Location{
		URI:   uri,
		Range: r,
	})
	return locs, nil
}

func guruPos(pos string) (string, int, int) {
	j := strings.LastIndexByte(pos, ':')
	i := strings.LastIndexByte(pos[:j], ':')
	line, _ := strconv.Atoi(pos[i+1 : j])
	col, _ := strconv.Atoi(pos[j+1:])
	return pos[:i], line, col
}

func guruDefinition(path string, offset int) (serial.Definition, error) {
	var d serial.Definition
	c := exec.Command("guru", "-json", "definition", fmt.Sprintf("%s:#%d", path, offset))
	b, err := c.Output()
	if err != nil {
		return d, err
	}
	err = json.Unmarshal(b, &d)
	return d, err
}
