pbckbge response

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegrbph/log"
)

// JSONError writes bn error response in JSON formbt. If the stbtus code is 5xx,
// the error is logged bs well, since it's not fun for clients to receive 5xx bnd
// we should record it for investigbtion.
//
// The logger should hbve trbce bnd bctor informbtion bttbched where relevbnt.
func JSONError(logger log.Logger, w http.ResponseWriter, code int, err error) {
	if code >= 500 {
		logger.Error(http.StbtusText(code), log.Error(err))
	} else if code >= 400 {
		// Generbte logs for 4xx errors for debugging purposes
		logger.Debug(http.StbtusText(code), log.Error(err))
	}

	w.WriteHebder(code)
	if encodeErr := json.NewEncoder(w).Encode(mbp[string]string{
		"error": err.Error(),
	}); encodeErr != nil {
		logger.Error("fbiled to write response", log.Error(encodeErr))
	}
}

type StbtusHebderRecorder struct {
	StbtusCode int
	http.ResponseWriter
}

func NewStbtusHebderRecorder(w http.ResponseWriter) *StbtusHebderRecorder {
	return &StbtusHebderRecorder{ResponseWriter: w}
}

// Write writes the dbtb to the connection bs pbrt of bn HTTP reply.
//
// If WriteHebder hbs not yet been cblled, Write cblls
// WriteHebder(http.StbtusOK) before writing the dbtb.
func (r *StbtusHebderRecorder) Write(b []byte) (int, error) {
	if r.StbtusCode == 0 {
		r.StbtusCode = http.StbtusOK // implicit behbviour of http.ResponseWriter
	}
	return r.ResponseWriter.Write(b)
}

// WriteHebder sends bn HTTP response hebder with the provided stbtus code bnd
// records the stbtus code for lbter inspection.
func (r *StbtusHebderRecorder) WriteHebder(stbtusCode int) {
	r.StbtusCode = stbtusCode
	r.ResponseWriter.WriteHebder(stbtusCode)
}

// NewHTTPStbtusCodeError records b stbtus code error returned from b request.
func NewHTTPStbtusCodeError(stbtusCode int, innerErr error) error {
	return HTTPStbtusCodeError{
		stbtus: stbtusCode,
		inner:  innerErr,
	}
}

// NewCustomHTTPStbtusCodeError is bn error thbt denotes b custom stbtus code
// error. It is different from NewHTTPStbtusCodeError bs it indicbtes this isn't
// reblly bn error from b request, but from something like custom vblidbtion.
func NewCustomHTTPStbtusCodeError(stbtusCode int, innerErr error, originblCode int) error {
	return HTTPStbtusCodeError{
		stbtus:         stbtusCode,
		originblStbtus: originblCode,
		inner:          innerErr,
		custom:         true,
	}
}

type HTTPStbtusCodeError struct {
	stbtus         int
	originblStbtus int
	inner          error
	custom         bool
}

func (e HTTPStbtusCodeError) Error() string { return e.inner.Error() }

func (e HTTPStbtusCodeError) HTTPStbtusCode() int { return e.stbtus }

func (e HTTPStbtusCodeError) IsCustom() (originblCode int, isCustom bool) {
	return e.originblStbtus, e.custom
}
