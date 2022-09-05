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

func TestIsEmptyPattern(t *testing.T) {
	testCases := []struct {
		name    string
		b       Basic
		isEmpty bool
	}{
		{
			name:    "empty basic is empty pattern",
			b:       Basic{},
			isEmpty: true,
		},
		{
			name: "parameter only is empty pattern",
			b: Basic{
				Parameters: []Parameter{{Field: FieldRepo, Value: "sg"}},
			},
			isEmpty: true,
		},
		{
			name: "multiple parameters only is empty pattern",
			b: Basic{
				Parameters: []Parameter{{Field: FieldRepo, Value: "sg"}, {Field: FieldFile, Value: "search"}},
			},
			isEmpty: true,
		},
		{
			name: "negated parameter only is empty pattern",
			b: Basic{
				Parameters: []Parameter{{Field: FieldRepo, Value: "sg", Negated: true}},
			},
			isEmpty: true,
		},
		{
			name: "parameter with pattern is not empty",
			b: Basic{
				Parameters: []Parameter{{Field: FieldRepo, Value: "sg"}, {Field: FieldFile, Value: "search"}},
				Pattern:    Pattern{Value: "test"},
			},
			isEmpty: false,
		},
		{
			name: "pattern only is not empty",
			b: Basic{
				Pattern: Pattern{Value: "test"},
			},
			isEmpty: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.b.IsEmptyPattern()
			if got != tc.isEmpty {
				t.Fail()
			}
		})
	}
}
