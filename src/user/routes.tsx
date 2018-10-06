import { userAreaRoutes } from '@sourcegraph/webapp/dist/user/area/routes'
import { UserAreaRoute } from '@sourcegraph/webapp/dist/user/area/UserArea'
import React from 'react'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'
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
        path: '/subscriptions/:subscriptionID',
        exact: true,
        render: props => <UserSubscriptionsProductSubscriptionPage {...props} />,
        condition: () => SHOW_BUSINESS_FEATURES,
    },
    {
        path: '/subscriptions',
        exact: true,
        render: props => <UserSubscriptionsProductSubscriptionsPage {...props} />,
        condition: () => SHOW_BUSINESS_FEATURES,
    },
]
