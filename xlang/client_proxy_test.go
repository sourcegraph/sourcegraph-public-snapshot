package xlang

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

func TestRewritePaths(t *testing.T) {
	u, err := uri.Parse("git://a.com/b?rev=c")
	if err != nil {
		t.Fatal(err)
	}
	c := clientProxyConn{context: contextID{rootPath: *u}}
	tests := map[string]string{
		"git://a.com/b?rev=c":     "file:///",
		"git://a.com/b?rev=c#f":   "file:///f",
		"git://a.com/b?rev=c#d/f": "file:///d/f",
	}
	for wantClientPath, wantServerPath := range tests {
		if serverPath, err := c.rewritePathFromClient(wantClientPath); err != nil {
			t.Fatal(err)
		} else if serverPath != wantServerPath {
			t.Errorf("rewritePathFromClient(%s): got server path %q, want %q", wantClientPath, serverPath, wantServerPath)
		}
		if clientPath, err := c.rewritePathFromServer(wantServerPath); err != nil {
			t.Fatal(err)
		} else if clientPath != wantClientPath {
			t.Errorf("rewritePathFromServer(%s): got client path %q, want %q", wantServerPath, clientPath, wantClientPath)
		}
	}
	// We disallow paths that are not part of the repo from initialize
	bad := []string{
		"git://a.com/bad?c#d/f",
		"git://a.com/?c#d/f",
		"git://a.com/b/../a?c#d/f",
		"git://a.com/b/..?c#d/f",
	}
	for _, p := range bad {
		_, err := c.rewritePathFromClient(p)
		if err == nil {
			t.Errorf("c.rewritePathFromClient(%v) should fail", p)
		}
	}
}
