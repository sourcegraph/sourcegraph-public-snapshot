package ui2

import (
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
