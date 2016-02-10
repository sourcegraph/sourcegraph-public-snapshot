package eventsutil

import (
	"testing"
)

func TestReturnOrg(t *testing.T) {
	test := returnOrganization("https://github.com/abc/xyz")
	if test != "abc" {
		t.Errorf("got org %q, want %q", test, "abc")
	}
	test = returnOrganization("http://www.github.com/abc/xyz")
	if test != "abc" {
		t.Errorf("got org %q, want %q", test, "abc")
	}
	test = returnOrganization("github.com/abc/xyz")
	if test != "abc" {
		t.Errorf("got org %q, want %q", test, "abc")
	}
	test = returnOrganization("www.github.com/abc/xyz")
	if test != "abc" {
		t.Errorf("got org %q, want %q", test, "abc")
	}
	test = returnOrganization("bitbucket.com/abc/xyz")
	if test != "" {
		t.Errorf("got org %q, want %q (no org, not github)", test, "")
	}
	test = returnOrganization("github.com/abc")
	if test != "" {
		t.Errorf("got org %q, want %q", test, "abc")
	}
	test = returnOrganization("https://github.com")
	if test != "" {
		t.Errorf("got org %q, want %q (no org given)", test, "")
	}
}
