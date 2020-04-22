package comby

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func TestMatchesUnmarshalling(t *testing.T) {
	// If we are not on CI skip the test if comby is not installed.
	if os.Getenv("CI") == "" && !exists() {
		t.Skip("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	files := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello foo")
}
`,
	}

	zipPath, cleanup, err := testutil.TempZipFromFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	cases := []struct {
		args Args
		want string
	}{
		{
			args: Args{
				Input:         ZipPath(zipPath),
				MatchTemplate: "func",
				FilePatterns:  []string{".go"},
				Matcher:       ".go",
			},
			want: "func",
		},
	}

	for _, test := range cases {
		m, _ := Matches(ctx, test.args)
		if err != nil {
			t.Fatal(err)
		}
		got := m[0].Matches[0].Matched
		if got != test.want {
			t.Errorf("got %v, want %v", got, test.want)
			continue
		}
	}
}

func TestMatchesInZip(t *testing.T) {
	// If we are not on CI skip the test if comby is not installed.
	if os.Getenv("CI") == "" && !exists() {
		t.Skip("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	files := map[string]string{

		"README.md": `# Hello World

Hello world example in go`,
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello foo")
}
`,
	}

	zipPath, cleanup, err := testutil.TempZipFromFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	cases := []struct {
		args Args
		want string
	}{
		{
			args: Args{
				Input:           ZipPath(zipPath),
				MatchTemplate:   "func",
				RewriteTemplate: "derp",
				FilePatterns:    []string{".go"},
				Matcher:         ".go",
			},
			want: `{"uri":"main.go","diff":"--- main.go\n+++ main.go\n@@ -2,6 +2,6 @@\n \n import \"fmt\"\n \n-func main() {\n+derp main() {\n \tfmt.Println(\"Hello foo\")\n }"}
`},
	}

	for _, test := range cases {
		b := new(bytes.Buffer)
		w := bufio.NewWriter(b)
		err := PipeTo(ctx, test.args, w)
		if err != nil {
			t.Fatal(err)
		}
		got := b.String()
		if got != test.want {
			t.Errorf("got %v, want %v", got, test.want)
			continue
		}
	}
}
