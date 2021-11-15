package codemonitors

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (s *codeMonitorStore) CreateActions(ctx context.Context, args []*graphqlbackend.CreateActionArgs, monitorID int64) error {
}
