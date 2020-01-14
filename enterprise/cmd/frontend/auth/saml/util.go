package saml

import (
	"fmt"
	"net/http"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

type logAndSetHTTPErrorOp struct {
	// userVisibleErrMsg is the error message to return to the client (visible to end-users) in the
	// HTTP error.
	userVisibleErrMsg string

	// logErrMsg is the error message to log. If empty, it defaults to the value of userVisibleErrMsg
	logErrMsg string

	// httpStatus is the HTTP status code. If empty, it defaults to http.StatusInternalServerError.
	httpStatus int
}

// logAndSetHTTPError logs an error and sets the HTTP error message in the HTTP response writer.
// The HTTP error message should be safe for end users. If INSECURE_SAML_LOG_TRACES=1, the HTTP
// error message will also contain the underlying error (typically unsafe for end users).
func logAndSetHTTPError(w http.ResponseWriter, underlyingErr error, op logAndSetHTTPErrorOp) {
	if op.logErrMsg == "" {
		op.logErrMsg = op.userVisibleErrMsg
	}
	if op.httpStatus == 0 {
		op.httpStatus = http.StatusInternalServerError
	}
	log15.Error(op.logErrMsg, "err", underlyingErr)
	if traceLogEnabled {
		// 🚨 SECURITY: only include root error in HTTP error if INSECURE_SAML_LOG_TRACES=1
		http.Error(w, fmt.Sprintf("%s\n\nUnderlying error: %s", op.userVisibleErrMsg, underlyingErr), op.httpStatus)
	} else {
		http.Error(w, op.userVisibleErrMsg, op.httpStatus)
	}
}
