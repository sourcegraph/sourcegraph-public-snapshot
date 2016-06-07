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
