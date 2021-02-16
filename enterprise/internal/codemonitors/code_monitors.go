package codemonitors

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (s *Store) CreateCodeMonitor(ctx context.Context, args *graphqlbackend.CreateCodeMonitorArgs) (m *Monitor, err error) {
	// Start transaction.
	var txStore *Store
	txStore, err = s.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = txStore.Done(err) }()

	// Create monitor.
	m, err = txStore.CreateMonitor(ctx, args.Monitor)
	if err != nil {
		return nil, err
	}

	// Create trigger.
	err = txStore.CreateTriggerQuery(ctx, m.ID, args.Trigger)
	if err != nil {
		return nil, err
	}

	// Create actions.
	err = txStore.CreateActions(ctx, args.Actions, m.ID)
	if err != nil {
		return nil, err
	}
	return m, err
}
