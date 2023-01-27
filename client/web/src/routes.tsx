import * as React from 'react'

import { Redirect, RouteComponentProps } from 'react-router'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { BatchChangesProps } from './batches'
import { CodeIntelligenceProps } from './codeintel'
import { communitySearchContextsRoutes } from './communitySearchContexts/routes'
import { BreadcrumbsProps, BreadcrumbSetters } from './components/Breadcrumbs'
import type { LayoutProps } from './Layout'
import { PageRoutes } from './routes.constants'
import { SearchPageWrapper } from './search/SearchPageWrapper'
import { getExperimentalFeatures } from './stores'
import { ThemePreferenceProps } from './theme'

const SiteAdminArea = lazyComponent(() => import('./site-admin/SiteAdminArea'), 'SiteAdminArea')
const SearchConsolePage = lazyComponent(() => import('./search/SearchConsolePage'), 'SearchConsolePage')
const SignInPage = lazyComponent(() => import('./auth/SignInPage'), 'SignInPage')
const SignUpPage = lazyComponent(() => import('./auth/SignUpPage'), 'SignUpPage')
const UnlockAccountPage = lazyComponent(() => import('./auth/UnlockAccount'), 'UnlockAccountPage')
const SiteInitPage = lazyComponent(() => import('./site-admin/init/SiteInitPage'), 'SiteInitPage')
const InstallGitHubAppSuccessPage = lazyComponent(
    () => import('./org/settings/codeHosts/InstallGitHubAppSuccessPage'),
    'InstallGitHubAppSuccessPage'
)
const RedirectToUserSettings = lazyComponent(
    () => import('./user/settings/RedirectToUserSettings'),
    'RedirectToUserSettings'
)
const RedirectToUserPage = lazyComponent(() => import('./user/settings/RedirectToUserPage'), 'RedirectToUserPage')

export interface LayoutRouteComponentProps<RouteParameters extends { [K in keyof RouteParameters]?: string }>
    extends RouteComponentProps<RouteParameters>,
        Omit<LayoutProps, 'match'>,
        ThemeProps,
        ThemePreferenceProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        CodeIntelligenceProps,
        BatchChangesProps {
    isSourcegraphDotCom: boolean
    isMacPlatform: boolean
}

// A version of LayoutRouteComponentProps that is compatible with react router v6
export type LayoutRouteComponentPropsRRV6<T extends { [K in keyof T]?: string }> = Omit<
    LayoutRouteComponentProps<T>,
    'location' | 'history' | 'match' | 'staticContext'
>

interface LayoutRoutePropsV5<Parameters_ extends { [K in keyof Parameters_]?: string }> {
    isV6: false
    path: string
    exact?: boolean
    render: (props: LayoutRouteComponentProps<Parameters_>) => React.ReactNode

    /**
     * A condition function that needs to return true if the route should be rendered
     *
     * @default () => true
     */
    condition?: (
        props: Omit<LayoutRouteComponentProps<Parameters_>, 'location' | 'history' | 'match' | 'staticContext'>
    ) => boolean
}

interface LayoutRoutePropsV6 {
    isV6: true
    path: string
    render: (props: LayoutRouteComponentPropsRRV6<{}>) => React.ReactNode

    /**
     * A condition function that needs to return true if the route should be rendered
     *
     * @default () => true
     */
    condition?: (
        props: Omit<LayoutRouteComponentProps<{}>, 'location' | 'history' | 'match' | 'staticContext'>
    ) => boolean
}

export type LayoutRouteProps<T extends { [K in keyof T]?: string }> = LayoutRoutePropsV5<T> | LayoutRoutePropsV6

// Force a hard reload so that we delegate to the serverside HTTP handler for a route.
function passThroughToServer(): React.ReactNode {
    window.location.reload()
    return null
}

/**
 * Holds all top-level routes for the app because both the navbar and the main content area need to
 * switch over matched path.
 *
 * See https://reacttraining.com/react-router/web/example/sidebar
 */
