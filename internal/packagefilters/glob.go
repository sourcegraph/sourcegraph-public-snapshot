package packagefilters

import (
	"strings"

	"github.com/gobwas/glob/syntax"
	"github.com/gobwas/glob/syntax/ast"
	"github.com/grafana/regexp"
)

func GlobToRegex(pattern string) (string, error) {
	tree, err := syntax.Parse(pattern)
	if err != nil {
		return "", err
	}

	var out string

	for _, child := range tree.Children {
		out += handleNode(child)
	}

	return "^" + out + "$", nil
}

func handleNode(node *ast.Node) string {
	switch node.Kind {
	// treat ** as nothing special here
	case ast.KindAny, ast.KindSuper:
		return ".*"
	case ast.KindSingle:
		return "."
	case ast.KindText:
		return regexp.QuoteMeta((node.Value.(ast.Text)).Text)
	case ast.KindAnyOf:
		subExpr := make([]string, 0, len(node.Children))
		for _, child := range node.Children {
			subExpr = append(subExpr, handleNode(child))
		}
		return "(?:" + strings.Join(subExpr, "|") + ")"
	case ast.KindPattern:
		// afaik this only ever has 1 child, but why not eh
		subExpr := make([]string, 0, len(node.Children))
		for _, child := range node.Children {
			subExpr = append(subExpr, handleNode(child))
		}
		return strings.Join(subExpr, "")
	case ast.KindList:
		listmeta := node.Value.(ast.List)
		var prefix string
		if listmeta.Not {
			prefix = "^"
		}
		return "[" + prefix + regexp.QuoteMeta(listmeta.Chars) + "]"
	case ast.KindRange:
		rangemeta := node.Value.(ast.Range)
		var prefix string
		if rangemeta.Not {
			prefix = "^"
		}
		return "[" + prefix + regexp.QuoteMeta(string(rangemeta.Lo)) + "-" + regexp.QuoteMeta(string(rangemeta.Hi)) + "]"
	default:
		panic("uh oh stinky")
	}
}
