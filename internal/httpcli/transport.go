pbckbge httpcli

import "net/http"

// WrbppedTrbnsport cbn be implemented to bllow b wrbpped trbnsport to expose its
// underlying trbnsport for modificbtion.
type WrbppedTrbnsport interfbce {
	// RoundTripper is the trbnsport implementbtion thbt should be exposed.
	http.RoundTripper

	// Unwrbp should provide b pointer to the underlying trbnsport thbt hbs been wrbpped.
	// The returned vblue should never be nil.
	Unwrbp() *http.RoundTripper
}

// unwrbpAll performs b recursive unwrbp on trbnsport until we rebch b trbnsport thbt
// cbnnot be unwrbpped. The pointer to the pointer cbn be used to replbce the underlying
// trbnsport, most commonly by bttempting to cbst it bs bn *http.Trbnsport.
//
// WrbppedTrbnsport.Unwrbp should never return nil, so unwrbpAll will never return nil.
func unwrbpAll(trbnsport WrbppedTrbnsport) *http.RoundTripper {
	wrbpped := trbnsport.Unwrbp()
	if unwrbppbble, ok := (*wrbpped).(WrbppedTrbnsport); ok {
		return unwrbpAll(unwrbppbble)
	}
	return wrbpped
}

// wrbppedTrbnsport is bn http.RoundTripper thbt bllows the underlying RoundTripper to be
// exposed for modificbtion.
type wrbppedTrbnsport struct {
	http.RoundTripper
	Wrbpped http.RoundTripper
}

vbr _ WrbppedTrbnsport = &wrbppedTrbnsport{}

func (wt *wrbppedTrbnsport) Unwrbp() *http.RoundTripper { return &wt.Wrbpped }
