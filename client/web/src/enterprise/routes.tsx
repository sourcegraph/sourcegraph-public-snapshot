import { Redirect } from 'react-router'
import { Navigate } from 'react-router-dom-v5-compat'

import { isErrorLike } from '@sourcegraph/common'
import { SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { isCodeInsightsEnabled } from '../insights/utils/is-code-insights-enabled'
import { LayoutRouteComponentPropsRRV6, LayoutRouteProps, routes } from '../routes'
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

export const enterpriseRoutes: readonly LayoutRouteProps<any>[] = [
    {
        isV6: true,
        path: EnterprisePageRoutes.BatchChanges,
        render: (props: LayoutRouteComponentPropsRRV6<{}>) => <GlobalBatchChangesArea {...props} />,
        // We also render this route on sourcegraph.com as a precaution in case anyone
        // follows an in-app link to /batch-changes from sourcegraph.com; the component
        // will just redirect the visitor to the marketing page
        condition: ({ batchChangesEnabled, isSourcegraphDotCom }) => batchChangesEnabled || isSourcegraphDotCom,
    },
    {
        isV6: true,
        path: EnterprisePageRoutes.CodeMonitoring,
        render: (props: LayoutRouteComponentPropsRRV6<{}>) => <GlobalCodeMonitoringArea {...props} />,
    },
    {
        isV6: true,
        path: EnterprisePageRoutes.Insights,
        render: (props: LayoutRouteComponentPropsRRV6<{}>) => <CodeInsightsRouter {...props} />,
        condition: props => isCodeInsightsEnabled(props.settingsCascade),
    },
    {
        isV6: true,
        path: EnterprisePageRoutes.Contexts,
        render: (props: LayoutRouteComponentPropsRRV6<{}>) => <SearchContextsListPage {...props} />,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        isV6: true,
        path: EnterprisePageRoutes.CreateContext,
        render: (props: LayoutRouteComponentPropsRRV6<{}>) => <CreateSearchContextPage {...props} />,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        isV6: true,
        path: EnterprisePageRoutes.EditContext,
        render: (props: LayoutRouteComponentPropsRRV6<{}>) => <EditSearchContextPage {...props} />,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        isV6: true,
        path: EnterprisePageRoutes.Context,
        render: (props: LayoutRouteComponentPropsRRV6<{}>) => <SearchContextPage {...props} />,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        isV6: true,
        path: EnterprisePageRoutes.SearchNotebook,
        render: () => <Navigate to={EnterprisePageRoutes.Notebooks} replace={true} />,
    },
    {
        isV6: true,
        path: EnterprisePageRoutes.NotebookCreate,
        render: (props: LayoutRouteComponentPropsRRV6<{}>) =>
            useExperimentalFeatures.getState().showSearchNotebook && props.authenticatedUser ? (
                <CreateNotebookPage {...props} authenticatedUser={props.authenticatedUser} />
            ) : (
                <Navigate to={EnterprisePageRoutes.Notebooks} replace={true} />
            ),
    },
    {
        isV6: true,
        path: EnterprisePageRoutes.Notebook,
        render: (props: LayoutRouteComponentPropsRRV6<{}>) => {
            const { showSearchNotebook } = useExperimentalFeatures.getState()

            return showSearchNotebook ? <NotebookPage {...props} /> : <Redirect to={PageRoutes.Search} />
        },
    },
    {
        isV6: true,
        path: EnterprisePageRoutes.Notebooks,
        render: (props: LayoutRouteComponentPropsRRV6<{}>) =>
            useExperimentalFeatures.getState().showSearchNotebook ? (
                <NotebooksListPage {...props} />
            ) : (
                <Navigate to={PageRoutes.Search} replace={true} />
            ),
    },
    ...routes,
]
