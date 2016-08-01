package tokenizer

import (
	"testing"
)

func TestGo(testing *testing.T) {

	testTokenizer(testing,
		&goTokenizer{},
		[]*test{
			{
				"keywords",
				"package main",
				[]Token{{8, 1, 9, "main"}},
			},
		})
}
