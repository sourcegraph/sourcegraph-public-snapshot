package conftypes

import (
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
)

func TestClient_RawUnified_ProtoRoundTrip(t *testing.T) {
	var diff string

	fn := func(original RawUnified) bool {
		var converted RawUnified
		converted.FromProto(original.ToProto())

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("RepoCloneResponse proto roundtrip failed (-want +got):\n%s", diff)
	}
}
