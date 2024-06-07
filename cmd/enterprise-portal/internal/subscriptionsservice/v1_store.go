package subscriptionsservice

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
)

// StoreV1 is the data layer carrier for subscriptions service v1. This interface
// is meant to abstract away and limit the exposure of the underlying data layer
// to the handler through a thin-wrapper.
type StoreV1 interface {
	ListEnterpriseSubscriptionLicenses(context.Context, []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter, int) ([]*dotcomdb.LicenseAttributes, error)
}

type storeV1 struct {
	db       *database.DB
	dotcomDB *dotcomdb.Reader
}

type NewStoreV1Options struct {
	DB       *database.DB
	DotcomDB *dotcomdb.Reader
}

// NewStoreV1 returns a new StoreV1 using the given client and database handles.
func NewStoreV1(opts NewStoreV1Options) StoreV1 {
	return &storeV1{
		db:       opts.DB,
		dotcomDB: opts.DotcomDB,
	}
}

func (s *storeV1) ListEnterpriseSubscriptionLicenses(ctx context.Context, filters []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter, limit int) ([]*dotcomdb.LicenseAttributes, error) {
	return s.dotcomDB.ListEnterpriseSubscriptionLicenses(ctx, filters, limit)
}
