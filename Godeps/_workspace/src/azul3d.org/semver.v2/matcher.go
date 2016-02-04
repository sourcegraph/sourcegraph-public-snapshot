// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package semver

import (
	"net/url"
)

// Matcher defines an object responsible for matching any given URL to an
// associated repository.
type Matcher interface {
	// Match should match the given URL to an associated repository (e.g. a git
	// repository).
	//
	// If the given URL is not a valid package URL, which may be the case very
	// often, then a nil repo and err=ErrNotPackageURL must be returned.
	//
	// If any *HTTPError is returned, then that error string is sent to the
	// client and the error's HTTP status code is written, the request is
	// considered handled.
	//
	// If any other error is returned, the request is left unhandled and the
	// error is directly returned to the caller of the Handle method.
	Match(u *url.URL) (r *Repo, err error)
}

// MatcherFunc implements the Matcher interface by simply invoking the
// function.
type MatcherFunc func(u *url.URL) (r *Repo, err error)

// Match simply invokes the function, m.
func (m MatcherFunc) Match(u *url.URL) (r *Repo, err error) {
	return m(u)
}
