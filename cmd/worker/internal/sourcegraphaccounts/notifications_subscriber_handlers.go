package sourcegraphaccounts

import (
	"context"

	"github.com/sourcegraph/log"
	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	notificationsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/notifications/v1"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type notificationsSubscriberHandlers struct {
	logger       log.Logger
	store        notificationsSubscriberStore
	samsProvider *schema.OpenIDConnectAuthProvider
}

func newNotificationsSubscriberHandlers(
	logger log.Logger,
	store notificationsSubscriberStore,
	samsProvider *schema.OpenIDConnectAuthProvider,
) *notificationsSubscriberHandlers {
	return &notificationsSubscriberHandlers{
		logger:       logger,
		store:        store,
		samsProvider: samsProvider,
	}
}

func (hs *notificationsSubscriberHandlers) onUserDeleted() func(ctx context.Context, data *notificationsv1.UserDeletedData) error {
	return func(ctx context.Context, data *notificationsv1.UserDeletedData) error {
		// Double check with SAMS that the user is deleted
		_, err := hs.store.GetSAMSUserByID(ctx, data.AccountID)
		if err == nil {
			hs.logger.Error("Received user deleted notification for a user still exists", log.String("accountID", data.AccountID))
			return nil
		} else if !errors.Is(err, sams.ErrNotFound) {
			return errors.Wrap(err, "get user by ID")
		}

		// NOTE: Do not query with "ClientID" nor limit to 1 record because they are
		// irrelevant in this context as long as the SAMS instance matches
		// ("ServiceID"). No matter how many times we have rotated the client ID, we
		// should still delete all the user records with the matching account ID.
		extAccts, err := hs.store.ListUserExternalAccounts(
			ctx,
			database.ExternalAccountsListOptions{
				ServiceType: hs.samsProvider.Type,
				ServiceID:   hs.samsProvider.Issuer,
				AccountID:   data.AccountID,
			},
		)
		if err != nil {
			return errors.Wrap(err, "list user external accounts")
		}

		userIDs := make([]int32, 0, len(extAccts))
		for _, extAcct := range extAccts {
			userIDs = append(userIDs, extAcct.UserID)
		}
		err = hs.store.HardDeleteUsers(ctx, userIDs)
		if err != nil {
			return errors.Wrap(err, "hard-delete users")
		}

		hs.logger.Debug("User hard-deleted",
			log.String("accountID", data.AccountID),
			log.Int("recordsCount", len(userIDs)),
		)
		return nil
	}
}
