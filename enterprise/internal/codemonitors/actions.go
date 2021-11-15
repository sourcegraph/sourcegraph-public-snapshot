package codemonitors

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (s *codeMonitorStore) CreateActions(ctx context.Context, args []*graphqlbackend.CreateActionArgs, monitorID int64) error {
	for _, a := range args {
		if a.Email != nil {
			e, err := s.CreateEmailAction(ctx, monitorID, &EmailActionArgs{
				Enabled:  a.Email.Enabled,
				Priority: a.Email.Priority,
				Header:   a.Email.Header,
			})
			if err != nil {
				return err
			}
			err = s.CreateRecipients(ctx, a.Email.Recipients, e.ID)
			if err != nil {
				return err
			}
		}
		// TODO(camdencheek): add other action types (webhooks) here
	}
	return nil
}
