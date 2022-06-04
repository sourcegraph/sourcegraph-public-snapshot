package parser

import (
	"context"
	"errors"

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
	if err != nil && !errors.Is(err, squirrel.UnrecognizedFileExtensionError) && !errors.Is(err, squirrel.UnsupportedLanguageError) && !errors.Is(err, squirrel.NoTopLevelSymbolsQuery) {
		return nil, err
	}
	if symbols != nil {
		return symbols, nil
	}

	return p.fallback.Parse(path, contents)
}

func (p *TreeSitterParser) Close() {
	p.fallback.Close()
}
