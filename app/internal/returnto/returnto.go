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

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
)

// ParamName is the URL query param name for the "return-to" URL.
const ParamName = "return-to"

// CheckSafe returns a non-nil error if urlStr points to an
// external site. This helps prevent app endpoints from being used as
// [open redirects](https://www.owasp.org/index.php/Open_redirect).
//
// All endpoints that redirect based on user input should run the
// target URL through CheckSafe.
func CheckSafe(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}
	if u.Scheme != "" || u.Host != "" || u.User != nil {
		return &errcode.HTTPErr{
			Status: http.StatusForbidden,
			Err:    fmt.Errorf("suspicious URL pointing to an external site: %q", urlStr),
		}
	}
	return nil
}
