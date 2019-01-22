import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { LayoutProps } from './Layout'
import { parseSearchURLQuery } from './search'
const MainPage = React.lazy(async () => ({
    default: (await import('./enterprise/dotcom/welcome/MainPage')).MainPage,
}))
const WelcomeSearchPage = React.lazy(async () => ({
    default: (await import('./enterprise/dotcom/welcome/WelcomeSearchPage')).WelcomeSearchPage,
}))
const WelcomeCodeIntelligencePage = React.lazy(async () => ({
    default: (await import('./enterprise/dotcom/welcome/WelcomeCodeIntelligencePage')).WelcomeCodeIntelligencePage,
}))
const SearchPage = React.lazy(async () => ({
    default: (await import('./search/input/SearchPage')).SearchPage,
}))
const SearchResults = React.lazy(async () => ({
    default: (await import('./search/results/SearchResults')).SearchResults,
}))
const SavedQueriesPage = React.lazy(async () => ({
    default: (await import('./search/saved-queries/SavedQueries')).SavedQueriesPage,
}))
const SiteAdminArea = React.lazy(async () => ({
    default: (await import('./site-admin/SiteAdminArea')).SiteAdminArea,
}))
const UserArea = React.lazy(async () => ({
    default: (await import('./user/area/UserArea')).UserArea,
}))
const APIConsole = React.lazy(async () => ({ default: (await import('./api/APIConsole')).APIConsole }))
const ResetPasswordPage = React.lazy(async () => ({
    default: (await import('./auth/ResetPasswordPage')).ResetPasswordPage,
}))
const SignInPage = React.lazy(async () => ({ default: (await import('./auth/SignInPage')).SignInPage }))
const SignUpPage = React.lazy(async () => ({ default: (await import('./auth/SignUpPage')).SignUpPage }))
const DiscussionsPage = React.lazy(async () => ({
    default: (await import('./discussions/DiscussionsPage')).DiscussionsPage,
}))
const DocSitePage = React.lazy(async () => ({ default: (await import('./docSite/DocSitePage')).DocSitePage }))
const ExploreArea = React.lazy(async () => ({ default: (await import('./explore/ExploreArea')).ExploreArea }))
const ExtensionsArea = React.lazy(async () => ({
    default: (await import('./extensions/ExtensionsArea')).ExtensionsArea,
}))
const SurveyPage = React.lazy(async () => ({ default: (await import('./marketing/SurveyPage')).SurveyPage }))
const OpenPage = React.lazy(async () => ({ default: (await import('./open/OpenPage')).OpenPage }))
const OrgsArea = React.lazy(async () => ({ default: (await import('./org/OrgsArea')).OrgsArea }))
const RepoContainer = React.lazy(async () => ({ default: (await import('./repo/RepoContainer')).RepoContainer }))
const ScopePage = React.lazy(async () => ({ default: (await import('./search/input/ScopePage')).ScopePage }))
const SiteInitPage = React.lazy(async () => ({ default: (await import('./site-admin/SiteInitPage')).SiteInitPage }))
const RedirectToUserPage = React.lazy(async () => ({
    default: (await import('./user/account/RedirectToUserPage')).RedirectToUserPage,
}))
const RedirectToUserSettings = React.lazy(async () => ({
    default: (await import('./user/account/RedirectToUserSettings')).RedirectToUserSettings,
}))

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
            window.context.sourcegraphDotComMode && !props.user ? <Redirect to="/start" /> : <Redirect to="/search" />,
        exact: true,
    },
    {
        path: '/start',
        render: props => <MainPage {...props} />,
        exact: true,
    },
    {
        path: '/welcome/search',
        render: props => <WelcomeSearchPage {...props} />,
        exact: true,
    },
    {
        path: '/welcome/code-intelligence',
        render: props => <WelcomeCodeIntelligencePage {...props} />,
        exact: true,
    },
    {
        path: '/search',
        render: (props: any) =>
            parseSearchURLQuery(props.location.search) ? <SearchResults {...props} /> : <SearchPage {...props} />,
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
        render: props => <SearchResults {...props} />,
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
        render: props => <DocSitePage {...props} path={props.location.pathname.slice(props.match.path.length + 1)} />,
    },
    repoRevRoute,
]
