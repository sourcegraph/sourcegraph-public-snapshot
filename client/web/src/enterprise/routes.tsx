import { Navigate, RouteObject } from 'react-router-dom'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { isCodeInsightsEnabled } from '../insights/utils/is-code-insights-enabled'
import { LegacyRoute } from '../LegacyRouteContext'
import { routes } from '../routes'
import { EnterprisePageRoutes } from '../routes.constants'

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
const GlobalCodyArea = lazyComponent(() => import('./cody/GlobalCodyArea'), 'GlobalCodyArea')
const OwnPage = lazyComponent(() => import('./own/OwnPage'), 'OwnPage')
const AppComingSoonPage = lazyComponent(() => import('./app/AppComingSoonPage'), 'AppComingSoonPage')

export const enterpriseRoutes: RouteObject[] = [
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
    {
        path: EnterprisePageRoutes.Notebooks + '/*',
        element: <LegacyRoute render={props => <GlobalNotebooksArea {...props} />} />,
    },
    {
        path: EnterprisePageRoutes.Cody,
        element: <GlobalCodyArea />,
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
    ...routes,
]
