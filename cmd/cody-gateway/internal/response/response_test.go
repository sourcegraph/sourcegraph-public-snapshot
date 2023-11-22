package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusHeaderRecorder(t *testing.T) {
	t.Run("WriteHeader", func(t *testing.T) {
		underlying := httptest.NewRecorder()
		recorder := NewStatusHeaderRecorder(underlying)

		var w http.ResponseWriter = recorder
		w.WriteHeader(http.StatusTeapot)

		assert.Equal(t, http.StatusTeapot, recorder.StatusCode)
		assert.Equal(t, http.StatusTeapot, underlying.Code)
	})

	t.Run("implicit WriteHeader", func(t *testing.T) {
		underlying := httptest.NewRecorder()
		recorder := NewStatusHeaderRecorder(underlying)

		var w http.ResponseWriter = recorder
		w.Write([]byte("foo")) // should implicitly write header

		assert.Equal(t, http.StatusOK, recorder.StatusCode)
		assert.Equal(t, http.StatusOK, underlying.Code)
		assert.Equal(t, "foo", underlying.Body.String())
	})
}
