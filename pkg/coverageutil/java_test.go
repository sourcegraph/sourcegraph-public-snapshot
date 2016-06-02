package coverageutil

import (
	"testing"
)

func TestJava(testing *testing.T) {

	testTokenizer(testing,
		&javaTokenizer{},
		[]*test{
			{
				"keywords and UTF-8",
				"package /* © */ main; class A {}",
				[]Token{{29, "A"}},
			},
			{
				"packages and imports",
				"package foo.bar.baz.qux; import foo.bar.*; import static X.Y.Z; import org.apache.commons.X;",
				[]Token{},
			},
			{
				"numeric literals",
				"123 123l 123L 12_3 12_3l 0xB 0XA 0b0000 0B1L -2.5E2f",
				[]Token{},
			},
			{
				"escape sequences",
				"\"\\0\" '\\00' '\\'' abc",
				[]Token{{16, "abc"}},
			},
		})
}
