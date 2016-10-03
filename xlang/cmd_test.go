package xlang

import "testing"

func TestURLCmd(t *testing.T) {
	got := guessTrackedErrorURL([]byte(`{"Error":"jsonrpc2: code 0 message: type/object not found at {Line:4 Character:9}","Method":"textDocument/hover","Mode":"go","Params":{"textDocument":{"uri":"git://github.com/gorilla/mux?757bef944d0f21880861c2dd9c871ca543023cba#mux.go"},"position":{"line":4,"character":9}},"RootPath":"git://github.com/gorilla/mux?757bef944d0f21880861c2dd9c871ca543023cba"}`))
	want := "github.com/gorilla/mux@757bef944d0f21880861c2dd9c871ca543023cba/-/blob/mux.go#L5:10"
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
