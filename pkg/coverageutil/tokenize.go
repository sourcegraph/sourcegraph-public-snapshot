// coverageutil contain aux functions and classes
// for coverage command
package coverageutil

import (
	"strings"
)

// Token is a source code token located at the given offset.
// Usually token matches language identifier
type Token struct {
	// Byte offset
	Offset uint32
	// Token text
	Text string
}

// Tokenizer produces tokens from source code.
type Tokenizer interface {
	// Init initializes tokenizer using given data
	Init(src []byte)
	// Next returns next token or nil if no more tokens can be produced
	// (EOF or unrecoverable error)
	Next() *Token
	// Done deallocates tokenizer's resources if needed
	Done()
}

// NOOP tokenizer
type noopTokenizer struct{}

func (noop *noopTokenizer) Init(src []byte) {
}

func (noop *noopTokenizer) Next() *Token {
	return nil
}

func (noop *noopTokenizer) Done() {
}

// singleton
var noop = &noopTokenizer{}

// tokenizerFactory makes tokenizers
type tokenizerFactory func() Tokenizer

// lookup return tokenizer factory that can handle source code
// written in `lang` and located in the file `path`.
type lookup func(lang, path string) tokenizerFactory

// tokenizer lookup function registry
var registry = make([]lookup, 0)

// register registers new tokenizer lookup function
func register(fn lookup) {
	registry = append(registry, fn)
}

// Lookup return tokenizer that can handle source code
// written in `lang` and located in the file `path`.
// It asks all the lookups registered if they can produce a tokenizer
// and returns noop if there is no match
func Lookup(lang, path string) Tokenizer {
	for _, fn := range registry {
		factory := fn(lang, path)
		if factory != nil {
			return factory()
		}
	}
	return noop
}

// newExtensionBasedLookup adds new tokenizer lookup function that
// produces tokenizers based on language and list of known extensions
func newExtensionBasedLookup(language string, extensions []string, factory tokenizerFactory) {
	register(func(lang, path string) tokenizerFactory {
		if language != lang {
			return nil
		}
		for _, extension := range extensions {
			if strings.HasSuffix(path, extension) {
				return factory
			}
		}
		return nil
	})
}
