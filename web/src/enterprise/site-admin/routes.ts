import { siteAdminAreaRoutes } from '../../site-admin/routes'
import { SiteAdminAreaRoute } from '../../site-admin/SiteAdminArea'
import { lazyComponent } from '../../util/lazyComponent'

export const enterpriseSiteAdminAreaRoutes: readonly SiteAdminAreaRoute[] = [
    ...siteAdminAreaRoutes,
    {
        path: '/license',
        render: lazyComponent(
            () => import('./productSubscription/SiteAdminProductSubscriptionPage'),
            'SiteAdminProductSubscriptionPage'
        ),
        exact: true,
    },
    {
        path: '/dotcom/customers',
        render: lazyComponent(
            () => import('./dotcom/customers/SiteAdminCustomersPage'),
            'SiteAdminProductCustomersPage'
        ),
        exact: true,
    },
    {
        path: '/dotcom/product/subscriptions/new',
        render: lazyComponent(
            () => import('./dotcom/productSubscriptions/SiteAdminCreateProductSubscriptionPage'),
            'SiteAdminCreateProductSubscriptionPage'
        ),
        exact: true,
    },
    {
        path: '/dotcom/product/subscriptions/:subscriptionUUID',
        render: lazyComponent(
            () => import('./dotcom/productSubscriptions/SiteAdminProductSubscriptionPage'),
            'SiteAdminProductSubscriptionPage'
        ),
        exact: true,
    },
    {
        path: '/dotcom/product/subscriptions',
        render: lazyComponent(
            () => import('./dotcom/productSubscriptions/SiteAdminProductSubscriptionsPage'),
            'SiteAdminProductSubscriptionsPage'
        ),
        exact: true,
    },
    {
        path: '/dotcom/product/licenses',
        render: lazyComponent(
            () => import('./dotcom/productSubscriptions/SiteAdminProductLicensesPage'),
            'SiteAdminProductLicensesPage'
        ),
        exact: true,
    },
    {
        path: '/auth/providers',
        render: lazyComponent(
            () => import('./SiteAdminAuthenticationProvidersPage'),
            'SiteAdminAuthenticationProvidersPage'
        ),
        exact: true,
    },
    {
        path: '/auth/external-accounts',
        render: lazyComponent(() => import('./SiteAdminExternalAccountsPage'), 'SiteAdminExternalAccountsPage'),
        exact: true,
    },
    {
        path: '/registry/extensions',
        render: lazyComponent(() => import('./SiteAdminRegistryExtensionsPage'), 'SiteAdminRegistryExtensionsPage'),
        exact: true,
    },
    {
        path: '/lsif-uploads',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminLsifUploadsPage'), 'SiteAdminLsifUploadsPage'),
    },
    {
        path: '/lsif-uploads/:id',
        exact: true,
        render: lazyComponent(() => import('./SiteAdminLsifUploadPage'), 'SiteAdminLsifUploadPage'),
    },
]
