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
	robotsTxtHelper(writer, true)
	writer.Flush()
	robots, _ := robotstxt.FromBytes(b.Bytes())
	for _, test := range tests {
		got := robots.TestAgent(test.url, "GoogleBot")
		if got != test.want {
			t.Errorf("%q: got %t, want %t", test.url, got, test.want)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_269(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
