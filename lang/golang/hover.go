package golang

import (
	"bytes"
	"errors"
	"fmt"
	"go/doc"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

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

	// godef for symbol info
	ofs, valid := offsetForPosition(contents, params.Position)
	if !valid {
		return nil, errors.New("invalid position")
	}
	def, err := godef(h.goEnv(), h.filePath(params.TextDocument.URI), int(ofs))
	if err != nil {
		return nil, err
	}

	// using def position, find its docs.
	uri, err := h.fileURI(def.Path)
	if err != nil {
		return nil, err
	}
	if uri != params.TextDocument.URI {
		// different file to input. This happens when the definition
		// lives in a different file to what we are hovering over.
		contents, err = ioutil.ReadFile(def.Path)
		if err != nil {
			return nil, err
		}
	}
	docstring, err := docAtPosition(lsp.Position{Line: def.Line - 1, Character: def.Column - 1}, contents)
	if err != nil {
		return nil, err
	}

	ms := []lsp.MarkedString{{Language: "go", Value: def.Info}}
	if docstring != "" {
		var htmlBuf bytes.Buffer
		doc.ToHTML(&htmlBuf, docstring, nil)
		ms = append(
			ms,
			lsp.MarkedString{
				Language: "text/html",
				Value:    htmlBuf.String(),
			},
			lsp.MarkedString{
				Language: "text/plain",
				Value:    docstring,
			},
		)
	}
	return &lsp.Hover{
		Contents: ms,
		Range:    r,
	}, nil
}

type godefResult struct {
	Path         string
	Line, Column int
	Info         string
}

func godef(env []string, path string, offset int) (*godefResult, error) {
	b, err := cmdOutput(env, exec.Command("godef", "-a", "-f", path, "-o", strconv.Itoa(offset)))
	if err != nil {
		return nil, fmt.Errorf("%v: %v", err, string(b))
	}

	lines := bytes.Split(b, []byte{'\n'})
	if len(lines) < 2 {
		return nil, fmt.Errorf("not enough lines in output: %v", string(b))
	}

	defpath, line, col, err := parseGodefPos(string(lines[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid position line: %s", string(lines[0]))
	}

	return &godefResult{
		Path:   defpath,
		Line:   line,
		Column: col,
		Info:   strings.TrimSpace(string(lines[1])),
	}, nil
}

func parseGodefPos(pos string) (path string, line int, col int, err error) {
	if !strings.Contains(pos, ":") {
		err = fmt.Errorf("expected pos %q to contain a colon", pos)
		return
	}
	j := strings.LastIndexByte(pos, ':')
	i := strings.LastIndexByte(pos[:j], ':')
	path = pos[:i]
	line, err = strconv.Atoi(pos[i+1 : j])
	if err != nil {
		return
	}
	col, err = strconv.Atoi(pos[j+1:])
	return
}
