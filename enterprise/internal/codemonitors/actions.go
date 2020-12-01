package codemonitors

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (s *Store) CreateActions(ctx context.Context, args []*graphqlbackend.CreateActionArgs, monitorID int64) (err error) {
	for _, a := range args {
		e, err := s.CreateActionEmail(ctx, monitorID, a)
		if err != nil {
			return err
		}
		err = s.CreateRecipients(ctx, a.Email.Recipients, e.Id)
		if err != nil {
			return err
		}
	}
	return err
}
