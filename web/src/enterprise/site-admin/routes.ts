import { siteAdminAreaRoutes } from '../../site-admin/routes'
import { SiteAdminAreaRoute } from '../../site-admin/SiteAdminArea'
import { asyncComponent } from '../../util/asyncComponent'

export const enterpriseSiteAdminAreaRoutes: ReadonlyArray<SiteAdminAreaRoute> = [
    ...siteAdminAreaRoutes,
    {
        path: '/license',
        render: asyncComponent(
            () => import('./productSubscription/SiteAdminProductSubscriptionPage'),
            'SiteAdminProductSubscriptionPage'
        ),
        exact: true,
    },
    {
        path: '/dotcom/customers',
        render: asyncComponent(
            () => import('./dotcom/customers/SiteAdminCustomersPage'),
            'SiteAdminProductCustomersPage'
        ),
        exact: true,
    },
    {
        path: '/dotcom/product/subscriptions/new',
        render: asyncComponent(
            () => import('./dotcom/productSubscriptions/SiteAdminCreateProductSubscriptionPage'),
            'SiteAdminCreateProductSubscriptionPage'
        ),
        exact: true,
    },
    {
        path: '/dotcom/product/subscriptions/:subscriptionUUID',
        render: asyncComponent(
            () => import('./dotcom/productSubscriptions/SiteAdminProductSubscriptionPage'),
            'SiteAdminProductSubscriptionPage'
        ),
        exact: true,
    },
    {
        path: '/dotcom/product/subscriptions',
        render: asyncComponent(
            () => import('./dotcom/productSubscriptions/SiteAdminProductSubscriptionsPage'),
            'SiteAdminProductSubscriptionsPage'
        ),
        exact: true,
    },
    {
        path: '/dotcom/product/licenses',
        render: asyncComponent(
            () => import('./dotcom/productSubscriptions/SiteAdminProductLicensesPage'),
            'SiteAdminProductLicensesPage'
        ),
        exact: true,
    },
    {
        path: '/auth/providers',
        render: asyncComponent(
            () => import('./SiteAdminAuthenticationProvidersPage'),
            'SiteAdminAuthenticationProvidersPage'
        ),
        exact: true,
    },
    {
        path: '/auth/external-accounts',
        render: asyncComponent(() => import('./SiteAdminExternalAccountsPage'), 'SiteAdminExternalAccountsPage'),
        exact: true,
    },
    {
        path: '/registry/extensions',
        render: asyncComponent(() => import('./SiteAdminRegistryExtensionsPage'), 'SiteAdminRegistryExtensionsPage'),
        exact: true,
    },
]
