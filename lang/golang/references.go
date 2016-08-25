package golang

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/cmd/guru/serial"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/jsonrpc2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func (h *Handler) handleReferences(req *jsonrpc2.Request, params lsp.ReferenceParams) ([]lsp.Location, error) {
	contents, err := h.readFile(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	ofs, valid := offsetForPosition(contents, params.Position)
	if !valid {
		return nil, errors.New("invalid position")
	}
	def, pkgs, err := guruReferrers(h.goEnv(), h.filePath(params.TextDocument.URI), int(ofs))
	if err != nil {
		return nil, err
	}

	// TODO(keegancsmith) PERF We do a lot of unnecessary duplicate work
	guruPosToLoc := func(pos string) (*lsp.Location, error) {
		filename, line, column := guruPos(pos)
		uri, err := h.fileURI(filename)
		if err != nil {
			return nil, err
		}
		contents, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		r, err := rangeAtPosition(lsp.Position{Line: line - 1, Character: column - 1}, contents)
		if err != nil {
			return nil, err
		}
		return &lsp.Location{
			URI:   uri,
			Range: r,
		}, nil
	}
	var locs []lsp.Location
	if params.Context.IncludeDeclaration {
		l, err := guruPosToLoc(def.ObjPos)
		if err != nil {
			return nil, err
		}
		locs = append(locs, *l)
	}
	for _, pkg := range pkgs {
		for _, r := range pkg.Refs {
			l, err := guruPosToLoc(r.Pos)
			if err != nil {
				return nil, err
			}
			locs = append(locs, *l)
		}
	}

	sort.Sort(locationList(locs))
	return locs, nil
}

type locationList []lsp.Location

func (l locationList) Less(a, b int) bool {
	if l[a].URI != l[b].URI {
		return l[a].URI < l[b].URI
	}
	if l[a].Range.Start.Line != l[b].Range.Start.Line {
		return l[a].Range.Start.Line < l[b].Range.Start.Line
	}
	return l[a].Range.Start.Character < l[b].Range.Start.Character
}

func (l locationList) Swap(a, b int) {
	l[a], l[b] = l[b], l[a]
}
func (l locationList) Len() int {
	return len(l)
}

func guruReferrers(env []string, path string, offset int) (*serial.ReferrersInitial, []*serial.ReferrersPackage, error) {
	b, err := cmdOutput(env, exec.Command("guru", "-json", "referrers", fmt.Sprintf("%s:#%d", path, offset)))
	if err != nil {
		return nil, nil, err
	}
	// TODO(keegancsmith) directly decode stdout rather than an intermediate buffer
	r := bytes.NewReader(b)
	d := json.NewDecoder(r)

	var def serial.ReferrersInitial
	err = d.Decode(&def)
	if err != nil {
		return nil, nil, err
	}
	var pkgs []*serial.ReferrersPackage
	for {
		var pkg serial.ReferrersPackage
		err = d.Decode(&pkg)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, err
		}
		pkgs = append(pkgs, &pkg)
	}
	return &def, pkgs, nil
}

func guruPos(pos string) (string, int, int) {
	j := strings.LastIndexByte(pos, ':')
	i := strings.LastIndexByte(pos[:j], ':')
	line, _ := strconv.Atoi(pos[i+1 : j])
	col, _ := strconv.Atoi(pos[j+1:])
	return pos[:i], line, col
}
