package database

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
)

func TestUpdateEmail(t *testing.T) {
	ctx, db, s := newTestStore(t)
	uid1 := insertTestUser(ctx, t, db, "u1", false)
	ctx1 := actor.WithActor(ctx, actor.FromUser(uid1))
	uid2 := insertTestUser(ctx, t, db, "u2", false)
	ctx2 := actor.WithActor(ctx, actor.FromUser(uid2))
	uid3 := insertTestUser(ctx, t, db, "u3", true)
	ctx3 := actor.WithActor(ctx, actor.FromUser(uid3))
	fixtures := s.insertTestMonitor(ctx1, t)
	_ = s.insertTestMonitor(ctx2, t) // user2 also has monitors

	ea, err := s.CreateEmailAction(ctx1, fixtures.monitor.ID, &EmailActionArgs{
		Priority: "NORMAL",
	})
	require.NoError(t, err)

	// User1 can update it
	_, err = s.UpdateEmailAction(ctx1, ea.ID, &EmailActionArgs{
		Priority: "CRITICAL",
	})
	require.NoError(t, err)

	// User2 cannot update it
	_, err = s.UpdateEmailAction(ctx2, ea.ID, &EmailActionArgs{
		Priority: "NORMAL",
	})
	require.Error(t, err)

	// User3 can update it
	_, err = s.UpdateEmailAction(ctx3, ea.ID, &EmailActionArgs{
		Priority: "CRITICAL",
	})
	require.NoError(t, err)

	ea, err = s.GetEmailAction(ctx1, ea.ID)
	require.NoError(t, err)
	require.Equal(t, ea.Priority, "CRITICAL")
}
