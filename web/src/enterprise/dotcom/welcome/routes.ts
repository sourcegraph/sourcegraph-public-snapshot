import { asyncComponent } from '../../../util/asyncComponent'
import { WelcomeAreaRoute } from './WelcomeArea'

export const welcomeAreaRoutes: ReadonlyArray<WelcomeAreaRoute> = [
    {
        path: '/',
        exact: true,
        render: asyncComponent(() => import('./WelcomeMainPage'), 'WelcomeMainPage'),
    },
    // We will add more pages here soon. The other pages (search, code intel, integrations) were
    // removed to avoid blocking shipping of the new main welcome page.
]
