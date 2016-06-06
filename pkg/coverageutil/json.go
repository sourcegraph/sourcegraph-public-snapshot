package coverageutil

import (
	"bytes"

	"github.com/sourcegraph/srclib-json-tokenizer/sgjson"
)

type jsonTokenizer struct {
	bytes  []byte
	tokens []sgjson.TokenInfo
	errors []string

	// index of the next unread token in tokens
	pointer int
}

func (jt *jsonTokenizer) Init(src []byte) {
	jt.bytes = src
	// jt.tokens = make([]sgjson.TokenInfo, 0)
	// jt.errors = make([]string, 0)
	jt.pointer = 0
	dec := sgjson.NewDecoder(bytes.NewReader(src))
	dec.UseNumber()
	unfiltered, err := dec.Tokenize()
	if err != nil {
		jt.errors = append(jt.errors, err.Error())
	} else {
		var filtered []sgjson.TokenInfo
		for _, t := range unfiltered {
			switch t.Token.(type) {
			case sgjson.Delim: // skip delimiters
				continue
			case string: // remove beginning and ending quotation marks
				t.Start++
				t.Endp--
			}
			filtered = append(filtered, t)
		}
		jt.tokens = filtered
	}
}

func (jt *jsonTokenizer) Next() *Token {
	// out of tokens
	if jt.pointer >= len(jt.tokens) {
		return nil
	}
	info := jt.tokens[jt.pointer]
	jt.pointer++
	return &Token{
		Offset: uint32(info.Start),
		Line:   info.Line,
		Text:   string(jt.bytes[info.Start:info.Endp]),
	}
}

func (jt *jsonTokenizer) Errors() []string {
	return jt.errors
}

func (jt *jsonTokenizer) Done() {}

func init() {
	factory := func() Tokenizer {
		return &jsonTokenizer{}
	}
	newExtensionBasedLookup("JSON", []string{".json"}, factory)
}
