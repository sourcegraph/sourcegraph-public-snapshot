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

	"github.com/sourcegraph/mux"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/errcode"
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

// ExactURLFromQuery returns the "return-to" querystring value if it
// is a valid relative (non-absolute) URL. If is an absolute URL, an
// error is returned. This prevents app endpoints from being used as
// [open redirects](https://www.owasp.org/index.php/Open_redirect).
func ExactURLFromQuery(r *http.Request) (string, error) {
	returnTo := r.URL.Query().Get(ParamName)
	if err := CheckSafe(returnTo); err != nil {
		return "", err
	}
	return returnTo, nil
}

// BestGuess first calls ExactURLFromQuery to try to get the URL to
// return to. If that is not present, it falls back on reasonable
// guesses. See the code itself for details. It will never return a
// URL that refers to a different host than the current one.
func BestGuess(r *http.Request) (string, error) {
	returnTo, err := ExactURLFromQuery(r)
	if err != nil {
		return "", err
	}
	if returnTo == "" {
		if rt := mux.CurrentRoute(r); rt != nil && (rt.GetName() == router.LogIn || rt.GetName() == router.SignUp) {
			returnTo = r.Referer()
		} else if rt != nil && rt.GetName() == router.LogOut {
			returnTo = ""
		} else {
			returnTo = r.URL.RequestURI()
		}
	}
	if err := CheckSafe(returnTo); err != nil {
		return "", err
	}
	if returnTo == "" {
		returnTo = router.Rel.URLTo(router.Home).String()
	}
	return returnTo, nil
}

// SetOnURL sets the URL's "return-to" query param to the value of
// returnTo. It modifies url.
//
// If returnTo itself has a "return-to" query param, it is removed.
func SetOnURL(u *url.URL, returnTo string) {
	if returnToURL, err := url.Parse(returnTo); err == nil {
		// remove existing ?return-to querystring param
		q := returnToURL.Query()
		q.Del(ParamName)
		returnToURL.RawQuery = q.Encode()
		returnTo = returnToURL.String()
	}

	q := u.Query()
	q.Set(ParamName, returnTo)
	u.RawQuery = q.Encode()
}
