package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/stretchr/testify/assert"
)

func TestStatusHeaderRecorder(t *testing.T) {
	logger := logtest.Scoped(t)

	t.Run("WriteHeader", func(t *testing.T) {
		underlying := httptest.NewRecorder()
		recorder := NewStatusHeaderRecorder(underlying, logger)

		var w http.ResponseWriter = recorder
		w.WriteHeader(http.StatusTeapot)

		assert.Equal(t, http.StatusTeapot, recorder.StatusCode)
		assert.Equal(t, http.StatusTeapot, underlying.Code)
	})

	t.Run("implicit WriteHeader", func(t *testing.T) {
		underlying := httptest.NewRecorder()
		recorder := NewStatusHeaderRecorder(underlying, logger)

		var w http.ResponseWriter = recorder
		w.Write([]byte("foo")) // should implicitly write header

		assert.Equal(t, http.StatusOK, recorder.StatusCode)
		assert.Equal(t, http.StatusOK, underlying.Code)
		assert.Equal(t, "foo", underlying.Body.String())
	})
}

func TestAutoFlusher_Write(t *testing.T) {
	rw := &responseWriterMock{}

	writer, err := NewAutoFlushingWriter(rw)
	assert.NoError(t, err)

	_, err = writer.Write([]byte("data"))
	assert.NoError(t, err)
	assert.Equal(t, 1, rw.flushCalledTimes)
	assert.Equal(t, []byte("data"), rw.written)
	_, err = writer.Write([]byte("value"))
	assert.NoError(t, err)
	assert.Equal(t, 2, rw.flushCalledTimes)
	assert.Equal(t, []byte("datavalue"), rw.written)
}

type responseWriterMock struct {
	written          []byte
	flushCalledTimes int
}

var _ http.Flusher = &responseWriterMock{}
var _ http.ResponseWriter = &responseWriterMock{}

func (m *responseWriterMock) Flush() {
	m.flushCalledTimes += 1
}

func (m *responseWriterMock) Write(p []byte) (int, error) {
	m.written = append(m.written, p...)
	return len(p), nil
}
func (m *responseWriterMock) Header() http.Header {
	return nil
}

func (m *responseWriterMock) WriteHeader(_ int) {
}
