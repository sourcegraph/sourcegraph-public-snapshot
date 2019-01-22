import React from 'react'
import { Redirect } from 'react-router'
import { LayoutRouteProps, routes } from '../routes'
import { welcomeAreaRoutes } from './dotcom/welcome/routes'
const WelcomeArea = React.lazy(async () => ({
    default: (await import('./dotcom/welcome/WelcomeArea')).WelcomeArea,
}))
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
    {
        path: '/start',
        render: () => <Redirect to="/welcome" />,
        exact: true,
    },
    {
        path: '/welcome',
        render: props => <WelcomeArea {...props} routes={welcomeAreaRoutes} />,
    },
    ...routes,
]
