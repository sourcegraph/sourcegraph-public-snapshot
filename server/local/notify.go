package local

import (
	"fmt"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

var Notify sourcegraph.NotifyServer = &notify{}

type notify struct{}

var _ sourcegraph.NotifyServer = (*notify)(nil)

func (s *notify) GenericEvent(ctx context.Context, e *sourcegraph.NotifyGenericEvent) (*pbtypes.Void, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *notify) Mention(ctx context.Context, m *sourcegraph.NotifyMention) (*pbtypes.Void, error) {
	return nil, fmt.Errorf("not implemented")
}
