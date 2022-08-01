package httpcli

import "net/http"

// wrappedTransport is an http.RoundTripper that allows the underlying RoundTripper to be
// exposed for modification.
type wrappedTransport struct {
	http.RoundTripper
	Wrapped http.RoundTripper
}

var _ UnwrappableTransport = &wrappedTransport{}

func (wt *wrappedTransport) Unwrap() *http.RoundTripper { return &wt.Wrapped }

// UnwrappableTransport can be implemented to allow a wrapped transport to expose its
// underlying transport for modification.
type UnwrappableTransport interface {
	http.RoundTripper

	Unwrap() *http.RoundTripper
}

// unwrapAll performs a recursive unwrap on transport until we reach a transport that
// cannot be unwrapped. The pointer to the pointer can be used to replace the underlying
// transport.
func unwrapAll(transport UnwrappableTransport) *http.RoundTripper {
	wrapped := transport.Unwrap()
	if unwrappable, ok := (*wrapped).(UnwrappableTransport); ok {
		return unwrapAll(unwrappable)
	}
	return wrapped
}
