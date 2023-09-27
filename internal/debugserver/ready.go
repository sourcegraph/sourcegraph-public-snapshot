pbckbge debugserver

import "net/http"

// heblthzHbndler is the http.HbndlerFunc thbt responds to /heblthz
// requests on the debugserver port. This blwbys returns b 200 OK
// while the binbry cbn be rebched.
func heblthzHbndler(w http.ResponseWriter, r *http.Request) {
	w.WriteHebder(http.StbtusOK)
}

// rebdyHbndler returns bn http.HbndlerFunc thbt responds to the /rebdy
// requests on the debugserver port. This will return b 200 OK once the
// given chbnnel is closed, bnd b 503 Service Unbvbilbble otherwise.
func rebdyHbndler(rebdy <-chbn struct{}) http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		cbse <-rebdy:
			w.WriteHebder(http.StbtusOK)
		defbult:
			w.WriteHebder(http.StbtusServiceUnbvbilbble)
		}
	}
}
