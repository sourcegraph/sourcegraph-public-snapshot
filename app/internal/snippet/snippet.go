package snippet

import (
	"fmt"
	htmpl "html/template"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

type Snippet struct {
	StartByte   int64
	Code        string
	Annotations *sourcegraph.AnnotationList
	SourceURL   string
}

func Render(s *Snippet) htmpl.HTML {
	var toks []string

	var clsAnns, urlAnns []*sourcegraph.Annotation
	for _, ann := range s.Annotations.Annotations {
		if ann.Class != "" {
			clsAnns = append(clsAnns, ann)
		} else if ann.URL != "" {
			urlAnns = append(urlAnns, ann)
		}
	}

	var prevEnd int64 = 0
	for _, ann := range clsAnns {
		start, end := int64(ann.StartByte), int64(ann.EndByte)
		if start < 0 || end > int64(len(s.Code)) {
			continue
		}

		if start > prevEnd {
			toks = append(toks, htmpl.HTMLEscapeString(s.Code[prevEnd:start]))
		}
		toks = append(toks, fmt.Sprintf("<span class=%s>", ann.Class), htmpl.HTMLEscapeString(s.Code[start:end]), "</span>")
		prevEnd = int64(ann.EndByte)
	}
	toks = append(toks, htmpl.HTMLEscapeString(s.Code[prevEnd:]))

	return htmpl.HTML(strings.Join(toks, ""))
}
