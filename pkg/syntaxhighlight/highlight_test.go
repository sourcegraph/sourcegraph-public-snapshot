package syntaxhighlight_test

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"github.com/sourcegraph/annotate"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/syntaxhighlight"
)

var updateExpected bool

func init() {
	flag.BoolVar(&updateExpected, "update-expected", false, "Updates the expected data")
}

func TestLexers(t *testing.T) {
	t.Skip("flaky")

	files, err := ioutil.ReadDir("testdata/case")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		t.Logf("Processing %s", file.Name())
		processFile(file.Name(), t)
	}
}

func TestUnknownFormat(t *testing.T) {
	source := "do bats eat Cats"
	collector := &syntaxhighlight.TokenCollectorAnnotator{}
	syntaxhighlight.Annotate([]byte(source), &syntaxhighlight.FallbackLexer{}, collector)
	expected := []syntaxhighlight.Token{
		{Text: "do", Type: syntaxhighlight.Keyword, Offset: 0},
		{Text: " ", Type: syntaxhighlight.Whitespace, Offset: 2},
		{Text: "bats", Type: syntaxhighlight.Name_Other, Offset: 3},
		{Text: " ", Type: syntaxhighlight.Whitespace, Offset: 7},
		{Text: "eat", Type: syntaxhighlight.Name_Other, Offset: 8},
		{Text: " ", Type: syntaxhighlight.Whitespace, Offset: 11},
		{Text: "Cats", Type: syntaxhighlight.Keyword_Type, Offset: 12},
	}

	if len(collector.Tokens) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d", len(expected), len(collector.Tokens))
	}

	diff := findTheDifference(collector.Tokens, expected)
	if diff != -1 {
		t.Fatalf("There is a difference at position %d. Expected %s but got %s",
			diff,
			expected[diff],
			collector.Tokens[diff])
	}
}

// compares tokens definitions from JSON file, the HTML-formatted tokens
func processFile(name string, t *testing.T) {
	source, err := ioutil.ReadFile(filepath.Join("testdata/case", name))
	if err != nil {
		t.Fatalf("Failed to read case %s: %s", name, err)
	}
	// make sure that line feeds are normalized
	re := regexp.MustCompile("\r\n|\r")
	source = re.ReplaceAll(source, []byte{'\n'})
	collector := &syntaxhighlight.TokenCollectorAnnotator{}
	lexer := syntaxhighlight.NewLexerByExtension(filepath.Ext(name))
	if lexer == nil {
		lexer = &syntaxhighlight.FallbackLexer{}
	}
	syntaxhighlight.Annotate(source, lexer, collector)
	tokens := collector.Tokens
	if updateExpected {
		b, err := json.MarshalIndent(tokens, "", "  ")
		if err != nil {
			t.Fatalf("could not marshal tokens for %s: %s", name, err)
		}
		err = ioutil.WriteFile(filepath.Join("testdata/expected/json", name+".json"), b, 0644)
		if err != nil {
			t.Fatalf("could not update expected data for %s: %s", name, err)
		}
	}
	expectedTokensData, err := ioutil.ReadFile(filepath.Join("testdata/expected/json", name+".json"))
	if err != nil {
		t.Fatalf("Failed to read expected tokens data %s: %s", name, err)
	}
	var expectedTokens []syntaxhighlight.Token
	if err := json.Unmarshal(expectedTokensData, &expectedTokens); err != nil {
		t.Fatalf("Failed to unmarshall expected tokens data %s: %s", name, err)
	}

	if len(tokens) != len(expectedTokens) {
		actualTokensData, _ := json.MarshalIndent(tokens, "", "  ")
		showDiff(string(expectedTokensData), string(actualTokensData), t)
		t.Fatalf("Expected %d tokens, got %d", len(expectedTokens), len(tokens))
	}

	diff := findTheDifference(tokens, expectedTokens)
	if diff != -1 {
		t.Fatalf("There is a difference at position %d. Expected %s but got %s",
			diff,
			expectedTokens[diff],
			tokens[diff])
	}

	annotations, _ := syntaxhighlight.Annotate(source, lexer, syntaxhighlight.NewHTMLAnnotator(syntaxhighlight.DefaultHTMLConfig))
	annotated, _ := annotate.Annotate(source, annotations, template.HTMLEscape)

	actual := strings.TrimSpace(string(annotated))
	if updateExpected {
		err = ioutil.WriteFile(filepath.Join("testdata/expected/html", name+".html"), []byte(actual), 0644)
		if err != nil {
			t.Fatalf("could not update expected data for %s: %s", name, err)
		}
	}

	expectedb, err := ioutil.ReadFile(filepath.Join("testdata/expected/html", name+".html"))
	if err != nil {
		t.Fatalf("Failed to read expected HTML file %s: %s", name, err)
	}
	// make sure that line feeds are normalized
	expected := strings.TrimSpace(string(re.ReplaceAll(expectedb, []byte{'\n'})))

	if actual != expected {
		showDiff(expected, actual, t)
		t.Fatalf("HTML representation does not match")
	}

}

func findTheDifference(a []syntaxhighlight.Token, b []syntaxhighlight.Token) int {
	for i, t := range a {
		if !tokenEquals(t, b[i]) {
			return i
		}
	}
	return -1
}

func tokenEquals(a syntaxhighlight.Token, b syntaxhighlight.Token) bool {
	if a.Text != b.Text {
		return false
	}
	if a.Type.Name != b.Type.Name {
		return false
	}
	if a.Offset != b.Offset {
		return false
	}
	return true
}

func showDiff(expected string, actual string, t *testing.T) {
	f1, err := ioutil.TempFile("", "diff")
	if err != nil {
		t.Fatalf("Unable to create temporary file: %s", err)
	}
	defer os.RemoveAll(f1.Name())
	f2, err := ioutil.TempFile("", "diff")
	if err != nil {
		t.Fatalf("Unable to create temporary file: %s", err)
	}
	defer os.RemoveAll(f2.Name())
	err = ioutil.WriteFile(f1.Name(), []byte(expected), 0600)
	if err != nil {
		t.Fatalf("Unable to create temporary file: %s", err)
	}
	err = ioutil.WriteFile(f2.Name(), []byte(actual), 0600)
	if err != nil {
		t.Fatalf("Unable to create temporary file: %s", err)
	}
	cmd := exec.Command("diff", "-u", f1.Name(), f2.Name())
	t.Log(cmd.Args)
	diff, _ := cmd.CombinedOutput()
	t.Log(string(diff))
}
