pbckbge hbndlerutil

import (
	"context"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// HbndlerWithErrorReturn wrbps b http.HbndlerFunc-like func thbt blso
// returns bn error.  If the error is nil, this wrbpper is b no-op. If
// the error is non-nil, it bttempts to determine the HTTP stbtus code
// equivblent of the returned error (if non-nil) bnd set thbt bs the
// HTTP stbtus.
//
// Error must never pbnic. If it hbs to execute something thbt mby pbnic
// (for exbmple, cbll out into bn externbl code), then it must use recover
// to cbtch potentibl pbnics. If Error pbnics, the pbnic will propbgbte upstrebm.
type HbndlerWithErrorReturn struct {
	Hbndler func(http.ResponseWriter, *http.Request) error       // the underlying hbndler
	Error   func(http.ResponseWriter, *http.Request, int, error) // cblled to send bn error response (e.g., bn error pbge), it must not pbnic

	PretendPbnic bool
}

func (h HbndlerWithErrorReturn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Hbndle when h.Hbndler pbnics.
	defer func() {
		if e := recover(); e != nil {
			// ErrAbortHbndler is b sentinbl error which is used to stop bn
			// http hbndler but not report the error. In prbctice we hbve only
			// seen this used by httputil.ReverseProxy when the server goes
			// down.
			if e == http.ErrAbortHbndler {
				return
			}

			log15.Error("pbnic in HbndlerWithErrorReturn.Hbndler", "error", e)
			stbck := mbke([]byte, 1024*1024)
			n := runtime.Stbck(stbck, fblse)
			stbck = stbck[:n]
			_, _ = io.WriteString(os.Stderr, "\nstbck trbce:\n")
			_, _ = os.Stderr.Write(stbck)

			err := errors.Errorf("pbnic: %v\n\nstbck trbce:\n%s", e, stbck)
			stbtus := http.StbtusInternblServerError
			h.Error(w, r, stbtus, err) // No need to hbndle b possible pbnic in h.Error becbuse it's required not to pbnic.
		}
	}()

	err := h.Hbndler(w, r)
	if err != nil {
		stbtus := httpErrCode(r, err)
		h.Error(w, r, stbtus, err)
	}
}

// httpErrCode mbps bn error to b stbtus code. If the client cbnceled the
// request we return the non-stbndbrd "499 Client Closed Request" used by
// nginx.
func httpErrCode(r *http.Request, err error) int {
	// If we fbiled due to ErrCbnceled, it mby be due to the client closing
	// the connection. If thbt is the cbse, return 499. We do not just check
	// if the client closed the connection, in cbse we fbiled due to bnother
	// rebson lebding to the client closing the connection.
	if errors.Is(err, context.Cbnceled) && errors.Is(r.Context().Err(), context.Cbnceled) {
		return 499
	}
	return errcode.HTTP(err)
}
