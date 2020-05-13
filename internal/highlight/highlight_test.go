package highlight

import (
	"html/template"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPreSpansToTable_Simple(t *testing.T) {
	input := `<pre>
<span>package</span>
</pre>

`
	want := `<table><tr><td class="line" data-line="1"></td><td class="code"><div><span>package</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div></div></td></tr></table>`
	got, err := preSpansToTable(input)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("\ngot:\n%s\nwant:\n%s\n", got, want)
	}
}

func TestPreSpansToTable_Complex(t *testing.T) {
	input := `<pre style="background-color:#ffffff;">
<span style="font-weight:bold;color:#a71d5d;">package</span><span style="color:#323232;"> errcode
</span><span style="color:#323232;">
</span><span style="font-weight:bold;color:#a71d5d;">import </span><span style="color:#323232;">(
</span><span style="color:#323232;">	</span><span style="color:#183691;">&quot;net/http&quot;
</span><span style="color:#323232;">	</span><span style="color:#183691;">&quot;github.com/sourcegraph/sourcegraph/internal/api/legacyerr&quot;
</span><span style="color:#323232;">)
</span><span style="color:#323232;">
</span><span style="color:#323232;">
</span></pre>
`

	want := `<table><tr><td class="line" data-line="1"></td><td class="code"><div><span style="font-weight:bold;color:#a71d5d;">package</span><span style="color:#323232;"> errcode
</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="3"></td><td class="code"><div><span style="font-weight:bold;color:#a71d5d;">import </span><span style="color:#323232;">(
</span></div></td></tr><tr><td class="line" data-line="4"></td><td class="code"><div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;net/http&#34;
</span></div></td></tr><tr><td class="line" data-line="5"></td><td class="code"><div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;github.com/sourcegraph/sourcegraph/internal/api/legacyerr&#34;
</span></div></td></tr><tr><td class="line" data-line="6"></td><td class="code"><div><span style="color:#323232;">)
</span></div></td></tr><tr><td class="line" data-line="7"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="8"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="9"></td><td class="code"><div></div></td></tr></table>`
	got, err := preSpansToTable(input)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("\ngot:\n%s\nwant:\n%s\n", got, want)
	}
}

func TestGeneratePlainTable(t *testing.T) {
	input := `line 1
line 2

`
	want := template.HTML(`<table><tr><td class="line" data-line="1"></td><td class="code"><span>line 1</span></td></tr><tr><td class="line" data-line="2"></td><td class="code"><span>line 2</span></td></tr><tr><td class="line" data-line="3"></td><td class="code"><span>
</span></td></tr><tr><td class="line" data-line="4"></td><td class="code"><span>
</span></td></tr></table>`)
	got, err := generatePlainTable(input)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("\ngot:\n%s\nwant:\n%s\n", got, want)
	}
}

func TestGeneratePlainTableSecurity(t *testing.T) {
	input := `<strong>line 1</strong>
<script>alert("line 2")</script>

`
	want := template.HTML(`<table><tr><td class="line" data-line="1"></td><td class="code"><span>&lt;strong&gt;line 1&lt;/strong&gt;</span></td></tr><tr><td class="line" data-line="2"></td><td class="code"><span>&lt;script&gt;alert(&#34;line 2&#34;)&lt;/script&gt;</span></td></tr><tr><td class="line" data-line="3"></td><td class="code"><span>
</span></td></tr><tr><td class="line" data-line="4"></td><td class="code"><span>
</span></td></tr></table>`)
	got, err := generatePlainTable(input)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("\ngot:\n%s\nwant:\n%s\n", got, want)
	}
}

func TestIssue6892(t *testing.T) {
	input := `<pre style="background-color:#1e1e1e;">

<span style="color:#9b9b9b;">import</span>
</pre>`
	want := `<table><tr><td class="line" data-line="1"></td><td class="code"><div><span>
</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#9b9b9b;">import</span></div></td></tr><tr><td class="line" data-line="3"></td><td class="code"><div></div></td></tr></table>`
	got, err := preSpansToTable(input)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("\ngot:\n%s\nwant:\n%s\n", got, want)
	}
}

