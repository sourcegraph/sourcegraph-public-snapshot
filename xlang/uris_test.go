package xlang

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
)

func TestAbsRelWorkspaceURI(t *testing.T) {
	root, err := uri.Parse("git://a.com/b?rev=c")
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]string{
		"git://a.com/b?rev=c":     "file:///",
		"git://a.com/b?rev=c#f":   "file:///f",
		"git://a.com/b?rev=c#d/f": "file:///d/f",
	}
	for wantClientPath, wantServerPath := range tests {
		if serverPath, err := relWorkspaceURI(*root, wantClientPath); err != nil {
			t.Fatal(err)
		} else if serverPath.String() != wantServerPath {
			t.Errorf("relWorkspaceURI(%s): got server path %q, want %q", wantClientPath, serverPath, wantServerPath)
		}
		if clientPath, err := absWorkspaceURI(*root, wantServerPath); err != nil {
			t.Fatal(err)
		} else if clientPath.String() != wantClientPath {
			t.Errorf("absWorkspaceURI(%s): got client path %q, want %q", wantServerPath, clientPath, wantClientPath)
		}
	}
	// We disallow paths that are not part of the repo from initialize
	bad := []string{
		"git://a.com/bad?c#d/f",
		"git://a.com/?c#d/f",
		"git://a.com/b/../a?c#d/f",
		"git://a.com/b/..?c#d/f",
		"git://a.com?c#/d",
		"git://a.com?c#../d",
		"git://a.com?c#..",
		"git://a.com?c#.",
		"git://a.com?c#./..",
		"git://a.com?c#a/../b",
		"git://a.com?c#a//b",
	}
	for _, p := range bad {
		_, err := relWorkspaceURI(*root, p)
		if err == nil {
			t.Errorf("relWorkspaceURI(%v) should fail", p)
		}
	}
}
