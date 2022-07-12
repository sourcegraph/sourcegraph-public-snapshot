package query

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDescription(t *testing.T) {
	ps := Parameters{
		Parameter{
			Field:      FieldRepo,
			Value:      "description(test)",
			Annotation: Annotation{Labels: IsPredicate},
		},
		Parameter{
			Field:      FieldRepo,
			Value:      "description(test input)",
			Annotation: Annotation{Labels: IsPredicate},
		},
	}

	want := []string{
		"(?:test)",
		"(?:test).*?(?:input)",
	}

	require.Equal(t, want, ps.Description())
}
