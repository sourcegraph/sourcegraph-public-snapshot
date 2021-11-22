package sqlite

import "testing"

func TestIsLiteralEquality(t *testing.T) {
	type TestCase struct {
		Regex       string
		WantOk      bool
		WantLiteral string
	}

	for _, test := range []TestCase{
		{Regex: `^foo$`, WantLiteral: "foo", WantOk: true},
		{Regex: `^[f]oo$`, WantLiteral: `foo`, WantOk: true},
		{Regex: `^\\$`, WantLiteral: `\`, WantOk: true},
		{Regex: `^\$`, WantOk: false},
		{Regex: `^\($`, WantLiteral: `(`, WantOk: true},
		{Regex: `\\`, WantOk: false},
		{Regex: `\$`, WantOk: false},
		{Regex: `\(`, WantOk: false},
		{Regex: `foo$`, WantOk: false},
		{Regex: `(^foo$|^bar$)`, WantOk: false},
	} {
		gotOk, gotLiteral, err := isLiteralEquality(test.Regex)
		if err != nil {
			t.Fatal(err)
		}
		if gotOk != test.WantOk {
			t.Errorf("isLiteralEquality(%s) returned %t, wanted %t", test.Regex, gotOk, test.WantOk)
		}
		if gotLiteral != test.WantLiteral {
			t.Errorf(
				"isLiteralEquality(%s) returned the literal %s, wanted %s",
				test.Regex,
				gotLiteral,
				test.WantLiteral,
			)
		}
	}
}
