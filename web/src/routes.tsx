import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { LayoutProps } from './Layout'
import { parseSearchURLQuery } from './search'
import { asyncComponent } from './util/asyncComponent'

const SearchPage = asyncComponent(
    () => import('./search/input/SearchPage'),
    'SearchPage',
    require.resolveWeak('./search/input/SearchPage')
)
const SearchResults = asyncComponent(
    () => import('./search/results/SearchResults'),
    'SearchResults',
    require.resolveWeak('./search/results/SearchResults')
)
const SiteAdminArea = asyncComponent(
    () => import('./site-admin/SiteAdminArea'),
    'SiteAdminArea',
    require.resolveWeak('./site-admin/SiteAdminArea')
)
const ExtensionsArea = asyncComponent(
    () => import('./extensions/ExtensionsArea'),
    'ExtensionsArea',
    require.resolveWeak('./extensions/ExtensionsArea')
)

export interface LayoutRouteComponentProps extends RouteComponentProps<any>, LayoutProps {}

export interface LayoutRouteProps {
    path: string
    exact?: boolean
    render: React.ComponentType<LayoutRouteComponentProps>

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
    render: asyncComponent(
        () => import('./repo/RepoContainer'),
        'RepoContainer',
        require.resolveWeak('./repo/RepoContainer')
    ),
}

/**
 * Holds all top-level routes for the app because both the navbar and the main content area need to
 * switch over matched path.
 *
 * See https://reacttraining.com/react-router/web/example/sidebar
 */
export const routes: ReadonlyArray<LayoutRouteProps> = [
    {
        path: '/',
        render: (props: any) =>
            window.context.sourcegraphDotComMode && !props.user ? (
                <Redirect to="/welcome" />
            ) : (
                <Redirect to="/search" />
            ),
        exact: true,
    },
    {
        path: '/search',
        render: (props: any) =>
            parseSearchURLQuery(props.location.search) ? (
                <SearchResults {...props} deployType={window.context.deployType} />
            ) : (
                <SearchPage {...props} />
            ),
        exact: true,
    },
    {
        path: '/search/searches',
        render: asyncComponent(
            () => import('./search/saved-queries/SavedQueries'),
            'SavedQueriesPage',
            require.resolveWeak('./search/saved-queries/SavedQueries')
        ),
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/open',
        render: asyncComponent(() => import('./open/OpenPage'), 'OpenPage', require.resolveWeak('./open/OpenPage')),
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/sign-in',
        render: asyncComponent(
            () => import('./auth/SignInPage'),
            'SignInPage',
            require.resolveWeak('./auth/SignInPage')
        ),
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/sign-up',
        render: asyncComponent(
            () => import('./auth/SignUpPage'),
            'SignUpPage',
            require.resolveWeak('./auth/SignUpPage')
        ),
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/settings',
        render: asyncComponent(
            () => import('./user/settings/RedirectToUserSettings'),
            'RedirectToUserSettings',
            require.resolveWeak('./user/settings/RedirectToUserSettings')
        ),
    },
    {
        path: '/user',
        render: asyncComponent(
            () => import('./user/settings/RedirectToUserPage'),
            'RedirectToUserPage',
            require.resolveWeak('./user/settings/RedirectToUserPage')
        ),
    },
    {
        path: '/organizations',
        render: asyncComponent(() => import('./org/OrgsArea'), 'OrgsArea', require.resolveWeak('./org/OrgsArea')),
    },
    {
        path: '/search',
        render: props => <SearchResults {...props} deployType={window.context.deployType} />,
        exact: true,
    },
    {
        path: '/site-admin/init',
        exact: true,
        render: asyncComponent(
            () => import('./site-admin/SiteInitPage'),
            'SiteInitPage',
            require.resolveWeak('./site-admin/SiteInitPage')
        ),
        forceNarrowWidth: false,
    },
    {
        path: '/site-admin',
        render: props => (
            <SiteAdminArea
                {...props}
                routes={props.siteAdminAreaRoutes}
                sideBarGroups={props.siteAdminSideBarGroups}
                overviewComponents={props.siteAdminOverviewComponents}
            />
        ),
    },
    {
        path: '/password-reset',
        render: asyncComponent(
            () => import('./auth/ResetPasswordPage'),
            'ResetPasswordPage',
            require.resolveWeak('./auth/ResetPasswordPage')
        ),
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/explore',
        render: asyncComponent(
            () => import('./explore/ExploreArea'),
            'ExploreArea',
            require.resolveWeak('./explore/ExploreArea')
        ),
        exact: true,
    },
    {
        path: '/discussions',
        render: asyncComponent(
            () => import('./discussions/DiscussionsPage'),
            'DiscussionsPage',
            require.resolveWeak('./discussions/DiscussionsPage')
        ),
        exact: true,
    },
    {
        path: '/search/scope/:id',
        render: asyncComponent(
            () => import('./search/input/ScopePage'),
            'ScopePage',
            require.resolveWeak('./search/input/ScopePage')
        ),
        exact: true,
    },
    {
        path: '/api/console',
        render: asyncComponent(() => import('./api/APIConsole'), 'APIConsole', require.resolveWeak('./api/APIConsole')),
        exact: true,
    },
    {
        path: '/users/:username',
        render: asyncComponent(
            () => import('./user/area/UserArea'),
            'UserArea',
            require.resolveWeak('./user/area/UserArea')
        ),
    },
    {
        path: '/survey/:score?',
        render: asyncComponent(
            () => import('./marketing/SurveyPage'),
            'SurveyPage',
            require.resolveWeak('./marketing/SurveyPage')
        ),
    },
    {
        path: '/extensions',
        render: props => <ExtensionsArea {...props} routes={props.extensionsAreaRoutes} />,
    },
    {
        path: '/help',
        render: () => {
            // Force a hard reload so that we delegate to the HTTP handler for /help, which handles
            // redirecting /help to https://docs.sourcegraph.com. That logic is not duplicated in
            // the web app because that would add complexity with no user benefit.
            //
            // TODO(sqs): This currently has a bug in dev mode where you can't go back to the app
            // after following the redirect. This will be fixed when we run docsite on
            // http://localhost:5080 in Procfile because then the redirect will be cross-domain and
            // won't reuse the same history stack.
            window.location.reload()
            return null
        },
    },
    {
        path: '/snippets',
        render: asyncComponent(
            () => import('./snippets/SnippetsPage'),
            'SnippetsPage',
            require.resolveWeak('./snippets/SnippetsPage')
        ),
    },
    repoRevRoute,
]
