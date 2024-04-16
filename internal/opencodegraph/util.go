package opencodegraph

import (
	"go/token"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/sourcegraph/schema"
)

func firstSubmatchNamesAndRanges(pattern *regexp.Regexp, content string) (names []string, ranges []schema.OpenCodeGraphRange) {
	fset := token.NewFileSet()
	f := fset.AddFile("x", 1, len(content))
	f.SetLinesForContent([]byte(content))

	ms := pattern.FindAllStringSubmatchIndex(content, -1)
	for _, m := range ms {
		mstart := m[2]
		mend := m[3]
		start := f.Position(f.Pos(mstart))
		end := f.Position(f.Pos(mend))

		name := string(content[mstart:mend])
		names = append(names, name)
		ranges = append(ranges, schema.OpenCodeGraphRange{
			Start: schema.OpenCodeGraphPosition{Line: start.Line - 1, Character: start.Column - 1},
			End:   schema.OpenCodeGraphPosition{Line: end.Line - 1, Character: end.Column - 1},
		})
	}

	return names, ranges
}
