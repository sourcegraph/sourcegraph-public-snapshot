package parser

import (
	"bytes"

	"github.com/sourcegraph/go-ctags"
)

type FilteringParser struct {
	parser      ctags.Parser
	maxFileSize int
	maxSymbols  int
}

func NewFilteringParser(parser ctags.Parser, maxFileSize int, maxSymbols int) ctags.Parser {
	return &FilteringParser{
		parser:      parser,
		maxFileSize: maxFileSize,
		maxSymbols:  maxSymbols,
	}
}

func (p *FilteringParser) Parse(path string, content []byte) ([]*ctags.Entry, error) {
	if len(content) > p.maxFileSize {
		// File is over 512KiB, don't parse it
		return nil, nil
	}

	// Check to see if first 256 bytes contain a 0x00. If so, we'll assume that
	// the file is binary and skip parsing. Otherwise, we'll have some non-zero
	// contents that passed our filters above to parse.
	if bytes.IndexByte(content[:min(len(content), 256)], 0x00) >= 0 {
		return nil, nil
	}

	entries, err := p.parser.Parse(path, content)
	if err != nil {
		return nil, err
	}

	if len(entries) > p.maxSymbols {
		// File has too many symbols, don't return any of them
		return nil, nil
	}

	return entries, nil
}

func (p *FilteringParser) Close() {
	p.parser.Close()
}
