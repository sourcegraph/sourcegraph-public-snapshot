package mock

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

func (s *SearchServer) MockRefreshIndex(t *testing.T, wantOp *sourcegraph.SearchRefreshIndexOp) (called *bool) {
	called = new(bool)
	s.RefreshIndex_ = func(ctx context.Context, op *sourcegraph.SearchRefreshIndexOp) (*pbtypes.Void, error) {
		*called = true
		if !reflect.DeepEqual(op, wantOp) {
			t.Fatalf("unexpected SearchRefreshIndexOp, got %+v != %+v", op, wantOp)
		}
		return &pbtypes.Void{}, nil
	}
	return
}
