package coverageutil

import "testing"

func TestJSON(t *testing.T) {
	testTokenizer(t,
		&jsonTokenizer{},
		[]*test{
			{
				"test array literal",
				`[3, "three", true, null]`,
				[]Token{{1, 1, `3`}, {5, 1, `three`}, {13, 1, `true`}, {19, 1, `null`}},
			},
			{
				"test object",
				`{"obj": [{"a": 1}]}`,
				[]Token{{2, 1, `obj`}, {11, 1, `a`}, {15, 1, `1`}},
			},

			{"test line break",
				`{"obj": [{"a":` + "\n" + ` 1}]}`,
				[]Token{{2, 1, `obj`}, {11, 1, `a`}, {16, 2, `1`}}},
		})
}

func TestJSONError(t *testing.T) {
	broken := `{}}`
	tokenizer := jsonTokenizer{}
	tokenizer.Init([]byte(broken))
	if tokenizer.Next() != nil {
		t.Fatalf("Expected an error when trying to parse unbalanced JSON: %s", broken)
	}
	if len(tokenizer.Errors()) == 0 {
		t.Fatalf("Expected a non-zero number of errors returned when trying to parse unbalanced JSON: %s", broken)
	}
}
