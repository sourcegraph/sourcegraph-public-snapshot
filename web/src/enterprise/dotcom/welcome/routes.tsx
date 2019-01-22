import React from 'react'
import { WelcomeAreaRoute } from './WelcomeArea'
const WelcomeMainPage = React.lazy(async () => ({
    default: (await import('./WelcomeMainPage')).WelcomeMainPage,
}))
const WelcomeSearchPage = React.lazy(async () => ({
    default: (await import('./WelcomeSearchPage')).WelcomeSearchPage,
}))
const WelcomeCodeIntelligencePage = React.lazy(async () => ({
    default: (await import('./WelcomeCodeIntelligencePage')).WelcomeCodeIntelligencePage,
}))
const WelcomeIntegrationsPage = React.lazy(async () => ({
    default: (await import('./WelcomeIntegrationsPage')).WelcomeIntegrationsPage,
}))

export const welcomeAreaRoutes: ReadonlyArray<WelcomeAreaRoute> = [
    {
        path: '/',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <WelcomeMainPage {...props} />,
    },
    {
        path: '/search',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <WelcomeSearchPage {...props} />,
    },
    {
        path: '/code-intelligence',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <WelcomeCodeIntelligencePage {...props} />,
    },
    {
        path: '/integrations',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <WelcomeIntegrationsPage {...props} />,
    },
]
