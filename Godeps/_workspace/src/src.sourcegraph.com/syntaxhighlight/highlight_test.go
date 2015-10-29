package syntaxhighlight_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"github.com/sourcegraph/annotate"

	"src.sourcegraph.com/syntaxhighlight"
)

func TestLexers(t *testing.T) {

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

	source := "do bats eat cats"
	collector := &syntaxhighlight.TokenCollectorAnnotator{}
	syntaxhighlight.Annotate([]byte(source), ``, ``, collector)
	if collector.Tokens == nil || len(collector.Tokens) != 1 {
		t.Fatalf("Expected one tokens, got %v", collector.Tokens)
	}
	if collector.Tokens[0].Text != source {
		t.Fatalf("Expected token source to be '%s', got '%s'", source, collector.Tokens[0].Text)
	}

	if collector.Tokens[0].Offset != 0 {
		t.Fatalf("Expected token source offset to be 0, got %d", collector.Tokens[0].Offset)
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
	syntaxhighlight.Annotate(source, filepath.Ext(name), ``, collector)
	tokens := collector.Tokens
	expectedTokensData, err := ioutil.ReadFile(filepath.Join("testdata/expected/json", name+".json"))
	if err != nil {
		t.Fatalf("Failed to read expected tokens data %s: %s", name, err)
	}
	var expectedTokens []syntaxhighlight.Token
	if err := json.Unmarshal(expectedTokensData, &expectedTokens); err != nil {
		t.Fatalf("Failed to unmarshall expected tokens data %s: %s", name, err)
	}

	if len(tokens) != len(expectedTokens) {
		actualTokensData, _ := json.MarshalIndent(tokens, ``, `  `)
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

	expectedb, err := ioutil.ReadFile(filepath.Join("testdata/expected/html", name+".html"))
	if err != nil {
		t.Fatalf("Failed to read expected HTML file %s: %s", name, err)
	}
	// make sure that line feeds are normalized
	expected := strings.TrimSpace(string(re.ReplaceAll(expectedb, []byte{'\n'})))

	annotations, _ := syntaxhighlight.Annotate(source,
		name,
		``,
		syntaxhighlight.NewHTMLAnnotator(syntaxhighlight.DefaultHTMLConfig))
	annotated, _ := annotate.Annotate(source, annotations, template.HTMLEscape)

	actual := strings.TrimSpace(string(annotated))

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
	f1, err := ioutil.TempFile(``, `diff`)
	if err != nil {
		t.Fatalf(`Unable to create temporary file: %s`, err)
	}
	defer os.RemoveAll(f1.Name())
	f2, err := ioutil.TempFile(``, `diff`)
	if err != nil {
		t.Fatalf(`Unable to create temporary file: %s`, err)
	}
	defer os.RemoveAll(f2.Name())
	err = ioutil.WriteFile(f1.Name(), []byte(expected), 0600)
	if err != nil {
		t.Fatalf(`Unable to create temporary file: %s`, err)
	}
	err = ioutil.WriteFile(f2.Name(), []byte(actual), 0600)
	if err != nil {
		t.Fatalf(`Unable to create temporary file: %s`, err)
	}
	cmd := exec.Command(`diff`, `-u`, f1.Name(), f2.Name())
	t.Log(cmd.Args)
	diff, _ := cmd.CombinedOutput()
	t.Log(string(diff))
}
