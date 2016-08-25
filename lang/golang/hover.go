package golang

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/doc"
	"io/ioutil"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func (h *Handler) handleHover(req *jsonrpc2.Request, params lsp.TextDocumentPositionParams) (*lsp.Hover, error) {
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
	uri, err := h.fileURI(def.Position.Path)
	if err != nil {
		return nil, err
	}
	if uri != params.TextDocument.URI {
		// different file to input. This happens when the definition
		// lives in a different file to what we are hovering over.
		contents, err = ioutil.ReadFile(def.Position.Path)
		if err != nil {
			return nil, err
		}
	}
	docstring, err := docAtPosition(lsp.Position{Line: def.Position.Line - 1, Character: def.Position.Column - 1}, contents)
	if err != nil {
		return nil, err
	}

	ms := []lsp.MarkedString{{Language: "go", Value: def.Type.Decl}}
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

	var unitName string
	if def.Type.IsField {
		// Looking backward for type declaration.
		// TODO(unknwon): this is not memory friendly hack,
		// use look backward by byte so only have read and not allocate
		// any new memory.
		lines := bytes.SplitN(contents, []byte("\n"), def.Position.Line)
		for i := len(lines) - 2; i >= 0; i-- {
			line := bytes.TrimSpace(lines[i])
			if !bytes.HasPrefix(line, []byte("type ")) {
				continue
			}
			line = line[5:]
			typName := string(bytes.SplitN(line, []byte(" "), 2)[0])
			unitName = typName + "/" + def.Type.Name
			break
		}
	} else {
		unitName = path.Join(def.Type.Receiver, def.Type.Name)
	}

	// Cut off path before '/src/' and trim suffix of file name.
	unit := uri[strings.Index(uri, "/src/")+5:]
	unit = strings.TrimSuffix(unit, "/"+path.Base(unit))

	ms = append(ms,
		lsp.MarkedString{
			Language: "text/unit",
			Value:    unit,
		},
		lsp.MarkedString{
			Language: "text/uri",
			Value:    uri,
		},
		lsp.MarkedString{
			Language: "text/name",
			Value:    unitName,
		},
	)

	return &lsp.Hover{
		Contents: ms,
		Range:    r,
	}, nil
}

type godefResult struct {
	Position struct {
		IsDir  bool   `json:"is_dir"`
		Path   string `json:"path"`
		Line   int    `json:"line"`
		Column int    `json:"column"`
	} `json:"position"`
	Type struct {
		Name     string `json:"name"`
		IsField  bool   `json:"is_field"`
		Receiver string `json:"receiver"`
		Decl     string `json:"decl"`
	} `json:"type"`
}

// TODO(unknwon): parse JSON output from godef to have better handling.
func godef(env []string, path string, offset int) (*godefResult, error) {
	b, err := cmdOutput(env, exec.Command("godef", "-json", "-t", "-f", path, "-o", strconv.Itoa(offset)))
	if err != nil {
		return nil, fmt.Errorf("%v: %v", err, string(b))
	}

	// Non-JSON response is an error.
	// Errors returned from godef are not in JSON foramt,
	// simply rely on JSON unmarshal error will lose the full
	// error string.
	if len(b) == 0 || b[0] != '{' {
		return nil, fmt.Errorf("error response: %s", b)
	}

	var result *godefResult
	if err = json.Unmarshal(b, &result); err != nil {
		return nil, fmt.Errorf("unmarshal JSON: %s", b)
	}
	result.Type.Decl = strings.SplitN(result.Type.Decl, "\n", 2)[0]
	return result, nil
}
