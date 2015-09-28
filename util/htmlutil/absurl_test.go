package htmlutil

import (
	"net/url"
	"testing"
)

func TestMakeUURLsAbsolute(t *testing.T) {
	absURL, err := url.Parse("https://sourcegraph.com/blog/")
	if err != nil {
		t.Error(err)
	}
	out, err := MakeURLsAbsolute(`abc<a href="foo">def</a>ghi<img src="/bar">jkl`, absURL)
	if err != nil {
		t.Error(err)
	}
	if got, want := out, `abc<a href="https://sourcegraph.com/blog/foo">def</a>ghi<img src="https://sourcegraph.com/bar">jkl`; got != want {
		t.Errorf("makeURLsAbsolute: got %q, want %q", got, want)
	}
}
