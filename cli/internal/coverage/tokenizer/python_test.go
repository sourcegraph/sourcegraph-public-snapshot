package tokenizer

import (
	"testing"
)

func TestPython(testing *testing.T) {
	testTokenizer(testing,
		&pythonTokenizer{},
		[]*test{
			{
				"strings",
				"'abc' \"abc\" '''abc''' \"\"\"abc\"\"\" '''abc",
				[]Token{},
			},
			{
				"raw and Unicode strings",
				"r'abc' rU\"abc\" U'''abc''' uR\"\"\"abc\"\"\" UR'''abc",
				[]Token{},
			},
			{
				"complex numbers",
				"2+10j 2+10q",
				[]Token{{10, 1, 12, "q"}},
			},
			{
				"comments",
				"#abc\ndefz#fgh",
				[]Token{{5, 2, 5, "defz"}},
			},
		})
}
