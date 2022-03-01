package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListRecipients(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, _, userCTX := newTestUser(ctx, t, db)
	fixtures := s.insertTestMonitor(userCTX, t)

	rs, err := s.ListRecipients(ctx, ListRecipientsOpts{EmailID: &fixtures.emails[0].ID})
	require.NoError(t, err)

	require.Equal(t, []*Recipient{fixtures.recipients[0]}, rs)
}
