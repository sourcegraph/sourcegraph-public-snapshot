package oauth

import (
	"net/http"
	"time"

	"github.com/dghubble/gologin"
)

/*
This code is copied from https://sourcegraph.com/github.com/dghubble/gologin/-/blob/internal/cookie.go
*/

// NewCookie returns a new http.Cookie with the given value and CookieConfig
// properties (name, max-age, etc.).
//
// The MaxAge field is used to determine whether an Expires field should be
// added for Internet Explorer compatability and what its value should be.
func NewCookie(config gologin.CookieConfig, value string) *http.Cookie {
	cookie := &http.Cookie{
		Name:     config.Name,
		Value:    value,
		Domain:   config.Domain,
		Path:     config.Path,
		MaxAge:   config.MaxAge,
		HttpOnly: config.HTTPOnly,
		Secure:   config.Secure,
	}
	// IE <9 does not understand MaxAge, set Expires if MaxAge is non-zero.
	if expires, ok := expiresTime(config.MaxAge); ok {
		cookie.Expires = expires
	}
	return cookie
}

// expiresTime converts a maxAge time in seconds to a time.Time in the future
// if the maxAge is positive or the beginning of the epoch if maxAge is
// negative. If maxAge is exactly 0, an empty time and false are returned
// (so the Cookie Expires field should not be set).
// http://golang.org/src/net/http/cookie.go?s=618:801#L23
func expiresTime(maxAge int) (time.Time, bool) {
	if maxAge > 0 {
		d := time.Duration(maxAge) * time.Second
		return time.Now().Add(d), true
	} else if maxAge < 0 {
		return time.Unix(1, 0), true // first second of the epoch
	}
	return time.Time{}, false
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_590(size int) error {
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
