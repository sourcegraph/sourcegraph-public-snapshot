import React from 'react'
import { userSettingsAreaRoutes } from '../../../user/settings/routes'
import { UserSettingsAreaRoute } from '../../../user/settings/UserSettingsArea'
import { SHOW_BUSINESS_FEATURES } from '../../dotcom/productSubscriptions/features'
import { authExp } from '../../site-admin/SiteAdminAuthenticationProvidersPage'
const UserSettingsExternalAccountsPage = React.lazy(async () => ({
    default: (await import('./UserSettingsExternalAccountsPage')).UserSettingsExternalAccountsPage,
}))
const UserSubscriptionsEditProductSubscriptionPage = React.lazy(async () => ({
    default: (await import('../productSubscriptions/UserSubscriptionsEditProductSubscriptionPage'))
        .UserSubscriptionsEditProductSubscriptionPage,
}))
const UserSubscriptionsNewProductSubscriptionPage = React.lazy(async () => ({
    default: (await import('../productSubscriptions/UserSubscriptionsNewProductSubscriptionPage'))
        .UserSubscriptionsNewProductSubscriptionPage,
}))
const UserSubscriptionsProductSubscriptionPage = React.lazy(async () => ({
    default: (await import('../productSubscriptions/UserSubscriptionsProductSubscriptionPage'))
        .UserSubscriptionsProductSubscriptionPage,
}))
const UserSubscriptionsProductSubscriptionsPage = React.lazy(async () => ({
    default: (await import('../productSubscriptions/UserSubscriptionsProductSubscriptionsPage'))
        .UserSubscriptionsProductSubscriptionsPage,
}))

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
