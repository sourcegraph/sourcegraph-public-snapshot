import { userSettingsAreaRoutes } from '../../../user/settings/routes'
import { UserSettingsAreaRoute } from '../../../user/settings/UserSettingsArea'
import { asyncComponent } from '../../../util/asyncComponent'
import { SHOW_BUSINESS_FEATURES } from '../../dotcom/productSubscriptions/features'
import { authExp } from '../../site-admin/SiteAdminAuthenticationProvidersPage'

export const enterpriseUserSettingsAreaRoutes: ReadonlyArray<UserSettingsAreaRoute> = [
    ...userSettingsAreaRoutes,
    {
        path: '/external-accounts',
        exact: true,
        render: asyncComponent(() => import('./UserSettingsExternalAccountsPage'), 'UserSettingsExternalAccountsPage'),
        condition: () => authExp,
    },
    {
        path: '/subscriptions/new',
        exact: true,
        render: asyncComponent(
            () => import('../productSubscriptions/UserSubscriptionsNewProductSubscriptionPage'),
            'UserSubscriptionsNewProductSubscriptionPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
    },
    {
        path: '/subscriptions/:subscriptionUUID',
        exact: true,
        render: asyncComponent(
            () => import('../productSubscriptions/UserSubscriptionsProductSubscriptionPage'),
            'UserSubscriptionsProductSubscriptionPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
    },
    {
        path: '/subscriptions/:subscriptionUUID/edit',
        exact: true,
        render: asyncComponent(
            () => import('../productSubscriptions/UserSubscriptionsEditProductSubscriptionPage'),
            'UserSubscriptionsEditProductSubscriptionPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
    },
    {
        path: '/subscriptions',
        exact: true,
        render: asyncComponent(
            () => import('../productSubscriptions/UserSubscriptionsProductSubscriptionsPage'),
            'UserSubscriptionsProductSubscriptionsPage'
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
    },
]
