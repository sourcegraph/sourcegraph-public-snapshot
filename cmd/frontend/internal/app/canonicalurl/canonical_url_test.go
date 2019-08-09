package canonicalurl

import (
	"net/url"
	"testing"
)

func TestFromURL(t *testing.T) {
	tests := []struct {
		currentURL *url.URL
		want       *url.URL
	}{
		{&url.URL{RawQuery: "utm_source=3"}, &url.URL{RawQuery: ""}},
		{&url.URL{RawQuery: "foo=3"}, &url.URL{RawQuery: "foo=3"}},
		{&url.URL{RawQuery: "foo=3&utm_source=4"}, &url.URL{RawQuery: "foo=3"}},
	}
	for _, test := range tests {
		curl := FromURL(test.currentURL)
		if *test.want != *curl {
			t.Errorf("%s: want %s, got %s", test.currentURL, test.want, curl)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_257(size int) error {
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
