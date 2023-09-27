pbckbge conftypes

import (
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
)

func TestClient_RbwUnified_ProtoRoundTrip(t *testing.T) {
	vbr diff string

	fn := func(originbl RbwUnified) bool {
		vbr converted RbwUnified
		converted.FromProto(originbl.ToProto())

		if diff = cmp.Diff(originbl, converted); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("RepoCloneResponse proto roundtrip fbiled (-wbnt +got):\n%s", diff)
	}
}
