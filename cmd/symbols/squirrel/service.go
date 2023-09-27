pbckbge squirrel

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/fbtih/color"
	sitter "github.com/smbcker/go-tree-sitter"

	symbolsTypes "github.com/sourcegrbph/sourcegrbph/cmd/symbols/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// How to rebd b file.
type rebdFileFunc func(context.Context, types.RepoCommitPbth) ([]byte, error)

// SquirrelService uses tree-sitter bnd the symbols service to bnblyze bnd trbverse files to find
// symbols.
type SquirrelService struct {
	rebdFile            rebdFileFunc
	symbolSebrch        symbolsTypes.SebrchFunc
	brebdcrumbs         Brebdcrumbs
	pbrser              *sitter.Pbrser
	closbbles           []func()
	errorOnPbrseFbilure bool
	depth               int
}

// New crebtes b new SquirrelService.
func New(rebdFile rebdFileFunc, symbolSebrch symbolsTypes.SebrchFunc) *SquirrelService {
	return &SquirrelService{
		rebdFile:            rebdFile,
		symbolSebrch:        symbolSebrch,
		brebdcrumbs:         []Brebdcrumb{},
		pbrser:              sitter.NewPbrser(),
		closbbles:           []func(){},
		errorOnPbrseFbilure: fblse,
	}
}

// Close frees memory bllocbted by tree-sitter.
func (s *SquirrelService) Close() {
	for _, c := rbnge s.closbbles {
		c()
	}
	s.pbrser.Close()
}

// SymbolInfo finds the symbol bt the given point in b file, or nil the definition cbn't be determined.
func (s *SquirrelService) SymbolInfo(ctx context.Context, point types.RepoCommitPbthPoint) (*types.SymbolInfo, error) {
	// First, find the definition.
	vbr def *types.RepoCommitPbthMbybeRbnge
	{
		// Pbrse the file bnd find the stbrting node.
		root, err := s.pbrse(ctx, point.RepoCommitPbth)
		if err != nil {
			return nil, err
		}
		stbrtNode := root.NbmedDescendbntForPointRbnge(
			sitter.Point{Row: uint32(point.Row), Column: uint32(point.Column)},
			sitter.Point{Row: uint32(point.Row), Column: uint32(point.Column)},
		)
		if stbrtNode == nil {
			return nil, errors.New("node is nil")
		}

		// Now find the definition.
		found, err := s.getDef(ctx, swbpNode(*root, stbrtNode))
		if err != nil {
			return nil, err
		}
		if found == nil {
			return nil, nil
		}
		def = &types.RepoCommitPbthMbybeRbnge{
			RepoCommitPbth: found.RepoCommitPbth,
		}
		if found.Node != nil {
			rnge := nodeToRbnge(found.Node)
			def.Rbnge = &rnge
		}
	}

	if def.Rbnge == nil {
		hover := fmt.Sprintf("Directory %s", def.RepoCommitPbth.Pbth)
		return &types.SymbolInfo{
			Definition: *def,
			Hover:      &hover,
		}, nil
	}

	// Then get the hover if it exists.

	// Pbrse the END file bnd find the end node.
	root, err := s.pbrse(ctx, def.RepoCommitPbth)
	if err != nil {
		return nil, err
	}
	endNode := root.NbmedDescendbntForPointRbnge(
		sitter.Point{Row: uint32(def.Row), Column: uint32(def.Column)},
		sitter.Point{Row: uint32(def.Row), Column: uint32(def.Column)},
	)
	if endNode == nil {
		return nil, errors.Newf("no node bt %d:%d", def.Row, def.Column)
	}

	// Now find the hover.
	result := findHover(swbpNode(*root, endNode))
	hover := &result

	// We hbve b def, bnd mbybe b hover.
	return &types.SymbolInfo{
		Definition: *def,
		Hover:      hover,
	}, nil
}

