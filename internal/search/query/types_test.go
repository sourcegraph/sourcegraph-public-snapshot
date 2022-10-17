package query

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepoHasDescription(t *testing.T) {
	ps := Parameters{
		Parameter{
			Field:      FieldRepo,
			Value:      "has.description(test)",
			Annotation: Annotation{Labels: IsPredicate},
		},
		Parameter{
			Field:      FieldRepo,
			Value:      "has.description(test input)",
			Annotation: Annotation{Labels: IsPredicate},
		},
	}

	want := []string{
		"(?:test)",
		"(?:test).*?(?:input)",
	}

	require.Equal(t, want, ps.RepoHasDescription())
}
