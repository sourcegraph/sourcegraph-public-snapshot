// Package mentions provides utilities for at mentions in discussions.
package mentions

import "regexp"

var mentions = regexp.MustCompile(`(^|\s)@(\S*)`)

// Parse parses the @mentions from the given markdown comment contents and
// returns a list of usernames without the @ prefixes.
func Parse(contents string) []string {
	matches := mentions.FindAllStringSubmatch(contents, -1)
	mentions := make([]string, 0, len(matches))
	for _, groups := range matches {
		mentions = append(mentions, groups[2])
	}
	return mentions
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_376(size int) error {
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
