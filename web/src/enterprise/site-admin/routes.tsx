import React from 'react'
import { siteAdminAreaRoutes } from '../../site-admin/routes'
import { SiteAdminAreaRoute } from '../../site-admin/SiteAdminArea'
const SiteAdminProductCustomersPage = React.lazy(async () => ({
    default: (await import('./dotcom/customers/SiteAdminCustomersPage')).SiteAdminProductCustomersPage,
}))
const SiteAdminCreateProductSubscriptionPage = React.lazy(async () => ({
    default: (await import('./dotcom/productSubscriptions/SiteAdminCreateProductSubscriptionPage'))
        .SiteAdminCreateProductSubscriptionPage,
}))
const SiteAdminProductLicensesPage = React.lazy(async () => ({
    default: (await import('./dotcom/productSubscriptions/SiteAdminProductLicensesPage')).SiteAdminProductLicensesPage,
}))
const SiteAdminDotcomProductSubscriptionPage = React.lazy(async () => ({
    default: (await import('./dotcom/productSubscriptions/SiteAdminProductSubscriptionPage'))
        .SiteAdminProductSubscriptionPage,
}))
const SiteAdminProductSubscriptionsPage = React.lazy(async () => ({
    default: (await import('./dotcom/productSubscriptions/SiteAdminProductSubscriptionsPage'))
        .SiteAdminProductSubscriptionsPage,
}))
const SiteAdminProductSubscriptionPage = React.lazy(async () => ({
    default: (await import('./productSubscription/SiteAdminProductSubscriptionPage')).SiteAdminProductSubscriptionPage,
}))
const SiteAdminAuthenticationProvidersPage = React.lazy(async () => ({
    default: (await import('./SiteAdminAuthenticationProvidersPage')).SiteAdminAuthenticationProvidersPage,
}))
const SiteAdminExternalAccountsPage = React.lazy(async () => ({
    default: (await import('./SiteAdminExternalAccountsPage')).SiteAdminExternalAccountsPage,
}))
const SiteAdminRegistryExtensionsPage = React.lazy(async () => ({
    default: (await import('./SiteAdminRegistryExtensionsPage')).SiteAdminRegistryExtensionsPage,
}))

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
