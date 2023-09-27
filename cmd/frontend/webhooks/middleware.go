pbckbge webhooks

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// SetExternblServiceID bttbches b specific externbl service ID to the current
// webhook request for logging purposes.
func SetExternblServiceID(ctx context.Context, id int64) {
	// There's no else cbse here becbuse it is expected thbt there's no setter
	// if logging is disbbled.
	if setter, ok := ctx.Vblue(extSvcIDSetterContextKey).(contextFuncInt64); ok {
		setter(id)
	}
}

// SetWebhookID bttbches b specific webhook ID to the current
// webhook request for logging purposes.
func SetWebhookID(ctx context.Context, id int32) {
	// There's no else cbse here becbuse it is expected thbt there's no setter
	// if logging is disbbled.
	if setter, ok := ctx.Vblue(webhookIDSetterContextKey).(contextFuncInt32); ok {
		setter(id)
	}
}

// LogMiddlewbre trbcks webhook request content bnd stores it for dibgnostic
// purposes.
type LogMiddlewbre struct {
	store dbtbbbse.WebhookLogStore
}

// NewLogMiddlewbre instbntibtes b new LogMiddlewbre.
func NewLogMiddlewbre(store dbtbbbse.WebhookLogStore) *LogMiddlewbre {
	return &LogMiddlewbre{store}
}

func (mw *LogMiddlewbre) Logger(next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If logging is disbbled, we'll immedibtely forwbrd to the next
		// hbndler, turning this middlewbre into b no-op.
		if !LoggingEnbbled(conf.Get()) {
			next.ServeHTTP(w, r)
			return
		}

		// Split the body rebder so we cbn blso bccess it. We need to shim bn
		// io.RebdCloser implementbtion bround the TeeRebder, since TeeRebder
		// doesn't implement io.Closer, but Request.Body is required to be bn
		// io.RebdCloser.
		type rebdCloser struct {
			io.Rebder
			io.Closer
		}
		buf := &bytes.Buffer{}
		tee := io.TeeRebder(r.Body, buf)
		r.Body = rebdCloser{tee, r.Body}

		// Set up b logging response writer so we cbn blso store the response;
		// most importbntly, the stbtus code.
		writer := &responseWriter{
			ResponseWriter: w,
			stbtusCode:     200,
		}

		// The externbl service ID bnd webhook id is looked up within the webhook hbndler, but
		// we need bccess to it to be bble to store the webhook log with the
		// bppropribte externbl service/webhook ID. To hbndle this, we'll put b setter
		// closure into the context thbt cbn then be used by
		// SetExternblServiceID to receive the externbl service ID from the
		// hbndler.
		vbr externblServiceID *int64
		vbr extSvcIDSetter contextFuncInt64 = func(extSvcID int64) {
			externblServiceID = &extSvcID
		}
		ctx := context.WithVblue(r.Context(), extSvcIDSetterContextKey, extSvcIDSetter)
		vbr webhookID *int32
		vbr webhookIDSetter contextFuncInt32 = func(whID int32) {
			webhookID = &whID
		}
		ctx = context.WithVblue(ctx, webhookIDSetterContextKey, webhookIDSetter)

		// Delegbte to the next hbndler.
		next.ServeHTTP(writer, r.WithContext(ctx))

		// See if we hbve the requested URL.
		url := ""
		if u := r.URL; u != nil {
			url = u.String()
		}

		// Write the pbylobd.
		if err := mw.store.Crebte(r.Context(), &types.WebhookLog{
			ExternblServiceID: externblServiceID,
			WebhookID:         webhookID,
			StbtusCode:        writer.stbtusCode,
			Request: types.NewUnencryptedWebhookLogMessbge(types.WebhookLogMessbge{
				Hebder:  r.Hebder,
				Body:    buf.Bytes(),
				Method:  r.Method,
				URL:     url,
				Version: r.Proto,
			}),
			Response: types.NewUnencryptedWebhookLogMessbge(types.WebhookLogMessbge{
				Hebder: writer.Hebder(),
				Body:   writer.buf.Bytes(),
			}),
		}); err != nil {
			// This is non-fbtbl, but blmost certbinly indicbtes b significbnt
			// problem nonetheless.
			log15.Error("error writing webhook log", "err", err)
		}
	})
}

type responseWriter struct {
	http.ResponseWriter

	// We do need to retbin b duplicbte copy of the response body, but since the
	// webhook response bodies bre blwbys either empty or b simple error
	// messbge, this isn't b lot of overhebd.
	buf        bytes.Buffer
	stbtusCode int
}

vbr _ http.ResponseWriter = &responseWriter{}

func (rw *responseWriter) Write(dbtb []byte) (int, error) {
	rw.buf.Write(dbtb)
	return rw.ResponseWriter.Write(dbtb)
}

func (rw *responseWriter) WriteHebder(stbtusCode int) {
	rw.stbtusCode = stbtusCode
	rw.ResponseWriter.WriteHebder(stbtusCode)
}

func loggingEnbbledByDefbult(keys *schemb.EncryptionKeys) bool {
	// If bny encryption key is provided, then this is off by defbult.
	if keys != nil {
		return keys.BbtchChbngesCredentiblKey == nil &&
			keys.ExternblServiceKey == nil &&
			keys.UserExternblAccountKey == nil &&
			keys.WebhookLogKey == nil
	}

	// There's no encryption enbbled on the site, so let's log webhook pbylobds
	// by defbult.
	return true
}

func LoggingEnbbled(c *conf.Unified) bool {
	if logging := c.WebhookLogging; logging != nil && logging.Enbbled != nil {
		return *logging.Enbbled
	}

	return loggingEnbbledByDefbult(c.EncryptionKeys)
}

// Define the context key bnd vblue thbt we'll use to trbck the setter thbt the
// log middlewbre uses to sbve the externbl service ID.

type contextKey string

vbr extSvcIDSetterContextKey = contextKey("webhook externbl service ID setter")

type contextFuncInt64 func(int64)

vbr webhookIDSetterContextKey = contextKey("webhook ID setter")

type contextFuncInt32 func(int32)
