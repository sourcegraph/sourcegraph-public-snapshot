package redispool

import "testing"

func TestSchemeMatcher(t *testing.T) {
	tests := []struct {
		urlMaybe  string
		hasScheme bool
	}{
		{"redis://foo.com", true},
		{"https://foo.com", true},
		{"redis://:password@foo.com/0", true},
		{"redis://foo.com/0?password=foo", true},
		{"foo:1234", false},
	}
	for _, test := range tests {
		hasScheme := schemeMatcher.MatchString(test.urlMaybe)
		if hasScheme != test.hasScheme {
			t.Errorf("for string %q, exp != got: %v != %v", test.urlMaybe, test.hasScheme, hasScheme)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_872(size int) error {
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
