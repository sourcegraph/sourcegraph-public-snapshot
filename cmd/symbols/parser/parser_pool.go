package parser

import (
	"context"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

type CtagsSource = uint8

const (
	Universal CtagsSource = iota
	Scip
)

type ParserFactory func(CtagsSource) (ctags.Parser, error)

type parserPool struct {
	symbolsSource map[string]any
	newParser     ParserFactory
	universalPool chan ctags.Parser
	scipPool      chan ctags.Parser
}

// TODO(SuperAuguste): have two numParserProcesses - one for universal and one for scip
func NewParserPool(newParser ParserFactory, numParserProcesses int) (*parserPool, error) {
	universalPool := make(chan ctags.Parser, numParserProcesses)
	scipPool := make(chan ctags.Parser, numParserProcesses)

	for i := 0; i < numParserProcesses; i++ {
		parser, err := newParser(Universal)
		if err != nil {
			return nil, err
		}
		universalPool <- parser
	}

	for i := 0; i < numParserProcesses; i++ {
		parser, err := newParser(Scip)
		if err != nil {
			return nil, err
		}
		scipPool <- parser
	}

	ppool := &parserPool{
		symbolsSource: conf.Get().SymbolsSource,
		newParser:     newParser,
		universalPool: universalPool,
		scipPool:      scipPool,
	}

	conf.Watch(func() {
		ppool.symbolsSource = conf.Get().SymbolsSource
	})

	return ppool, nil
}

// Get a parser from the pool. Once this parser is no longer in use, the Done method
// MUST be called with either the parser or a nil value (when countering an error).
// Nil values will be recreated on-demand via the factory supplied when constructing
// the pool. This method always returns a non-nil parser with a nil error value.
//
// This method blocks until a parser is available or the given context is canceled.
func (p *parserPool) Get(ctx context.Context, source CtagsSource) (ctags.Parser, error) {
	var pool chan ctags.Parser
	if source == Universal {
		pool = p.universalPool
	} else {
		pool = p.scipPool
	}
	select {
	case parser := <-pool:
		if parser != nil {
			return parser, nil
		}

		return p.newParser(source)

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *parserPool) Done(parser ctags.Parser, source CtagsSource) {
	var pool chan ctags.Parser
	if source == Universal {
		pool = p.universalPool
	} else {
		pool = p.scipPool
	}
	pool <- parser
}
