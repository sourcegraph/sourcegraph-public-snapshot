import React from 'react'
import { Redirect } from 'react-router'
import { LayoutRouteProps, routes } from '../routes'
import { asyncComponent } from '../util/asyncComponent'
import { welcomeAreaRoutes } from './dotcom/welcome/routes'

const WelcomeArea = asyncComponent(() => import('./dotcom/welcome/WelcomeArea'), 'WelcomeArea')
const NewProductSubscriptionPageOrRedirectUser = asyncComponent(
    () => import('./user/productSubscriptions/NewProductSubscriptionPageOrRedirectUser'),
    'NewProductSubscriptionPageOrRedirectUser'
)

export const enterpriseRoutes: ReadonlyArray<LayoutRouteProps> = [
    {
        // Allow unauthenticated viewers to view the "new subscription" page to price out a subscription (instead
        // of just dumping them on a sign-in page).
        path: '/subscriptions/new',
        exact: true,
        render: props => <NewProductSubscriptionPageOrRedirectUser {...props} />,
    },
    {
        // Redirect from old /user/subscriptions/new -> /subscriptions/new.
        path: '/user/subscriptions/new',
        exact: true,
        render: () => <Redirect to="/subscriptions/new" />,
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
