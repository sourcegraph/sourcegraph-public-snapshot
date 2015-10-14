package sourcecode

import (
	"bytes"
	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"

	"github.com/sourcegraph/annotate"

	"src.sourcegraph.com/syntaxhighlight"
)

// NilAnnotator is a special kind of annotator that always returns nil, but stores
// within itself the snippet of source code that is passed through it as tokens.
//
// This functionality is useful when one wishes to obtain the tokenized source as a data
// structure, as opposed to an annotated string, allowing full control over rendering and
// displaying it.
type NilAnnotator struct {
	Code       *sourcegraph.SourceCode
	byteOffset int
	line       int
	lines      int
}

// Instantiates new NilAnnotator from a given source code.
// Annotator will contain list of line spans (start byte - end byte) in the source code
func NewNilAnnotator(e *vcsclient.FileWithRange) *NilAnnotator {
	lines := make([]*sourcegraph.SourceCodeLine, 0, bytes.Count(e.Contents, []byte("\n"))+1)
	last := len(e.Contents) - 1

	offset := 0
	index := bytes.IndexByte(e.Contents, '\n')
	for index >= 0 {
		lines = append(lines, newSourceLine(offset, offset+index, e.StartByte))
		offset += index + 1
		if offset == last+1 {
			break
		}
		index = bytes.IndexByte(e.Contents[offset:], '\n')
	}
	if offset <= last {
		lines = append(lines, newSourceLine(offset, last, e.StartByte))
	}
	ann := NilAnnotator{
		Code: &sourcegraph.SourceCode{
			Lines: lines,
		},
		byteOffset: int(e.StartByte),
		line:       0,
		lines:      len(lines),
	}
	return &ann
}

func (a *NilAnnotator) Annotate(token syntaxhighlight.Token) (*annotate.Annotation, error) {
	start := int32(token.Offset) + int32(a.byteOffset)
	for a.line < a.lines {
		line := a.Code.Lines[a.line]
		if line.StartByte <= start && line.EndByte >= start {
			chunks := strings.Split(token.Text, "\n")
			for index, chunk := range chunks {
				if a.line+index >= a.lines {
					break
				}
				l := int32(len(chunk))
				a.addToken(a.Code.Lines[a.line+index],
					&sourcegraph.SourceCodeToken{
						StartByte: int32(start),
						EndByte:   int32(start) + l,
						Class:     syntaxhighlight.DefaultHTMLConfig.GetTokenClass(token),
						Label:     chunk,
					})
				start += l
			}
			a.line += len(chunks) - 1
			return nil, nil
		}
		a.line++
	}
	return nil, nil
}

func (a *NilAnnotator) Init() error {
	return nil
}

func (a *NilAnnotator) Done() error {
	return nil
}

func (a *NilAnnotator) addToken(line *sourcegraph.SourceCodeLine, t *sourcegraph.SourceCodeToken) {
	if (*line).Tokens == nil {
		(*line).Tokens = make([]*sourcegraph.SourceCodeToken, 0, 1)
	}
	(*line).Tokens = append((*line).Tokens, t)
}

func newSourceLine(start int, end int, base int64) *sourcegraph.SourceCodeLine {
	ret := sourcegraph.SourceCodeLine{StartByte: int32(start) + int32(base), EndByte: int32(end) + int32(base)}
	return &ret
}
