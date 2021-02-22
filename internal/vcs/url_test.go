package vcs

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestParseURL(t *testing.T) {
	type parseURLTest struct {
		in      string
		wantURL *url.URL
		wantStr string // expected result of reserializing the URL; empty means same as "in".
	}

	newParseURLTest := func(in, scheme, user, host, path, str, rawquery string) *parseURLTest {
		var userinfo *url.Userinfo

		if user != "" {
			if strings.Contains(user, ":") {
				username := strings.Split(user, ":")[0]
				password := strings.Split(user, ":")[1]
				userinfo = url.UserPassword(username, password)
			} else {
				userinfo = url.User(user)
			}
		}
		if str == "" {
			str = in
		}

		return &parseURLTest{
			in: in,
			wantURL: &url.URL{
				Scheme:   scheme,
				Host:     host,
				Path:     path,
				User:     userinfo,
				RawQuery: rawquery,
			},
			wantStr: str,
		}
	}

	// https://www.kernel.org/pub/software/scm/git/docs/git-clone.html
	tests := []*parseURLTest{
		newParseURLTest(
			"user@host.xz:path/to/repo.git/",
			"ssh", "user", "host.xz", "path/to/repo.git/",
			"ssh://user@host.xz/path/to/repo.git/", "",
		),
		newParseURLTest(
			"host.xz:path/to/repo.git/",
			"ssh", "", "host.xz", "path/to/repo.git/",
			"ssh://host.xz/path/to/repo.git/", "",
		),
		newParseURLTest(
			"host.xz:/path/to/repo.git/",
			"ssh", "", "host.xz", "/path/to/repo.git/",
			"ssh://host.xz/path/to/repo.git/", "",
		),
		newParseURLTest(
			"host.xz:path/to/repo-with_specials.git/",
			"ssh", "", "host.xz", "path/to/repo-with_specials.git/",
			"ssh://host.xz/path/to/repo-with_specials.git/", "",
		),
		newParseURLTest(
			"git://host.xz/path/to/repo.git/",
			"git", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"git://host.xz:1234/path/to/repo.git/",
			"git", "", "host.xz:1234", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"http://host.xz/path/to/repo.git/",
			"http", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"http://host.xz:1234/path/to/repo.git/",
			"http", "", "host.xz:1234", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"https://host.xz/path/to/repo.git/",
			"https", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"https://host.xz:1234/path/to/repo.git/",
			"https", "", "host.xz:1234", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"ftp://host.xz/path/to/repo.git/",
			"ftp", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"ftp://host.xz:1234/path/to/repo.git/",
			"ftp", "", "host.xz:1234", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"ftps://host.xz/path/to/repo.git/",
			"ftps", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"ftps://host.xz:1234/path/to/repo.git/",
			"ftps", "", "host.xz:1234", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"rsync://host.xz/path/to/repo.git/",
			"rsync", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"ssh://user@host.xz:1234/path/to/repo.git/",
			"ssh", "user", "host.xz:1234", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"ssh://host.xz:1234/path/to/repo.git/",
			"ssh", "", "host.xz:1234", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"ssh://host.xz/path/to/repo.git/",
			"ssh", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"git+ssh://host.xz/path/to/repo.git/",
			"git+ssh", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		// Tests with query strings
		newParseURLTest(
			"https://host.xz/organization/repo.git?ref=",
			"https", "", "host.xz", "/organization/repo.git",
			"", "ref=",
		),
		newParseURLTest(
			"https://host.xz/organization/repo.git?ref=test",
			"https", "", "host.xz", "/organization/repo.git",
			"", "ref=test",
		),
		newParseURLTest(
			"https://host.xz/organization/repo.git?ref=feature/test",
			"https", "", "host.xz", "/organization/repo.git",
			"", "ref=feature/test",
		),
		newParseURLTest(
			"git@host.xz:organization/repo.git?ref=test",
			"ssh", "git", "host.xz", "organization/repo.git",
			"ssh://git@host.xz/organization/repo.git?ref=test", "ref=test",
		),
		newParseURLTest(
			"git@host.xz:organization/repo.git?ref=feature/test",
			"ssh", "git", "host.xz", "organization/repo.git",
			"ssh://git@host.xz/organization/repo.git?ref=feature/test", "ref=feature/test",
		),
		// Tests with user+password and some with query strings
		newParseURLTest(
			"https://user:password@host.xz/organization/repo.git/",
			"https", "user:password", "host.xz", "/organization/repo.git/",
			"", "",
		),
		newParseURLTest(
			"https://user:password@host.xz/organization/repo.git?ref=test",
			"https", "user:password", "host.xz", "/organization/repo.git",
			"", "ref=test",
		),
		newParseURLTest(
			"https://user:password@host.xz/organization/repo.git?ref=feature/test",
			"https", "user:password", "host.xz", "/organization/repo.git",
			"", "ref=feature/test",
		),
		newParseURLTest(
			"user-1234@host.xz:path/to/repo.git/",
			"ssh", "user-1234", "host.xz", "path/to/repo.git/",
			"ssh://user-1234@host.xz/path/to/repo.git/", "",
		),
		newParseURLTest(
			"/path/to/repo.git/",
			"file", "", "", "/path/to/repo.git/",
			"file:///path/to/repo.git/", "",
		),
		newParseURLTest(
			"file:///path/to/repo.git/",
			"file", "", "", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"perforce://admin:pa$$word@ssl:192.168.1.100:1666//Sourcegraph/",
			"perforce", "admin:pa$$word", "ssl:192.168.1.100:1666", "//Sourcegraph/",
			"perforce://admin:pa$$word@ssl:192.168.1.100:1666//Sourcegraph/", "",
		),
	}

	for _, tt := range tests {
		got, err := ParseURL(tt.in)
		if err != nil {
			t.Errorf("ParseURL(%q) = unexpected err %q, want %q", tt.in, err, tt.wantURL)
			continue
		}
		if !reflect.DeepEqual(got, tt.wantURL) {
			t.Errorf("ParseURL(%q)\ngot  %q\nwant %q", tt.in, got, tt.wantURL)
		}
		str := got.String()
		if str != tt.wantStr {
			t.Errorf("Parse(%q).String() = %q, want %q", tt.in, str, tt.wantStr)
		}
	}
}
