package query

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestSubstitute(t *testing.T) {
	test := func(input string) string {
		q, _ := ParseLiteral(input)
		var result string
		VisitPredicate(q, func(field, name, value string) {
			if field == FieldRepo && name == "contains" {
				result = "contains value is " + value
			}
		})
		return result
	}

	autogold.Want("VisitPredicate visits predicates",
		"contains value is file:foo").
		Equal(t, test("repo:contains(file:foo)"))
}
