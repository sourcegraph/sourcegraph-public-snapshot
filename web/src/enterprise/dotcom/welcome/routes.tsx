import React from 'react'
import { asyncComponent } from '../../../util/asyncComponent'
import { WelcomeAreaRoute } from './WelcomeArea'

const WelcomeMainPage = asyncComponent(() => import('./WelcomeMainPage'), 'WelcomeMainPage')

export const welcomeAreaRoutes: ReadonlyArray<WelcomeAreaRoute> = [
    {
        path: '/',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <WelcomeMainPage {...props} />,
    },
    // We will add more pages here soon. The other pages (search, code intel, integrations) were
    // removed to avoid blocking shipping of the new main welcome page.
]
