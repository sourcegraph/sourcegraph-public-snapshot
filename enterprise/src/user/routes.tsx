import React from 'react'
import { userAreaRoutes } from '../../../packages/webapp/src/user/area/routes'
import { UserAreaRoute } from '../../../packages/webapp/src/user/area/UserArea'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'
import { UserSubscriptionsEditProductSubscriptionPage } from './productSubscriptions/UserSubscriptionsEditProductSubscriptionPage'
import { UserSubscriptionsNewProductSubscriptionPage } from './productSubscriptions/UserSubscriptionsNewProductSubscriptionPage'
import { UserSubscriptionsProductSubscriptionPage } from './productSubscriptions/UserSubscriptionsProductSubscriptionPage'
import { UserSubscriptionsProductSubscriptionsPage } from './productSubscriptions/UserSubscriptionsProductSubscriptionsPage'

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
