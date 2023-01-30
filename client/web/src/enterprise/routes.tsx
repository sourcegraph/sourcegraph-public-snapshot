import { Redirect } from 'react-router'
import { Navigate } from 'react-router-dom-v5-compat'

import { isErrorLike } from '@sourcegraph/common'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { isCodeInsightsEnabled } from '../insights/utils/is-code-insights-enabled'
import { LayoutRouteProps, routes } from '../routes'
import { EnterprisePageRoutes, PageRoutes } from '../routes.constants'
import { useExperimentalFeatures } from '../stores'

const NotebookPage = lazyComponent(() => import('../notebooks/notebookPage/NotebookPage'), 'NotebookPage')
const CreateNotebookPage = lazyComponent(
    () => import('../notebooks/createPage/CreateNotebookPage'),
    'CreateNotebookPage'
)
const NotebooksListPage = lazyComponent(() => import('../notebooks/listPage/NotebooksListPage'), 'NotebooksListPage')
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

const isSearchContextsManagementEnabled = (settingsCascade: SettingsCascadeOrError): boolean =>
    !isErrorLike(settingsCascade.final) && settingsCascade.final?.experimentalFeatures?.showSearchContext !== false

export const enterpriseRoutes: readonly LayoutRouteProps[] = [
    {
        path: EnterprisePageRoutes.BatchChanges,
        render: props => <GlobalBatchChangesArea {...props} />,
        // We also render this route on sourcegraph.com as a precaution in case anyone
        // follows an in-app link to /batch-changes from sourcegraph.com; the component
        // will just redirect the visitor to the marketing page
        condition: ({ batchChangesEnabled, isSourcegraphDotCom }) => batchChangesEnabled || isSourcegraphDotCom,
    },
    {
        path: EnterprisePageRoutes.CodeMonitoring,
        render: props => <GlobalCodeMonitoringArea {...props} />,
    },
    {
        path: EnterprisePageRoutes.Insights,
        render: props => <CodeInsightsRouter {...props} />,
        condition: props => isCodeInsightsEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.Contexts,
        render: props => <SearchContextsListPage {...props} />,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.CreateContext,
        render: props => <CreateSearchContextPage {...props} />,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.EditContext,
        render: props => <EditSearchContextPage {...props} />,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.Context,
        render: props => <SearchContextPage {...props} />,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.SearchNotebook,
        render: () => <Navigate to={EnterprisePageRoutes.Notebooks} replace={true} />,
    },
    {
        path: EnterprisePageRoutes.NotebookCreate,
        render: props =>
            useExperimentalFeatures.getState().showSearchNotebook && props.authenticatedUser ? (
                <CreateNotebookPage {...props} authenticatedUser={props.authenticatedUser} />
            ) : (
                <Navigate to={EnterprisePageRoutes.Notebooks} replace={true} />
            ),
    },
    {
        path: EnterprisePageRoutes.Notebook,
        render: props => {
            const { showSearchNotebook } = useExperimentalFeatures.getState()

            return showSearchNotebook ? <NotebookPage {...props} /> : <Redirect to={PageRoutes.Search} />
        },
    },
    {
        path: EnterprisePageRoutes.Notebooks,
        render: props =>
            useExperimentalFeatures.getState().showSearchNotebook ? (
                <NotebooksListPage {...props} />
            ) : (
                <Navigate to={PageRoutes.Search} replace={true} />
            ),
    },
    ...routes,
]
