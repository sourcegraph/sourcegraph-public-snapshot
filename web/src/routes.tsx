import { RouteProps } from 'react-router'
import { PasswordResetPage } from './auth/PasswordResetPage'
import { SignInPage } from './auth/SignInPage'
import { RepositoryResolver } from './repo/RepositoryResolver'
import { SearchResults } from './search/SearchResults'
import { SettingsPage } from './settings/SettingsPage'

export interface LayoutRouteProps extends RouteProps {
    component: React.ComponentType<any>
}

/**
 * Holds all top-level routes for the app because both the navbar and the main content area need to
 * switch over matched path.
 *
 * See https://reacttraining.com/react-router/web/example/sidebar
 */
export const routes: LayoutRouteProps[] = [
    {
        path: '/search',
        component: SearchResults,
        exact: true
    },
    {
        path: '/sign-in',
        component: SignInPage,
        exact: true
    },
    {
        path: '/sign-up',
        component: SignInPage,
        exact: true
    },
    {
        path: '/settings',
        component: SettingsPage
    },
    {
        path: '/password-reset',
        component: PasswordResetPage,
        exact: true
    },
    {
        path: '/:repoRev+/-/blob/:filePath+',
        component: RepositoryResolver
    },
    {
        path: '/:repoRev+/-/tree/:filePath+',
        component: RepositoryResolver
    },
    {
        path: '/:repoRev+',
        component: RepositoryResolver
    }
]
