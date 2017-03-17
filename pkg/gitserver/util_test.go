package gitserver

import "testing"

func TestNormalizeRepo(t *testing.T) {
	if normalizeRepo("FooBar.git") != "FooBar" {
		t.Fail()
	}
	if normalizeRepo("gitHub.Com/FooBar.git") != "github.com/foobar" {
		t.Fail()
	}
	if normalizeRepo("myServer.Com/FooBar.git") != "myserver.com/FooBar" {
		t.Fail()
	}
}
