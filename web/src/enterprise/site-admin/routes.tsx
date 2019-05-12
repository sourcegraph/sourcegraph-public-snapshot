import React from 'react'
import { siteAdminAreaRoutes } from '../../site-admin/routes'
import { SiteAdminAreaRoute } from '../../site-admin/SiteAdminArea'
import { asyncComponent } from '../../util/asyncComponent'

const SiteAdminProductCustomersPage = asyncComponent(
    () => import('./dotcom/customers/SiteAdminCustomersPage'),
    'SiteAdminProductCustomersPage'
)
const SiteAdminCreateProductSubscriptionPage = asyncComponent(
    () => import('./dotcom/productSubscriptions/SiteAdminCreateProductSubscriptionPage'),
    'SiteAdminCreateProductSubscriptionPage'
)
const SiteAdminProductLicensesPage = asyncComponent(
    () => import('./dotcom/productSubscriptions/SiteAdminProductLicensesPage'),
    'SiteAdminProductLicensesPage'
)
const SiteAdminDotcomProductSubscriptionPage = asyncComponent(
    () => import('./dotcom/productSubscriptions/SiteAdminProductSubscriptionPage'),
    'SiteAdminProductSubscriptionPage'
)
const SiteAdminProductSubscriptionsPage = asyncComponent(
    () => import('./dotcom/productSubscriptions/SiteAdminProductSubscriptionsPage'),
    'SiteAdminProductSubscriptionsPage'
)
const SiteAdminProductSubscriptionPage = asyncComponent(
    () => import('./productSubscription/SiteAdminProductSubscriptionPage'),
    'SiteAdminProductSubscriptionPage'
)
const SiteAdminAuthenticationProvidersPage = asyncComponent(
    () => import('./SiteAdminAuthenticationProvidersPage'),
    'SiteAdminAuthenticationProvidersPage'
)
const SiteAdminExternalAccountsPage = asyncComponent(
    () => import('./SiteAdminExternalAccountsPage'),
    'SiteAdminExternalAccountsPage'
)
const SiteAdminRegistryExtensionsPage = asyncComponent(
    () => import('./SiteAdminRegistryExtensionsPage'),
    'SiteAdminRegistryExtensionsPage'
)

export const enterpriseSiteAdminAreaRoutes: ReadonlyArray<SiteAdminAreaRoute> = [
    ...siteAdminAreaRoutes,
    {
        path: '/license',
        render: props => <SiteAdminProductSubscriptionPage {...props} />,
        exact: true,
    },
    {
        path: '/dotcom/customers',
        render: props => <SiteAdminProductCustomersPage {...props} />,
        exact: true,
    },
    {
        path: '/dotcom/product/subscriptions/new',
        render: props => <SiteAdminCreateProductSubscriptionPage {...props} />,
        exact: true,
    },
    {
        path: '/dotcom/product/subscriptions/:subscriptionUUID',
        render: props => <SiteAdminDotcomProductSubscriptionPage {...props} />,
        exact: true,
    },
    {
        path: '/dotcom/product/subscriptions',
        render: props => <SiteAdminProductSubscriptionsPage {...props} />,
        exact: true,
    },
    {
        path: '/dotcom/product/licenses',
        render: props => <SiteAdminProductLicensesPage {...props} />,
        exact: true,
    },
    {
        path: '/auth/providers',
        render: props => <SiteAdminAuthenticationProvidersPage {...props} />,
        exact: true,
    },
    {
        path: '/auth/external-accounts',
        render: props => <SiteAdminExternalAccountsPage {...props} />,
        exact: true,
    },
    {
        path: '/registry/extensions',
        render: props => <SiteAdminRegistryExtensionsPage {...props} />,
        exact: true,
    },
]
