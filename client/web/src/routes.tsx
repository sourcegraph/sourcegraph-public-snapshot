import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BatchChangesProps } from './batches'
import { CodeIntelligenceProps } from './codeintel'
import { communitySearchContextsRoutes } from './communitySearchContexts/routes'
import { BreadcrumbsProps, BreadcrumbSetters } from './components/Breadcrumbs'
import type { LayoutProps } from './Layout'
import type { ExtensionAlertProps } from './repo/RepoContainer'
import { ThemePreferenceProps } from './theme'
import { UserExternalServicesOrRepositoriesUpdateProps } from './util'
import { lazyComponent } from './util/lazyComponent'

const SearchPage = lazyComponent(() => import('./search/home/SearchPage'), 'SearchPage')
const StreamingSearchResults = lazyComponent(
    () => import('./search/results/StreamingSearchResults'),
    'StreamingSearchResults'
)
const SiteAdminArea = lazyComponent(() => import('./site-admin/SiteAdminArea'), 'SiteAdminArea')
const ExtensionsArea = lazyComponent(() => import('./extensions/ExtensionsArea'), 'ExtensionsArea')
const SearchConsolePage = lazyComponent(() => import('./search/SearchConsolePage'), 'SearchConsolePage')
const SearchNotebookPage = lazyComponent(() => import('./search/notebook/SearchNotebookPage'), 'SearchNotebookPage')
const SignInPage = lazyComponent(() => import('./auth/SignInPage'), 'SignInPage')
const SignUpPage = lazyComponent(() => import('./auth/SignUpPage'), 'SignUpPage')
const PostSignUpPage = lazyComponent(() => import('./auth/PostSignUpPage'), 'PostSignUpPage')
const SiteInitPage = lazyComponent(() => import('./site-admin/init/SiteInitPage'), 'SiteInitPage')

export interface LayoutRouteComponentProps<RouteParameters extends { [K in keyof RouteParameters]?: string }>
    extends RouteComponentProps<RouteParameters>,
        Omit<LayoutProps, 'match'>,
        ThemeProps,
        ThemePreferenceProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        ExtensionAlertProps,
        CodeIntelligenceProps,
        BatchChangesProps,
        UserExternalServicesOrRepositoriesUpdateProps {
    isSourcegraphDotCom: boolean
    isMacPlatform: boolean
}

export interface LayoutRouteProps<Parameters_ extends { [K in keyof Parameters_]?: string }> {
    path: string
    exact?: boolean
    render: (props: LayoutRouteComponentProps<Parameters_>) => React.ReactNode

    /**
     * A condition function that needs to return true if the route should be rendered
     *
     * @default () => true
     */
    condition?: (props: LayoutRouteComponentProps<Parameters_>) => boolean
}

// Force a hard reload so that we delegate to the serverside HTTP handler for a route.
function passThroughToServer(): React.ReactNode {
    window.location.reload()
    return null
}

export enum PageRoutes {
    INDEX = '/',
    SEARCH = '/search',
    SEARCH_CONSOLE = '/search/console',
    SEARCH_NOTEBOOK = '/search/notebook',
    SIGN_IN = '/sign-in',
    SIGN_UP = '/sign-up',
    WELCOME = '/welcome',
    SETTINGS = '/settings',
    USER = '/user',
    ORGANIZATIONS = '/organizations',
    SITE_ADMIN = '/site-admin',
    SITE_ADMIN_INIT = '/site-admin/init',
    PASSWORD_RESET = '/password-reset',
    API_CONSOLE = '/api/console',
    USER_AREA = '/users/:username',
    SURVEY = '/survey/:score?',
    EXTENSIONS = '/extensions',
    HELP = '/help',
    DEBUG = '/-/debug/*',
    REPO_CONTAINER = '/:repoRevAndRest+',
}

/**
 * Holds all top-level routes for the app because both the navbar and the main content area need to
 * switch over matched path.
 *
 * See https://reacttraining.com/react-router/web/example/sidebar
 */
