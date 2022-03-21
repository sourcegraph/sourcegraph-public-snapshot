package query

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestToZoektQuery(t *testing.T) {
	test := func(input string) string {
		q, err := ParseLiteral(input)
		if err != nil {
			return err.Error()
		}
		b, err := ToBasicQuery(q)
		if err != nil {
			return err.Error()
		}
		zoektQuery, err := b.ToZoektQuery(false, true, true)
		if err != nil {
			return err.Error()
		}
		return zoektQuery.String()
	}

	autogold.Want("basic string",
		`file_substr:"a"`).
		Equal(t, test(`a`))

	autogold.Want("basic and-expression",
		`(or (and file_substr:"a" file_substr:"b" (not file_substr:"c")) file_substr:"d")`).
		Equal(t, test(`a and b and not c or d`))
}
