package sourcecode

import (
	"bytes"
	"strings"

	"github.com/sourcegraph/annotate"
	"github.com/sourcegraph/syntaxhighlight"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
)

// NilAnnotator is a special kind of annotator that always returns nil, but stores
// within itself the snippet of source code that is passed through it as tokens.
//
// This functionality is useful when one wishes to obtain the tokenized source as a data
// structure, as opposed to an annotated string, allowing full control over rendering and
// displaying it.
type NilAnnotator struct {
	Config     syntaxhighlight.HTMLConfig
	Code       *sourcegraph.SourceCode
	byteOffset int
}

func NewNilAnnotator(e *vcsclient.FileWithRange) *NilAnnotator {
	ann := NilAnnotator{
		Config: syntaxhighlight.DefaultHTMLConfig,
		Code: &sourcegraph.SourceCode{
			Lines: make([]*sourcegraph.SourceCodeLine, 0, bytes.Count(e.Contents, []byte("\n"))),
		},
		byteOffset: int(e.StartByte),
	}
	ann.addLine(ann.byteOffset)
	return &ann
}

func (a *NilAnnotator) addToken(t *sourcegraph.SourceCodeToken) {
	line := a.Code.Lines[len(a.Code.Lines)-1]
	if line.Tokens == nil {
		line.Tokens = make([]*sourcegraph.SourceCodeToken, 0, 1)
	}
	// If this token and the previous one are both strings, merge them.
	if n := len(line.Tokens); t.Class == a.Config.Whitespace && n > 0 {
		if line.Tokens[n-1].Class == a.Config.Whitespace {
			line.Tokens[n-1].Label += t.Label
			return
		}
	}
	line.Tokens = append(line.Tokens, t)
}

func (a *NilAnnotator) addLine(startByte int) {
	a.Code.Lines = append(a.Code.Lines, &sourcegraph.SourceCodeLine{StartByte: int32(startByte)})
	if len(a.Code.Lines) > 1 {
		lastLine := a.Code.Lines[len(a.Code.Lines)-2]
		lastLine.EndByte = int32(startByte) - 1
	}
}

func (a *NilAnnotator) addMultilineToken(startByte int, unsafeHTML string, class string) {
	lines := strings.Split(unsafeHTML, "\n")
	for n, unsafeHTML := range lines {
		if len(unsafeHTML) > 0 {
			a.addToken(&sourcegraph.SourceCodeToken{
				StartByte: int32(startByte),
				EndByte:   int32(startByte + len(unsafeHTML)),
				Class:     class,
				Label:     unsafeHTML,
			})
			startByte += len(unsafeHTML)
		}
		if n < len(lines)-1 {
			a.addLine(startByte)
		}
	}
}

func (a *NilAnnotator) Annotate(start int, kind syntaxhighlight.Kind, tokText string) (*annotate.Annotation, error) {
	class := ((syntaxhighlight.HTMLConfig)(a.Config)).Class(kind)
	start += a.byteOffset

	switch {
	// New line
	case tokText == "\n":
		a.addLine(start + 1)

	// Multiline token (ie. block comments, string literals)
	case strings.Contains(tokText, "\n"):
		// Here we pass the unescaped string so we can calculate line lenghts correctly.
		// This method is expected to take responsibility of escaping any token text.
		a.addMultilineToken(start+1, tokText, class)

	// Token
	default:
		a.addToken(&sourcegraph.SourceCodeToken{
			StartByte: int32(start),
			EndByte:   int32(start + len(tokText)),
			Class:     class,
			Label:     tokText,
		})
	}

	return nil, nil
}