func TestUnhighlightLongLines_Simple(t *testing.T) {
	input := `<table><tr><td class="line" data-line="1"></td><td class="code"><div><span>under 40 bytes</span><span> spans are kept
</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span>this line is over 40 bytes</span><span> so spans are not kept
</span></div></td></tr></table>`

	want := `<table><tbody><tr><td class="line" data-line="1"></td><td class="code"><div><span>under 40 bytes</span><span> spans are kept
</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span>this line is over 40 bytes so spans are not kept
</span></div></td></tr></tbody></table>`
	got, err := unhighlightLongLines(input, 40)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("\ngot:\n%s\nwant:\n%s\n", got, want)
	}
}

func TestUnhighlightLongLines_Complex(t *testing.T) {
	input := `<table><tr><td class="line" data-line="1"></td><td class="code"><div><span style="font-weight:bold;color:#a71d5d;">package</span><span style="color:#323232;"> spans on short lines like this are kept
</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="3"></td><td class="code"><div><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="font-weight:bold;color:#a71d5d;">spans on uber long lines like this are dropped </span><span style="color:#323232;">spans on uber
</span></div></td></tr><tr><td class="line" data-line="4"></td><td class="code"><div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;net/http&#34;
</span></div></td></tr><tr><td class="line" data-line="5"></td><td class="code"><div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;github.com/sourcegraph/sourcegraph/internal/api/legacyerr&#34;
</span></div></td></tr><tr><td class="line" data-line="6"></td><td class="code"><div><span style="color:#323232;">)
</span></div></td></tr><tr><td class="line" data-line="7"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="8"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="9"></td><td class="code"><div></div></td></tr></table>`

	want := `<table><tbody><tr><td class="line" data-line="1"></td><td class="code"><div><span style="font-weight:bold;color:#a71d5d;">package</span><span style="color:#323232;"> spans on short lines like this are kept
</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="3"></td><td class="code"><div><span>spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber long lines like this are dropped spans on uber
</span></div></td></tr><tr><td class="line" data-line="4"></td><td class="code"><div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;net/http&#34;
</span></div></td></tr><tr><td class="line" data-line="5"></td><td class="code"><div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;github.com/sourcegraph/sourcegraph/internal/api/legacyerr&#34;
</span></div></td></tr><tr><td class="line" data-line="6"></td><td class="code"><div><span style="color:#323232;">)
</span></div></td></tr><tr><td class="line" data-line="7"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="8"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="9"></td><td class="code"><div></div></td></tr></tbody></table>`
	got, err := unhighlightLongLines(input, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("\ngot:\n%s\nwant:\n%s\n", got, want)
	}
}

func TestParseLinesFromHighlight(t *testing.T) {
	input := `<table><tr><td class="line" data-line="1"></td><td class="code"><div><span style="font-weight:bold;color:#a71d5d;">package</span><span style="color:#323232;"> spans on short lines like this are kept
</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="3"></td><td class="code"><div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;net/http&#34;
</span></div></td></tr><tr><td class="line" data-line="4"></td><td class="code"><div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;github.com/sourcegraph/sourcegraph/internal/api/legacyerr&#34;
</span></div></td></tr><tr><td class="line" data-line="5"></td><td class="code"><div><span style="color:#323232;">)
</span></div></td></tr><tr><td class="line" data-line="6"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="7"></td><td class="code"><div><span style="color:#323232;">
</span></div></td></tr><tr><td class="line" data-line="8"></td><td class="code"><div></div></td></tr></table>`

	want := []string{
		`<div><span style="font-weight:bold;color:#a71d5d;">package</span><span style="color:#323232;"> spans on short lines like this are kept
</span></div>`,
		`<div><span style="color:#323232;">
</span></div>`,
		`<div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;net/http&#34;
</span></div>`,
		`<div><span style="color:#323232;">	</span><span style="color:#183691;">&#34;github.com/sourcegraph/sourcegraph/internal/api/legacyerr&#34;
</span></div>`,
		`<div><span style="color:#323232;">)
</span></div>`,
		`<div><span style="color:#323232;">
</span></div>`,
		`<div><span style="color:#323232;">
</span></div>`,
		`<div></div>`}
	have, err := ParseLinesFromHighlight(input)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(have, want); diff != "" {
		t.Fatal(diff)
	}
}
