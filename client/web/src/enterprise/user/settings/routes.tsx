import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { userSettingsAreaRoutes } from '../../../user/settings/routes'
import { UserSettingsAreaRoute } from '../../../user/settings/UserSettingsArea'
import { SHOW_BUSINESS_FEATURES } from '../../dotcom/productSubscriptions/features'
import type { ExecutorsUserAreaProps } from '../../executors/ExecutorsUserArea'
import { authExp } from '../../site-admin/SiteAdminAuthenticationProvidersPage'

import { UserEventLogsPageProps } from './UserEventLogsPage'

const ExecutorsUserArea = lazyComponent<ExecutorsUserAreaProps, 'ExecutorsUserArea'>(
    () => import('../../executors/ExecutorsUserArea'),
    'ExecutorsUserArea'
)

export const enterpriseUserSettingsAreaRoutes: readonly UserSettingsAreaRoute[] = [
    ...userSettingsAreaRoutes,
    {
        path: '/permissions',
        exact: true,
        render: lazyComponent(() => import('./auth/UserSettingsPermissionsPage'), 'UserSettingsPermissionsPage'),
        condition: ({ authenticatedUser }) => authenticatedUser.siteAdmin,
    },
    {
        path: '/event-log',
        exact: true,
        render: lazyComponent<UserEventLogsPageProps, 'UserEventLogsPage'>(
            () => import('./UserEventLogsPage'),
            'UserEventLogsPage'
        ),
    },
    {
        path: '/external-accounts',
        exact: true,
        render: lazyComponent(() => import('./UserSettingsExternalAccountsPage'), 'UserSettingsExternalAccountsPage'),
        condition: () => authExp,
    },
    {
        path: '/executors',
        render: props => <ExecutorsUserArea {...props} namespaceID={props.user.id} />,
        condition: ({ user: { viewerCanAdminister } }) => viewerCanAdminister,
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
        path: '/subscriptions/:subscriptionUUID',
        exact: true,
        render: lazyComponent(
            () => import('../productSubscriptions/UserSubscriptionsProductSubscriptionPage'),
            'UserSubscriptionsProductSubscriptionPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
    },
    {
        path: '/subscriptions',
        exact: true,
        render: lazyComponent(
            () => import('../productSubscriptions/UserSubscriptionsProductSubscriptionsPage'),
            'UserSubscriptionsProductSubscriptionsPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
    },
]
