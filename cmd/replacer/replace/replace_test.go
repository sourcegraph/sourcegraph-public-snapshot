package replace_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/replacer/protocol"
	"github.com/sourcegraph/sourcegraph/cmd/replacer/replace"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func TestReplace(t *testing.T) {
	// If we are not on CI skip the test if comby is not installed.
	if _, err := exec.LookPath("comby"); os.Getenv("CI") == "" && err != nil {
		t.Skip("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
	}

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

	cases := []struct {
		arg  protocol.RewriteSpecification
		want string
	}{
		{protocol.RewriteSpecification{
			MatchTemplate:   "func",
			RewriteTemplate: "derp",
			FileExtension:   ".go",
		}, `
{"uri":"main.go","diff":"--- main.go\n+++ main.go\n@@ -2,6 +2,6 @@\n \n import \"fmt\"\n \n-func main() {\n+derp main() {\n \tfmt.Println(\"Hello foo\")\n }"}
`},
	}

	store, cleanup, err := testutil.NewStore(files)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	ts := httptest.NewServer(&replace.Service{Store: store})
	defer ts.Close()

	for _, test := range cases {
		req := protocol.Request{
			Repo:                 "foo",
			URL:                  "u",
			Commit:               "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			RewriteSpecification: test.arg,
			FetchTimeout:         "2000ms",
		}
		got, err := doReplace(ts.URL, &req)
		if err != nil {
			t.Errorf("%v failed: %s", test.arg, err)
			continue
		}

		// We have an extra newline to make expected readable
		if len(test.want) > 0 {
			test.want = test.want[1:]
		}

		if got != test.want {
			d, err := testutil.Diff(test.want, got)
			if err != nil {
				t.Fatal(err)
			}
			t.Errorf("%v unexpected response:\n%s", test.arg, d)
			continue
		}

	}
}

func TestReplace_badrequest(t *testing.T) {
	cases := []protocol.Request{
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			// No MatchTemplate
		},
	}

	store, cleanup, err := testutil.NewStore(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	ts := httptest.NewServer(&replace.Service{Store: store})
	defer ts.Close()

	for _, p := range cases {
		_, err := doReplace(ts.URL, &p)
		if err == nil {
			t.Fatalf("%v expected to fail", p)
		}
		if !strings.HasPrefix(err.Error(), "non-200 response: code=400 ") {
			t.Fatalf("%v expected to have HTTP 400 response. Got %s", p, err)
		}
	}
}

func doReplace(u string, p *protocol.Request) (string, error) {
	form := url.Values{
		"Repo":            []string{string(p.Repo)},
		"URL":             []string{string(p.URL)},
		"Commit":          []string{string(p.Commit)},
		"FetchTimeout":    []string{p.FetchTimeout},
		"MatchTemplate":   []string{p.RewriteSpecification.MatchTemplate},
		"RewriteTemplate": []string{p.RewriteSpecification.RewriteTemplate},
		"FileExtension":   []string{p.RewriteSpecification.FileExtension},
	}
	resp, err := http.PostForm(u, form)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("non-200 response: code=%d body=%s", resp.StatusCode, string(body))
	}

	return string(body), err
}