export const routes: readonly LayoutRouteProps<any>[] = [
    {
        path: PageRoutes.INDEX,
        render: () => <Redirect to="/search" />,
        exact: true,
    },
    {
        path: PageRoutes.SEARCH,
        render: props => (props.parsedSearchQuery ? <StreamingSearchResults {...props} /> : <SearchPage {...props} />),
        exact: true,
    },
    {
        path: PageRoutes.SEARCH_CONSOLE,
        render: props =>
            props.showMultilineSearchConsole ? <SearchConsolePage {...props} /> : <Redirect to="/search" />,
        exact: true,
    },
    {
        path: PageRoutes.SEARCH_NOTEBOOK,
        render: props => (props.showSearchNotebook ? <SearchNotebookPage {...props} /> : <Redirect to="/search" />),
        exact: true,
    },
    {
        path: PageRoutes.SIGN_IN,
        render: props => <SignInPage {...props} context={window.context} />,
        exact: true,
    },
    {
        path: PageRoutes.SIGN_UP,
        render: props => <SignUpPage {...props} context={window.context} />,
        exact: true,
    },
    {
        path: PageRoutes.WELCOME,
        render: props =>
            /**
             * Welcome flow is allowed when:
             * 1. user is authenticated
             * 2. it's a DotComMode instance
             * AND
             * instance has enabled enablePostSignupFlow experimental feature
             * OR
             * user authenticated has a AllowUserViewPostSignup tag
             */

            !!props.authenticatedUser &&
            window.context.sourcegraphDotComMode &&
            (window.context.experimentalFeatures.enablePostSignupFlow ||
                props.authenticatedUser?.tags.includes('AllowUserViewPostSignup')) ? (
                <PostSignUpPage
                    authenticatedUser={props.authenticatedUser}
                    telemetryService={props.telemetryService}
                    context={window.context}
                    onUserExternalServicesOrRepositoriesUpdate={props.onUserExternalServicesOrRepositoriesUpdate}
                    setSelectedSearchContextSpec={props.setSelectedSearchContextSpec}
                />
            ) : (
                <Redirect to="/search" />
            ),

        exact: true,
    },
    {
        path: PageRoutes.SETTINGS,
        render: lazyComponent(() => import('./user/settings/RedirectToUserSettings'), 'RedirectToUserSettings'),
    },
    {
        path: PageRoutes.USER,
        render: lazyComponent(() => import('./user/settings/RedirectToUserPage'), 'RedirectToUserPage'),
    },
    {
        path: PageRoutes.ORGANIZATIONS,
        render: lazyComponent(() => import('./org/OrgsArea'), 'OrgsArea'),
    },
    {
        path: PageRoutes.SITE_ADMIN_INIT,
        exact: true,
        render: props => <SiteInitPage {...props} context={window.context} />,
    },
    {
        path: PageRoutes.SITE_ADMIN,
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
        path: PageRoutes.PASSWORD_RESET,
        render: lazyComponent(() => import('./auth/ResetPasswordPage'), 'ResetPasswordPage'),
        exact: true,
    },
    {
        path: PageRoutes.API_CONSOLE,
        render: lazyComponent(() => import('./api/ApiConsole'), 'ApiConsole'),
        exact: true,
    },
    {
        path: PageRoutes.USER_AREA,
        render: lazyComponent(() => import('./user/area/UserArea'), 'UserArea'),
    },
    {
        path: PageRoutes.SURVEY,
        render: lazyComponent(() => import('./marketing/SurveyPage'), 'SurveyPage'),
    },
    {
        path: PageRoutes.EXTENSIONS,
        render: props => <ExtensionsArea {...props} routes={props.extensionsAreaRoutes} />,
    },
    {
        path: PageRoutes.HELP,
        render: passThroughToServer,
    },
    {
        path: PageRoutes.DEBUG,
        render: passThroughToServer,
    },
    ...communitySearchContextsRoutes,
    {
        path: PageRoutes.REPO_CONTAINER,
        render: lazyComponent(() => import('./repo/RepoContainer'), 'RepoContainer'),
    },
]
