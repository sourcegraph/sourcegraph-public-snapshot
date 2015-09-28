package sourcecode

import (
	"testing"

	"github.com/kr/pretty"
	"github.com/sourcegraph/syntaxhighlight"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
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
					&sourcegraph.SourceCodeLine{
						Tokens: []*sourcegraph.SourceCodeToken{
							&sourcegraph.SourceCodeToken{Label: "a"},
							&sourcegraph.SourceCodeToken{Label: "b"},
							&sourcegraph.SourceCodeToken{Label: "c"},
							&sourcegraph.SourceCodeToken{Label: "d"},
							&sourcegraph.SourceCodeToken{Label: "e"},
						},
					},
					&sourcegraph.SourceCodeLine{},
					&sourcegraph.SourceCodeLine{
						Tokens: []*sourcegraph.SourceCodeToken{
							&sourcegraph.SourceCodeToken{Label: "c"},
						},
					},
				},
			},
			want: [][]string{[]string{"a", "b", "c", "d", "e"}, []string{}, []string{"c"}},
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
				[]string{"/* I am"},
				[]string{"a multiline"},
				[]string{"comment"},
				[]string{"*/"},
			},
		},
		{
			src: "a := `I am\na multiline\nstring literal\n`",
			want: [][]string{
				[]string{"a", " ", ":", "=", " ", "`I am"},
				[]string{"a multiline"},
				[]string{"string literal"},
				[]string{"`"},
			},
		},
	} {
		e := newFileWithRange([]byte(tt.src))
		ann := NewNilAnnotator(e)
		_, err := syntaxhighlight.Annotate(e.Contents, ann)
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
		tt.wantEndByte = tt.wantStartByte + int64(len(tt.src))
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
