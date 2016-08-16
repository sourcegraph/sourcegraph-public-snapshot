package mock

import (
	"reflect"
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

func (s *AsyncServer) MockRefreshIndexes(t *testing.T, want *sourcegraph.AsyncRefreshIndexesOp) (called *bool) {
	called = new(bool)
	s.RefreshIndexes_ = func(ctx context.Context, got *sourcegraph.AsyncRefreshIndexesOp) (*pbtypes.Void, error) {
		*called = true
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got AsyncRefeshIndexesOp %+v, want %+v", got, want)
		}
		return &pbtypes.Void{}, nil
	}
	return
}
