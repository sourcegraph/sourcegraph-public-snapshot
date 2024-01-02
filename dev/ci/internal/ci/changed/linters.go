package changed

import (
	"strings"
)

var diffsWithLinters = []Diff{
	Go,
	Dockerfiles,
	Docs,
	Client,
	Shell,
	Protobuf,
}

// GetTargets evaluates the lint targets to run over the given CI diff.
func GetLinterTargets(diff Diff) (targets []string) {
	for _, d := range diffsWithLinters {
		if diff.Has(d) {
			targets = append(targets, strings.ToLower(d.String()))
		}
	}
	return
}
