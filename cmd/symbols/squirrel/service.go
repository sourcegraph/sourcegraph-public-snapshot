package squirrel

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/fatih/color"
	sitter "github.com/smacker/go-tree-sitter"

	symbolsTypes "github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// How to read a file.
type readFileFunc func(context.Context, types.RepoCommitPath) ([]byte, error)

// SquirrelService uses tree-sitter and the symbols service to analyze and traverse files to find
// symbols.
type SquirrelService struct {
	readFile            readFileFunc
	symbolSearch        symbolsTypes.SearchFunc
	breadcrumbs         Breadcrumbs
	parser              *sitter.Parser
	closables           []func()
	errorOnParseFailure bool
	depth               int
}

// New creates a new SquirrelService.
func New(readFile readFileFunc, symbolSearch symbolsTypes.SearchFunc) *SquirrelService {
	return &SquirrelService{
		readFile:            readFile,
		symbolSearch:        symbolSearch,
		breadcrumbs:         []Breadcrumb{},
		parser:              sitter.NewParser(),
		closables:           []func(){},
		errorOnParseFailure: false,
	}
}

// Close frees memory allocated by tree-sitter.
func (s *SquirrelService) Close() {
	for _, c := range s.closables {
		c()
	}
	s.parser.Close()
}

// SymbolInfo finds the symbol at the given point in a file, or nil the definition can't be determined.
func (s *SquirrelService) SymbolInfo(ctx context.Context, point types.RepoCommitPathPoint) (*types.SymbolInfo, error) {
	// First, find the definition.
	var def *types.RepoCommitPathMaybeRange
	{
		// Parse the file and find the starting node.
		root, err := s.parse(ctx, point.RepoCommitPath)
		if err != nil {
			return nil, err
		}
		startNode := root.NamedDescendantForPointRange(
			sitter.Point{Row: uint32(point.Row), Column: uint32(point.Column)},
			sitter.Point{Row: uint32(point.Row), Column: uint32(point.Column)},
		)
		if startNode == nil {
			return nil, errors.New("node is nil")
		}

		// Now find the definition.
		found, err := s.getDef(ctx, swapNode(*root, startNode))
		if err != nil {
			return nil, err
		}
		if found == nil {
			return nil, nil
		}
		def = &types.RepoCommitPathMaybeRange{
			RepoCommitPath: found.RepoCommitPath,
		}
		if found.Node != nil {
			rnge := nodeToRange(found.Node)
			def.Range = &rnge
		}
	}

	if def.Range == nil {
		hover := fmt.Sprintf("Directory %s", def.RepoCommitPath.Path)
		return &types.SymbolInfo{
			Definition: *def,
			Hover:      &hover,
		}, nil
	}

	// Then get the hover if it exists.

	// Parse the END file and find the end node.
	root, err := s.parse(ctx, def.RepoCommitPath)
	if err != nil {
		return nil, err
	}
	endNode := root.NamedDescendantForPointRange(
		sitter.Point{Row: uint32(def.Row), Column: uint32(def.Column)},
		sitter.Point{Row: uint32(def.Row), Column: uint32(def.Column)},
	)
	if endNode == nil {
		return nil, errors.Newf("no node at %d:%d", def.Row, def.Column)
	}

	// Now find the hover.
	result := findHover(swapNode(*root, endNode))
	hover := &result

	// We have a def, and maybe a hover.
	return &types.SymbolInfo{
		Definition: *def,
		Hover:      hover,
	}, nil
}

// DirOrNode is a union type that can either be a directory or a node. It's returned by getDef().
//
// - It's usually   a Node, e.g. when finding the definition of an identifier
// - It's sometimes a Dir , e.g. when finding the definition of a  Go package
type DirOrNode struct {
	Dir  *types.RepoCommitPath
	Node *Node
}

func (dirOrNode *DirOrNode) String() string {
	if dirOrNode.Dir != nil {
		return dirOrNode.Dir.String()
	}
	return dirOrNode.Node.String()
}

func (s *SquirrelService) getDef(ctx context.Context, node Node) (*Node, error) {
	switch node.LangSpec.name {
	case "java":
		return s.getDefJava(ctx, node)
	case "starlark":
		return s.getDefStarlark(ctx, node)
	case "python":
		return s.getDefPython(ctx, node)
	// case "go":
	// case "csharp":
	// case "python":
	// case "javascript":
	// case "typescript":
	// case "cpp":
	// case "ruby":
	default:
		// Language not implemented yet
		return nil, nil
	}
}

const defaultMaxSquirrelDepth = 100

var maxSquirrelDepth = func() int {
	maxDepth := os.Getenv("SRC_SQUIRREL_MAX_STACK_DEPTH")
	if maxDepth == "" {
		return defaultMaxSquirrelDepth
	}

	v, err := strconv.Atoi(maxDepth)
	if err != nil {
		panic(fmt.Sprintf("invalid value for SRC_SQUIRREL_MAX_STACK_DEPTH: %s", err))
	}

	return v
}()

func (s *SquirrelService) onCall(node Node, arg fmt.Stringer, ret func() fmt.Stringer) func() {
	caller := ""
	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		caller = details.Name()
		caller = caller[strings.LastIndex(caller, ".")+1:]
	}

	msg := fmt.Sprintf("%s(%v) => %s", caller, color.New(color.FgCyan).Sprint(arg), color.New(color.Faint).Sprint("..."))
	s.breadcrumbWithOpts(node, func() string { return msg }, 3)

	s.depth += 1
	if s.depth > maxSquirrelDepth {
		panic(errors.New("max squirrel stack depth exceeded"))
	}

	return func() {
		s.depth -= 1

		msg = fmt.Sprintf("%s(%v) => %v", caller, color.New(color.FgCyan).Sprint(arg), color.New(color.FgYellow).Sprint(ret()))
	}
}

// breadcrumb adds a breadcrumb.
func (s *SquirrelService) breadcrumb(node Node, message string) {
	s.breadcrumbWithOpts(node, func() string { return message }, 2)
}

func (s *SquirrelService) breadcrumbWithOpts(node Node, message func() string, callerN int) {
	caller := ""
	pc, _, _, ok := runtime.Caller(callerN)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		//TODO(burmudar): linter reports that caller is never used
		caller = details.Name()
		caller = caller[strings.LastIndex(caller, ".")+1:] //nolint:staticcheck //ignore for now that this value is never used
	}
	file, line := details.FileLine(pc)

	breadcrumb := Breadcrumb{
		RepoCommitPathRange: types.RepoCommitPathRange{
			RepoCommitPath: node.RepoCommitPath,
			Range:          nodeToRange(node.Node),
		},
		length:  nodeLength(node.Node),
		message: message,
		number:  len(s.breadcrumbs) + 1,
		depth:   s.depth,
		file:    file,
		line:    line,
	}

	s.breadcrumbs = append(s.breadcrumbs, breadcrumb)
}
