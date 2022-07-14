import React from 'react'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { userSettingsAreaRoutes } from '../../../user/settings/routes'
import { UserSettingsAreaRoute } from '../../../user/settings/UserSettingsArea'
import { SHOW_BUSINESS_FEATURES } from '../../dotcom/productSubscriptions/features'
import { authExp } from '../../site-admin/utils'

const UserSettingsPermissionsPage = lazyComponent(
    () => import('./auth/UserSettingsPermissionsPage'),
    'UserSettingsPermissionsPage'
)
const UserEventLogsPage = React.lazy(() => import('./UserEventLogsPage'))

export const enterpriseUserSettingsAreaRoutes: readonly UserSettingsAreaRoute[] = [
    ...userSettingsAreaRoutes,
    {
        path: '/permissions',
        exact: true,
        render: props => <UserSettingsPermissionsPage {...props} />,
        condition: ({ authenticatedUser }) => authenticatedUser.siteAdmin,
    },
    {
        path: '/event-log',
        exact: true,
        render: props => <UserEventLogsPage {...props} />,
    },
    {
        path: '/external-accounts',
        exact: true,
        render: lazyComponent(() => import('./UserSettingsExternalAccountsPage'), 'UserSettingsExternalAccountsPage'),
        condition: () => authExp,
    },
    {
        path: '/batch-changes',
        exact: true,
        render: lazyComponent(
            () => import('../../batches/settings/BatchChangesSettingsArea'),
            'BatchChangesSettingsArea'
        ),
        condition: ({ batchChangesEnabled, user: { viewerCanAdminister } }) =>
            batchChangesEnabled && viewerCanAdminister,
    },
    {
        path: '/subscriptions/new',
        exact: true,
        render: lazyComponent(
            () => import('../productSubscriptions/UserSubscriptionsNewProductSubscriptionPage'),
            'UserSubscriptionsNewProductSubscriptionPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
    },
    {
        path: '/subscriptions/:subscriptionUUID',
        exact: true,
        render: lazyComponent(
            () => import('../productSubscriptions/UserSubscriptionsProductSubscriptionPage'),
            'UserSubscriptionsProductSubscriptionPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
    },
    {
        path: '/subscriptions/:subscriptionUUID/edit',
        exact: true,
        render: lazyComponent(
            () => import('../productSubscriptions/UserSubscriptionsEditProductSubscriptionPage'),
            'UserSubscriptionsEditProductSubscriptionPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
    },
    {
        path: '/subscriptions',
        exact: true,
        render: React.lazy(() => import('../productSubscriptions/UserSubscriptionsProductSubscriptionsPage')),
        condition: () => SHOW_BUSINESS_FEATURES,
    },
]
