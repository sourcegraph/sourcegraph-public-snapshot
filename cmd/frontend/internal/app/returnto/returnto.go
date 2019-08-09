// Package returnto determines the "return-to" URL for a given
// request.
//
// The "return-to" URL is the URL that the user agent should be
// redirected to, after a certain process is completed.
//
// For example, if the user is viewing the page at "/a" and then goes
// through the login process, they should be redirected back to "/a"
// after successfully logging in. This is accomplished through the
// following steps:
//
// 1. The "login" link in the navigation bar links to, e.g.,
//    "/login?return-to=/a".
//
// 2. The login form, when submitted, POSTS to "/login?return-to=/a"
//    (where the "/a" is taken from the URL's "return-to" query
//    param).
//
// 3. The login submission handler redirects to "/a" after a
//    successful login (where the "/a" is taken from the POST request's
//    "return-to" query param).
package returnto

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ParamName is the URL query param name for the "return-to" URL.
const ParamName = "return-to"

// URLFromRequest determines the proper return-to URL to use from the
// given request. It uses the URL passed in the "return-to" URL query
// parameter. If it's empty, the URL path "/" is returned. If it is
// invalid, an error is returned.
func URLFromRequest(r *http.Request, paramName string) (*url.URL, error) {
	v := r.URL.Query().Get(paramName)
	if v == "" {
		return &url.URL{Path: "/"}, nil
	}
	u, err := url.Parse(v)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "" || u.Host != "" || u.User != nil || u.Opaque != "" {
		return nil, fmt.Errorf("suspicious return-to URL pointing to an external site: %q", v)
	}

	// Remove any nested return-to URLs to avoid an infinite loop.
	q := u.Query()
	if q.Get(paramName) != "" {
		q.Del(paramName)
		u.RawQuery = q.Encode()
	}

	if u.Path == "" {
		u.Path = "/"
	}
	if !strings.HasPrefix(u.Path, "/") {
		return nil, fmt.Errorf("return-to URL must have absolute path, starting with '/': %s", v)
	}
	return u, nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_276(size int) error {
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
