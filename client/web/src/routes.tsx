import { useEffect } from 'react'

import { Navigate } from 'react-router-dom-v5-compat'

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
const OrgsArea = lazyComponent(() => import('./org/OrgsArea'), 'OrgsArea')
const ResetPasswordPage = lazyComponent(() => import('./auth/ResetPasswordPage'), 'ResetPasswordPage')
const ApiConsole = lazyComponent(() => import('./api/ApiConsole'), 'ApiConsole')
const UserArea = lazyComponent(() => import('./user/area/UserArea'), 'UserArea')
const SurveyPage = lazyComponent(() => import('./marketing/page/SurveyPage'), 'SurveyPage')
const RepoContainer = lazyComponent(() => import('./repo/RepoContainer'), 'RepoContainer')

export interface LayoutRouteComponentProps
    extends Omit<LayoutProps, 'match'>,
        ThemeProps,
        ThemePreferenceProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        CodeIntelligenceProps,
        BatchChangesProps {
    isSourcegraphDotCom: boolean
    isMacPlatform: boolean
}

export interface LayoutRouteProps {
    path: string
    render: (props: LayoutRouteComponentProps) => React.ReactNode

    /**
     * A condition function that needs to return true if the route should be rendered
     *
     * @default () => true
     */
    condition?: (props: LayoutRouteComponentProps) => boolean
}

// Force a hard reload so that we delegate to the serverside HTTP handler for a route.
const PassThroughToServer: React.FC = () => {
    useEffect(() => {
        window.location.reload()
    })
    return null
}

/**
 * Holds all top-level routes for the app because both the navbar and the main content area need to
 * switch over matched path.
 *
 * See https://reacttraining.com/react-router/web/example/sidebar
 */
export const routes: readonly LayoutRouteProps[] = (
    [
        {
            path: PageRoutes.Index,
            render: () => <Navigate replace={true} to={PageRoutes.Search} />,
        },
        {
            path: PageRoutes.Search,
            render: props => <SearchPageWrapper {...props} />,
        },
        {
            path: PageRoutes.SearchConsole,
            render: props => {
                const { showMultilineSearchConsole } = getExperimentalFeatures()

                return showMultilineSearchConsole ? (
                    <SearchConsolePage {...props} />
                ) : (
                    <Navigate replace={true} to={PageRoutes.Search} />
                )
            },
        },
        {
            path: PageRoutes.SignIn,
            render: props => <SignInPage {...props} context={window.context} />,
        },
        {
            path: PageRoutes.SignUp,
            render: props => <SignUpPage {...props} context={window.context} />,
        },
        {
            path: PageRoutes.UnlockAccount,
            render: props => <UnlockAccountPage {...props} context={window.context} />,
        },
        {
            path: PageRoutes.Welcome,
            // This route is deprecated after we removed the post-sign-up page experimental feature, but we keep it for now to not break links.
            render: () => <Navigate replace={true} to={PageRoutes.Search} />,
        },
        {
            path: PageRoutes.InstallGitHubAppSuccess,
            render: () => <InstallGitHubAppSuccessPage />,
        },
        {
            path: PageRoutes.Settings,
            render: props => <RedirectToUserSettings {...props} />,
        },
        {
            path: PageRoutes.User,
            render: props => <RedirectToUserPage {...props} />,
        },
        {
            path: PageRoutes.Organizations,
            render: props => <OrgsArea {...props} />,
        },
        {
            path: PageRoutes.SiteAdminInit,
            render: props => <SiteInitPage {...props} context={window.context} />,
        },
        {
            path: PageRoutes.SiteAdmin,
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
            path: PageRoutes.PasswordReset,
            render: props => <ResetPasswordPage {...props} />,
        },
        {
            path: PageRoutes.ApiConsole,
            render: () => <ApiConsole />,
        },
        {
            path: PageRoutes.UserArea,
            render: props => <UserArea {...props} />,
        },
        {
            path: PageRoutes.Survey,
            render: props => <SurveyPage {...props} />,
        },
        {
            path: PageRoutes.Help,
            render: () => <PassThroughToServer />,
        },
        {
            path: PageRoutes.Debug,
            render: () => <PassThroughToServer />,
        },
        ...communitySearchContextsRoutes,
        {
            path: PageRoutes.RepoContainer,
            render: props => <RepoContainer {...props} />,
        },
    ] as readonly (LayoutRouteProps | undefined)[]
).filter(Boolean) as readonly LayoutRouteProps[]
