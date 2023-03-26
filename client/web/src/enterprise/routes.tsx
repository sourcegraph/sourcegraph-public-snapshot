import { Navigate, RouteObject } from 'react-router-dom'

import { isDefined } from '@sourcegraph/common'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { isCodeInsightsEnabled } from '../insights/utils/is-code-insights-enabled'
import { LegacyRoute } from '../LegacyRouteContext'
import { routes } from '../routes'
import { EnterprisePageRoutes } from '../routes.constants'

const GlobalNotebooksArea = !process.env.DISABLE_NOTEBOOKS
    ? lazyComponent(() => import('../notebooks/GlobalNotebooksArea'), 'GlobalNotebooksArea')
    : null
const GlobalBatchChangesArea = !process.env.DISABLE_BATCH_CHANGES
    ? lazyComponent(() => import('./batches/global/GlobalBatchChangesArea'), 'GlobalBatchChangesArea')
    : null
const GlobalCodeMonitoringArea = !process.env.DISABLE_CODE_MONITORING
    ? lazyComponent(() => import('./code-monitoring/global/GlobalCodeMonitoringArea'), 'GlobalCodeMonitoringArea')
    : null
const CodeInsightsRouter = !process.env.DISABLE_CODE_INSIGHTS
    ? lazyComponent(() => import('./insights/CodeInsightsRouter'), 'CodeInsightsRouter')
    : null
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
const CodySearchPage = lazyComponent(() => import('./cody/search/CodySearchPage'), 'CodySearchPage')
const OwnPage = lazyComponent(() => import('./own/OwnPage'), 'OwnPage')
const AppComingSoonPage = lazyComponent(() => import('./app/AppComingSoonPage'), 'AppComingSoonPage')
const AppAuthCallbackPage = lazyComponent(() => import('./app/AppAuthCallbackPage'), 'AppAuthCallbackPage')

export const enterpriseRoutes: RouteObject[] = [
    GlobalBatchChangesArea && {
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
    GlobalCodeMonitoringArea && {
        path: EnterprisePageRoutes.CodeMonitoring,
        element: <LegacyRoute render={props => <GlobalCodeMonitoringArea {...props} />} />,
    },
    CodeInsightsRouter && {
        path: EnterprisePageRoutes.Insights,
        element: (
            <LegacyRoute
                render={props => <CodeInsightsRouter {...props} />}
                condition={props => isCodeInsightsEnabled(props.settingsCascade)}
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
    GlobalNotebooksArea && {
        path: EnterprisePageRoutes.Notebooks + '/*',
        element: <LegacyRoute render={props => <GlobalNotebooksArea {...props} />} />,
    },
    {
        path: EnterprisePageRoutes.CodySearch,
        element: <CodySearchPage />,
    },
    {
        path: EnterprisePageRoutes.Own,
        element: <OwnPage />,
    },
    {
        path: EnterprisePageRoutes.AppComingSoon,
        element: (
            <LegacyRoute render={() => <AppComingSoonPage />} condition={({ isSourcegraphApp }) => isSourcegraphApp} />
        ),
    },
    {
        path: EnterprisePageRoutes.AppAuthCallback,
        element: (
            <LegacyRoute
                render={() => <AppAuthCallbackPage />}
                condition={({ isSourcegraphApp }) => isSourcegraphApp}
            />
        ),
    },
    ...routes,
].filter(isDefined)
