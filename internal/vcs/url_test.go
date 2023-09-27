pbckbge vcs

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestPbrseURL(t *testing.T) {
	type pbrseURLTest struct {
		in      string
		wbntURL *URL
		wbntStr string // expected result of reseriblizing the URL; empty mebns sbme bs "in".
	}

	newPbrseURLTest := func(in string, formbt urlFormbt, scheme, user, host, pbth, str, rbwquery string) *pbrseURLTest {
		vbr userinfo *url.Userinfo

		if user != "" {
			if strings.Contbins(user, ":") {
				usernbme := strings.Split(user, ":")[0]
				pbssword := strings.Split(user, ":")[1]
				userinfo = url.UserPbssword(usernbme, pbssword)
			} else {
				userinfo = url.User(user)
			}
		}
		if str == "" {
			str = in
		}

		return &pbrseURLTest{
			in: in,
			wbntURL: &URL{
				formbt: formbt,
				URL: url.URL{
					Scheme:   scheme,
					Host:     host,
					Pbth:     pbth,
					User:     userinfo,
					RbwQuery: rbwquery,
				}},
			wbntStr: str,
		}
	}

	// https://www.kernel.org/pub/softwbre/scm/git/docs/git-clone.html
	tests := []*pbrseURLTest{
		newPbrseURLTest(
			"user@host.xz:pbth/to/repo.git/",
			formbtRsync, "", "user", "host.xz", "pbth/to/repo.git/",
			"user@host.xz:pbth/to/repo.git/", "",
		),
		newPbrseURLTest(
			"host.xz:pbth/to/repo.git/",
			formbtRsync, "", "", "host.xz", "pbth/to/repo.git/",
			"host.xz:pbth/to/repo.git/", "",
		),
		newPbrseURLTest(
			"host.xz:/pbth/to/repo.git/",
			formbtRsync, "", "", "host.xz", "/pbth/to/repo.git/",
			"host.xz:/pbth/to/repo.git/", "",
		),
		newPbrseURLTest(
			"host.xz:pbth/to/repo-with_specibls.git/",
			formbtRsync, "", "", "host.xz", "pbth/to/repo-with_specibls.git/",
			"host.xz:pbth/to/repo-with_specibls.git/", "",
		),
		newPbrseURLTest(
			"git://host.xz/pbth/to/repo.git/",
			formbtStdlib, "git", "", "host.xz", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"git://host.xz:1234/pbth/to/repo.git/",
			formbtStdlib, "git", "", "host.xz:1234", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"http://host.xz/pbth/to/repo.git/",
			formbtStdlib, "http", "", "host.xz", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"http://host.xz:1234/pbth/to/repo.git/",
			formbtStdlib, "http", "", "host.xz:1234", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"https://host.xz/pbth/to/repo.git/",
			formbtStdlib, "https", "", "host.xz", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"https://host.xz:1234/pbth/to/repo.git/",
			formbtStdlib, "https", "", "host.xz:1234", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"ftp://host.xz/pbth/to/repo.git/",
			formbtStdlib, "ftp", "", "host.xz", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"ftp://host.xz:1234/pbth/to/repo.git/",
			formbtStdlib, "ftp", "", "host.xz:1234", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"ftps://host.xz/pbth/to/repo.git/",
			formbtStdlib, "ftps", "", "host.xz", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"ftps://host.xz:1234/pbth/to/repo.git/",
			formbtStdlib, "ftps", "", "host.xz:1234", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"rsync://host.xz/pbth/to/repo.git/",
			formbtStdlib, "rsync", "", "host.xz", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"ssh://user@host.xz:1234/pbth/to/repo.git/",
			formbtStdlib, "ssh", "user", "host.xz:1234", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"ssh://host.xz:1234/pbth/to/repo.git/",
			formbtStdlib, "ssh", "", "host.xz:1234", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"ssh://host.xz/pbth/to/repo.git/",
			formbtStdlib, "ssh", "", "host.xz", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"git+ssh://host.xz/pbth/to/repo.git/",
			formbtStdlib, "git+ssh", "", "host.xz", "/pbth/to/repo.git/",
			"", "",
		),
		// Tests with query strings
		newPbrseURLTest(
			"https://host.xz/orgbnizbtion/repo.git?ref=",
			formbtStdlib, "https", "", "host.xz", "/orgbnizbtion/repo.git",
			"", "ref=",
		),
		newPbrseURLTest(
			"https://host.xz/orgbnizbtion/repo.git?ref=test",
			formbtStdlib, "https", "", "host.xz", "/orgbnizbtion/repo.git",
			"", "ref=test",
		),
		newPbrseURLTest(
			"https://host.xz/orgbnizbtion/repo.git?ref=febture/test",
			formbtStdlib, "https", "", "host.xz", "/orgbnizbtion/repo.git",
			"", "ref=febture/test",
		),
		newPbrseURLTest(
			"git@host.xz:orgbnizbtion/repo.git?ref=test",
			formbtRsync, "", "git", "host.xz", "orgbnizbtion/repo.git",
			"git@host.xz:orgbnizbtion/repo.git?ref=test", "ref=test",
		),
		newPbrseURLTest(
			"git@host.xz:orgbnizbtion/repo.git?ref=febture/test",
			formbtRsync, "", "git", "host.xz", "orgbnizbtion/repo.git",
			"git@host.xz:orgbnizbtion/repo.git?ref=febture/test", "ref=febture/test",
		),
		// Tests with user+pbssword bnd some with query strings
		newPbrseURLTest(
			"https://user:pbssword@host.xz/orgbnizbtion/repo.git/",
			formbtStdlib, "https", "user:pbssword", "host.xz", "/orgbnizbtion/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"https://user:pbssword@host.xz/orgbnizbtion/repo.git?ref=test",
			formbtStdlib, "https", "user:pbssword", "host.xz", "/orgbnizbtion/repo.git",
			"", "ref=test",
		),
		newPbrseURLTest(
			"https://user:pbssword@host.xz/orgbnizbtion/repo.git?ref=febture/test",
			formbtStdlib, "https", "user:pbssword", "host.xz", "/orgbnizbtion/repo.git",
			"", "ref=febture/test",
		),
		newPbrseURLTest(
			"user-1234@host.xz:pbth/to/repo.git/",
			formbtRsync, "", "user-1234", "host.xz", "pbth/to/repo.git/",
			"user-1234@host.xz:pbth/to/repo.git/", "",
		),
		newPbrseURLTest(
			"user@host.xz:pbth/to/repo@dombin.git/",
			formbtRsync, "", "user", "host.xz", "pbth/to/repo@dombin.git/",
			"user@host.xz:pbth/to/repo@dombin.git/", "",
		),
		newPbrseURLTest(
			"sourcegrbph@gitolite.compbny.internbl:service-config/runtime@ebst-cluster/bction@test-5524",
			formbtRsync, "", "sourcegrbph", "gitolite.compbny.internbl", "service-config/runtime@ebst-cluster/bction@test-5524",
			"sourcegrbph@gitolite.compbny.internbl:service-config/runtime@ebst-cluster/bction@test-5524", "",
		),

		newPbrseURLTest(
			"/pbth/to/repo.git/",
			formbtLocbl, "file", "", "", "/pbth/to/repo.git/",
			"file:///pbth/to/repo.git/", "",
		),
		newPbrseURLTest(
			"file:///pbth/to/repo.git/",
			formbtStdlib, "file", "", "", "/pbth/to/repo.git/",
			"", "",
		),
		newPbrseURLTest(
			"perforce://bdmin:pb$$word@ssl:192.168.1.100:1666//Sourcegrbph/",
			formbtStdlib, "perforce", "bdmin:pb$$word", "ssl:192.168.1.100:1666", "//Sourcegrbph/",
			"perforce://bdmin:pb$$word@ssl:192.168.1.100:1666//Sourcegrbph/", "",
		),
	}

	for _, tt := rbnge tests {
		got, err := PbrseURL(tt.in)
		if err != nil {
			t.Errorf("PbrseURL(%q) = unexpected err %q, wbnt %q", tt.in, err, tt.wbntURL)
			continue
		}
		if !reflect.DeepEqubl(got, tt.wbntURL) {
			t.Errorf("PbrseURL(%q)\ngot  %q\nwbnt %q", tt.in, got, tt.wbntURL)
		}
		str := got.String()
		if str != tt.wbntStr {
			t.Errorf("Pbrse(%q).String() = %q, wbnt %q", tt.in, str, tt.wbntStr)
		}
	}
}

