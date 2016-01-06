package runner

import "github.com/drone/drone-exec/parser"

func Load(tree *parser.Tree) *Build {
	return &Build{tree: tree}
}
