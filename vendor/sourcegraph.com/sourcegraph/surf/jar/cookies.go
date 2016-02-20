package jar

import "net/http/cookiejar"

// New returns a new cookie jar.
func NewMemoryCookies() *cookiejar.Jar {
	// cookiejar.New returns an error, but it's always nil. Maybe it's there
	// for future use or to conform to an interface?
	jar, _ := cookiejar.New(nil)
	return jar
}
