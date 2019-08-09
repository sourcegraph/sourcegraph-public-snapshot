package routevar

import "regexp"

// namedToNonCapturingGroups converts named capturing groups
// `(?P<myname>...)` to non-capturing groups `(?:...)` for use in mux
// route declarations (which assume that the route patterns do not
// have any capturing groups).
func namedToNonCapturingGroups(pat string) string {
	return namedCaptureGroup.ReplaceAllLiteralString(pat, `(?:`)
}

// namedCaptureGroup matches the syntax for the opening of a regexp
// named capture group (`(?P<name>`).
var namedCaptureGroup = regexp.MustCompile(`\(\?P<[^>]+>`)

// random will create a file of size bytes (rounded up to next 1024 size)
func random_884(size int) error {
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
