package reposource

import (
	"github.com/sourcegraph/sourcegraph/schema"
)

// CompileGerritNameTransformations compiles a list of GerritNameTransformation into common NameTransformation,
// it halts and returns when any compile error occurred.
func CompileGerritNameTransformations(ts []*schema.GerritNameTransformation) (NameTransformations, error) {
	nts := make([]NameTransformation, len(ts))
	for i, t := range ts {
		nt, err := NewNameTransformation(NameTransformationOptions{
			Regex:       t.Regex,
			Replacement: t.Replacement,
		})
		if err != nil {
			return nil, err
		}
		nts[i] = nt
	}
	return nts, nil
}
