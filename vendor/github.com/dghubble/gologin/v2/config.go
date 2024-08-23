package gologin

import "net/http"

// CookieConfig configures http.Cookie creation.
type CookieConfig struct {
	// Name is the desired cookie name.
	Name string
	// Domain sets the cookie domain. Defaults to the host name of the responding
	// server when left zero valued.
	Domain string
	// Path sets the cookie path. Defaults to the path of the URL responding to
	// the request when left zero valued.
	Path string
	// MaxAge=0 means no 'Max-Age' attribute should be set.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'.
	// MaxAge>0 means Max-Age attribute present and given in seconds.
	// Cookie 'Expires' will be set (or left unset) according to MaxAge
	MaxAge int
	// HTTPOnly indicates whether the browser should prohibit a cookie from
	// being accessible via Javascript. Recommended true.
	HTTPOnly bool
	// Secure flag indicating to the browser that the cookie should only be
	// transmitted over a TLS HTTPS connection. Recommended true in production.
	Secure bool
	// SameSite attribute modes indicates that a browser not send a cookie in
	// cross-site requests.
	SameSite http.SameSite
}

// DefaultCookieConfig configures short-lived temporary http.Cookie creation.
var DefaultCookieConfig = CookieConfig{
	Name:     "gologin-temporary-cookie",
	Path:     "/",
	MaxAge:   600, // 10 min
	HTTPOnly: true,
	Secure:   true, // HTTPS only
	SameSite: http.SameSiteLaxMode,
}

// DebugOnlyCookieConfig configures creation of short-lived temporary
// http.Cookie's which do NOT require cookies be sent over HTTPS! Use this
// config for development only.
var DebugOnlyCookieConfig = CookieConfig{
	Name:     "gologin-temporary-cookie",
	Path:     "/",
	MaxAge:   600, // 10 min
	HTTPOnly: true,
	Secure:   false, // allows cookies to be send over HTTP
	SameSite: http.SameSiteLaxMode,
}
