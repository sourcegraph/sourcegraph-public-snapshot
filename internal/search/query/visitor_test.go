package query

import (
	"testing"

	"github.com/hexops/autogold/v2"
)

func TestSubstitute(t *testing.T) {
	test := func(input string) string {
		q, _ := ParseLiteral(input)
		var result string
		VisitPredicate(q, func(field, name, value string, negated bool) {
			if field == FieldRepo && name == "contains.file" {
				result = "contains.file value is " + value
			}
		})
		return result
	}

	autogold.Expect("contains.file value is path:foo").
		Equal(t, test("repo:contains.file(path:foo)"))
}

func TestVisitTypedPredicate(t *testing.T) {
	cases := []struct {
		query  string
		output autogold.Value
	}{{
		"repo:test",
		autogold.Expect([]*RepoContainsFilePredicate{}),
	}, {
		"repo:test repo:contains.file(path:test)",
		autogold.Expect([]*RepoContainsFilePredicate{{Path: "test"}}),
	}, {
		"repo:test repo:has.file(path:test)",
		autogold.Expect([]*RepoContainsFilePredicate{{Path: "test"}}),
	}, {
		"repo:test repo:contains.file(test)",
		autogold.Expect([]*RepoContainsFilePredicate{{Path: "test"}}),
	}}

	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			q, _ := ParseLiteral(tc.query)
			var result []*RepoContainsFilePredicate
			VisitTypedPredicate(q, func(pred *RepoContainsFilePredicate) {
				result = append(result, pred)
			})
			tc.output.Equal(t, result)
		})
	}
}
