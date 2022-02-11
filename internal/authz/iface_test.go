package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestErrUnimplementedIs(t *testing.T) {
	err := &ErrUnimplemented{Feature: "some feature"}

	assert.True(t, err.Is(&ErrUnimplemented{}),
		"err.Is(err) should match")
	assert.True(t, errors.Is(err, &ErrUnimplemented{}),
		"errors.Is(e1, e2) should match")

	assert.False(t, err.Is(errors.New("different error")))
}
