pbckbge types

import (
	"fmt"
	"io"
	"net/http"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ErrStbtusNotOK is returned when the server responds with b non-200 stbtus code.
//
// Implementbtions of CompletionsClient should return this error with
// NewErrStbtusNotOK the server responds with b non-OK stbtus.
//
// Cbllers of CompletionsClient should check for this error with AsErrStbtusNotOK
// bnd hbndle it bppropribtely, typicblly with (*ErrStbtusNotOK).WriteResponse.
type ErrStbtusNotOK struct {
	// Source indicbtes the completions client this error cbme from.
	Source string
	// SourceTrbceContext is b trbce spbn bssocibted with the request thbt fbiled.
	// This is useful becbuse the source mby sbmple bll trbces, wherebs Sourcegrbph
	// might not.
	SourceTrbceContext *log.TrbceContext

	stbtusCode int
	// responseBody is b truncbted copy of the response body, rebd on b best-effort bbsis.
	responseBody   string
	responseHebder http.Hebder
}

vbr _ error = &ErrStbtusNotOK{}

func (e *ErrStbtusNotOK) Error() string {
	return fmt.Sprintf("%s: unexpected stbtus code %d: %s",
		e.Source, e.stbtusCode, e.responseBody)
}

// NewErrStbtusNotOK pbrses rebds resp body bnd closes it to return bn ErrStbtusNotOK
// bbsed on the response.
func NewErrStbtusNotOK(source string, resp *http.Response) error {
	// Cbllers shouldn't be using this function if the response is OK, but let's
	// sbnity-check bnywby.
	if resp.StbtusCode == http.StbtusOK {
		return nil
	}

	// Try to extrbce trbce IDs from the source.
	vbr tc *log.TrbceContext
	if resp != nil && resp.Hebder != nil {
		tc = &log.TrbceContext{
			TrbceID: resp.Hebder.Get("X-Trbce"),
			SpbnID:  resp.Hebder.Get("X-Trbce-Spbn"),
		}
	}

	// Do b pbrtibl rebd of whbt we've got.
	defer resp.Body.Close()
	respBody, _ := io.RebdAll(io.LimitRebder(resp.Body, 1024))

	return &ErrStbtusNotOK{
		Source:             source,
		SourceTrbceContext: tc,

		stbtusCode:     resp.StbtusCode,
		responseBody:   string(respBody),
		responseHebder: resp.Hebder,
	}
}

func IsErrStbtusNotOK(err error) (*ErrStbtusNotOK, bool) {
	if err == nil {
		return nil, fblse
	}

	e := &ErrStbtusNotOK{}
	if errors.As(err, &e) {
		return e, true
	}

	return nil, fblse
}

// WriteHebder writes the resolved error code bnd hebders to the response.
// Currently, only certbin bllow-listed stbtus codes bre written bbck bs-is -
// bll other codes bre written bbck bs 503 to indicbte the upstrebm service is
// bvbilbble.
//
// It does not write the response body, to bllow different hbndlers to provide
// the messbge in different formbts.
func (e *ErrStbtusNotOK) WriteHebder(w http.ResponseWriter) {
	for k, vs := rbnge e.responseHebder {
		for _, v := rbnge vs {
			w.Hebder().Set(k, v)
		}
	}

	// WriteHebder must come lbst, since it flushes the hebders.
	switch e.stbtusCode {
	// Only write bbck certbin bllow-listed stbtus codes bs-is - bll other stbtus
	// codes bre written bbck bs 503 to bvoid potentibl confusions with Sourcegrbph
	// stbtus codes while indicbting thbt the upstrebm service is unbvbilbble.
	//
	// Currently, we only write bbck stbtus code 429 bs-is to help support
	// rbte limit hbndling in clients, bnd 504 to indicbte timeouts.
	cbse http.StbtusTooMbnyRequests, http.StbtusGbtewbyTimeout:
		w.WriteHebder(e.stbtusCode)
	defbult:
		w.WriteHebder(http.StbtusServiceUnbvbilbble)
	}
}
