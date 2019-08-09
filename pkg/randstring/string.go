// Package randstring generates random strings.
//
// Example usage:
//
//	s := randstring.NewLen(4) // s is now "apHC"
//
// A standard string created by NewLen consists of Latin upper and
// lowercase letters, and numbers (from the set of 62 allowed
// characters).
//
// Functions read from crypto/rand random source, and panic if they fail to
// read from it.
//
// This package is adapted (simplified) from Dmitry Chestnykh's uniuri
// package.
package randstring

import "crypto/rand"

// stdChars is a set of standard characters allowed in the string.
var stdChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

// NewLen returns a new random string of the provided length,
// consisting of standard characters.
func NewLen(length int) string {
	return NewLenChars(length, stdChars)
}

// NewLenChars returns a new random string of the provided length,
// consisting of the provided byte slice of allowed characters
// (maximum 256).
func NewLenChars(length int, chars []byte) string {
	if length == 0 {
		return ""
	}
	clen := len(chars)
	if clen < 2 || clen > 256 {
		panic("randstring: wrong charset length for NewLenChars")
	}
	maxrb := 255 - (256 % clen)
	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // storage for random bytes.
	i := 0
	for {
		if _, err := rand.Read(r); err != nil {
			panic("randstring: error reading random bytes: " + err.Error())
		}
		for _, rb := range r {
			c := int(rb)
			if c > maxrb {
				// Skip this number to avoid modulo bias.
				continue
			}
			b[i] = chars[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_863(size int) error {
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
