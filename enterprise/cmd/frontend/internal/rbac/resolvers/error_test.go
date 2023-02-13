package resolvers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrIDIsZero(t *testing.T) {
	e := ErrIDIsZero{}

	assert.Equal(t, e.Error(), "invalid node id")
}
