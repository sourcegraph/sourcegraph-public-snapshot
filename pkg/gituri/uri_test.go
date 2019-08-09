package gituri

import (
	"net/url"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := map[string]url.URL{}
	for uriStr, want := range tests {
		t.Run(strings.Replace(uriStr, "/", "-", -1), func(t *testing.T) {
			uri, err := Parse(uriStr)
			if err != nil {
				t.Fatal(err)
			}
			if uri.URL != want {
				t.Errorf("got %+v, want %+v", uri.URL, want)
			}
		})
	}
}

func TestParse_error(t *testing.T) {
	tests := map[string]string{
		"github.com/foo/bar":   "must be absolute",
		"/github.com/foo/bar":  "must be absolute",
		"//github.com/foo/bar": "must be absolute",
		"%":                    "invalid",
	}
	for uriStr, want := range tests {
		t.Run(strings.Replace(uriStr, "/", "-", -1), func(t *testing.T) {
			uri, err := Parse(uriStr)
			if err == nil {
				t.Fatalf("got nil error, want %q", want)
			}
			if uri != nil {
				t.Error("got non-nil URL, want nil")
			}
			if !strings.Contains(err.Error(), want) {
				t.Errorf("got %q, want it to contain %q", err, want)
			}
		})
	}
}

func TestURI_CloneURL(t *testing.T) {
	want := "https://github.com/foo/bar"
	uriStrs := []string{
		"https://github.com/foo/bar",
		"https://github.com/foo/bar?v",
		"https://github.com/foo/bar?v#f",
		"https://github.com/foo/bar#f",
	}
	for _, uriStr := range uriStrs {
		t.Run(strings.Replace(uriStr, "/", "-", -1), func(t *testing.T) {
			uri, err := Parse(uriStr)
			if err != nil {
				t.Fatal(err)
			}
			if uri.CloneURL().String() != want {
				t.Errorf("got %s, want %s", uri.CloneURL(), want)
			}
		})
	}
}

func TestURI_Rev(t *testing.T) {
	tests := map[string]string{
		"https://github.com/foo/bar":     "",
		"https://github.com/foo/bar?v":   "v",
		"https://github.com/foo/bar?v#":  "v",
		"https://github.com/foo/bar?v#f": "v",
	}
	for uriStr, want := range tests {
		t.Run(strings.Replace(uriStr, "/", "-", -1), func(t *testing.T) {
			uri, err := Parse(uriStr)
			if err != nil {
				t.Fatal(err)
			}
			if uri.Rev() != want {
				t.Errorf("got %s, want %s", uri.Rev(), want)
			}
		})
	}
}

func TestURI_FilePath(t *testing.T) {
	tests := map[string]string{
		"https://github.com/foo/bar":          "",
		"https://github.com/foo/bar?v":        "",
		"https://github.com/foo/bar?v#":       "",
		"https://github.com/foo/bar?v#.":      "",
		"https://github.com/foo/bar?v#f":      "f",
		"https://github.com/foo/bar?v#/f":     "f",
		"https://github.com/foo/bar?v#f/d":    "f/d",
		"https://github.com/foo/bar?v#f/..":   "",
		"https://github.com/foo/bar?v#f/d/..": "f",
		"https://github.com/foo/bar?v#//":     "",
		"https://github.com/foo/bar?v#d%2Ff":  "d/f",
	}
	for uriStr, want := range tests {
		t.Run(strings.Replace(uriStr, "/", "-", -1), func(t *testing.T) {
			uri, err := Parse(uriStr)
			if err != nil {
				t.Fatal(err)
			}
			if uri.FilePath() != want {
				t.Errorf("got %s, want %s", uri.FilePath(), want)
			}
		})
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_825(size int) error {
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
