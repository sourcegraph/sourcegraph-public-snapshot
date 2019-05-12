import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { LayoutProps } from './Layout'
import { parseSearchURLQuery } from './search'
import { asyncComponent } from './util/asyncComponent'

const SearchPage = asyncComponent(() => import('./search/input/SearchPage'), 'SearchPage')
const SearchResults = asyncComponent(() => import('./search/results/SearchResults'), 'SearchResults')
const SavedQueriesPage = asyncComponent(() => import('./search/saved-queries/SavedQueries'), 'SavedQueriesPage')
const SiteAdminArea = asyncComponent(() => import('./site-admin/SiteAdminArea'), 'SiteAdminArea')
const UserArea = asyncComponent(() => import('./user/area/UserArea'), 'UserArea')
const APIConsole = asyncComponent(() => import('./api/APIConsole'), 'APIConsole')
const ResetPasswordPage = asyncComponent(() => import('./auth/ResetPasswordPage'), 'ResetPasswordPage')
const SignInPage = asyncComponent(() => import('./auth/SignInPage'), 'SignInPage')
const SignUpPage = asyncComponent(() => import('./auth/SignUpPage'), 'SignUpPage')
const DiscussionsPage = asyncComponent(() => import('./discussions/DiscussionsPage'), 'DiscussionsPage')
const ExploreArea = asyncComponent(() => import('./explore/ExploreArea'), 'ExploreArea')
const ExtensionsArea = asyncComponent(() => import('./extensions/ExtensionsArea'), 'ExtensionsArea')
const SurveyPage = asyncComponent(() => import('./marketing/SurveyPage'), 'SurveyPage')
const OpenPage = asyncComponent(() => import('./open/OpenPage'), 'OpenPage')
const OrgsArea = asyncComponent(() => import('./org/OrgsArea'), 'OrgsArea')
const RepoContainer = asyncComponent(() => import('./repo/RepoContainer'), 'RepoContainer')
const ScopePage = asyncComponent(() => import('./search/input/ScopePage'), 'ScopePage')
const SiteInitPage = asyncComponent(() => import('./site-admin/SiteInitPage'), 'SiteInitPage')
const RedirectToUserPage = asyncComponent(() => import('./user/settings/RedirectToUserPage'), 'RedirectToUserPage')
const RedirectToUserSettings = asyncComponent(
    () => import('./user/settings/RedirectToUserSettings'),
    'RedirectToUserSettings'
)
const SnippetsPage = asyncComponent(() => import('./snippets/SnippetsPage'), 'SnippetsPage')

export interface LayoutRouteComponentProps extends RouteComponentProps<any>, LayoutProps {}

export interface LayoutRouteProps {
    path: string
    exact?: boolean
    render: (props: LayoutRouteComponentProps) => React.ReactNode

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
    render: props => <RepoContainer {...props} />,
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
        render: props => <SavedQueriesPage {...props} />,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/open',
        render: props => <OpenPage {...props} />,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/sign-in',
        render: props => <SignInPage {...props} />,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/sign-up',
        render: props => <SignUpPage {...props} />,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/settings',
        render: props => <RedirectToUserSettings {...props} />,
    },
    {
        path: '/user',
        render: props => <RedirectToUserPage {...props} />,
    },
    {
        path: '/organizations',
        render: props => <OrgsArea {...props} />,
    },
    {
        path: '/search',
        render: props => <SearchResults {...props} deployType={window.context.deployType} />,
        exact: true,
    },
    {
        path: '/site-admin/init',
        exact: true,
        render: props => <SiteInitPage {...props} />,
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
        render: props => <ResetPasswordPage {...props} />,
        exact: true,
        forceNarrowWidth: true,
    },
    {
        path: '/explore',
        render: props => <ExploreArea {...props} />,
        exact: true,
    },
    {
        path: '/discussions',
        render: props => <DiscussionsPage {...props} />,
        exact: true,
    },
    {
        path: '/search/scope/:id',
        render: props => <ScopePage {...props} />,
        exact: true,
    },
    {
        path: '/api/console',
        render: props => <APIConsole {...props} />,
        exact: true,
    },
    {
        path: '/users/:username',
        render: props => <UserArea {...props} />,
    },
    {
        path: '/survey/:score?',
        render: props => <SurveyPage {...props} />,
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
        render: props => <SnippetsPage {...props} />,
    },
    repoRevRoute,
]
