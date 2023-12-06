import { useEffect } from 'react'

import { Navigate, type RouteObject } from 'react-router-dom'

import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { communitySearchContextsRoutes } from './communitySearchContexts/routes'
import { type LegacyLayoutRouteContext, LegacyRoute } from './LegacyRouteContext'
import { PageRoutes } from './routes.constants'
import { SearchPageWrapper } from './search/SearchPageWrapper'

const SiteAdminArea = lazyComponent(() => import('./site-admin/SiteAdminArea'), 'SiteAdminArea')
const SearchConsolePage = lazyComponent(() => import('./search/SearchConsolePage'), 'SearchConsolePage')
const SignInPage = lazyComponent(() => import('./auth/SignInPage'), 'SignInPage')
const RequestAccessPage = lazyComponent(() => import('./auth/RequestAccessPage'), 'RequestAccessPage')
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
const TeamsArea = lazyComponent(() => import('./team/TeamsArea'), 'TeamsArea')
const CodySidebarStoreProvider = lazyComponent(() => import('./cody/sidebar/Provider'), 'CodySidebarStoreProvider')
const GetCodyPage = lazyComponent(() => import('./get-cody/GetCodyPage'), 'GetCodyPage')
const PostSignUpPage = lazyComponent(() => import('./auth/PostSignUpPage'), 'PostSignUpPage')

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
export const routes: RouteObject[] = [
    {
        path: PageRoutes.Index,
        element: <Navigate replace={true} to={PageRoutes.Search} />,
    },
    {
        path: PageRoutes.Search,
        element: <LegacyRoute render={props => <SearchPageWrapper {...props} />} />,
    },
    {
        path: PageRoutes.SearchConsole,
        element: <LegacyRoute render={props => <SearchConsolePageOrRedirect {...props} />} />,
    },
    {
        path: PageRoutes.SignIn,
        element: <LegacyRoute render={props => <SignInPage {...props} context={window.context} />} />,
    },
    {
        path: PageRoutes.RequestAccess,
        element: <RequestAccessPage />,
    },
    {
        path: PageRoutes.SignUp,
        element: (
            <LegacyRoute
                render={props => (
                    <SignUpPage
                        {...props}
                        context={window.context}
                        telemetryRecorder={window.context.telemetryRecorder}
                    />
                )}
            />
        ),
    },
    {
        path: PageRoutes.UnlockAccount,
        element: <LegacyRoute render={props => <UnlockAccountPage {...props} context={window.context} />} />,
    },
    {
        path: PageRoutes.Welcome,
        // This route is deprecated after we removed the post-sign-up page experimental feature, but we keep it for now to not break links.
        element: <Navigate replace={true} to={PageRoutes.Search} />,
    },
    {
        path: PageRoutes.InstallGitHubAppSuccess,
        element: <InstallGitHubAppSuccessPage />,
    },
    {
        path: PageRoutes.Settings,
        element: <LegacyRoute render={props => <RedirectToUserSettings {...props} />} />,
    },
    {
        path: PageRoutes.User,
        element: <LegacyRoute render={props => <RedirectToUserPage {...props} />} />,
    },
    {
        path: PageRoutes.Teams,
        element: <LegacyRoute render={props => <TeamsArea {...props} />} />,
    },
    {
        path: PageRoutes.Organizations,
        element: (
            <LegacyRoute
                render={props => <OrgsArea {...props} telemetryRecorder={window.context.telemetryRecorder} />}
            />
        ),
    },
    {
        path: PageRoutes.SiteAdminInit,
        element: <LegacyRoute render={props => <SiteInitPage {...props} context={window.context} />} />,
    },
    {
        path: PageRoutes.SiteAdmin,
        element: (
            <LegacyRoute
                render={props => (
                    <SiteAdminArea
                        {...props}
                        routes={props.siteAdminAreaRoutes}
                        sideBarGroups={props.siteAdminSideBarGroups}
                        overviewComponents={props.siteAdminOverviewComponents}
                        codeInsightsEnabled={window.context.codeInsightsEnabled}
                        telemetryRecorder={window.context.telemetryRecorder}
                    />
                )}
            />
        ),
    },
    {
        path: PageRoutes.PasswordReset,
        element: <LegacyRoute render={props => <ResetPasswordPage {...props} />} />,
    },
    {
        path: PageRoutes.ApiConsole,
        element: <ApiConsole />,
    },
    {
        path: PageRoutes.UserArea,
        element: (
            <LegacyRoute
                render={props => <UserArea {...props} telemetryRecorder={window.context.telemetryRecorder} />}
            />
        ),
    },
    {
        path: PageRoutes.Survey,
        element: <LegacyRoute render={props => <SurveyPage {...props} />} />,
    },
    {
        path: PageRoutes.Help,
        element: <PassThroughToServer />,
    },
    {
        path: PageRoutes.Debug,
        element: <PassThroughToServer />,
    },
    ...communitySearchContextsRoutes,
    {
        path: PageRoutes.RepoContainer,
        element: (
            <LegacyRoute
                render={props => (
                    <CodySidebarStoreProvider
                        authenticatedUser={props.authenticatedUser}
                        telemetryRecorder={window.context.telemetryRecorder}
                    >
                        <RepoContainer {...props} telemetryRecorder={window.context.telemetryRecorder} />
                    </CodySidebarStoreProvider>
                )}
            />
        ),
        // In RR6, the useMatches hook will only give you the location that is matched
        // by the path rule and not the path rule instead. Since we need to be able to
        // detect if we're inside the repo container reliably inside the Layout, we
        // expose this information in the handle object instead.
        handle: { isRepoContainer: true },
    },
    {
        path: PageRoutes.GetCody,
        element: <LegacyRoute render={props => <GetCodyPage {...props} context={window.context} />} />,
    },
    {
        path: PageRoutes.PostSignUp,
        element: <LegacyRoute render={props => <PostSignUpPage {...props} />} />,
    },
]

function SearchConsolePageOrRedirect(props: LegacyLayoutRouteContext): JSX.Element {
    const showMultilineSearchConsole = useExperimentalFeatures(features => features.showMultilineSearchConsole)

    return showMultilineSearchConsole ? (
        <SearchConsolePage {...props} telemetryRecorder={window.context.telemetryRecorder} />
    ) : (
        <Navigate replace={true} to={PageRoutes.Search} />
    )
}
