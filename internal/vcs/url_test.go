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
		wantURL *URL
		wantStr string // expected result of reserializing the URL; empty means same as "in".
	}

	newParseURLTest := func(in string, format urlFormat, scheme, user, host, path, str, rawquery string) *parseURLTest {
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
			wantURL: &URL{
				format: format,
				URL: url.URL{
					Scheme:   scheme,
					Host:     host,
					Path:     path,
					User:     userinfo,
					RawQuery: rawquery,
				}},
			wantStr: str,
		}
	}

	// https://www.kernel.org/pub/software/scm/git/docs/git-clone.html
	tests := []*parseURLTest{
		newParseURLTest(
			"user@host.xz:path/to/repo.git/",
			formatRsync, "", "user", "host.xz", "path/to/repo.git/",
			"user@host.xz:path/to/repo.git/", "",
		),
		newParseURLTest(
			"host.xz:path/to/repo.git/",
			formatRsync, "", "", "host.xz", "path/to/repo.git/",
			"host.xz:path/to/repo.git/", "",
		),
		newParseURLTest(
			"host.xz:/path/to/repo.git/",
			formatRsync, "", "", "host.xz", "/path/to/repo.git/",
			"host.xz:/path/to/repo.git/", "",
		),
		newParseURLTest(
			"host.xz:path/to/repo-with_specials.git/",
			formatRsync, "", "", "host.xz", "path/to/repo-with_specials.git/",
			"host.xz:path/to/repo-with_specials.git/", "",
		),
		newParseURLTest(
			"git://host.xz/path/to/repo.git/",
			formatStdlib, "git", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"git://host.xz:1234/path/to/repo.git/",
			formatStdlib, "git", "", "host.xz:1234", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"http://host.xz/path/to/repo.git/",
			formatStdlib, "http", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"http://host.xz:1234/path/to/repo.git/",
			formatStdlib, "http", "", "host.xz:1234", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"https://host.xz/path/to/repo.git/",
			formatStdlib, "https", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"https://host.xz:1234/path/to/repo.git/",
			formatStdlib, "https", "", "host.xz:1234", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"ftp://host.xz/path/to/repo.git/",
			formatStdlib, "ftp", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"ftp://host.xz:1234/path/to/repo.git/",
			formatStdlib, "ftp", "", "host.xz:1234", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"ftps://host.xz/path/to/repo.git/",
			formatStdlib, "ftps", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"ftps://host.xz:1234/path/to/repo.git/",
			formatStdlib, "ftps", "", "host.xz:1234", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"rsync://host.xz/path/to/repo.git/",
			formatStdlib, "rsync", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"ssh://user@host.xz:1234/path/to/repo.git/",
			formatStdlib, "ssh", "user", "host.xz:1234", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"ssh://host.xz:1234/path/to/repo.git/",
			formatStdlib, "ssh", "", "host.xz:1234", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"ssh://host.xz/path/to/repo.git/",
			formatStdlib, "ssh", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"git+ssh://host.xz/path/to/repo.git/",
			formatStdlib, "git+ssh", "", "host.xz", "/path/to/repo.git/",
			"", "",
		),
		// Tests with query strings
		newParseURLTest(
			"https://host.xz/organization/repo.git?ref=",
			formatStdlib, "https", "", "host.xz", "/organization/repo.git",
			"", "ref=",
		),
		newParseURLTest(
			"https://host.xz/organization/repo.git?ref=test",
			formatStdlib, "https", "", "host.xz", "/organization/repo.git",
			"", "ref=test",
		),
		newParseURLTest(
			"https://host.xz/organization/repo.git?ref=feature/test",
			formatStdlib, "https", "", "host.xz", "/organization/repo.git",
			"", "ref=feature/test",
		),
		newParseURLTest(
			"git@host.xz:organization/repo.git?ref=test",
			formatRsync, "", "git", "host.xz", "organization/repo.git",
			"git@host.xz:organization/repo.git?ref=test", "ref=test",
		),
		newParseURLTest(
			"git@host.xz:organization/repo.git?ref=feature/test",
			formatRsync, "", "git", "host.xz", "organization/repo.git",
			"git@host.xz:organization/repo.git?ref=feature/test", "ref=feature/test",
		),
		// Tests with user+password and some with query strings
		newParseURLTest(
			"https://user:password@host.xz/organization/repo.git/",
			formatStdlib, "https", "user:password", "host.xz", "/organization/repo.git/",
			"", "",
		),
		newParseURLTest(
			"https://user:password@host.xz/organization/repo.git?ref=test",
			formatStdlib, "https", "user:password", "host.xz", "/organization/repo.git",
			"", "ref=test",
		),
		newParseURLTest(
			"https://user:password@host.xz/organization/repo.git?ref=feature/test",
			formatStdlib, "https", "user:password", "host.xz", "/organization/repo.git",
			"", "ref=feature/test",
		),
		newParseURLTest(
			"user-1234@host.xz:path/to/repo.git/",
			formatRsync, "", "user-1234", "host.xz", "path/to/repo.git/",
			"user-1234@host.xz:path/to/repo.git/", "",
		),
		newParseURLTest(
			"user@host.xz:path/to/repo@domain.git/",
			formatRsync, "", "user", "host.xz", "path/to/repo@domain.git/",
			"user@host.xz:path/to/repo@domain.git/", "",
		),
		newParseURLTest(
			"sourcegraph@gitolite.company.internal:service-config/runtime@east-cluster/action@test-5524",
			formatRsync, "", "sourcegraph", "gitolite.company.internal", "service-config/runtime@east-cluster/action@test-5524",
			"sourcegraph@gitolite.company.internal:service-config/runtime@east-cluster/action@test-5524", "",
		),

		newParseURLTest(
			"/path/to/repo.git/",
			formatLocal, "file", "", "", "/path/to/repo.git/",
			"file:///path/to/repo.git/", "",
		),
		newParseURLTest(
			"file:///path/to/repo.git/",
			formatStdlib, "file", "", "", "/path/to/repo.git/",
			"", "",
		),
		newParseURLTest(
			"perforce://admin:pa$$word@ssl:192.168.1.100:1666//Sourcegraph/",
			formatStdlib, "perforce", "admin:pa$$word", "ssl:192.168.1.100:1666", "//Sourcegraph/",
			"perforce://admin:pa$$word@ssl:192.168.1.100:1666//Sourcegraph/", "",
		),
		// iap://
		newParseURLTest(
			"iap://admin:pa$$word@gitlab.sgdev.org/test/repo.git",
			formatStdlib, "iap", "admin:pa$$word", "gitlab.sgdev.org", "/test/repo.git",
			"iap://admin:pa$$word@gitlab.sgdev.org/test/repo.git", "",
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

// Test copied and adjusted from TestJoinPath in go1.19
func TestJoinPath(t *testing.T) {
	tests := []struct {
		base string
		elem []string
		out  string
	}{
		{
			base: "https://go.googlesource.com",
			elem: []string{"go"},
			out:  "https://go.googlesource.com/go",
		},
		{
			base: "https://go.googlesource.com/a/b/c",
			elem: []string{"../../../go"},
			out:  "https://go.googlesource.com/go",
		},
		{
			base: "https://go.googlesource.com/",
			elem: []string{"../go"},
			out:  "https://go.googlesource.com/go",
		},
		{
			base: "https://go.googlesource.com",
			elem: []string{"../go"},
			out:  "https://go.googlesource.com/go",
		},
		{
			base: "https://go.googlesource.com",
			elem: []string{"../go", "../../go", "../../../go"},
			out:  "https://go.googlesource.com/go",
		},
		{
			base: "https://go.googlesource.com/../go",
			elem: nil,
			out:  "https://go.googlesource.com/go",
		},
		{
			base: "https://go.googlesource.com/",
			elem: []string{"./go"},
			out:  "https://go.googlesource.com/go",
		},
		{
			base: "https://go.googlesource.com//",
			elem: []string{"/go"},
			out:  "https://go.googlesource.com/go",
		},
		{
			base: "https://go.googlesource.com//",
			elem: []string{"/go", "a", "b", "c"},
			out:  "https://go.googlesource.com/go/a/b/c",
		},
		{
			base: "https://go.googlesource.com",
			elem: []string{"go/"},
			out:  "https://go.googlesource.com/go/",
		},
		{
			base: "https://go.googlesource.com",
			elem: []string{"go//"},
			out:  "https://go.googlesource.com/go/",
		},
		{
			base: "https://go.googlesource.com",
			elem: nil,
			out:  "https://go.googlesource.com/",
		},
		{
			base: "https://go.googlesource.com/",
			elem: nil,
			out:  "https://go.googlesource.com/",
		},
		{
			base: "https://go.googlesource.com/a%2fb",
			elem: []string{"c"},
			out:  "https://go.googlesource.com/a%2fb/c",
		},
		{
			base: "https://go.googlesource.com/a%2fb",
			elem: []string{"c%2fd"},
			out:  "https://go.googlesource.com/a%2fb/c%2fd",
		},
		{
			base: "https://go.googlesource.com/a/b",
			elem: []string{"/go"},
			out:  "https://go.googlesource.com/a/b/go",
		},
		{
			base: "/",
			elem: nil,
			out:  "file:///",
		},
		{
			base: "a",
			elem: nil,
			out:  "file://a",
		},
		{
			base: "a",
			elem: []string{"b"},
			out:  "file://a/b",
		},
		{
			base: "a",
			elem: []string{"../b"},
			out:  "file://b",
		},
		{
			base: "a",
			elem: []string{"../../b"},
			out:  "file://b",
		},
		{
			base: "",
			elem: []string{"a"},
			out:  "file://a",
		},
		{
			base: "",
			elem: []string{"../a"},
			out:  "file://a",
		},
	}

	for _, tt := range tests {
		wantErr := "nil"
		if tt.out == "" {
			wantErr = "non-nil error"
		}
		var out string
		u, err := ParseURL(tt.base)
		if err == nil {
			u = u.JoinPath(tt.elem...)
			out = u.String()
		}
		if out != tt.out || (err == nil) != (tt.out != "") {
			t.Errorf("Parse(%q).JoinPath(%q) = %q, %v, want %q, %v", tt.base, tt.elem, out, err, tt.out, wantErr)
		}
	}
}
