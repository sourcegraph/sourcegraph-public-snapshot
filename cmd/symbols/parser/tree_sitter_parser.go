package parser

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/squirrel"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type TreeSitterParser struct {
	fallback SimpleParser
	parser   *sitter.Parser
}

func NewTreeSitterParser(parser SimpleParser) SimpleParser {
	return &TreeSitterParser{
		fallback: parser,
		parser:   sitter.NewParser(),
	}
}

func (p *TreeSitterParser) Parse(path string, contents []byte) (result.Symbols, error) {
	repoCommitPath := types.RepoCommitPath{
		Repo:   "N/A",
		Commit: "N/A",
		Path:   path,
	}
	symbols, err := squirrel.GetSymbols(context.Background(), p.parser, repoCommitPath, contents)
	if err == nil {
		return symbols, nil
	}

	symbols, err = p.fallback.Parse(path, contents)
	if err != nil {
		return nil, err
	}

	return symbols, nil
}

func (p *TreeSitterParser) Close() {
	p.fallback.Close()
}
