import React from 'react'
import { Redirect } from 'react-router'
import { LayoutRouteProps, routes } from '../routes'
const WelcomeMainPage = React.lazy(async () => ({
    default: (await import('./dotcom/welcome/WelcomeMainPage')).WelcomeMainPage,
}))
const WelcomeSearchPage = React.lazy(async () => ({
    default: (await import('./dotcom/welcome/WelcomeSearchPage')).WelcomeSearchPage,
}))
const WelcomeCodeIntelligencePage = React.lazy(async () => ({
    default: (await import('./dotcom/welcome/WelcomeCodeIntelligencePage')).WelcomeCodeIntelligencePage,
}))
const WelcomeIntegrationsPage = React.lazy(async () => ({
    default: (await import('./dotcom/welcome/WelcomeIntegrationsPage')).WelcomeIntegrationsPage,
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
        render: props => <WelcomeMainPage {...props} />,
        exact: true,
    },
    {
        path: '/welcome/search',
        render: props => <WelcomeSearchPage {...props} />,
        exact: true,
    },
    {
        path: '/welcome/code-intelligence',
        render: props => <WelcomeCodeIntelligencePage {...props} />,
        exact: true,
    },
    {
        path: '/welcome/integrations',
        render: props => <WelcomeIntegrationsPage {...props} />,
        exact: true,
    },
    ...routes,
]
