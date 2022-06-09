package parser

import (
	"log"
	"os"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
)

func NewCtagsParserFactory(config types.CtagsConfig) ParserFactory {
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

	return func() (ctags.Parser, error) {
		parser, err := ctags.New(options)
		if err != nil {
			return nil, err
		}
		return NewFilteringParser(parser, config.MaxFileSize, config.MaxSymbols), nil
	}
}
