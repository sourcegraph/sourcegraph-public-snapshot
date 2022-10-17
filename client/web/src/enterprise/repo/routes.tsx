import { Redirect } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { RepoContainerRoute } from '../../repo/RepoContainer'
import { RepoRevisionContainerRoute } from '../../repo/RepoRevisionContainer'
import { repoContainerRoutes, repoRevisionContainerRoutes } from '../../repo/routes'

const RepositoryCodeIntelArea = lazyComponent(
    () => import('../codeintel/repo/RepositoryCodeIntelArea'),
    'RepositoryCodeIntelArea'
)

const RepositoryBatchChangesArea = lazyComponent(
    () => import('../batches/repo/RepositoryBatchChangesArea'),
    'RepositoryBatchChangesArea'
)

export const enterpriseRepoContainerRoutes: readonly RepoContainerRoute[] = [
    ...repoContainerRoutes,

    {
        path: '/-/code-intelligence',
        exact: false,
        render: props => <Redirect to={props.location.pathname.replace('/code-intelligence', '/code-graph')} />,
    },
    {
        path: '/-/code-graph',
        exact: false,
        render: context => <RepositoryCodeIntelArea {...context} />,
    },

    {
        path: '/-/batch-changes',
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        render: context => <RepositoryBatchChangesArea {...context} />,
    },
]

export const enterpriseRepoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[] = repoRevisionContainerRoutes
