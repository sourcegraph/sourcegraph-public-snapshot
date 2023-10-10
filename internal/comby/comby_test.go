package comby

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestMatchesUnmarshalling(t *testing.T) {
	// If we are not on CI skip the test if comby is not installed.
	if os.Getenv("CI") == "" && !Exists() {
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

	zipPath := tempZipFromFiles(t, files)

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
		m, err := Matches(ctx, logtest.Scoped(t), test.args)
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
	if os.Getenv("CI") == "" && !Exists() {
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

	zipPath := tempZipFromFiles(t, files)

	cases := []struct {
		args Args
		want string
	}{
		{
			args: Args{
				Input:           ZipPath(zipPath),
				MatchTemplate:   "func",
				RewriteTemplate: "derp",
				ResultKind:      Diff,
				FilePatterns:    []string{".go"},
				Matcher:         ".go",
			},
			want: `{"uri":"main.go","diff":"--- main.go\n+++ main.go\n@@ -2,6 +2,6 @@\n \n import \"fmt\"\n \n-func main() {\n+derp main() {\n \tfmt.Println(\"Hello foo\")\n }"}
`},
	}

	for _, test := range cases {
		var b bytes.Buffer
		err := runWithoutPipes(ctx, test.args, &b)
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

func Test_stdin(t *testing.T) {
	// If we are not on CI skip the test if comby is not installed.
	if os.Getenv("CI") == "" && !Exists() {
		t.Skip("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
	}

	test := func(args Args) string {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var b bytes.Buffer
		err := runWithoutPipes(ctx, args, &b)
		if err != nil {
			t.Fatal(err)
		}

		return b.String()
	}

	autogold.Expect(`{"uri":null,"diff":"--- /dev/null\n+++ /dev/null\n@@ -1,1 +1,1 @@\n-yes\n+no"}
`).
		Equal(t, test(Args{
			Input:           FileContent("yes\n"),
			MatchTemplate:   "yes",
			RewriteTemplate: "no",
			ResultKind:      Diff,
			FilePatterns:    []string{".go"},
			Matcher:         ".go",
		}))
}

func TestReplacements(t *testing.T) {
	// If we are not on CI skip the test if comby is not installed.
	if os.Getenv("CI") == "" && !Exists() {
		t.Skip("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	files := map[string]string{
		"main.go": `package tuesday`,
	}

	zipPath := tempZipFromFiles(t, files)

	cases := []struct {
		args Args
		want string
	}{
		{
			args: Args{
				Input:           ZipPath(zipPath),
				MatchTemplate:   "tuesday",
				RewriteTemplate: "wednesday",
				ResultKind:      Replacement,
				FilePatterns:    []string{".go"},
				Matcher:         ".go",
			},
			want: "package wednesday",
		},
	}

	for _, test := range cases {
		r, err := Replacements(ctx, logtest.Scoped(t), test.args)
		if err != nil {
			t.Fatal(err)
		}
		got := r[0].Content
		if got != test.want {
			t.Errorf("got %v, want %v", got, test.want)
			continue
		}
	}
}

func tempZipFromFiles(t *testing.T, files map[string]string) string {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	for name, content := range files {
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:   name,
			Method: zip.Store,
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(w, content); err != nil {
			t.Fatal(err)
		}
	}

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "test.zip")
	if err := os.WriteFile(path, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}

	return path
}

func runWithoutPipes(ctx context.Context, args Args, b *bytes.Buffer) (err error) {
	if !Exists() {
		return errors.New("comby is not installed")
	}

	rawArgs := rawArgs(args)
	cmd := exec.CommandContext(ctx, combyPath, rawArgs...)
	// Ensure forked child processes are killed
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if content, ok := args.Input.(FileContent); ok {
		cmd.Stdin = bytes.NewReader(content)
	}
	cmd.Stdout = b
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed with stdout %s", stderrBuf.String())
	}
	return nil
}
