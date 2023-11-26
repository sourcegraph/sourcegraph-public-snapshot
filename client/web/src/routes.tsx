import { useEffect } from 'react'

import { Navigate, useNavigate, type RouteObject } from 'react-router-dom'

import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { communitySearchContextsRoutes } from './communitySearchContexts/routes'
import { type LegacyLayoutRouteContext, LegacyRoute } from './LegacyRouteContext'
import { PageRoutes } from './routes.constants'
import { isSearchJobsEnabled } from './search-jobs/utility'
import { SearchPageWrapper } from './search/SearchPageWrapper'

const SiteAdminArea = lazyComponent(() => import('./site-admin/SiteAdminArea'), 'SiteAdminArea')
const LicenseManagementArea = lazyComponent(
    () => import('./productSubscription/LicenseManagementArea'),
    'LicenseManagementArea'
)
const SearchConsolePage = lazyComponent(() => import('./search/SearchConsolePage'), 'SearchConsolePage')
const SignInPage = lazyComponent(() => import('./auth/SignInPage'), 'SignInPage')
const RequestAccessPage = lazyComponent(() => import('./auth/RequestAccessPage'), 'RequestAccessPage')
const SignUpPage = lazyComponent(() => import('./auth/SignUpPage'), 'SignUpPage')
const UnlockAccountPage = lazyComponent(() => import('./auth/UnlockAccount'), 'UnlockAccountPage')
const SiteInitPage = lazyComponent(() => import('./site-admin/init/SiteInitPage'), 'SiteInitPage')
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

