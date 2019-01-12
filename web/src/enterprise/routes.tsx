import React from 'react'
import { LayoutRouteProps, routes } from '../routes'
const NewProductSubscriptionPageOrRedirectUser = React.lazy(async () => ({
    default: (await import('./user/productSubscriptions/NewProductSubscriptionPageOrRedirectUser'))
        .NewProductSubscriptionPageOrRedirectUser,
}))

export const enterpriseRoutes: ReadonlyArray<LayoutRouteProps> = [
    {
        // Allow unauthenticated viewers to view the "new subscription" page to price out a subscription (instead
        // of just dumping them on a sign-in page).
        path: '/user/subscriptions/new',
        exact: true,
        render: props => <NewProductSubscriptionPageOrRedirectUser {...props} />,
    },
    ...routes,
]
