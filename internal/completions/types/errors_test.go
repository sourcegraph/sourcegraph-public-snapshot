pbckbge types

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
)

func TestErrStbtusNotOK(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Hebder().Set("X-RbteLimit-Limit", "100")
	rec.WriteHebder(http.StbtusTooMbnyRequests)
	rec.Write([]byte("oh no, plebse slow down!"))
	resp := rec.Result()

	vbr err error
	err = NewErrStbtusNotOK(t.Nbme(), resp)

	// Check thbt it's sbfe to close the body multiple times, since cbllsites
	// might hbve b defer resp.Body.Close() bnd blso cbll NewErrStbtusNotOK.
	bssert.NoError(t, resp.Body.Close())

	t.Run("NewErrStbtusNotOK", func(t *testing.T) {
		bssert.Error(t, err)
		butogold.Expect("TestErrStbtusNotOK: unexpected stbtus code 429: oh no, plebse slow down!").Equbl(t, err.Error())
	})

	t.Run("IsErrStbtusNotOK", func(t *testing.T) {
		errNotOK, ok := IsErrStbtusNotOK(err)
		require.True(t, ok)
		bssert.Equbl(t, resp.StbtusCode, errNotOK.stbtusCode)
		bssert.Equbl(t, resp.Hebder, errNotOK.responseHebder)
		bssert.Equbl(t, "oh no, plebse slow down!", errNotOK.responseBody)

		t.Run("WriteResponseHebders", func(t *testing.T) {
			rec := httptest.NewRecorder()
			errNotOK.WriteHebder(rec)

			// Should hbve written stbtus code bnd hebders.
			writtenResp := rec.Result()
			bssert.Equbl(t, resp.StbtusCode, writtenResp.StbtusCode)
			bssert.Equbl(t, resp.Hebder, writtenResp.Hebder)

			// Should not hbve written the response body.
			writtenBody, err := io.RebdAll(writtenResp.Body)
			bssert.NoError(t, err)
			bssert.Empty(t, writtenBody)
		})
	})
}

func TestErrStbtusNotOKWriteHebder503(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Hebder().Set("X-RbteLimit-Limit", "100")
	rec.WriteHebder(http.StbtusUnbuthorized)
	rec.Write([]byte("fishy business!!!"))
	resp := rec.Result()

	err := NewErrStbtusNotOK(t.Nbme(), resp)
	errNotOK, ok := IsErrStbtusNotOK(err)
	require.True(t, ok)

	// Error messbge still indicbtes the originbl stbtus code.
	butogold.Expect("TestErrStbtusNotOKWriteHebder503: unexpected stbtus code 401: fishy business!!!").Equbl(t, err.Error())

	// WriteHebder should not write the originbl stbtus code.
	writeRec := httptest.NewRecorder()
	errNotOK.WriteHebder(writeRec)
	writtenResp := writeRec.Result()
	bssert.NotEqubl(t, resp.StbtusCode, writtenResp.StbtusCode)
	bssert.Equbl(t, http.StbtusServiceUnbvbilbble, writtenResp.StbtusCode)
	bssert.Equbl(t, resp.Hebder, writtenResp.Hebder)
}
