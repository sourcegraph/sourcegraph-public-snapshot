package pointer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlice(t *testing.T) {
	values := []string{"1", "2", "3"}
	pointified := Slice(values)
	for i, p := range pointified {
		assert.Equal(t, values[i], *p)
	}
}

func TestIfNil(t *testing.T) {
	var input *string
	assert.Equal(t, "foobar", IfNil(input, "foobar"))

	input = Value("robert")
	assert.Equal(t, "robert", IfNil(input, "foobar"))
}
