import React from 'react'
import { userSettingsAreaRoutes } from '../../../user/settings/routes'
import { UserSettingsAreaRoute } from '../../../user/settings/UserSettingsArea'
import { asyncComponent } from '../../../util/asyncComponent'
import { SHOW_BUSINESS_FEATURES } from '../../dotcom/productSubscriptions/features'
import { authExp } from '../../site-admin/SiteAdminAuthenticationProvidersPage'

const UserSettingsExternalAccountsPage = asyncComponent(
    () => import('./UserSettingsExternalAccountsPage'),
    'UserSettingsExternalAccountsPage'
)
const UserSubscriptionsEditProductSubscriptionPage = asyncComponent(() => import("../productSubscriptions/UserSubscriptionsEditProductSubscriptionPage"), "UserSubscriptionsEditProductSubscriptionPage")
const UserSubscriptionsNewProductSubscriptionPage = asyncComponent(() => import("../productSubscriptions/UserSubscriptionsNewProductSubscriptionPage"), "UserSubscriptionsNewProductSubscriptionPage")
const UserSubscriptionsProductSubscriptionPage = asyncComponent(() => import("../productSubscriptions/UserSubscriptionsProductSubscriptionPage"), "UserSubscriptionsProductSubscriptionPage")
const UserSubscriptionsProductSubscriptionsPage = asyncComponent(() => import("../productSubscriptions/UserSubscriptionsProductSubscriptionsPage"), "UserSubscriptionsProductSubscriptionsPage")

export const enterpriseUserSettingsAreaRoutes: ReadonlyArray<UserSettingsAreaRoute> = [
    ...userSettingsAreaRoutes,
    {
        path: '/external-accounts',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserSettingsExternalAccountsPage {...props} />,
        condition: () => authExp,
    },
    {
        path: '/subscriptions/new',
        exact: true,
        render: props => <UserSubscriptionsNewProductSubscriptionPage {...props} />,
        condition: () => SHOW_BUSINESS_FEATURES,
    },
    {
        path: '/subscriptions/:subscriptionUUID',
        exact: true,
        render: props => <UserSubscriptionsProductSubscriptionPage {...props} />,
        condition: () => SHOW_BUSINESS_FEATURES,
    },
    {
        path: '/subscriptions/:subscriptionUUID/edit',
        exact: true,
        render: props => <UserSubscriptionsEditProductSubscriptionPage {...props} />,
        condition: () => SHOW_BUSINESS_FEATURES,
    },
    {
        path: '/subscriptions',
        exact: true,
        render: props => <UserSubscriptionsProductSubscriptionsPage {...props} />,
        condition: () => SHOW_BUSINESS_FEATURES,
    },
]
