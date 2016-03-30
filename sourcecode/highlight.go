package sourcecode

import (
	"github.com/sourcegraph/annotate"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/syntaxhighlight"
)

func SyntaxHighlight(fileName string, src []byte) ([]*annotate.Annotation, error) {
	htmlAnn := syntaxhighlight.NewHTMLAnnotator(syntaxhighlight.DefaultHTMLConfig)
	return runAnnotator(htmlAnn, fileName, src)
}

func runAnnotator(annotator syntaxhighlight.Annotator, fileName string, src []byte) ([]*annotate.Annotation, error) {
	anns, err := syntaxhighlight.Annotate(src, fileName, ``, annotator)
	if err != nil {
		return nil, err
	}
	return anns, nil
}
