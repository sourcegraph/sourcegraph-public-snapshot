import { RouteProps } from 'react-router'
import { PasswordResetPage } from './auth/PasswordResetPage'
import { SignInPage } from './auth/SignInPage'
import { CommentsPage } from './comments/CommentsPage'
import { RepositoryResolver } from './repo/RepositoryResolver'
import { enableSearch2 } from './search'
import { SearchResults } from './search/SearchResults'
import { SearchResults as SearchResults2 } from './search2/SearchResults'
import { SettingsPage } from './settings/SettingsPage'

export interface LayoutRouteProps extends RouteProps {
    component: React.ComponentType<any>

    /**
     * Whether or not to force the width of the page to be narrow. Otherwise
     * this is controlled by the user's full-width toggle state, which is not
     * accessible on all pages.
     */
    forceNarrowWidth?: boolean
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
        component: enableSearch2 ? SearchResults2 : SearchResults,
        exact: true,
    },
    {
        path: '/c/:ulid',
        component: CommentsPage,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/sign-in',
        component: SignInPage,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/sign-up',
        component: SignInPage,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/settings',
        component: SettingsPage,
        forceNarrowWidth: true,
    },
    {
        path: '/password-reset',
        component: PasswordResetPage,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/:repoRev+/-/blob/:filePath+',
        component: RepositoryResolver,
    },
    {
        path: '/:repoRev+/-/tree/:filePath+',
        component: RepositoryResolver,
    },
    {
        path: '/:repoRev+',
        component: RepositoryResolver,
    },
]
