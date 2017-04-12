package proxy

import "testing"

func TestShortenMethodName(t *testing.T) {
	cases := map[string]string{
		"workspace/xreferences":       "w/xr",
		"workspace/symbol":            "w/s",
		"textDocument/documentSymbol": "td/ds",
		"textDocument/hover":          "td/h",
		"$/cancelRequest":             "/cr",
	}
	for method, want := range cases {
		got := shortenMethodName(method)
		if got != want {
			t.Errorf("shortenMethodName(%q) == %q != %q", method, got, want)
		}
	}
}
