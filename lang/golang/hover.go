package golang

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

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
	def, err := godef(h.filePath("gopath"), h.filePath(params.TextDocument.URI), int(ofs))
	if err != nil {
		return nil, err
	}

	return &lsp.Hover{
		Contents: []lsp.MarkedString{{Language: "go", Value: def.Info}},
		Range:    r,
	}, nil
}

type godefResult struct {
	Path         string
	Line, Column int
	Info         string
}

func godef(gopath, path string, offset int) (*godefResult, error) {
	start := time.Now()
	c := exec.Command("godef", "-a", "-f", path, "-o", strconv.Itoa(offset))
	c.Env = []string{"GOPATH=" + gopath}
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "GOPATH=") {
			c.Env = append(c.Env, e)
		}
	}
	b, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%v: %v", err, string(b))
	}
	fmt.Printf("TIME: %v %s\n", time.Since(start), strings.Join(c.Args, " "))

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
