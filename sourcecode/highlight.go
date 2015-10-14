package sourcecode

import (
	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"

	"github.com/sourcegraph/annotate"

	"src.sourcegraph.com/syntaxhighlight"
)

func SyntaxHighlight(fileName string, src []byte) ([]*annotate.Annotation, error) {
	htmlAnn := syntaxhighlight.NewHTMLAnnotator(syntaxhighlight.DefaultHTMLConfig)
	return runAnnotator(htmlAnn, fileName, src)
}

// Tokenize takes a file entry and returns its contents as a tokenized structure.
func Tokenize(e *vcsclient.FileWithRange) *sourcegraph.SourceCode {
	nilAnn := NewNilAnnotator(e)
	// TODO(sqs!): error check?
	runAnnotator(nilAnn, e.Name, e.Contents)
	return nilAnn.Code
}

// TokenizePlain takes a file entry and returns its contents as a tokenized structure.
// This function assumes that the file does not need syntax highlighting and returns
// pure string tokens.
func TokenizePlain(e *vcsclient.FileWithRange) *sourcegraph.SourceCode {
	lines := strings.Split(string(e.Contents), "\n")
	code := sourcegraph.SourceCode{
		Lines: make([]*sourcegraph.SourceCodeLine, len(lines)),
	}
	for i, line := range lines {
		code.Lines[i] = &sourcegraph.SourceCodeLine{
			Tokens: []*sourcegraph.SourceCodeToken{{Label: line}},
		}
	}
	return &code
}

func runAnnotator(annotator syntaxhighlight.Annotator, fileName string, src []byte) ([]*annotate.Annotation, error) {
	anns, err := syntaxhighlight.Annotate(src, fileName, ``, annotator)
	if err != nil {
		return nil, err
	}
	return anns, nil
}
