import { lazyComponent } from '../../../util/lazyComponent'
import { WelcomeAreaRoute } from './WelcomeArea'

export const welcomeAreaRoutes: ReadonlyArray<WelcomeAreaRoute> = [
    {
        path: '/',
        exact: true,
        render: lazyComponent(() => import('./WelcomeMainPage'), 'WelcomeMainPage'),
    },
    // We will add more pages here soon. The other pages (search, code intel, integrations) were
    // removed to avoid blocking shipping of the new main welcome page.
]
