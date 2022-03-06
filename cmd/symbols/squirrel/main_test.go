package squirrel

import (
	"fmt"
	"strings"
	"testing"
)

func TestHover(t *testing.T) {
	golang := `
// not a comment line

// comment line 1
// comment line 2
func f() {}
`

	_ = golang

	tests := []struct {
		path     string
		contents string
		want     string
	}{
		{"test.go", golang, "comment line 1\ncomment line 2\n"},
	}

	for _, test := range tests {
		payload, err := localCodeIntel(test.path, test.contents)
		fatalIfError(t, err)

		ok := false
		for _, symbol := range payload.Symbols {
			if symbol.Hover == nil {
				continue
			}

			got := *symbol.Hover

			if !strings.Contains(got, test.want) {
				continue
			} else {
				ok = true
				break
			}
		}

		if !ok {
			comments := []string{}
			for _, symbol := range payload.Symbols {
				if symbol.Hover == nil {
					continue
				}
				comments = append(comments, *symbol.Hover)
			}
			fmt.Printf("did not find comment %q. All comments:\n", test.want)
			for _, comment := range comments {
				fmt.Printf("%q\n", comment)
			}
			t.FailNow()
		}
	}
}

func fatalIfError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
