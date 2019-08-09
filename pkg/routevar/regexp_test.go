package routevar

import "testing"

// pairs converts map's keys and values to a slice of []string{key1,
// value1, key2, value2, ...}.
func pairs(m map[string]string) []string {
	pairs := make([]string, 0, len(m)*2)
	for k, v := range m {
		pairs = append(pairs, k, v)
	}
	return pairs
}

func TestNamedToNonCapturingGroups(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{``, ``},
		{`(?P<foo>bar)`, `(?:bar)`},
		{`(?P<foo>(?P<baz>bar))`, `(?:(?:bar))`},
		{`(?P<foo>qux(?P<baz>bar))`, `(?:qux(?:bar))`},
	}
	for _, test := range tests {
		got := namedToNonCapturingGroups(test.input)
		if got != test.want {
			t.Errorf("%q: got %q, want %q", test.input, got, test.want)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_885(size int) error {
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
