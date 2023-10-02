package protocol_test

import (
	"fmt"
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/stretchr/testify/require"
)

func TestRequestProtoRoundtrip(t *testing.T) {
	err := quick.Check(func(r1 protocol.Request) bool {
		p1 := r1.ToProto()

		var r2 protocol.Request
		r2.FromProto(p1)
		require.Equal(t, r1, r2)

		p2 := r2.ToProto()
		require.Equal(t, p1, p2)

		return true
	}, nil)

	if err != nil {
		t.Fatal(err)
	}
}

func TestProtoFileMatchProtoRoundTrip(t *testing.T) {
	var errString string
	err := quick.Check(func(original protocol.FileMatch) bool {

		var converted protocol.FileMatch
		converted.FromProto(original.ToProto())

		if diff := cmp.Diff(original, converted); diff != "" {
			errString = fmt.Sprintf("unexpected diff (-want +got):\n%s", diff)
			return false
		}

		return true

	}, nil)

	if err != nil {
		t.Fatal(errString)
	}
}
