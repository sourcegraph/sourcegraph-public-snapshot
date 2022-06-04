package parser

import (
	"log"
	"os"
	"strings"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type CtagsParser struct {
	parser ctags.Parser
}

func NewCtagsParser(config types.CtagsConfig) (SimpleParser, error) {
	options := ctags.Options{
		Bin:                config.Command,
		PatternLengthLimit: config.PatternLengthLimit,
	}
	if config.LogErrors {
		options.Info = log.New(os.Stderr, "ctags: ", log.LstdFlags)
	}
	if config.DebugLogs {
		options.Debug = log.New(os.Stderr, "DBUG ctags: ", log.LstdFlags)
	}

	parser, err := ctags.New(options)
	if err != nil {
		return nil, err
	}

	return &CtagsParser{parser: parser}, nil
}

func (p *CtagsParser) Parse(path string, contents []byte) (result.Symbols, error) {
	entries, err := p.parser.Parse(path, contents)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(contents), "\n")

	symbols := result.Symbols{}
	for _, entry := range entries {
		symbol := ctagsEntryToSymbol(entry, path, lines)
		if symbol == nil {
			continue
		}

		symbols = append(symbols, *symbol)
	}

	return symbols, nil
}

func (p *CtagsParser) Close() {
	p.parser.Close()
}
