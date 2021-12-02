package parser

import (
	"log"
	"os"

	"github.com/sourcegraph/go-ctags"
)

func NewCtagsParserFactory(ctagsCommand string, patternLengthLimit int, logErrors, debugLogs bool) ParserFactory {
	options := ctags.Options{
		Bin:                ctagsCommand,
		PatternLengthLimit: patternLengthLimit,
	}
	if logErrors {
		options.Info = log.New(os.Stderr, "ctags: ", log.LstdFlags)
	}
	if debugLogs {
		options.Debug = log.New(os.Stderr, "DBUG ctags: ", log.LstdFlags)
	}

	return func() (ctags.Parser, error) { return ctags.New(options) }
}
