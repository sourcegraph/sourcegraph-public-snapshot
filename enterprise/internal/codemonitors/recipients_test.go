package codemonitors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllRecipientsForEmailIDInt64(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, id, _, userCTX := newTestUser(ctx, t, db)
	_, err := s.insertTestMonitor(userCTX, t)
	require.NoError(t, err)

	var (
		wantEmailID     int64 = 1
		wantRecipientID int64 = 1
	)
	rs, err := s.ListRecipients(ctx, ListRecipientsOpts{EmailID: &wantEmailID})
	require.NoError(t, err)

	want := []*Recipient{{
		ID:              wantRecipientID,
		Email:           wantEmailID,
		NamespaceUserID: &id,
		NamespaceOrgID:  nil,
	}}
	require.Equal(t, want, rs)
}
