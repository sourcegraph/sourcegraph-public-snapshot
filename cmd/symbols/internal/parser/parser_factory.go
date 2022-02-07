package parser

import "github.com/sourcegraph/go-ctags"

type ParserFactory func() (ctags.Parser, error)
