package query

import (
	"encoding/json"
	"testing"

	"github.com/hexops/autogold/v2"
)

func TestStringHuman(t *testing.T) {
	cases := []string{
		"a b c",
		"this or that",
		"this or that and here or there",
		"not x and y",
		"repo:foo a and b",
		`repo:"quoted\""`,
		`"ab\"cd"`,
		`repo:foo file:bar baz and qux`,
		`/abcd\// patterntype:regexp`,
		"repo:foo file:bar",
		"(repo:foo or repo:bar) or (repo:baz or repo:qux) (a or b)",
		"(repo:foo or repo:bar file:a) or (repo:baz or repo:qux and file:b) a and b",
		"repo:foo (not b) (not c) a",
		"repo:foo a -content:b -content:c",
		"-repo:modspeed -file:pogspeed Arizonan not Phoenicians",
		"r:alias",
		`/bo/u\gros/`,
		`filePath.Clean( AND NOT filepath.Clean(filePath.Join("/",`,
	}

	test := func(input string) string {
		q, _ := ParseStandard(input)
		j, _ := json.MarshalIndent(struct {
			Input  string
			Result string
		}{
			Input:  input,
			Result: StringHuman(q),
		}, "", "  ")
		return string(j)
	}

	for _, c := range cases {
		t.Run("printer", func(t *testing.T) {
			autogold.ExpectFile(t, autogold.Raw(test(c)))
		})
	}
}
