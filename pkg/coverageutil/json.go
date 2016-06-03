package coverageutil

import (
	"bytes"

	"github.com/sourcegraph/srclib-json-tokenizer/sgjson"
)

type jsonTokenizer struct {
	bytes  []byte
	tokens []sgjson.TokenInfo

	//index of the next unread token in tokens
	pointer int
}

func (jt *jsonTokenizer) Init(src []byte) {
	dec := sgjson.NewDecoder(bytes.NewReader(src))
	dec.UseNumber()
	unfiltered, err := dec.Tokenize()

	if err != nil {
		return
	}

	var filtered []sgjson.TokenInfo

	for _, t := range unfiltered {
		switch t.Token.(type) {

		//skip delimiters
		case sgjson.Delim:
			continue

		//remove beginning and ending quotation marks
		case string:
			t.Start++
			t.Endp--
		}

		filtered = append(filtered, t)
	}
	jt.bytes = src
	jt.tokens = filtered
	jt.pointer = 0
}

func (jt *jsonTokenizer) Next() *Token {
	//error in tokenizing or out of tokens
	if jt.pointer >= len(jt.tokens) {
		return nil
	}

	info := jt.tokens[jt.pointer]
	jt.pointer++

	out := &Token{}
	out.Offset = uint32(info.Start)
	out.Text = string(jt.bytes[info.Start:info.Endp])
	out.Line = info.Line

	return out
}

func (jt *jsonTokenizer) Done() {}

func init() {
	var factory = func() Tokenizer {
		return &jsonTokenizer{}
	}
	newExtensionBasedLookup("JSON", []string{".json"}, factory)
}
