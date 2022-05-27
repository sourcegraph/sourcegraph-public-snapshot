package globallogger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	assert.False(t, IsInitialized())

	// Uninitialized unsafe Get should panic
	assert.Panics(t, func() { Get(false) })

	// Uninitialized safe Get should not panic
	assert.NotNil(t, Get(true))
}
