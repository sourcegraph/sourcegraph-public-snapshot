package ui2

import (
	"html/template"
	"testing"
)

func TestPreSpansToTable_Simple(t *testing.T) {
	input := `<pre>
<span>package</span>
</pre>

`
	want := `<table><tr><td>1</td><td><span>package</span></td></tr><tr><td>2</td><td></td></tr></table>`
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
<span style="font-weight:bold;color:#a71d5d;">package</span><span style="color:#323232;"> errcode</span>

<span style="font-weight:bold;color:#a71d5d;">import</span><span style="color:#323232;"> </span><span style="color:#323232;">(</span>
<span style="color:#323232;">	</span><span style="color:#183691;">&quot;</span><span style="color:#183691;">net/http</span><span style="color:#183691;">&quot;</span>

<span style="color:#323232;">	</span><span style="color:#183691;">&quot;</span><span style="color:#183691;">sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr</span><span style="color:#183691;">&quot;</span>
<span style="color:#323232;">)</span>

</pre>

`
	want := `<table><tr><td>1</td><td><span style="font-weight:bold;color:#a71d5d;">package</span><span style="color:#323232;"> errcode</span></td></tr><tr><td>2</td><td></td></tr><tr><td>3</td><td><span style="font-weight:bold;color:#a71d5d;">import</span><span style="color:#323232;"> </span><span style="color:#323232;">(</span></td></tr><tr><td>4</td><td><span style="color:#323232;">	</span><span style="color:#183691;">&#34;</span><span style="color:#183691;">net/http</span><span style="color:#183691;">&#34;</span></td></tr><tr><td>5</td><td></td></tr><tr><td>6</td><td><span style="color:#323232;">	</span><span style="color:#183691;">&#34;</span><span style="color:#183691;">sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr</span><span style="color:#183691;">&#34;</span></td></tr><tr><td>7</td><td><span style="color:#323232;">)</span></td></tr><tr><td>8</td><td></td></tr><tr><td>9</td><td></td></tr></table>`
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
	want := template.HTML(`<table><tr><td>1</td><td><span>line 1</span></td></tr><tr><td>2</td><td><span>line 2</span></td></tr><tr><td>3</td><td><span></span></td></tr><tr><td>4</td><td><span></span></td></tr></table>`)
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
	want := template.HTML(`<table><tr><td>1</td><td><span>&lt;strong&gt;line 1&lt;/strong&gt;</span></td></tr><tr><td>2</td><td><span>&lt;script&gt;alert(&#34;line 2&#34;)&lt;/script&gt;</span></td></tr><tr><td>3</td><td><span></span></td></tr><tr><td>4</td><td><span></span></td></tr></table>`)
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
	want := `<table><tr><td>1</td><td></td></tr><tr><td>2</td><td><span style="color:#9b9b9b;">import</span></td></tr><tr><td>3</td><td></td></tr></table>`
	got, err := preSpansToTable(input)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("\ngot:\n%s\nwant:\n%s\n", got, want)
	}
}
