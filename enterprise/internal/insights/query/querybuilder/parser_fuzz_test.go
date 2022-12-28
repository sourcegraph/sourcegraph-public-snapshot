package querybuilder

import (
	"testing"
)

func FuzzTestParseQuery(f *testing.F) {
	testCases := []struct {
		name  string
		query string
		fail  bool
	}{
		{
			"valid query",
			"select:file test",
			false,
		},
		{
			"valid literal query",
			"select:file i++",
			false,
		},
	}
	for _, tc := range testCases {
		f.Add(tc.query)
	}
	f.Fuzz(func(t *testing.T, query string) {
		q, err := ParseQuery(query, "literal")
		t.Log(q)

		if err != nil {
			t.Log(err)
		}
	})
}
