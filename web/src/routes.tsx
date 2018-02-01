import * as React from 'react'
import { RouteProps } from 'react-router'
import { APIConsole } from './api/APIConsole'
import { PasswordResetPage } from './auth/PasswordResetPage'
import { SignInPage } from './auth/SignInPage'
import { SignUpPage } from './auth/SignUpPage'
import { CommentsPage } from './comments/CommentsPage'
import { ThreadPage } from './comments/ThreadPage'
import { ErrorNotSupportedPage } from './components/ErrorNotSupportedPage'
import { ExplorePage } from './explore/ExplorePage'
import { OpenPage } from './open/OpenPage'
import { OrgsArea } from './org/OrgsArea'
import { RepoContainer } from './repo/RepoContainer'
import { parseSearchURLQuery } from './search'
import { SavedQueriesPage } from './search/SavedQueries'
import { ScopePage } from './search/ScopePage'
import { SearchPage } from './search/SearchPage'
import { SearchResults } from './search/SearchResults'
import { SettingsArea } from './settings/SettingsArea'
import { SiteAdminArea } from './site-admin/SiteAdminArea'
import { SiteInitPage } from './site-admin/SiteInitPage'
import { canListAllRepositories } from './util/features'

export interface LayoutRouteProps extends RouteProps {
    component?: React.ComponentType<any>
    render?: (props: any) => React.ReactNode

    /**
     * Whether or not to force the width of the page to be narrow.
     */
    forceNarrowWidth?: boolean
}

/**
 * Holds properties for repository+ routes.
 */
export const repoRevRoute: LayoutRouteProps = {
    path: '/:repoRevAndRest+',
    component: RepoContainer,
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
        render: (props: any) =>
            parseSearchURLQuery(props.location.search) ? <SearchResults {...props} /> : <SearchPage {...props} />,
        exact: true,
    },
    {
        path: '/search/queries',
        component: SavedQueriesPage,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/c/:ulid',
        component: CommentsPage,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/open',
        component: OpenPage,
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
        component: SignUpPage,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/settings',
        component: SettingsArea,
    },
    {
        path: '/organizations',
        component: OrgsArea,
    },
    {
        path: '/search',
        component: SearchResults,
        exact: true,
    },
    {
        path: '/site-admin/init',
        exact: true,
        component: SiteInitPage,
        forceNarrowWidth: false,
    },
    {
        path: '/site-admin',
        component: SiteAdminArea,
    },
    {
        path: '/threads/:threadID',
        exact: true,
        component: ThreadPage,
        forceNarrowWidth: true,
    },
    {
        path: '/password-reset',
        component: PasswordResetPage,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/explore',
        component: canListAllRepositories ? ExplorePage : ErrorNotSupportedPage,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/search/scope/:id',
        component: ScopePage,
        exact: true,
    },
    {
        path: '/api/console',
        component: APIConsole,
        exact: true,
    },
    repoRevRoute,
]
