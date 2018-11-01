import React from 'react'
import { siteAdminAreaRoutes } from '../../../packages/webapp/src/site-admin/routes'
import { SiteAdminAreaRoute } from '../../../packages/webapp/src/site-admin/SiteAdminArea'
import { SiteAdminProductCustomersPage } from './dotcom/customers/SiteAdminCustomersPage'
import { SiteAdminCreateProductSubscriptionPage } from './dotcom/productSubscriptions/SiteAdminCreateProductSubscriptionPage'
import { SiteAdminProductLicensesPage } from './dotcom/productSubscriptions/SiteAdminProductLicensesPage'
import { SiteAdminProductSubscriptionPage as SiteAdminDotcomProductSubscriptionPage } from './dotcom/productSubscriptions/SiteAdminProductSubscriptionPage'
import { SiteAdminProductSubscriptionsPage } from './dotcom/productSubscriptions/SiteAdminProductSubscriptionsPage'
import { SiteAdminProductSubscriptionPage } from './productSubscription/SiteAdminProductSubscriptionPage'
import { SiteAdminAuthenticationProvidersPage } from './SiteAdminAuthenticationProvidersPage'
import { SiteAdminExternalAccountsPage } from './SiteAdminExternalAccountsPage'
import { SiteAdminRegistryExtensionsPage } from './SiteAdminRegistryExtensionsPage'

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
