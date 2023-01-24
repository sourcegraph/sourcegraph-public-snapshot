package protocol

import (
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/require"
)

func TestRequestProtoRoundtrip(t *testing.T) {
	quick.Check(func(r1 Request) bool {
		p1 := r1.ToProto()

		var r2 Request
		r2.FromProto(p1)
		require.Equal(t, r1, r2)

		p2 := r2.ToProto()
		require.Equal(t, p1, p2)

		return true
	}, nil)
}