const GlobalNotebooksArea = lazyComponent(() => import('./notebooks/GlobalNotebooksArea'), 'GlobalNotebooksArea')
const GlobalBatchChangesArea = lazyComponent(
    () => import('./enterprise/batches/global/GlobalBatchChangesArea'),
    'GlobalBatchChangesArea'
)
const GlobalCodeMonitoringArea = lazyComponent(
    () => import('./enterprise/code-monitoring/global/GlobalCodeMonitoringArea'),
    'GlobalCodeMonitoringArea'
)
const CodeInsightsRouter = lazyComponent(() => import('./enterprise/insights/CodeInsightsRouter'), 'CodeInsightsRouter')
const SearchContextsListPage = lazyComponent(
    () => import('./enterprise/searchContexts/SearchContextsListPage'),
    'SearchContextsListPage'
)
const CreateSearchContextPage = lazyComponent(
    () => import('./enterprise/searchContexts/CreateSearchContextPage'),
    'CreateSearchContextPage'
)
const EditSearchContextPage = lazyComponent(
    () => import('./enterprise/searchContexts/EditSearchContextPage'),
    'EditSearchContextPage'
)
const SearchContextPage = lazyComponent(
    () => import('./enterprise/searchContexts/SearchContextPage'),
    'SearchContextPage'
)
const CodySearchPage = lazyComponent(() => import('./cody/search/CodySearchPage'), 'CodySearchPage')
const CodyChatPage = lazyComponent(() => import('./cody/chat/CodyChatPage'), 'CodyChatPage')
const CodyManagementPage = lazyComponent(() => import('./cody/management/CodyManagementPage'), 'CodyManagementPage')
const CodySubscriptionPage = lazyComponent(
    () => import('./cody/subscription/CodySubscriptionPage'),
    'CodySubscriptionPage'
)
const OwnPage = lazyComponent(() => import('./enterprise/own/OwnPage'), 'OwnPage')
const SearchJob = lazyComponent(() => import('./enterprise/search-jobs/SearchJobsPage'), 'SearchJobsPage')

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
        path: PageRoutes.GetCody,
        element: <LegacyRoute render={props => <GetCodyPage {...props} context={window.context} />} />,
    },
    {
        path: PageRoutes.PostSignUp,
        element: <LegacyRoute render={props => <PostSignUpPage {...props} />} />,
    },
    {
        path: PageRoutes.BatchChanges,
        element: (
            <LegacyRoute
                render={props => <GlobalBatchChangesArea {...props} />}
                // We also render this route on sourcegraph.com as a precaution in case anyone
                // follows an in-app link to /batch-changes from sourcegraph.com; the component
                // will just redirect the visitor to the marketing page
                condition={({ batchChangesEnabled, isSourcegraphDotCom }) => batchChangesEnabled || isSourcegraphDotCom}
            />
        ),
    },
    {
        path: PageRoutes.CodeMonitoring,
        element: (
            <LegacyRoute
                render={props => <GlobalCodeMonitoringArea {...props} />}
                condition={({ isSourcegraphDotCom }) => !isSourcegraphDotCom}
            />
        ),
    },
    {
        path: PageRoutes.Insights,
        element: (
            <LegacyRoute
                render={props => <CodeInsightsRouter {...props} />}
                condition={({ codeInsightsEnabled }) => !!codeInsightsEnabled}
            />
        ),
    },
    {
        path: PageRoutes.SearchJobs,
        element: (
            <LegacyRoute
                render={props => (
                    <SearchJob
                        isAdmin={props.authenticatedUser?.siteAdmin ?? false}
                        telemetryService={props.telemetryService}
                    />
                )}
                condition={isSearchJobsEnabled}
            />
        ),
    },
    {
        path: PageRoutes.Contexts,
        element: <LegacyRoute render={props => <SearchContextsListPage {...props} />} />,
    },
    {
        path: PageRoutes.CreateContext,
        element: <LegacyRoute render={props => <CreateSearchContextPage {...props} />} />,
    },
    {
        path: PageRoutes.EditContext,
        element: <LegacyRoute render={props => <EditSearchContextPage {...props} />} />,
    },
    {
        path: PageRoutes.Context,
        element: <LegacyRoute render={props => <SearchContextPage {...props} />} />,
    },
    {
        path: PageRoutes.SearchNotebook,
        element: <Navigate to={PageRoutes.Notebooks} replace={true} />,
    },
    {
        path: PageRoutes.Notebooks + '/*',
        element: <LegacyRoute render={props => <GlobalNotebooksArea {...props} />} />,
    },
    {
        path: PageRoutes.CodySearch,
        element: <LegacyRoute render={props => <CodySearchPage {...props} />} />,
    },
    // TODO: [TEMPORARY] remove this redirect route when the marketing page is added.
    {
        path: '/cody/*',
        element: (
            <LegacyRoute
                render={() => {
                    const chatID = window.location.pathname.split('/').pop()
                    const navigate = useNavigate()

                    useEffect(() => {
                        navigate(`/cody/chat/${chatID}`)
                    }, [navigate, chatID])

                    return <div />
                }}
                condition={() => !window.location.pathname.startsWith('/cody/chat')}
            />
        ),
    },
    {
        path: PageRoutes.Cody + '/*',
        element: <LegacyRoute render={props => <CodyChatPage {...props} context={window.context} />} />,
    },
    {
        path: PageRoutes.CodyManagement,
        element: <LegacyRoute render={props => <CodyManagementPage {...props} />} />,
    },
    {
        path: PageRoutes.CodySubscription,
        element: <LegacyRoute render={props => <CodySubscriptionPage {...props} />} />,
    },
    {
        path: PageRoutes.Own,
        element: <OwnPage />,
    },
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
        element: <LegacyRoute render={props => <SignUpPage {...props} context={window.context} />} />,
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
        element: <LegacyRoute render={props => <OrgsArea {...props} />} />,
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
                    />
                )}
            />
        ),
    },
    {
        path: PageRoutes.LicenseManagement,
        element: <LegacyRoute render={props => <LicenseManagementArea {...props} />} />,
    },
    {
        path: PageRoutes.PasswordReset,
        element: <LegacyRoute render={props => <ResetPasswordPage {...props} context={window.context} />} />,
    },
    {
        path: PageRoutes.ApiConsole,
        element: <ApiConsole />,
    },
    {
        path: PageRoutes.UserArea,
        element: <LegacyRoute render={props => <UserArea {...props} />} />,
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
                    <CodySidebarStoreProvider authenticatedUser={props.authenticatedUser}>
                        <RepoContainer {...props} />
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
]

function SearchConsolePageOrRedirect(props: LegacyLayoutRouteContext): JSX.Element {
    const showMultilineSearchConsole = useExperimentalFeatures(features => features.showMultilineSearchConsole)

    return showMultilineSearchConsole ? (
        <SearchConsolePage {...props} />
    ) : (
        <Navigate replace={true} to={PageRoutes.Search} />
    )
}
