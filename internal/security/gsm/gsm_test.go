package gsm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchGSM(t *testing.T) {

	t.Run("Error without GOOGLE_PROJECT_ID", func(t *testing.T) {
		_, err := getSecretFromGSM("really-a-password")
		assert.ErrorContains(t, err, "no GOOGLE_PROJECT_ID defined")
	})

	t.Run("Get() after Lock() panics", func(t *testing.T) {
		assert.Panics(t, func() {
			Lock()
			Get("foo", "Foo secret for Baz")
		})
	})

	t.Run("Double Get() panics", func(t *testing.T) {
		assert.Panics(t, func() {
			Get("foo", "Foo secret for Baz")
			Get("foo", "Foo secret for Baz")
		})
	})

}
