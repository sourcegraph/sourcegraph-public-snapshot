package sourcecode_test

import (
	"html/template"
	"testing"

	"src.sourcegraph.com/sourcegraph/sourcecode"
	"src.sourcegraph.com/sourcegraph/ui/uiconf"
)

func TestCustomRegexStyle(t *testing.T) {
	// Sample input to process is "(Change).URL html/template.URL", with the goal of having
	// a regexp that makes the two "URL" words bold.
	//
	// This regexp contains two non-capturing groups (that get skipped)
	// and two capturing groups (that make the matched text bold).
	// It is kept simple to be more readable as an example.
	var exampleRegexp = `^(?:\(Change\))\.(URL)(?: html/template)\.(URL)$`

	err := uiconf.Flags.DefQualifiedNameBold.UnmarshalFlag(exampleRegexp)
	if err != nil {
		t.Fatal(err)
	}
	defer uiconf.Flags.DefQualifiedNameBold.UnmarshalFlag("")

	var (
		in   = template.HTML(`(Change).<wbr><span class="name">URL</span>  html/<wbr>template.<wbr>URL`)
		want = template.HTML(`(Change).<span class="def-qualified-name-bold">URL</span> html/template.<span class="def-qualified-name-bold">URL</span>`)
	)

	if got := sourcecode.OverrideStyleViaRegexpFlags(in); got != want {
		t.Fatalf("got != want:\n%q\n%q", got, want)
	}
}
