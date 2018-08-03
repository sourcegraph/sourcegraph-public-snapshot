import * as React from 'react'
import { RouteProps } from 'react-router'
import { APIConsole } from './api/APIConsole'
import { ResetPasswordPage } from './auth/ResetPasswordPage'
import { SignInPage } from './auth/SignInPage'
import { SignUpPage } from './auth/SignUpPage'
import { ErrorNotSupportedPage } from './components/ErrorNotSupportedPage'
import { ExplorePage } from './explore/ExplorePage'
import { ExtensionsArea } from './extensions/ExtensionsArea'
import { SurveyPage } from './marketing/SurveyPage'
import { OpenPage } from './open/OpenPage'
import { OrgsArea } from './org/OrgsArea'
import { RepoContainer } from './repo/RepoContainer'
import { parseSearchURLQuery } from './search'
import { ScopePage } from './search/input/ScopePage'
import { SearchPage } from './search/input/SearchPage'
import { SearchResults } from './search/results/SearchResults'
import { SavedQueriesPage } from './search/saved-queries/SavedQueries'
import { SiteAdminArea } from './site-admin/SiteAdminArea'
import { SiteInitPage } from './site-admin/SiteInitPage'
import { RedirectToUserAccount } from './user/account/RedirectToUserAccount'
import { UserArea } from './user/area/UserArea'
import { canListAllRepositories } from './util/features'

interface LayoutRouteProps extends RouteProps {
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
        path: '/search/searches',
        component: SavedQueriesPage,
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
        component: RedirectToUserAccount,
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
        path: '/password-reset',
        component: ResetPasswordPage,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/explore',
        component: canListAllRepositories ? ExplorePage : ErrorNotSupportedPage,
        exact: true,
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
    {
        path: '/users/:username',
        component: UserArea,
    },
    {
        path: '/survey/:score?',
        component: SurveyPage,
    },
    ...(window.context.platformEnabled
        ? [
              {
                  path: '/extensions',
                  component: ExtensionsArea,
              },
          ]
        : []),
    repoRevRoute,
]