export const routes: readonly LayoutRouteProps<any>[] = (
    [
        {
            isV6: true,
            path: PageRoutes.Index,
            render: () => <Redirect to={PageRoutes.Search} />,
        },
        {
            isV6: true,
            path: PageRoutes.Search,
            render: (props: LayoutRouteComponentPropsRRV6<{}>) => <SearchPageWrapper {...props} />,
        },
        {
            isV6: true,
            path: PageRoutes.SearchConsole,
            render: (props: LayoutRouteComponentPropsRRV6<{}>) => {
                const { showMultilineSearchConsole } = getExperimentalFeatures()

                return showMultilineSearchConsole ? (
                    <SearchConsolePage {...props} />
                ) : (
                    <Redirect to={PageRoutes.Search} />
                )
            },
        },
        {
            isV6: true,
            path: PageRoutes.SignIn,
            render: (props: LayoutRouteComponentPropsRRV6<{}>) => <SignInPage {...props} context={window.context} />,
        },
        {
            isV6: true,
            path: PageRoutes.SignUp,
            render: (props: LayoutRouteComponentPropsRRV6<{}>) => <SignUpPage {...props} context={window.context} />,
        },
        {
            isV6: true,
            path: PageRoutes.UnlockAccount,
            render: (props: LayoutRouteComponentPropsRRV6<{}>) => (
                <UnlockAccountPage {...props} context={window.context} />
            ),
        },
        {
            isV6: true,
            path: PageRoutes.Welcome,
            // This route is deprecated after we removed the post-sign-up page experimental feature, but we keep it for now to not break links.
            render: () => <Redirect to={PageRoutes.Search} />,
        },
        {
            isV6: true,
            path: PageRoutes.InstallGitHubAppSuccess,
            render: () => <InstallGitHubAppSuccessPage />,
        },
        {
            isV6: true,
            path: PageRoutes.Settings,
            render: (props: LayoutRouteComponentPropsRRV6<{}>) => <RedirectToUserSettings {...props} />,
        },
        {
            isV6: true,
            path: PageRoutes.User,
            render: (props: LayoutRouteComponentPropsRRV6<{}>) => <RedirectToUserPage {...props} />,
        },
        {
            isV6: false,
            path: PageRoutes.Organizations,
            render: lazyComponent(() => import('./org/OrgsArea'), 'OrgsArea'),
        },
        {
            isV6: false,
            path: PageRoutes.SiteAdminInit,
            exact: true,
            render: (props: LayoutRouteComponentProps<{}>) => <SiteInitPage {...props} context={window.context} />,
        },
        {
            isV6: false,
            path: PageRoutes.SiteAdmin,
            render: (props: LayoutRouteComponentProps<{}>) => (
                <SiteAdminArea
                    {...props}
                    routes={props.siteAdminAreaRoutes}
                    sideBarGroups={props.siteAdminSideBarGroups}
                    overviewComponents={props.siteAdminOverviewComponents}
                />
            ),
        },
        {
            isV6: false,
            path: PageRoutes.PasswordReset,
            render: lazyComponent(() => import('./auth/ResetPasswordPage'), 'ResetPasswordPage'),
            exact: true,
        },
        {
            isV6: false,
            path: PageRoutes.ApiConsole,
            render: lazyComponent(() => import('./api/ApiConsole'), 'ApiConsole'),
            exact: true,
        },
        {
            isV6: false,
            path: PageRoutes.UserArea,
            render: lazyComponent(() => import('./user/area/UserArea'), 'UserArea'),
        },
        {
            isV6: false,
            path: PageRoutes.Survey,
            render: lazyComponent(() => import('./marketing/page/SurveyPage'), 'SurveyPage'),
        },
        {
            isV6: false,
            path: PageRoutes.Help,
            render: passThroughToServer,
        },
        {
            isV6: false,
            path: PageRoutes.Debug,
            render: passThroughToServer,
        },
        ...communitySearchContextsRoutes,
        {
            path: PageRoutes.RepoContainer,
            render: lazyComponent(() => import('./repo/RepoContainer'), 'RepoContainer'),
        },
    ] as readonly (LayoutRouteProps<any> | undefined)[]
).filter(Boolean) as readonly LayoutRouteProps<any>[]
