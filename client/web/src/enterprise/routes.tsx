import { useEffect } from 'react'

import { Navigate, type RouteObject, useNavigate } from 'react-router-dom'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { LegacyRoute } from '../LegacyRouteContext'
import { routes } from '../routes'
import { EnterprisePageRoutes } from '../routes.constants'
import { isSearchJobsEnabled } from '../search-jobs/utility'

const GlobalNotebooksArea = lazyComponent(() => import('../notebooks/GlobalNotebooksArea'), 'GlobalNotebooksArea')
const GlobalBatchChangesArea = lazyComponent(
    () => import('./batches/global/GlobalBatchChangesArea'),
    'GlobalBatchChangesArea'
)
const GlobalCodeMonitoringArea = lazyComponent(
    () => import('./code-monitoring/global/GlobalCodeMonitoringArea'),
    'GlobalCodeMonitoringArea'
)
const CodeInsightsRouter = lazyComponent(() => import('./insights/CodeInsightsRouter'), 'CodeInsightsRouter')
const SearchContextsListPage = lazyComponent(
    () => import('./searchContexts/SearchContextsListPage'),
    'SearchContextsListPage'
)
const CreateSearchContextPage = lazyComponent(
    () => import('./searchContexts/CreateSearchContextPage'),
    'CreateSearchContextPage'
)
const EditSearchContextPage = lazyComponent(
    () => import('./searchContexts/EditSearchContextPage'),
    'EditSearchContextPage'
)
const SearchContextPage = lazyComponent(() => import('./searchContexts/SearchContextPage'), 'SearchContextPage')
const CodySearchPage = lazyComponent(() => import('../cody/search/CodySearchPage'), 'CodySearchPage')
const CodyChatPage = lazyComponent(() => import('../cody/chat/CodyChatPage'), 'CodyChatPage')
const OwnPage = lazyComponent(() => import('./own/OwnPage'), 'OwnPage')
const AppAuthCallbackPage = lazyComponent(() => import('./app/AppAuthCallbackPage'), 'AppAuthCallbackPage')
const AppSetup = lazyComponent(() => import('./app/setup/AppSetupWizard'), 'AppSetupWizard')
const SearchJob = lazyComponent(() => import('./search-jobs/SearchJobsPage'), 'SearchJobsPage')

export const enterpriseRoutes: RouteObject[] = [
    {
        path: `${EnterprisePageRoutes.AppSetup}/*`,
        handle: { isFullPage: true },
        element: (
            <LegacyRoute
                render={props => <AppSetup telemetryService={props.telemetryService} />}
                condition={({ isCodyApp }) => isCodyApp}
            />
        ),
    },
    {
        path: EnterprisePageRoutes.BatchChanges,
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
        path: EnterprisePageRoutes.CodeMonitoring,
        element: <LegacyRoute render={props => <GlobalCodeMonitoringArea {...props} />} />,
    },
    {
        path: EnterprisePageRoutes.Insights,
        element: (
            <LegacyRoute
                render={props => <CodeInsightsRouter {...props} />}
                condition={({ codeInsightsEnabled }) => !!codeInsightsEnabled}
            />
        ),
    },
    {
        path: EnterprisePageRoutes.SearchJobs,
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
        path: EnterprisePageRoutes.Contexts,
        element: <LegacyRoute render={props => <SearchContextsListPage {...props} />} />,
    },
    {
        path: EnterprisePageRoutes.CreateContext,
        element: <LegacyRoute render={props => <CreateSearchContextPage {...props} />} />,
    },
    {
        path: EnterprisePageRoutes.EditContext,
        element: <LegacyRoute render={props => <EditSearchContextPage {...props} />} />,
    },
    {
        path: EnterprisePageRoutes.Context,
        element: <LegacyRoute render={props => <SearchContextPage {...props} />} />,
    },
    {
        path: EnterprisePageRoutes.SearchNotebook,
        element: <Navigate to={EnterprisePageRoutes.Notebooks} replace={true} />,
    },
    {
        path: EnterprisePageRoutes.Notebooks + '/*',
        element: <LegacyRoute render={props => <GlobalNotebooksArea {...props} />} />,
    },
    {
        path: EnterprisePageRoutes.CodySearch,
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
        path: EnterprisePageRoutes.Cody + '/*',
        element: <LegacyRoute render={props => <CodyChatPage {...props} context={window.context} />} />,
    },
    {
        path: EnterprisePageRoutes.Own,
        element: <OwnPage />,
    },
    {
        path: EnterprisePageRoutes.AppAuthCallback,
        element: <LegacyRoute render={() => <AppAuthCallbackPage />} condition={({ isCodyApp }) => isCodyApp} />,
    },
    ...routes,
]
