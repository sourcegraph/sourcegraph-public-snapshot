package app

import "testing"

func TestSanitizeFormattedCode(t *testing.T) {
	got := string(sanitizeFormattedCode(`abc<span class="defn-popover fun">def</span>ghi<a href="/.def" class="defn-popover XXX">jkl</a>mno<div>pqr</div>stu`))
	want := `abc<span class="defn-popover fun">def</span>ghi<a href="/.def">jkl</a>mnopqrstu`
	if got != want {
		t.Errorf("sanitizeFormattedCode: got %q, want %q", got, want)
	}
}
