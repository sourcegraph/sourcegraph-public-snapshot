package sourcegraphaccounts

import (
	"context"
	"strings"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/v2/testutil/require"
	"github.com/sourcegraph/log/logtest"
	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	notificationsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/notifications/v1"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestNotificationsSubscriberHandlers_onUserDeleted(t *testing.T) {
	ctx := context.Background()
	samsProvider := &schema.OpenIDConnectAuthProvider{}
	t.Run("user still exists", func(t *testing.T) {
		logger, exportLogs := logtest.Captured(t)
		store := NewMockNotificationsSubscriberStore()
		store.GetSAMSUserByIDFunc.SetDefaultReturn(&clientsv1.User{}, nil)
		hs := newNotificationsSubscriberHandlers(logger, store, samsProvider)
		err := hs.onUserDeleted()(ctx, &notificationsv1.UserDeletedData{AccountID: "018d21f2-14d7-7c3b-8714-10cdd32dd1ab"})
		require.NoError(t, err)

		foundLog := false
		for _, log := range exportLogs() {
			if strings.Contains(log.Message, "deleted notification for a user still exists") {
				foundLog = true
				break
			}
		}
		require.True(t, foundLog)
	})

	t.Run("perform user deletion", func(t *testing.T) {
		store := NewMockNotificationsSubscriberStore()
		store.GetSAMSUserByIDFunc.SetDefaultReturn(nil, sams.ErrNotFound)
		store.ListUserExternalAccountsFunc.SetDefaultReturn(
			[]*extsvc.Account{
				{UserID: 1},
			},
			nil,
		)
		hs := newNotificationsSubscriberHandlers(logtest.NoOp(t), store, samsProvider)
		err := hs.onUserDeleted()(ctx, &notificationsv1.UserDeletedData{AccountID: "018d21f2-14d7-7c3b-8714-10cdd32dd1ab"})
		require.NoError(t, err)
		mockrequire.Called(t, store.ListUserExternalAccountsFunc)
		mockrequire.Called(t, store.HardDeleteUsersFunc)
	})
}
