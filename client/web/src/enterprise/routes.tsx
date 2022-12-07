import { Redirect } from 'react-router'

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

const isSearchContextsManagementEnabled = (settingsCascade: SettingsCascadeOrError): boolean =>
    !isErrorLike(settingsCascade.final) && settingsCascade.final?.experimentalFeatures?.showSearchContext !== false

export const enterpriseRoutes: readonly LayoutRouteProps<any>[] = [
    {
        path: EnterprisePageRoutes.BatchChanges,
        render: lazyComponent(() => import('./batches/global/GlobalBatchChangesArea'), 'GlobalBatchChangesArea'),
        // We also render this route on sourcegraph.com as a precaution in case anyone
        // follows an in-app link to /batch-changes from sourcegraph.com; the component
        // will just redirect the visitor to the marketing page
        condition: ({ batchChangesEnabled, isSourcegraphDotCom }) => batchChangesEnabled || isSourcegraphDotCom,
    },
    {
        path: EnterprisePageRoutes.Stats,
        render: lazyComponent(() => import('./search/stats/SearchStatsPage'), 'SearchStatsPage'),
    },
    {
        path: EnterprisePageRoutes.CodeMonitoring,
        render: lazyComponent(
            () => import('./code-monitoring/global/GlobalCodeMonitoringArea'),
            'GlobalCodeMonitoringArea'
        ),
    },
    {
        path: EnterprisePageRoutes.Insights,
        render: lazyComponent(() => import('./insights/CodeInsightsRouter'), 'CodeInsightsRouter'),
        condition: props => isCodeInsightsEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.Contexts,
        render: lazyComponent(() => import('./searchContexts/SearchContextsListPage'), 'SearchContextsListPage'),
        exact: true,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.CreateContext,
        render: lazyComponent(() => import('./searchContexts/CreateSearchContextPage'), 'CreateSearchContextPage'),
        exact: true,
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.EditContext,
        render: lazyComponent(() => import('./searchContexts/EditSearchContextPage'), 'EditSearchContextPage'),
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.Context,
        render: lazyComponent(() => import('./searchContexts/SearchContextPage'), 'SearchContextPage'),
        condition: props => isSearchContextsManagementEnabled(props.settingsCascade),
    },
    {
        path: EnterprisePageRoutes.SearchNotebook,
        render: () => <Redirect to={EnterprisePageRoutes.Notebooks} />,
        exact: true,
    },
    {
        path: EnterprisePageRoutes.NotebookCreate,
        render: props =>
            useExperimentalFeatures.getState().showSearchNotebook && props.authenticatedUser ? (
                <CreateNotebookPage {...props} authenticatedUser={props.authenticatedUser} />
            ) : (
                <Redirect to={EnterprisePageRoutes.Notebooks} />
            ),
        exact: true,
    },
    {
        path: EnterprisePageRoutes.Notebook,
        render: props => {
            const { showSearchNotebook, showSearchContext } = useExperimentalFeatures.getState()

            return showSearchNotebook ? (
                <NotebookPage {...props} showSearchContext={showSearchContext ?? false} />
            ) : (
                <Redirect to={PageRoutes.Search} />
            )
        },
        exact: true,
    },
    {
        path: EnterprisePageRoutes.Notebooks,
        render: props =>
            useExperimentalFeatures.getState().showSearchNotebook ? (
                <NotebooksListPage {...props} />
            ) : (
                <Redirect to={PageRoutes.Search} />
            ),
        exact: true,
    },
    ...routes,
]
