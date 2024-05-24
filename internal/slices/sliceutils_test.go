package slices_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/slices"
)

func TestMap(t *testing.T) {
	input := []int{1, 2, 3}

	res := slices.Map(input, func(i int) int { return i + 1 })
	expected := []int{2, 3, 4}
	assert.Equal(t, expected, res)
}
