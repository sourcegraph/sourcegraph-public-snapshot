package autogold

import (
	"fmt"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/hexops/valast"
)

func diff(got, want string, opts []Option) string {
	edits := myers.ComputeEdits(span.URIFromPath("out"), string(want), got)
	return fmt.Sprint(gotextdiff.ToUnified("want", "got", string(want), edits))
}

// Raw denotes a raw string.
type Raw string

func stringify(v interface{}, opts []Option) string {
	var (
		allowRaw, trailingNewline bool
		valastOpt                 = &valast.Options{}
	)
	for _, opt := range opts {
		opt := opt.(*option)
		if opt.exportedOnly {
			valastOpt.ExportedOnly = true
		}
		if opt.forPackageName != "" {
			valastOpt.PackageName = opt.forPackageName
		}
		if opt.forPackagePath != "" {
			valastOpt.PackagePath = opt.forPackagePath
		}
		if isBazel() {
			valastOpt.PackagePathToName = bazelPackagePathToName
		}
		if opt.allowRaw {
			allowRaw = true
		}
		if opt.trailingNewline {
			trailingNewline = true
		}
	}
	if v, ok := v.(Raw); ok && allowRaw {
		return string(v)
	}
	s := valast.StringWithOptions(v, valastOpt)
	if trailingNewline {
		return s + "\n"
	}
	return s
}
