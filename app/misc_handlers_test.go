package app

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/temoto/robotstxt-go"
)

func TestRobotsTxt(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{
			"/",
			true,
		},
		{
			"/careers",
			true,
		},
		{
			"/careers/",
			true,
		},
		{
			"/security/",
			true,
		},
		{
			"/privacy/",
			true,
		},
		{
			"/legal/",
			true,
		},
		{
			"/pricing/",
			true,
		},
		{
			"/blog/announcing-the-sourcegraph-chrome-extension-for/",
			true,
		},
		{
			"/blog/announcing-the-sourcegraph-chrome-extension-for",
			true,
		},
		{
			"/blog",
			true,
		},
		{
			"/blog/",
			true,
		},
		{
			"/github.com/golang/go/-/info/GoPackage/net/http/httptrace/-/ClientTrace",
			true,
		},
		{
			"/sourcegraph/sourcegraph/-/info/GoPackage/github.com/mediocregopher/radix.v2/util/-/Cmder",
			true,
		},
		{
			"/github.com/golang/go@master/-/tree/src/net/http",
			false,
		},
		{
			"/github.com/golang/go@12343554523113413/-/info/GoPackage/net/http/httptrace/-/ClientTrace",
			false,
		},
		{
			"/github.com/golang/go", // hosted on github, repo view
			false,
		},
		{
			"/github.com/golang/go/", // hosted on github, repo view with trailing
			false,
		},
		{
			"/sourcegraph/sourcegraph", // hosted on sourcegraph, repo view
			false,
		},
		{
			"/sourcegraph/sourcegraph/", // hosted on sourcegraph, repo view with trailing
			false,
		},
		{
			"/github.com/shurcooL/htmlg@master/-/blob/htmlg.go", // hosted on github, blob
			false,
		},
		{
			"https://sourcegraph.com/sourcegraph/sourcegraph@master/-/def/GoPackage/sourcegraph.com/sourcegraph/sourcegraph/app/internal/gitserver/-/AddHandlers", //definition view hosted on sourcegraph, should block because actually blob
			false,
		},
		{
			"/github.com/golang/go@12343554523113413/-/def/GoPackage/net/http/httptrace/-/ClientTrace",
			false, //definition view hosted on github, should block because actually blob
		},
	}
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	robotsTxtHelper(writer, true, "https://www.sourcegraph.com/sitemap.xml")
	writer.Flush()
	robots, _ := robotstxt.FromBytes(b.Bytes())
	for _, test := range tests {
		got := robots.TestAgent(test.url, "GoogleBot")
		if got != test.want {
			t.Errorf("%q: got %q, want %q", test.url, got, test.want)
		}
	}
}
