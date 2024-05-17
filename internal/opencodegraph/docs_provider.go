package opencodegraph

import (
	"context"
	"go/token"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	RegisterProvider(docsProvider{})
}

type docsProvider struct{}

func (docsProvider) Name() string { return "docs" }

func (docsProvider) Capabilities(ctx context.Context, params schema.CapabilitiesParams) (*schema.CapabilitiesResult, error) {
	return &schema.CapabilitiesResult{}, nil
}

func (docsProvider) Annotations(ctx context.Context, params schema.AnnotationsParams) (*schema.AnnotationsResult, error) {
	var result schema.AnnotationsResult

	for _, p := range docsPatterns {
		if p.PathPattern != nil && !p.PathPattern.MatchString(params.File) {
			continue
		}

		var rs []schema.OpenCodeGraphRange
		if p.ContentPattern != nil {
			ms := p.ContentPattern.FindAllStringIndex(params.Content, -1)
			if len(ms) == 0 {
				continue
			}

			fset := token.NewFileSet()
			f := fset.AddFile(params.File, 1, len(params.Content))
			f.SetLinesForContent([]byte(params.Content))

			for _, m := range ms {
				mstart := m[0]
				mend := m[1]
				start := f.Position(f.Pos(mstart))
				end := f.Position(f.Pos(mend))
				rs = append(rs, schema.OpenCodeGraphRange{
					Start: schema.OpenCodeGraphPosition{Line: start.Line - 1, Character: start.Column - 1},
					End:   schema.OpenCodeGraphPosition{Line: end.Line - 1, Character: end.Column - 1},
				})
			}
		} else {
			rs = append(rs, schema.OpenCodeGraphRange{
				Start: schema.OpenCodeGraphPosition{Line: 0, Character: 0},
				End:   schema.OpenCodeGraphPosition{Line: 0, Character: 0},
			})
		}

		id := p.Id
		result.Items = append(result.Items, &schema.OpenCodeGraphItem{
			Id:         id,
			Title:      "📘 Docs: " + p.Title,
			Url:        p.URL,
			Preview:    true,
			PreviewUrl: p.URL,
		})
		for _, r := range rs {
			result.Annotations = append(result.Annotations, &schema.OpenCodeGraphAnnotation{
				Item:  schema.OpenCodeGraphItemRef{Id: id},
				Range: r,
			})
		}
	}

	return &result, nil
}

var docsPatterns = []struct {
	Id                          string
	PathPattern, ContentPattern *regexp.Regexp
	Title                       string
	URL                         string
}{
	{
		Id:             "telemetry",
		PathPattern:    regexp.MustCompile(`\.tsx?$`),
		ContentPattern: regexp.MustCompile(`eventLogger.log(?:ViewEvent)?`),
		Title:          "Telemetry",
		URL:            "https://docs-legacy.sourcegraph.com/dev/background-information/telemetry#sourcegraph-web-app",
	},
	{
		Id:             "css",
		PathPattern:    regexp.MustCompile(`\.tsx?$`),
		ContentPattern: regexp.MustCompile(`import styles from.*\.module\.s?css'`),
		Title:          "CSS in client/web",
		URL:            "https://docs-legacy.sourcegraph.com/dev/background-information/web/styling",
	},
	{
		Id:          "bazel",
		PathPattern: regexp.MustCompile(`(^|/)BUILD\.bazel$`),
		Title:       "Bazel at Sourcegraph",
		URL:         "https://docs-legacy.sourcegraph.com/dev/background-information/bazel",
	},
}
