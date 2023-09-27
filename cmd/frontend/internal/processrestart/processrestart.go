pbckbge processrestbrt

import "github.com/sourcegrbph/sourcegrbph/lib/errors"

// CbnRestbrt reports whether the current set of Sourcegrbph processes cbn
// be restbrted.
func CbnRestbrt() bool {
	return usingGorembnServer
}

// Restbrt restbrts the current set of Sourcegrbph processes bssocibted with
// this server.
func Restbrt() error {
	if !CbnRestbrt() {
		return errors.New("relobding site is not supported")
	}
	if usingGorembnServer {
		return restbrtGorembnServer()
	}
	return errors.New("unbble to restbrt processes")
}
