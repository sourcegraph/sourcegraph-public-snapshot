import React from 'react'
import { userAreaRoutes } from '../../user/area/routes'
import { UserAreaRoute } from '../../user/area/UserArea'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'
const UserSubscriptionsEditProductSubscriptionPage = React.lazy(async () => ({
    default: (await import('./productSubscriptions/UserSubscriptionsEditProductSubscriptionPage'))
        .UserSubscriptionsEditProductSubscriptionPage,
}))
const UserSubscriptionsNewProductSubscriptionPage = React.lazy(async () => ({
    default: (await import('./productSubscriptions/UserSubscriptionsNewProductSubscriptionPage'))
        .UserSubscriptionsNewProductSubscriptionPage,
}))
const UserSubscriptionsProductSubscriptionPage = React.lazy(async () => ({
    default: (await import('./productSubscriptions/UserSubscriptionsProductSubscriptionPage'))
        .UserSubscriptionsProductSubscriptionPage,
}))
const UserSubscriptionsProductSubscriptionsPage = React.lazy(async () => ({
    default: (await import('./productSubscriptions/UserSubscriptionsProductSubscriptionsPage'))
        .UserSubscriptionsProductSubscriptionsPage,
}))

export const enterpriseUserAreaRoutes: ReadonlyArray<UserAreaRoute> = [
    ...userAreaRoutes,
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
