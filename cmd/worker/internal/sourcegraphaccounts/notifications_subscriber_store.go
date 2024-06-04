package sourcegraphaccounts

import (
	"context"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// notificationsSubscriberStore is the data layer carrier for notifications
// subscriber. This interface is meant to abstract away and limit the exposure
// of the underlying data layer to the handler through a thin-wrapper.
type notificationsSubscriberStore interface {
	// GetSAMSUserByID returns the SAMS user with the given ID. It returns
	// sams.ErrNotFound if no such user exists.
	//
	// Required scope: profile
	GetSAMSUserByID(ctx context.Context, accountID string) (*clientsv1.User, error)

	// ListUserExternalAccounts returns the external accounts satisfying the given
	// options.
	ListUserExternalAccounts(ctx context.Context, opts database.ExternalAccountsListOptions) ([]*extsvc.Account, error)
	// HardDeleteUsers hard-deletes the users with the given IDs.
	HardDeleteUsers(ctx context.Context, userIDs []int32) error
}

type notificationsSubscriberStoreImpl struct {
	samsClient *sams.ClientV1
	db         database.DB
}

func newNotificationsSubscriberStore(samsClient *sams.ClientV1, db database.DB) notificationsSubscriberStore {
	return &notificationsSubscriberStoreImpl{
		samsClient: samsClient,
		db:         db,
	}
}

func (s *notificationsSubscriberStoreImpl) GetSAMSUserByID(ctx context.Context, accountID string) (*clientsv1.User, error) {
	return s.samsClient.Users().GetUserByID(ctx, accountID)
}

func (s *notificationsSubscriberStoreImpl) ListUserExternalAccounts(ctx context.Context, opts database.ExternalAccountsListOptions) ([]*extsvc.Account, error) {
	return s.db.UserExternalAccounts().List(ctx, opts)
}

func (s *notificationsSubscriberStoreImpl) HardDeleteUsers(ctx context.Context, userIDs []int32) error {
	return s.db.Users().HardDeleteList(ctx, userIDs)
}
