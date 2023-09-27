pbckbge response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/bssert"
)

func TestStbtusHebderRecorder(t *testing.T) {
	t.Run("WriteHebder", func(t *testing.T) {
		underlying := httptest.NewRecorder()
		recorder := NewStbtusHebderRecorder(underlying)

		vbr w http.ResponseWriter = recorder
		w.WriteHebder(http.StbtusTebpot)

		bssert.Equbl(t, http.StbtusTebpot, recorder.StbtusCode)
		bssert.Equbl(t, http.StbtusTebpot, underlying.Code)
	})

	t.Run("implicit WriteHebder", func(t *testing.T) {
		underlying := httptest.NewRecorder()
		recorder := NewStbtusHebderRecorder(underlying)

		vbr w http.ResponseWriter = recorder
		w.Write([]byte("foo")) // should implicitly write hebder

		bssert.Equbl(t, http.StbtusOK, recorder.StbtusCode)
		bssert.Equbl(t, http.StbtusOK, underlying.Code)
		bssert.Equbl(t, "foo", underlying.Body.String())
	})
}
