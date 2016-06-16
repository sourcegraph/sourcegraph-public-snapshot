package search

import (
	"reflect"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"testing"
)

func TestBagOfWords(t *testing.T) {
	type testCase struct {
		def      graph.Def
		expected map[string]int
	}
	cases := []testCase{
		{
			def: graph.Def{
				DefKey: graph.DefKey{
					Path: "testOne/testNumberOne",
					Repo: "github.com/firstCase/caseNumOne",
				},
			},
			expected: map[string]int{
				"Number":        11,
				"One":           12,
				"testOne":       2,
				"testNumberOne": 32,
				"test":          12,
				"":              17,
				"firstCase":     1,
				"caseNumOne":    3,
			},
		},
		{
			def: graph.Def{
				DefKey: graph.DefKey{
					Path: "test_two/test_number_two",
					Repo: "github.com/secondCase/caseNumTwo",
				},
			},
			expected: map[string]int{
				"test_two":        2,
				"test_number_two": 32,
				"test":            12,
				"number":          11,
				"two":             12,
				"":                17,
				"secondCase":      1,
				"caseNumTwo":      3,
			},
		},
		{
			def: graph.Def{
				DefKey: graph.DefKey{
					Path: "test/three",
					Repo: "github.com/thirdCase/caseNumThree",
				},
			},
			expected: map[string]int{
				"three":        43,
				"test":         3,
				"":             17,
				"thirdCase":    1,
				"caseNumThree": 3,
			},
		},
		{
			def: graph.Def{
				DefKey: graph.DefKey{
					Path: "test/test",
					Repo: "github.com/fourthCase/caseNumFour",
				},
			},
			expected: map[string]int{
				"test":        46,
				"":            17,
				"fourthCase":  1,
				"caseNumFour": 3,
			},
		},
	}

	for i, c := range cases {
		eq := reflect.DeepEqual(BagOfWords(&c.def), c.expected)
		if !eq {
			t.Error("Miscalculated bag of words scoring for test case at index ", i)
		}
	}
}
