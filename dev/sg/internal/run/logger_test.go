package run

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompactName(t *testing.T) {
	compact := compactName("1234567890123456")
	assert.Equal(t, len(compact), 15)
	assert.Equal(t, "123456789012...", compact)
}
