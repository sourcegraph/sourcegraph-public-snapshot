package golang

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
	"sourcegraph.com/sourcegraph/srclib-go/gog"
	"sourcegraph.com/sourcegraph/srclib-go/gog/definfo"
)

func (h *Session) handleSymbol(req *jsonrpc2.Request, params lsp.WorkspaceSymbolParams) ([]lsp.SymbolInformation, error) {
	q, err := parseSymbolQuery(params.Query)
	if err != nil {
		return nil, err
	}
	defFilter := func(_ *gog.Def) bool { return false }
	switch q.Type {
	case "exported":
		defFilter = func(d *gog.Def) bool { return d.DefInfo.Exported }
	default:
		return nil, fmt.Errorf("unrecognized symbol query type %s", q.Type)
	}
	pkgs, err := expandPackages(h.goEnv(), q.Packages)
	if err != nil {
		return nil, err
	}
	o, err := runGog(h.goEnv(), pkgs)
	if err != nil {
		return nil, err
	}

	var symbols []lsp.SymbolInformation
	for _, d := range o.Defs {
		if !defFilter(d) {
			continue
		}
		uri, err := h.fileURI(d.File)
		if err != nil {
			return nil, err
		}
		// TODO(keegancsmith) duplicated IO + ignoring error for
		// convenience of packages
		var content []byte
		content, _ = ioutil.ReadFile(d.File)
		s := lsp.SymbolInformation{
			Name: d.Name,
			Kind: gogKindToLSP(d.DefInfo.Kind),
			Location: lsp.Location{
				URI: uri,
				Range: lsp.Range{
					Start: offsetToPosition(content, int(d.IdentSpan[0])),
					End:   offsetToPosition(content, int(d.IdentSpan[1])),
				},
			},
			ContainerName: d.DefInfo.Receiver + d.DefInfo.FieldOfStruct,
		}
		symbols = append(symbols, s)
	}
	return symbols, nil
}

type symbolQuery struct {
	// Type is the type of symbol query we are performing.
	Type string

	// Packages is which go packages to inspect. A empty slice indicates
	// all packages.
	Packages []string
}

func parseSymbolQuery(q string) (*symbolQuery, error) {
	parts := strings.Fields(q)
	if len(parts) < 1 {
		return nil, errors.New("empty symbol query")
	}
	return &symbolQuery{
		Type:     parts[0],
		Packages: parts[1:],
	}, nil
}

func runGog(env, pkgs []string) (*gog.Output, error) {
	c := exec.Command("gog", pkgs...)
	c.Env = env
	b, err := c.Output()
	if err != nil {
		return nil, err
	}
	var o gog.Output
	err = json.Unmarshal(b, &o)
	return &o, err
}

func expandPackages(env, pkgs []string) ([]string, error) {
	args := append([]string{"list"}, pkgs...)
	c := exec.Command("go", args...)
	c.Env = env
	b, err := c.Output()
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(b)), nil
}

func gogKindToLSP(kind string) lsp.SymbolKind {
	switch kind {
	case definfo.Package:
		return lsp.SKPackage
	case definfo.Field:
		return lsp.SKField
	case definfo.Func:
		return lsp.SKFunction
	case definfo.Method:
		return lsp.SKMethod
	case definfo.Var:
		return lsp.SKVariable
	case definfo.Type:
		return lsp.SKClass
	case definfo.Interface:
		return lsp.SKInterface
	case definfo.Const:
		return lsp.SKConstant
	default:
		// This should not happen
		return -1
	}
}

func offsetToPosition(content []byte, offset int) lsp.Position {
	var p lsp.Position
	for i, b := range content {
		if i == offset {
			break
		}
		if b == '\n' {
			p.Line++
			p.Character = 0
		} else {
			p.Character++
		}
	}
	return p
}
