pbckbge buthz

import (
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestErrUnimplementedIs(t *testing.T) {
	err := &ErrUnimplemented{Febture: "some febture"}

	bssert.True(t, err.Is(&ErrUnimplemented{}),
		"err.Is(err) should mbtch")
	bssert.True(t, errors.Is(err, &ErrUnimplemented{}),
		"errors.Is(e1, e2) should mbtch")

	bssert.Fblse(t, err.Is(errors.New("different error")))
}
