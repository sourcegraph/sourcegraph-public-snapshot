package sourcecode

import (
	"testing"

	"github.com/kr/pretty"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"src.sourcegraph.com/syntaxhighlight"
)

// codeEquals tests the equality between the given SourceCode entry and an
// array of lines containing arrays of tokens as their string representation.
func codeEquals(code *sourcegraph.SourceCode, want [][]string) bool {
	if len(code.Lines) != len(want) {
		return false
	}
	for i, line := range code.Lines {
		for j, t := range line.Tokens {
			if t.Label != want[i][j] {
				return false
			}
		}
	}
	return true
}

func TestCodeEquals(t *testing.T) {
	for _, tt := range []struct {
		code *sourcegraph.SourceCode
		want [][]string
	}{
		{
			code: &sourcegraph.SourceCode{
				Lines: []*sourcegraph.SourceCodeLine{
					{
						Tokens: []*sourcegraph.SourceCodeToken{
							{Label: "a"},
							{Label: "b"},
							{Label: "c"},
							{Label: "d"},
							{Label: "e"},
						},
					},
					{},
					{
						Tokens: []*sourcegraph.SourceCodeToken{
							{Label: "c"},
						},
					},
				},
			},
			want: [][]string{{"a", "b", "c", "d", "e"}, {}, {"c"}},
		},
	} {
		if !codeEquals(tt.code, tt.want) {
			t.Errorf("Expected: %# v, Got: %# v\n", tt.code, tt.want)
		}
	}
}

func newFileWithRange(src []byte) *vcsclient.FileWithRange {
	return &vcsclient.FileWithRange{
		TreeEntry: &vcsclient.TreeEntry{Contents: []byte(src)},
		FileRange: vcsclient.FileRange{StartByte: 0, EndByte: int64(len(src))},
	}
}

func TestNilAnnotator_multiLineTokens(t *testing.T) {
	for _, tt := range []struct {
		src  string
		want [][]string
	}{
		{
			src: "/* I am\na multiline\ncomment\n*/",
			want: [][]string{
				{"/* I am"},
				{"a multiline"},
				{"comment"},
				{"*/"},
			},
		},
		{
			src: "a := `I am\na multiline\nstring literal\n`",
			want: [][]string{
				{"a", " ", ":=", " ", "`I am"},
				{"a multiline"},
				{"string literal"},
				{"`"},
			},
		},
	} {
		e := newFileWithRange([]byte(tt.src))
		ann := NewNilAnnotator(e)
		_, err := syntaxhighlight.Annotate(e.Contents, ``, `text\x-gosrc`, ann)
		if err != nil {
			t.Fatal(err)
		}
		if !codeEquals(ann.Code, tt.want) {
			t.Errorf("Expected %# v\n\nGot %# v", tt.want, pretty.Formatter(ann.Code.Lines))
		}
	}
}

// The start byte of the first tokenized line should match the start byte of
// the entry's FileRange, and the end byte of the last line should match the end
// byte of the same FileRange.
func TestTokenize_startAndEndBytes(t *testing.T) {
	for _, tt := range []struct {
		src           string
		wantStartByte int64
		wantEndByte   int64
	}{
		{
			src:           "type SingleLineTreeEntry string",
			wantStartByte: 0,
		},
		{
			src:           "type SingleLineTreeEntry string",
			wantStartByte: 1993,
		},
		{
			src:           "foo int\nbar int\nbaz int",
			wantStartByte: 0,
		},
	} {
		tt.wantEndByte = tt.wantStartByte + int64(len(tt.src)-1)
		e := newFileWithRange([]byte(tt.src))
		e.FileRange.StartByte, e.FileRange.EndByte = tt.wantStartByte, tt.wantEndByte

		sc := Tokenize(e)
		firstLine, lastLine := sc.Lines[0], sc.Lines[len(sc.Lines)-1]
		if int64(firstLine.StartByte) != tt.wantStartByte {
			t.Errorf("Expected first line start byte to be %d\nGot %d", tt.wantStartByte, firstLine.StartByte)
		}
		if int64(lastLine.EndByte) != tt.wantEndByte {
			t.Errorf("Expected last line end byte to be %d\nGot %d", tt.wantEndByte, lastLine.EndByte)
		}
	}
}