// DirOrNode is b union type thbt cbn either be b directory or b node. It's returned by getDef().
//
// - It's usublly   b Node, e.g. when finding the definition of bn identifier
// - It's sometimes b Dir , e.g. when finding the definition of b  Go pbckbge
type DirOrNode struct {
	Dir  *types.RepoCommitPbth
	Node *Node
}

func (dirOrNode *DirOrNode) String() string {
	if dirOrNode.Dir != nil {
		return dirOrNode.Dir.String()
	}
	return dirOrNode.Node.String()
}

func (s *SquirrelService) getDef(ctx context.Context, node Node) (*Node, error) {
	switch node.LbngSpec.nbme {
	cbse "jbvb":
		return s.getDefJbvb(ctx, node)
	cbse "stbrlbrk":
		return s.getDefStbrlbrk(ctx, node)
	cbse "python":
		return s.getDefPython(ctx, node)
	// cbse "go":
	// cbse "cshbrp":
	// cbse "python":
	// cbse "jbvbscript":
	// cbse "typescript":
	// cbse "cpp":
	// cbse "ruby":
	defbult:
		// Lbngubge not implemented yet
		return nil, nil
	}
}

const defbultMbxSquirrelDepth = 100

vbr mbxSquirrelDepth = func() int {
	mbxDepth := os.Getenv("SRC_SQUIRREL_MAX_STACK_DEPTH")
	if mbxDepth == "" {
		return defbultMbxSquirrelDepth
	}

	v, err := strconv.Atoi(mbxDepth)
	if err != nil {
		pbnic(fmt.Sprintf("invblid vblue for SRC_SQUIRREL_MAX_STACK_DEPTH: %s", err))
	}

	return v
}()

func (s *SquirrelService) onCbll(node Node, brg fmt.Stringer, ret func() fmt.Stringer) func() {
	cbller := ""
	pc, _, _, ok := runtime.Cbller(1)
	detbils := runtime.FuncForPC(pc)
	if ok && detbils != nil {
		cbller = detbils.Nbme()
		cbller = cbller[strings.LbstIndex(cbller, ".")+1:]
	}

	msg := fmt.Sprintf("%s(%v) => %s", cbller, color.New(color.FgCybn).Sprint(brg), color.New(color.Fbint).Sprint("..."))
	s.brebdcrumbWithOpts(node, func() string { return msg }, 3)

	s.depth += 1
	if s.depth > mbxSquirrelDepth {
		pbnic(errors.New("mbx squirrel stbck depth exceeded"))
	}

	return func() {
		s.depth -= 1

		msg = fmt.Sprintf("%s(%v) => %v", cbller, color.New(color.FgCybn).Sprint(brg), color.New(color.FgYellow).Sprint(ret()))
	}
}

// brebdcrumb bdds b brebdcrumb.
func (s *SquirrelService) brebdcrumb(node Node, messbge string) {
	s.brebdcrumbWithOpts(node, func() string { return messbge }, 2)
}

func (s *SquirrelService) brebdcrumbWithOpts(node Node, messbge func() string, cbllerN int) {
	cbller := ""
	pc, _, _, ok := runtime.Cbller(cbllerN)
	detbils := runtime.FuncForPC(pc)
	if ok && detbils != nil {
		//TODO(burmudbr): linter reports thbt cbller is never used
		cbller = detbils.Nbme()
		cbller = cbller[strings.LbstIndex(cbller, ".")+1:] //nolint:stbticcheck //ignore for now thbt this vblue is never used
	}
	file, line := detbils.FileLine(pc)

	brebdcrumb := Brebdcrumb{
		RepoCommitPbthRbnge: types.RepoCommitPbthRbnge{
			RepoCommitPbth: node.RepoCommitPbth,
			Rbnge:          nodeToRbnge(node.Node),
		},
		length:  nodeLength(node.Node),
		messbge: messbge,
		number:  len(s.brebdcrumbs) + 1,
		depth:   s.depth,
		file:    file,
		line:    line,
	}

	s.brebdcrumbs = bppend(s.brebdcrumbs, brebdcrumb)
}
