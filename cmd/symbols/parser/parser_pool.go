package parser

import (
	"context"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ParserType = uint8

const (
	UnknownCtags ParserType = iota
	NoCtags
	UniversalCtags
	ScipCtags
)

func paserNameToParserType(name string) (ParserType, error) {
	switch name {
	case "off":
		return NoCtags, nil
	case "universal-ctags":
		return UniversalCtags, nil
	case "scip-ctags":
		return ScipCtags, nil
	default:
		return UnknownCtags, errors.Errorf("unknown parser type: %s", name)
	}
}

type ParserFactory func(ParserType) (ctags.Parser, error)

type parserPool struct {
	newParser ParserFactory
	pool      map[ParserType]chan ctags.Parser
}

func NewParserPool(newParser ParserFactory, numParserProcesses int) (*parserPool, error) {
	pool := make(map[ParserType]chan ctags.Parser)

	// NOTE: We obviously don't make `NoCtags` available in the pool.
	for _, parserType := range []ParserType{UniversalCtags, ScipCtags} {
		pool[parserType] = make(chan ctags.Parser, numParserProcesses)
		for i := 0; i < numParserProcesses; i++ {
			parser, err := newParser(parserType)
			if err != nil {
				return nil, err
			}
			pool[parserType] <- parser
		}
	}

	parserPool := &parserPool{
		newParser: newParser,
		pool:      pool,
	}

	return parserPool, nil
}

// Get a parser from the pool. Once this parser is no longer in use, the Done method
// MUST be called with either the parser or a nil value (when countering an error).
// Nil values will be recreated on-demand via the factory supplied when constructing
// the pool. This method always returns a non-nil parser with a nil error value.
//
// This method blocks until a parser is available or the given context is canceled.
func (p *parserPool) Get(ctx context.Context, source ParserType) (ctags.Parser, error) {
	if source == NoCtags || source == UnknownCtags {
		return nil, errors.New("NoCtags is not a valid ParserType")
	}

	pool := p.pool[source]

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

func (p *parserPool) Done(parser ctags.Parser, source ParserType) {
	pool := p.pool[source]
	pool <- parser
}
