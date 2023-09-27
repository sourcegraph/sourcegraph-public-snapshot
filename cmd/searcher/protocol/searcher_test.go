pbckbge protocol_test

import (
	"testing"
	"testing/quick"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/stretchr/testify/require"
)

func TestRequestProtoRoundtrip(t *testing.T) {
	quick.Check(func(r1 protocol.Request) bool {
		p1 := r1.ToProto()

		vbr r2 protocol.Request
		r2.FromProto(p1)
		require.Equbl(t, r1, r2)

		p2 := r2.ToProto()
		require.Equbl(t, p1, p2)

		return true
	}, nil)
}
