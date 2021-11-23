package codemonitors

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (s *codeMonitorStore) CreateActions(ctx context.Context, args []*graphqlbackend.CreateActionArgs, monitorID int64) error {
	for _, a := range args {
		e, err := s.CreateEmailAction(ctx, monitorID, a)
		if err != nil {
			return err
		}
		err = s.CreateRecipients(ctx, a.Email.Recipients, e.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
