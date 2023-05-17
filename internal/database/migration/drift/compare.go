package drift

import (
	"sort"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func CompareSchemaDescriptions(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary {
	s := []Summary{}
	for _, f := range []func(schemaName, version string, actual, expected schemas.SchemaDescription) []Summary{
		compareExtensions,
		compareEnums,
		compareFunctions,
		compareSequences,
		compareTables,
		compareViews,
	} {
		s = append(s, f(schemaName, version, actual, expected)...)
	}

	return s
}

// compareNamedLists invokes the given primary callback with a pair of differing elements from slices
// `as` and `bs`, respectively, with the same name. If there is a missing element from `as`, there will
// be an invocation of this callback with a nil value for its first parameter. If any invocation of the
// function returns true, the output of this function will be true.
func compareNamedLists[T schemas.Namer](
	as []T,
	bs []T,
	primaryCallback func(a *T, b T) Summary,
) []Summary {
	return compareNamedListsStrict(as, bs, primaryCallback, noopAdditionalCallback[T])
}

// compareNamedListsStrict invokes the given primary callback with a pair of differing elements from
// slices `as` and `bs`, respectively, with the same name. If there is a missing element from `as`, there
// will be an invocation of this callback with a nil value for its first parameter. Elements for which there
// is no analog in `bs` will be collected and sent to an invocation of the additions callback. If any
// invocation of either function returns true, the output of this function will be true.
func compareNamedListsStrict[T schemas.Namer](
	as []T,
	bs []T,
	primaryCallback func(a *T, b T) Summary,
	additionsCallback func(additional []T) []Summary,
) []Summary {
	wrappedPrimaryCallback := func(a *T, b T) []Summary {
		if v := primaryCallback(a, b); v != nil {
			return singleton(v)
		}

		return nil
	}

	return compareNamedListsMultiStrict(as, bs, wrappedPrimaryCallback, additionsCallback)
}

// compareNamedListsMulti invokes the given primary callback with a pair of differing elements from slices
// `as` and `bs`, respectively, with the same name. Similar `compareNamedLists`, but this version expects
// multiple `Summary` values from the callback.
func compareNamedListsMulti[T schemas.Namer](
	as []T,
	bs []T,
	primaryCallback func(a *T, b T) []Summary,
) []Summary {
	return compareNamedListsMultiStrict(as, bs, primaryCallback, noopAdditionalCallback[T])
}

// compareNamedListsMultiStrict invokes the given primary callback with a pair of differing elements from
// slices `as` and `bs`, respectively, with the same name. Similar `compareNamedListsStrict`, but
// this version expects multiple `Summary` values from the callback.
func compareNamedListsMultiStrict[T schemas.Namer](
	as []T,
	bs []T,
	primaryCallback func(a *T, b T) []Summary,
	additionsCallback func(additional []T) []Summary,
) []Summary {
	am := schemas.GroupByName(as)
	bm := schemas.GroupByName(bs)
	additional := make([]T, 0, len(am))
	summaries := []Summary(nil)

	for _, k := range keys(am) {
		av := schemas.Normalize(am[k])

		if bv, ok := bm[k]; ok {
			bv = schemas.Normalize(bv)

			if cmp.Diff(schemas.PreComparisonNormalize(av), schemas.PreComparisonNormalize(bv)) != "" {
				summaries = append(summaries, primaryCallback(&av, bv)...)
			}
		} else {
			additional = append(additional, av)
		}
	}
	for _, k := range keys(bm) {
		bv := schemas.Normalize(bm[k])

		if _, ok := am[k]; !ok {
			summaries = append(summaries, primaryCallback(nil, bv)...)
		}
	}

	if len(additional) > 0 {
		summaries = append(summaries, additionsCallback(additional)...)
	}

	return summaries
}

func noopAdditionalCallback[T schemas.Namer](_ []T) []Summary {
	return nil
}

// keys returns the ordered keys of the given map.
func keys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
