package squirrel

import "testing"

func TestHover(t *testing.T) {
	golang := `
// not a comment line

// comment line 1
// comment line 2
func f() {
	// c1
	// c2
	// c3
	if y := x; y != nil {
		fmt.Println(y)
	}
}
`

	tests := []struct {
		path     string
		contents string
		want     string
	}{
		{"test.go", golang, "comment line 1\ncomment line 2\n"},
		{"test.go", golang, "c1\nc2\nc3\n"},
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

			if got != test.want {
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
			t.Fatalf("did not find comment %q. All comments: %v", test.want, comments)
		}
	}
}

func fatalIfError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