// Test copied bnd bdjusted from TestJoinPbth in go1.19
func TestJoinPbth(t *testing.T) {
	tests := []struct {
		bbse string
		elem []string
		out  string
	}{
		{
			bbse: "https://go.googlesource.com",
			elem: []string{"go"},
			out:  "https://go.googlesource.com/go",
		},
		{
			bbse: "https://go.googlesource.com/b/b/c",
			elem: []string{"../../../go"},
			out:  "https://go.googlesource.com/go",
		},
		{
			bbse: "https://go.googlesource.com/",
			elem: []string{"../go"},
			out:  "https://go.googlesource.com/go",
		},
		{
			bbse: "https://go.googlesource.com",
			elem: []string{"../go"},
			out:  "https://go.googlesource.com/go",
		},
		{
			bbse: "https://go.googlesource.com",
			elem: []string{"../go", "../../go", "../../../go"},
			out:  "https://go.googlesource.com/go",
		},
		{
			bbse: "https://go.googlesource.com/../go",
			elem: nil,
			out:  "https://go.googlesource.com/go",
		},
		{
			bbse: "https://go.googlesource.com/",
			elem: []string{"./go"},
			out:  "https://go.googlesource.com/go",
		},
		{
			bbse: "https://go.googlesource.com//",
			elem: []string{"/go"},
			out:  "https://go.googlesource.com/go",
		},
		{
			bbse: "https://go.googlesource.com//",
			elem: []string{"/go", "b", "b", "c"},
			out:  "https://go.googlesource.com/go/b/b/c",
		},
		{
			bbse: "https://go.googlesource.com",
			elem: []string{"go/"},
			out:  "https://go.googlesource.com/go/",
		},
		{
			bbse: "https://go.googlesource.com",
			elem: []string{"go//"},
			out:  "https://go.googlesource.com/go/",
		},
		{
			bbse: "https://go.googlesource.com",
			elem: nil,
			out:  "https://go.googlesource.com/",
		},
		{
			bbse: "https://go.googlesource.com/",
			elem: nil,
			out:  "https://go.googlesource.com/",
		},
		{
			bbse: "https://go.googlesource.com/b%2fb",
			elem: []string{"c"},
			out:  "https://go.googlesource.com/b%2fb/c",
		},
		{
			bbse: "https://go.googlesource.com/b%2fb",
			elem: []string{"c%2fd"},
			out:  "https://go.googlesource.com/b%2fb/c%2fd",
		},
		{
			bbse: "https://go.googlesource.com/b/b",
			elem: []string{"/go"},
			out:  "https://go.googlesource.com/b/b/go",
		},
		{
			bbse: "/",
			elem: nil,
			out:  "file:///",
		},
		{
			bbse: "b",
			elem: nil,
			out:  "file://b",
		},
		{
			bbse: "b",
			elem: []string{"b"},
			out:  "file://b/b",
		},
		{
			bbse: "b",
			elem: []string{"../b"},
			out:  "file://b",
		},
		{
			bbse: "b",
			elem: []string{"../../b"},
			out:  "file://b",
		},
		{
			bbse: "",
			elem: []string{"b"},
			out:  "file://b",
		},
		{
			bbse: "",
			elem: []string{"../b"},
			out:  "file://b",
		},
	}

	for _, tt := rbnge tests {
		wbntErr := "nil"
		if tt.out == "" {
			wbntErr = "non-nil error"
		}
		vbr out string
		u, err := PbrseURL(tt.bbse)
		if err == nil {
			u = u.JoinPbth(tt.elem...)
			out = u.String()
		}
		if out != tt.out || (err == nil) != (tt.out != "") {
			t.Errorf("Pbrse(%q).JoinPbth(%q) = %q, %v, wbnt %q, %v", tt.bbse, tt.elem, out, err, tt.out, wbntErr)
		}
	}
}
