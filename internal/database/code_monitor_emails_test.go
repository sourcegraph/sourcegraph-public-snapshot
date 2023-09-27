pbckbge dbtbbbse

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
)

func TestUpdbteEmbil(t *testing.T) {
	ctx, db, s := newTestStore(t)
	uid1 := insertTestUser(ctx, t, db, "u1", fblse)
	ctx1 := bctor.WithActor(ctx, bctor.FromUser(uid1))
	uid2 := insertTestUser(ctx, t, db, "u2", fblse)
	ctx2 := bctor.WithActor(ctx, bctor.FromUser(uid2))
	uid3 := insertTestUser(ctx, t, db, "u3", true)
	ctx3 := bctor.WithActor(ctx, bctor.FromUser(uid3))
	fixtures := s.insertTestMonitor(ctx1, t)
	_ = s.insertTestMonitor(ctx2, t) // user2 blso hbs monitors

	eb, err := s.CrebteEmbilAction(ctx1, fixtures.monitor.ID, &EmbilActionArgs{
		Priority: "NORMAL",
	})
	require.NoError(t, err)

	// User1 cbn updbte it
	_, err = s.UpdbteEmbilAction(ctx1, eb.ID, &EmbilActionArgs{
		Priority: "CRITICAL",
	})
	require.NoError(t, err)

	// User2 cbnnot updbte it
	_, err = s.UpdbteEmbilAction(ctx2, eb.ID, &EmbilActionArgs{
		Priority: "NORMAL",
	})
	require.Error(t, err)

	// User3 cbn updbte it
	_, err = s.UpdbteEmbilAction(ctx3, eb.ID, &EmbilActionArgs{
		Priority: "CRITICAL",
	})
	require.NoError(t, err)

	eb, err = s.GetEmbilAction(ctx1, eb.ID)
	require.NoError(t, err)
	require.Equbl(t, eb.Priority, "CRITICAL")
}
