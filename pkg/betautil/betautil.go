package betautil

// Betas must be added and written exactly the same in two locations:
//
//  pkg/betautil/betautil.go
//  ui/web_modules/sourcegraph/util/betautil.js
//
// Beta strings should be human-readable, and should not be suffixed with
// "Beta". Add both a constant and an entry to the map below.
//
// Example const:
//
//  AwesomeStuff = "Awesome Stuff"
//

const (
	JavaScript = "JavaScript"
	Python     = "Python"
	Desktop    = "Desktop"
	Universe   = "Universe"
)

// Betas is a map of beta strings which is used to determine if a beta string
// is valid or not. It only indicates that a beta program exists, not whether
// or not it is active.
var Betas = map[string]bool{
	JavaScript: true,
	Python:     true,
	Desktop:    true,
	Universe:   true,
}
