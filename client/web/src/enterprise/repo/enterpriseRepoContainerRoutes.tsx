import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { RedirectRoute } from '../../components/RedirectRoute'
import { RepoContainerRoute } from '../../repo/RepoContainer'
import { repoContainerRoutes } from '../../repo/repoContainerRoutes'
import { CodeIntelConfigurationPolicyPage } from '../codeintel/configuration/pages/CodeIntelConfigurationPolicyPage'

const RepositoryCodeIntelArea = lazyComponent(
    () => import('../codeintel/repo/RepositoryCodeIntelArea'),
    'RepositoryCodeIntelArea'
)

const RepositoryBatchChangesArea = lazyComponent(
    () => import('../batches/repo/RepositoryBatchChangesArea'),
    'RepositoryBatchChangesArea'
)

const RepositoryOwnPage = lazyComponent(() => import('../own/RepositoryOwnPage'), 'RepositoryOwnPage')

const CodyConfigurationPage = lazyComponent(
    () => import('../cody/configuration/pages/CodyConfigurationPage'),
    'CodyConfigurationPage'
)

export const enterpriseRepoContainerRoutes: readonly RepoContainerRoute[] = [
    ...repoContainerRoutes,

    // Code graph routes
    {
        path: '/-/code-intelligence/*',
        render: () => (
            <RedirectRoute
                getRedirectURL={({ location }) => location.pathname.replace('/code-intelligence', '/code-graph')}
            />
        ),
    },
    {
        path: '/-/code-graph/*',
        render: context => <RepositoryCodeIntelArea {...context} />,
    },

    // Cody configuration routes
    {
        path: '/cody/configuration',
        render: props => <CodyConfigurationPage {...props} />,
        condition: () => Boolean(window.context?.embeddingsEnabled),
    },
    {
        path: '/cody/configuration/:id',
        render: props => <CodeIntelConfigurationPolicyPage domain={'embeddings'} {...props} />,
    },

    {
        path: '/-/batch-changes',
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        render: context => <RepositoryBatchChangesArea {...context} />,
    },
    {
        path: '/-/own',
        condition: ({ isSourcegraphDotCom }) => !isSourcegraphDotCom,
        render: context => <RepositoryOwnPage {...context} />,
    },
]
