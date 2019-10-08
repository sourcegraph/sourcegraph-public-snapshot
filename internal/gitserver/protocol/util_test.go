package protocol

import "testing"

func TestNormalizeRepo(t *testing.T) {
	if NormalizeRepo("FooBar.git") != "FooBar" {
		t.Fail()
	}
	if NormalizeRepo("gitHub.Com/FooBar.git") != "github.com/foobar" {
		t.Fail()
	}
	if NormalizeRepo("myServer.Com/FooBar.git") != "myserver.com/FooBar" {
		t.Fail()
	}
	if NormalizeRepo("myServer.Com/FooBar/.git") != "myserver.com/FooBar" {
		t.Fail()
	}
}
